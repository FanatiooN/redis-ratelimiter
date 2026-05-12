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
	var status *LimitStatus

	err := r.client.Watch(ctx, func(tx *redis.Tx) error {
		currentTime := time.Now().UnixNano()
		currentTimeStr := strconv.FormatInt(currentTime, 10)

		lastTime := currentTime - rule.Window.Nanoseconds()
		lastTimeStr := strconv.FormatInt(lastTime, 10)

		err := tx.ZRemRangeByScore(ctx, uniquePath, "0", lastTimeStr).Err()
		if err != nil {
			return err
		}

		cnt, err := tx.ZCard(ctx, uniquePath).Result()
		if err != nil {
			return err
		}

		limit := rule.Limit
		remaining := limit - int(cnt)
		allowed := remaining > 0

		_, err = tx.TxPipelined(ctx, func(pipeliner redis.Pipeliner) error {
			if allowed {
				pipeliner.ZAdd(ctx, uniquePath, redis.Z{
					Score:  float64(currentTime),
					Member: currentTimeStr,
				})
				pipeliner.Expire(ctx, uniquePath, rule.Window)

				remaining--
			}

			return nil
		})
		if err != nil {
			return err
		}

		retryAfter := time.Time{}

		if !allowed {
			oldest, err := tx.ZRangeWithScores(ctx, uniquePath, 0, 0).Result()
			if err != nil {
				return err
			}

			if len(oldest) > 0 {
				oldestTime := time.Unix(0, int64(oldest[0].Score))
				retryAfter = oldestTime.Add(rule.Window)
			}
		}

		status = &LimitStatus{
			Limit:      limit,
			Remaining:  remaining,
			RetryAfter: retryAfter,
			Allowed:    allowed,
		}

		return nil
	})

	return status, err
}
