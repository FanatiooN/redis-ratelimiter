package main

import (
	"context"
	"fmt"
	"time"
)

type Limiter interface {
	IsAllowed(ctx context.Context, user, path string) (LimitStatus, error)
}

type RateLimiter struct {
	rules        map[string]LimiterRule
	defaultRule  LimiterRule
	limitPerUser map[string][]time.Time
}

type LimitStatus struct {
	Limit      int
	Remaining  int
	RetryAfter time.Time
	Allowed    bool
}

type LimiterRule struct {
	window time.Duration
	limit  int
}

func New(window time.Duration, limit int) *RateLimiter {
	return &RateLimiter{
		rules:        make(map[string]LimiterRule),
		defaultRule:  LimiterRule{window: window, limit: limit},
		limitPerUser: make(map[string][]time.Time),
	}
}

func (r *RateLimiter) AddRule(ctx context.Context, path string, rule LimiterRule) error {
	r.rules[path] = rule
	return nil
}

func (r *RateLimiter) IsAllowed(ctx context.Context, user, path string) (LimitStatus, error) {
	rule, ok := r.rules[path]
	if !ok {
		rule = r.defaultRule
	}

	uniquePath := user + path
	currentUserLimit := r.limitPerUser[uniquePath]

	newTimestamps := make([]time.Time, 0, len(currentUserLimit)+1)
	currentTime := time.Now()
	lastTime := currentTime.Add(-rule.window)
	limit := rule.limit

	for _, timestamp := range currentUserLimit {
		if timestamp.After(lastTime) {
			newTimestamps = append(newTimestamps, timestamp)
		}
	}

	if len(newTimestamps) < limit {
		newTimestamps = append(newTimestamps, currentTime)
	}

	r.limitPerUser[uniquePath] = newTimestamps

	remaining := limit - len(currentUserLimit)
	retryAfter := time.Time{}
	if len(currentUserLimit) > 0 {
		retryAfter = currentUserLimit[0].Add(rule.window)
	}

	allowed := remaining > 0

	return LimitStatus{
		Limit:      limit,
		Remaining:  remaining,
		RetryAfter: retryAfter,
		Allowed:    allowed,
	}, nil
}

func main() {
	ctx := context.Background()

	rl := New(time.Minute, 20)
	rl.AddRule(ctx, "path",
		LimiterRule{
			window: time.Second * 30,
			limit:  50,
		})

	for i := 0; i < 100; i++ {
		status, _ := rl.IsAllowed(ctx, "user", "path")
		fmt.Printf("%v\n", status)
	}
}
