package middleware

import (
	"golang.org/x/time/rate"
)

type TokenBucketLimiter struct {
	*rate.Limiter
}

func NewTokenBucketLimiter(r rate.Limit, b int) *TokenBucketLimiter {
	t := TokenBucketLimiter{}
	t.Limiter = rate.NewLimiter(r, b)
	return &t
}
func (t *TokenBucketLimiter) Limit() bool {
	return !t.Allow()
}
