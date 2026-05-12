package main

import (
	"context"
	"fmt"
	"redis-ratelimiter/ratelimiter"
	"time"
)

func main() {
	ctx := context.Background()

	defaultRule := ratelimiter.LimiterRule{
		Window: time.Minute,
		Limit:  50,
	}
	rl := ratelimiter.New(defaultRule, "localhost", 6379)

	authRule := ratelimiter.LimiterRule{
		Window: time.Second * 30,
		Limit:  3,
	}
	rl.AddRule("auth", authRule)

	for i := 0; i < 100; i++ {
		status, _ := rl.IsAllowed(ctx, "user", "path")
		fmt.Printf("%v\n", status)
	}
}
