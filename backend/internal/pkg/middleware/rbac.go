package middleware

import "net/http"

// Well-known roles in descending privilege order.
const (
	RoleAdmin     = "admin"
	RoleOwner     = "owner"
	RoleDeveloper = "developer"
	RoleObserver  = "observer"
)

// RequireRole returns middleware that permits only the listed roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetUserRole(r.Context())
			if _, ok := allowed[role]; !ok {
				rid := GetRequestID(r.Context())
				WriteError(w, http.StatusForbidden, 2003, "insufficient permissions", rid)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
