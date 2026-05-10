package ratelimiter

import (
	"context"
)

type Limiter interface {
	IsAllowed(ctx context.Context, user, path string) (*LimitStatus, error)
}

type RateLimiter struct {
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

func (r *RateLimiter) IsAllowed(ctx context.Context, user, path string) (*LimitStatus, error) {
	rule := r.ruleFor(ctx, path)
	uniquePath := user + ":" + path
	return r.redis.IsAllowed(ctx, uniquePath, rule)
}
