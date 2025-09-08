package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/evrintobing17/expense-management-backend/internal/auth"
	"github.com/evrintobing17/expense-management-backend/internal/domain"
)

type contextKey string

const (
	userIDKey   contextKey = "userID"
	userRoleKey contextKey = "userRole"
)

func AuthMiddleware(authService auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Authorization header format must be: Bearer {token}", http.StatusUnauthorized)
				return
			}

			token := parts[1]
			userID, role, err := authService.ValidateToken(context.Background(), token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			ctx = context.WithValue(ctx, userRoleKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ManagerOnlyMiddleware ensures only users with manager role can access the endpoint
func ManagerOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(userRoleKey).(domain.Role)
		if !ok || role != domain.RoleManager {
			http.Error(w, "Access denied. Manager role required.", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper functions to get values from context
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDKey).(int)
	return userID, ok
}

func GetUserRoleFromContext(ctx context.Context) (domain.Role, bool) {
	role, ok := ctx.Value(userRoleKey).(domain.Role)
	return role, ok
}
