package redis_ratelimiter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")

	buf := &bytes.Buffer{}

	if err := json.NewEncoder(buf).Encode(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal server error"}`))

		return
	}

	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"message": message})
}

func Middleware(limiter Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			user := r.RemoteAddr
			status, err := limiter.IsAllowed(r.Context(), user, path)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, "internal server error")
				return
			}
			if !status.Allowed {
				WriteError(w, http.StatusTooManyRequests, fmt.Sprintf("rate limit exceeded, retry after %v", status.RetryAfter.Format("01-02 15:04:05")))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
