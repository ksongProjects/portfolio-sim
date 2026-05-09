package providers

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu         sync.Mutex
	maxPerSec  int
	maxPerHour int
	secCount   int
	hourCount  int
	secReset   time.Time
	hourReset  time.Time
}

func NewRateLimiter(maxPerSec, maxPerHour int) *RateLimiter {
	now := time.Now()
	return &RateLimiter{
		maxPerSec:  maxPerSec,
		maxPerHour: maxPerHour,
		secReset:   now.Add(time.Second),
		hourReset:  now.Add(time.Hour),
	}
}

func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	if now.After(r.secReset) {
		r.secCount = 0
		r.secReset = now.Add(time.Second)
	}
	if now.After(r.hourReset) {
		r.hourCount = 0
		r.hourReset = now.Add(time.Hour)
	}

	if r.secCount >= r.maxPerSec || r.hourCount >= r.maxPerHour {
		return false
	}

	r.secCount++
	r.hourCount++
	return true
}

func (r *RateLimiter) Wait() {
	for {
		if r.Allow() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}