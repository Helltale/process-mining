package presentation

import "net/http"

// withCORS — middleware, разрешающий CORS-запросы с фронта
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//TODO: вынести в конфиг
		// Разрешаем CORS для фронта на порту 5174
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5174")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Отвечаем на preflight-запрос
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Продолжаем обработку
		next.ServeHTTP(w, r)
	})
}
