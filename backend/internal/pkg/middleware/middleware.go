package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// contextKey is a private type for context keys in this package.
type contextKey string

const RequestIDKey contextKey = "request_id"

// Envelope is the unified JSON response format.
type Envelope struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, code int, message string, data interface{}, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Envelope{
		Code:      code,
		Message:   message,
		Data:      data,
		RequestID: requestID,
	})
}

// WriteError writes an error response.
func WriteError(w http.ResponseWriter, status int, code int, message string, requestID string) {
	WriteJSON(w, status, code, message, nil, requestID)
}

// GetRequestID extracts the request ID from context.
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(RequestIDKey).(string); ok {
		return v
	}
	return ""
}

// RequestID generates or propagates X-Request-Id.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Recovery catches panics and returns 500.
func Recovery(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						zap.Any("error", rec),
						zap.String("stack", string(debug.Stack())),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)
					rid, _ := r.Context().Value(RequestIDKey).(string)
					WriteError(w, http.StatusInternalServerError, 4001, "internal server error", rid)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Logging logs each request.
func Logging(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(ww, r)
			rid, _ := r.Context().Value(RequestIDKey).(string)
			log.Info("http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.status),
				zap.Duration("latency", time.Since(start)),
				zap.String("request_id", rid),
				zap.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

// CORS adds Cross-Origin headers.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	origins := strings.Join(allowedOrigins, ", ")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-Id, X-Idempotency-Key, X-User-Id, X-User-Role")
			w.Header().Set("Access-Control-Max-Age", "86400")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Chain applies middlewares in order (first listed = outermost).
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// ContentTypeJSON enforces Content-Type: application/json on mutating methods.
func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			ct := r.Header.Get("Content-Type")
			// Strip optional parameters like charset
			if idx := strings.Index(ct, ";"); idx != -1 {
				ct = strings.TrimSpace(ct[:idx])
			}
			if ct != "application/json" {
				rid := GetRequestID(r.Context())
				WriteError(w, http.StatusUnsupportedMediaType, 2001, "Content-Type must be application/json", rid)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// statusWriter wraps ResponseWriter to capture status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
