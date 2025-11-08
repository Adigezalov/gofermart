package middleware

import (
	"log"
	"net/http"
	"time"
)

// ResponseWriter обертка для захвата статус кода
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// NewResponseWriter создает новую обертку ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // по умолчанию 200
	}
}

// WriteHeader перехватывает статус код
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write перехватывает размер ответа
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// LoggingMiddleware middleware для логирования HTTP запросов
type LoggingMiddleware struct {
	enableColors bool
}

// NewLoggingMiddleware создает новый экземпляр LoggingMiddleware
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		enableColors: true, // по умолчанию включаем цвета
	}
}

// NewLoggingMiddlewareWithOptions создает middleware с настройками
func NewLoggingMiddlewareWithOptions(enableColors bool) *LoggingMiddleware {
	return &LoggingMiddleware{
		enableColors: enableColors,
	}
}

// LogRequest логирует HTTP запросы
func (m *LoggingMiddleware) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем обертку для ResponseWriter
		wrapped := NewResponseWriter(w)

		// Выполняем запрос
		next.ServeHTTP(wrapped, r)

		// Вычисляем время выполнения
		duration := time.Since(start)

		// Получаем User-Agent
		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			userAgent = "Unknown"
		}

		// Получаем Content-Type для POST/PUT запросов
		contentType := r.Header.Get("Content-Type")

		// Определяем цвет для статус кода
		statusColor := ""
		resetColor := ""
		if m.enableColors {
			statusColor = getStatusColor(wrapped.statusCode)
			resetColor = "\033[0m"
		}

		// Логируем информацию о запросе
		if contentType != "" {
			log.Printf(
				"%s %s %s%d%s %v %s %d bytes [%s]",
				r.Method,
				r.RequestURI,
				statusColor,
				wrapped.statusCode,
				resetColor,
				duration,
				getClientIP(r),
				wrapped.size,
				contentType,
			)
		} else {
			log.Printf(
				"%s %s %s%d%s %v %s %d bytes",
				r.Method,
				r.RequestURI,
				statusColor,
				wrapped.statusCode,
				resetColor,
				duration,
				getClientIP(r),
				wrapped.size,
			)
		}
	})
}

// LogRequestFunc версия middleware как функция для удобства
func (m *LoggingMiddleware) LogRequestFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.LogRequest(next).ServeHTTP(w, r)
	}
}

// getStatusColor возвращает ANSI цвет для статус кода
func getStatusColor(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "\033[32m" // зеленый для 2xx
	case statusCode >= 300 && statusCode < 400:
		return "\033[33m" // желтый для 3xx
	case statusCode >= 400 && statusCode < 500:
		return "\033[31m" // красный для 4xx
	case statusCode >= 500:
		return "\033[35m" // пурпурный для 5xx
	default:
		return "\033[0m" // без цвета
	}
}

// getClientIP извлекает реальный IP клиента
func getClientIP(r *http.Request) string {
	// Проверяем заголовки прокси
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Возвращаем RemoteAddr
	return r.RemoteAddr
}
