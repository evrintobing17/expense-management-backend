package usecase

import (
	"context"
	"testing"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/evrintobing17/expense-management-backend/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
}

func TestInitUseCase(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) TestCreateExpense() {
	ctx := context.Background()

	mockAuth := new(mocks.AuthService)

	usecase := NewAuthUseCase(mockAuth)

	s.T().Run("success CreateExpense", func(t *testing.T) {
		userResponse := &domain.User{
			ID:    1,
			Email: "test@example.com",
			Name:  "manager",
			Role:  "manager",
		}
		mockAuth.On("Login", mock.Anything, mock.Anything, mock.Anything).Return("some token", userResponse, nil).Once()

		token, result, err := usecase.Login(ctx, "TEST@example.com", "PWD")
		s.Nil(err)
		s.NotNil(result)
		s.NotNil(token)
	})

}
