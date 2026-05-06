package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Log incoming request
		zap.S().Infof("incoming request: method=%s path=%s remote_addr=%s user_agent=%s",
			r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Log response
		duration := time.Since(start)
		zap.S().Infof("request completed: httpCode=%d method=%s path=%s remote_addr=%s user_agent=%s duration=%s",
			wrapped.statusCode, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent(), duration)
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
