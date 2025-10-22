package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	// Добавляет уникальный ID к каждому запросу
	r.Use(middleware.RequestID)
	// Определяет реальный IP клиента за прокси/балансировщиком
	r.Use(middleware.RealIP)
	// Ловит panic и возвращает 500 ошибку вместо падения сервера
	r.Use(middleware.Recoverer)
	// Кастомное логирование запросов
	r.Use(h.loggingMiddleware)

	// Проверка бд подключения
	r.Get("/ping", h.PingDatabaseHandler)

	// Получение групп, генерация UUID и т.п.
	r.Route("/api", func(api chi.Router) {
		api.Post("/keys/generate", h.GenerateKeysHandler)
		api.Post("/keys/validate", h.ValidateKeyHandler)
		api.Get("/groups", h.GetGroupsHandler)
	})
	return r
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = b
	return r.ResponseWriter.Write(b)
}

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		//дефолтные значения
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		//запускаем
		next.ServeHTTP(recorder, r)
		//получаем время завершения
		duration := time.Since(start)

		h.logger.Info("http", fmt.Sprintf("%s %s %d %s | Client: %s | START: %v, Time: %v",
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			http.StatusText(recorder.statusCode),
			r.RemoteAddr,
			start.Format("18.05.1998 12:55:01"),
			duration,
		))
	})
}
