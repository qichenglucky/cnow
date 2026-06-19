package middleware

import (
	"context"
	"net/http"
	"strconv"
)

type ctxKey string

const (
	userIDKey   ctxKey = "user_id"
	userRoleKey ctxKey = "user_role"
)

// AuthRequired rejects requests that don't carry X-User-Id.
func AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-User-Id") == "" {
			rid := GetRequestID(r.Context())
			WriteError(w, http.StatusUnauthorized, 2002, "missing X-User-Id header", rid)
			return
		}
		ctx := injectIdentity(r.Context(), r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// InjectIdentity populates context from headers without rejecting (for optional auth).
func InjectIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := injectIdentity(r.Context(), r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func injectIdentity(ctx context.Context, r *http.Request) context.Context {
	if uid := r.Header.Get("X-User-Id"); uid != "" {
		if id, err := strconv.ParseInt(uid, 10, 64); err == nil {
			ctx = context.WithValue(ctx, userIDKey, id)
		}
	}
	if role := r.Header.Get("X-User-Role"); role != "" {
		ctx = context.WithValue(ctx, userRoleKey, role)
	}
	return ctx
}

// GetUserID returns the authenticated user id from context (0 if absent).
func GetUserID(ctx context.Context) int64 {
	if v, ok := ctx.Value(userIDKey).(int64); ok {
		return v
	}
	return 0
}

// GetUserRole returns the authenticated user role from context ("" if absent).
func GetUserRole(ctx context.Context) string {
	if v, ok := ctx.Value(userRoleKey).(string); ok {
		return v
	}
	return ""
}
