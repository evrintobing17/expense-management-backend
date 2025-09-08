package auth

import (
	"context"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
)

type AuthUseCase interface {
	Login(ctx context.Context, email, password string) (string, *domain.UserResponse, error)
}
