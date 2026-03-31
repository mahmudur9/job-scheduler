package middleware

import (
	"log"
	"net/http"
)

func RecoveryMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			if err := recover(); err != nil {
				log.Println("panic recovered:", err)
				http.Error(w, "internal server error", 500)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
