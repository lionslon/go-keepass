package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	// Для хранения данных ответа
	responseData struct {
		status int
		size   int
	}

	// Обертка над http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

/*Реализовываем интерфейс*/
func (m *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := m.ResponseWriter.Write(b)
	m.responseData.size += size // захватываем размер
	return size, err
}

func (m *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	m.ResponseWriter.WriteHeader(statusCode)
	m.responseData.status = statusCode // захватываем код статуса
}

func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: http.StatusOK,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}

		//Вызов целевого handler
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		// Логируем данные запроса и результат
		log.Info("",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", duration),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
		)
	})
}
