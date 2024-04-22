package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Prthmesh6/rate_limiter/limiter"
	"github.com/Prthmesh6/rate_limiter/models"

	"golang.org/x/time/rate"
)

func rateLimiterPerClient(lmt *limiter.Limiter, next func(writer http.ResponseWriter, request *http.Request)) http.Handler {
	var (
		mu      sync.Mutex
		clients = make(map[string]*models.Client)
	)
	expire := func() {
		for {
			time.Sleep(time.Minute)
			// Lock the mutex to protect this section from race conditions.
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.LastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}

	go expire()

	middle := func(w http.ResponseWriter, r *http.Request) {
		// Extract the IP address from the request.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Lock the mutex to protect this section from race conditions.
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &models.Client{Limiter: rate.NewLimiter(rate.Limit(lmt.Limit), lmt.Max)}
		}
		clients[ip].LastSeen = time.Now()
		if !clients[ip].Limiter.Allow() {
			mu.Unlock()

			Message := models.Message{
				Status: "Request Failed",
				Body:   "Try again later",
			}

			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&Message)
			return
		}
		mu.Unlock()
		next(w, r)
	}

	return http.HandlerFunc(middle)
}

func endpointHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	Message := models.Message{
		Status: "Successful",
		Body:   "API successfully reached server",
	}
	err := json.NewEncoder(writer).Encode(&Message)
	if err != nil {
		return
	}
}

func NewLimiter(limit float64, max int) *limiter.Limiter {
	return &limiter.Limiter{
		Limit:   limit,
		Max:     max,
		Message: "you reached the limit",
	}

}

// LimitFuncHandler is a middleware that performs rate-limiting given request handler function.
func RateFuncHandler(lmt *limiter.Limiter, nextFunc func(http.ResponseWriter, *http.Request)) http.Handler {
	return rateLimiterPerClient(lmt, http.HandlerFunc(nextFunc))
}

func main() {
	lmt := NewLimiter(1, 2)
	http.Handle("/ping", rateLimiterPerClient(lmt, endpointHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("There was an error listening on port :8080", err)
	}
}
