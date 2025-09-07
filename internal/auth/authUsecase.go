package auth

import "github.com/evrintobing17/expense-management-backend/internal/domain"

type AuthUseCase interface {
	Login(email, password string) (string, *domain.UserResponse, error)
}
