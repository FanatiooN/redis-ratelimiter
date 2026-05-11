package ratelimiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(addr string, port int) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%v:%v", addr, port),
	})

	return &Redis{client: rdb}
}

func (r *Redis) IsAllowed(ctx context.Context, uniquePath string, rule LimiterRule) (*LimitStatus, error) {
	currentTime := time.Now().UnixNano()
	currentTimeStr := strconv.FormatInt(currentTime, 10)

	lastTime := currentTime - rule.Window.Nanoseconds()
	lastTimeStr := strconv.FormatInt(lastTime, 10)

	r.client.ZRemRangeByScore(ctx, uniquePath, "0", lastTimeStr)

	cnt, err := r.client.ZCard(ctx, uniquePath).Result()
	if err != nil {
		return nil, err
	}

	limit := rule.Limit

	remaining := limit - int(cnt)
	allowed := remaining > 0

	if allowed {
		r.client.ZAdd(ctx, uniquePath, redis.Z{
			Score:  float64(currentTime),
			Member: currentTimeStr,
		})
		r.client.Expire(ctx, uniquePath, rule.Window)

		remaining--
	}

	retryAfter := time.Time{}

	if !allowed {
		oldest, _ := r.client.ZRangeWithScores(ctx, uniquePath, 0, 0).Result()

		oldestTime := time.Unix(0, int64(oldest[0].Score))
		retryAfter = oldestTime.Add(rule.Window)
	}

	return &LimitStatus{
		Limit:      limit,
		Remaining:  remaining,
		RetryAfter: retryAfter,
		Allowed:    allowed,
	}, nil
}
