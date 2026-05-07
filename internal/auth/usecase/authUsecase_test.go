package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/evrintobing17/expense-management-backend/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		mockAuth := new(mocks.AuthService)
		uc := NewAuthUseCase(mockAuth)
		user := &domain.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "manager",
			Role:  "manager",
		}
		mockAuth.On("Login", mock.Anything, "TEST@example.com", "PWD").Return("some token", user, nil).Once()

		token, result, err := uc.Login(ctx, "TEST@example.com", "PWD")
		require.NoError(t, err)
		require.Equal(t, "some token", token)
		require.Equal(t, &domain.UserResponse{
			ID:    1,
			Email: "test@example.com",
			Name:  "manager",
			Role:  "manager",
		}, result)
	})

	t.Run("auth service error", func(t *testing.T) {
		mockAuth := new(mocks.AuthService)
		uc := NewAuthUseCase(mockAuth)
		expectedErr := errors.New("invalid credentials")
		mockAuth.On("Login", mock.Anything, "test@example.com", "wrong").Return("", (*domain.User)(nil), expectedErr).Once()

		token, result, err := uc.Login(ctx, "test@example.com", "wrong")
		require.ErrorIs(t, err, expectedErr)
		require.Empty(t, token)
		require.Nil(t, result)
	})
}
