package ratelimiter

import (
	"context"
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

func New(rule LimiterRule) *RateLimiter {
	return &RateLimiter{
		rules:        make(map[string]LimiterRule),
		defaultRule:  rule,
		limitPerUser: make(map[string][]time.Time),
	}
}

func (r *RateLimiter) AddRule(ctx context.Context, path string, rule LimiterRule) error {
	r.rules[path] = rule
	return nil
}

func (r *RateLimiter) ruleFor(ctx context.Context, path string) LimiterRule {
	rule, ok := r.rules[path]
	if ok {
		return rule
	}
	return r.defaultRule
}

func (r *RateLimiter) IsAllowed(ctx context.Context, user, path string) (LimitStatus, error) {

	rule := r.ruleFor(ctx, path)

	uniquePath := user + path
	currentUserLimit := r.limitPerUser[uniquePath]

	newTimestamps := make([]time.Time, 0, len(currentUserLimit)+1)
	currentTime := time.Now()
	lastTime := currentTime.Add(-rule.Window)
	limit := rule.Limit

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
	allowed := remaining > 0

	retryAfter := time.Time{}
	if len(currentUserLimit) > 0 {
		retryAfter = currentUserLimit[0].Add(rule.Window)
	}

	return LimitStatus{
		Limit:      limit,
		Remaining:  remaining,
		RetryAfter: retryAfter,
		Allowed:    allowed,
	}, nil
}
