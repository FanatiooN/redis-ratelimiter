package ratelimiter

import "time"

type LimitStatus struct {
	Limit      int
	Remaining  int
	RetryAfter time.Time
	Allowed    bool
}

type LimiterRule struct {
	Window time.Duration
	Limit  int
}
