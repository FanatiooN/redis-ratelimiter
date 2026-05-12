package redis_ratelimiter

import (
	"context"
	"sync"
)

type Limiter interface {
	IsAllowed(ctx context.Context, user, path string) (*LimitStatus, error)
}

type RateLimiter struct {
	mu          sync.RWMutex
	rules       map[string]LimiterRule
	defaultRule LimiterRule
	redis       *Redis
}

func New(rule LimiterRule, addr string, port int) *RateLimiter {
	return &RateLimiter{
		rules:       make(map[string]LimiterRule),
		defaultRule: rule,
		redis:       NewRedis(addr, port),
	}
}

func (r *RateLimiter) AddRule(path string, rule LimiterRule) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rules[path] = rule
}

func (r *RateLimiter) ruleFor(path string) LimiterRule {
	r.mu.Lock()
	defer r.mu.Unlock()

	rule, ok := r.rules[path]
	if ok {
		return rule
	}

	return r.defaultRule
}

func (r *RateLimiter) IsAllowed(ctx context.Context, user, path string) (*LimitStatus, error) {
	rule := r.ruleFor(path)
	uniquePath := user + ":" + path

	return r.redis.IsAllowed(ctx, uniquePath, rule)
}
