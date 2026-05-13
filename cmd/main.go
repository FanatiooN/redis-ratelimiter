package main

import (
	"fmt"
	"net/http"
	redis_ratelimiter "redis-ratelimiter"
	"time"
)

func main() {
	defaultRule := redis_ratelimiter.LimiterRule{
		Window: time.Minute,
		Limit:  10,
	}
	rl := redis_ratelimiter.New(defaultRule, "localhost", 6379)

	authRule := redis_ratelimiter.LimiterRule{
		Window: time.Second * 30,
		Limit:  3,
	}
	rl.AddRule("/auth", authRule)

	mw := redis_ratelimiter.Middleware(rl)

	deflt := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redis_ratelimiter.WriteJSON(w, http.StatusOK, map[string]string{"message": "hello world"})
	})

	auth := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redis_ratelimiter.WriteJSON(w, http.StatusOK, map[string]string{"message": "auth ok"})
	})

	http.Handle("/", mw(deflt))
	http.Handle("/auth", mw(auth))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
