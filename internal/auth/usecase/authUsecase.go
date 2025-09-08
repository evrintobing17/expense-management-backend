package usecase

import (
	"context"

	"github.com/evrintobing17/expense-management-backend/internal/auth"
	"github.com/evrintobing17/expense-management-backend/internal/domain"
)

type authUseCase struct {
	authService auth.AuthService
}

func NewAuthUseCase(authService auth.AuthService) auth.AuthUseCase {
	return &authUseCase{authService: authService}
}

func (uc *authUseCase) Login(ctx context.Context, email, password string) (string, *domain.UserResponse, error) {
	token, user, err := uc.authService.Login(ctx, email, password)
	if err != nil {
		return "", nil, err
	}

	userResponse := &domain.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}

	return token, userResponse, nil
}
