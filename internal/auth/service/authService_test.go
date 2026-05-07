package service

import (
	"context"
	"errors"
	"testing"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/evrintobing17/expense-management-backend/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := new(mocks.UserRepository)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
		require.NoError(t, err)
		user := &domain.User{
			ID:           1,
			Email:        "user@example.com",
			Role:         domain.RoleEmployee,
			PasswordHash: string(hashedPassword),
		}
		mockRepo.On("FindByEmail", mock.Anything, "user@example.com").Return(user, nil).Once()
		svc := NewAuthService(mockRepo, "test-secret")

		token, gotUser, loginErr := svc.Login(ctx, "user@example.com", "secret")
		require.NoError(t, loginErr)
		require.NotEmpty(t, token)
		require.Equal(t, user, gotUser)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(mocks.UserRepository)
		mockRepo.On("FindByEmail", mock.Anything, "user@example.com").Return((*domain.User)(nil), errors.New("db error")).Once()
		svc := NewAuthService(mockRepo, "test-secret")

		token, gotUser, loginErr := svc.Login(ctx, "user@example.com", "secret")
		require.ErrorIs(t, loginErr, ErrInvalidCredentials)
		require.Empty(t, token)
		require.Nil(t, gotUser)
	})

	t.Run("invalid password", func(t *testing.T) {
		mockRepo := new(mocks.UserRepository)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
		require.NoError(t, err)
		user := &domain.User{
			ID:           1,
			Email:        "user@example.com",
			Role:         domain.RoleEmployee,
			PasswordHash: string(hashedPassword),
		}
		mockRepo.On("FindByEmail", mock.Anything, "user@example.com").Return(user, nil).Once()
		svc := NewAuthService(mockRepo, "test-secret")

		token, gotUser, loginErr := svc.Login(ctx, "user@example.com", "wrong")
		require.ErrorIs(t, loginErr, ErrInvalidCredentials)
		require.Empty(t, token)
		require.Nil(t, gotUser)
	})
}

func TestValidateToken(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mocks.UserRepository)
	svc := NewAuthService(mockRepo, "test-secret")

	// Generate a valid token directly through concrete type to test parsing.
	concrete := svc.(*authService)
	validToken, err := concrete.generateToken(ctx, &domain.User{
		ID:    10,
		Email: "user@example.com",
		Role:  domain.RoleManager,
	})
	require.NoError(t, err)

	userID, role, validateErr := svc.ValidateToken(ctx, validToken)
	require.NoError(t, validateErr)
	require.Equal(t, 10, userID)
	require.Equal(t, domain.RoleManager, role)

	_, _, validateErr = svc.ValidateToken(ctx, validToken+"broken")
	require.Error(t, validateErr)
}
