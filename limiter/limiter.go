package limiter

type Limiter struct {
	Limit   float64
	Max     int
	Message string
}

func NewLimiter(limit float64, max int) *Limiter {
	return &Limiter{
		Limit:   limit,
		Max:     max,
		Message: "you reached the limit",
	}
}
