package models

import (
	"time"

	"golang.org/x/time/rate"
)

type ExpiryOptions struct {
	DefaultExpiryTtl time.Duration
}

type Client struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}
