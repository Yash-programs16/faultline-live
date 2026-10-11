package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"faultline-live/backend/internal/config"
	"faultline-live/backend/internal/response"
	"faultline-live/backend/internal/telemetry"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
	RoutePatternKey contextKey = "route_pattern"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func Logging(logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := r.Context().Value(RequestIDKey)
			
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			next.ServeHTTP(wrapped, r)
			
			duration := time.Since(start)
			logger.Info("request",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Recovery(logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := r.Context().Value(RequestIDKey)
					logger.Error("panic recovered",
						"request_id", requestID,
						"error", err,
						"path", r.URL.Path,
					)
					response.InternalError(w, "Internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func CORS(allowedOrigins []string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false
			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Auth(cfg *config.AuthConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "Missing authorization header")
				return
			}
			
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Unauthorized(w, "Invalid authorization format")
				return
			}
			
			tokenString := parts[1]
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(cfg.JWTSecret), nil
			})
			
			if err != nil || !token.Valid {
				response.Unauthorized(w, "Invalid or expired token")
				return
			}
			
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.Unauthorized(w, "Invalid token claims")
				return
			}
			
			userID, ok := claims["user_id"].(float64)
			if !ok {
				response.Unauthorized(w, "Invalid token user ID")
				return
			}
			
			ctx := context.WithValue(r.Context(), UserIDKey, int64(userID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	return userID, ok
}

func GetRequestID(r *http.Request) string {
	if id, ok := r.Context().Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

func Metrics(telemetry *telemetry.Telemetry) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			routePattern := mux.CurrentRoute(r)
			var pattern string
			if routePattern != nil {
				pattern, _ = routePattern.GetPathTemplate()
			}
			if pattern == "" {
				pattern = r.URL.Path
			}
			
			ctx := context.WithValue(r.Context(), RoutePatternKey, pattern)
			
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r.WithContext(ctx))
			
			duration := time.Since(start)
			telemetry.RecordRequest(pattern, r.Method, wrapped.statusCode, duration)
		})
	}
}

func GetRoutePattern(r *http.Request) string {
	if pattern, ok := r.Context().Value(RoutePatternKey).(string); ok {
		return pattern
	}
	return r.URL.Path
}

// RateLimiter provides simple in-memory rate limiting
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, times := range rl.requests {
			valid := times[:0]
			for _, t := range times {
				if now.Sub(t) < rl.window {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = valid
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	times := rl.requests[key]
	valid := times[:0]
	for _, t := range times {
		if now.Sub(t) < rl.window {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[key] = valid
		return false
	}

	valid = append(valid, now)
	rl.requests[key] = valid
	return true
}

func RateLimit(limiter *RateLimiter) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP + path as key
			ip := strings.Split(r.RemoteAddr, ":")[0]
			key := ip + ":" + r.URL.Path

			if !limiter.Allow(key) {
				response.TooManyRequests(w, "Rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}