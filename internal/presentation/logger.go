package presentation

import (
	"log"
	"net/http"
	"time"
)

// LogRequest — middleware для логирования всех входящих HTTP-запросов
func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("➡️ %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("✅ %s %s — %s", r.Method, r.URL.Path, time.Since(start))
	})
}
