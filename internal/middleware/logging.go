package middleware

import (
	"net/http"
	"time"
)

type Logger interface {
	Info(string)
}

func LoggingMiddleware(logger Logger) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()

			next.ServeHTTP(w, r)

			logger.Info(r.Method + " " + r.URL.Path + " " + time.Since(start).String()) // Log route
		})
	}
}
