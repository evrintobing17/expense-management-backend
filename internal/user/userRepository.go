package user

import (
	"context"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
)

type UserRepository interface {
	FindByID(ctx context.Context, id int) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}
