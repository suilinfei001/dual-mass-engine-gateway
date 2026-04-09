package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/quality-gateway/shared/pkg/logger"
)

// MiddlewareFunc is a function that wraps an http.Handler.
type MiddlewareFunc func(http.Handler) http.Handler

// LoggingMiddleware logs HTTP requests.
func LoggingMiddleware(log *logger.Logger) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWrapper{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			log.Info("HTTP request",
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.Int("status", wrapped.status),
				logger.String("duration", duration.String()),
				logger.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

// responseWrapper wraps http.ResponseWriter to capture status code.
type responseWrapper struct {
	http.ResponseWriter
	status  int
	written bool
}

func (w *responseWrapper) WriteHeader(status int) {
	if !w.written {
		w.status = status
		w.written = true
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// RecoveryMiddleware recovers from panics.
func RecoveryMiddleware(log *logger.Logger) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					log.Error("Panic recovered",
						logger.Any("error", rvr),
						logger.String("path", r.URL.Path),
					)
					InternalError(w, "Internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware adds CORS headers.
func CORSMiddleware(allowedOrigins []string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware validates Bearer token.
func AuthMiddleware(token string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				Unauthorized(w, "Missing authorization header")
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				Unauthorized(w, "Invalid authorization header format")
				return
			}

			receivedToken := strings.TrimPrefix(authHeader, "Bearer ")
			if receivedToken != token {
				Unauthorized(w, "Invalid token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ContentTypeMiddleware ensures JSON content type for requests with body.
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			ct := r.Header.Get("Content-Type")
			if ct != "" && !strings.Contains(ct, "application/json") {
				BadRequest(w, "Content-Type must be application/json")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware adds a timeout to requests.
func TimeoutMiddleware(timeout time.Duration) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.TimeoutHandler(next, timeout, "Request timeout").ServeHTTP(w, r)
		})
	}
}

// Chain chains multiple middleware together.
func Chain(middlewares ...MiddlewareFunc) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
