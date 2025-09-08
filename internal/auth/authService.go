package auth

import (
	"context"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, *domain.User, error)
	ValidateToken(ctx context.Context, tokenString string) (int, domain.Role, error)
}
