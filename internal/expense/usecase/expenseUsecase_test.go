package usecase

import (
	"context"
	"testing"
	"time"

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
	userID := 1
	amountIDR := 10000
	description := "test"
	receiptURL := "test"

	mockExpense := new(mocks.ExpenseRepository)
	mockApproval := new(mocks.ApprovalRepository)

	usecase := NewExpenseUseCase(mockExpense, mockApproval)

	s.T().Run("success CreateExpense", func(t *testing.T) {
		mockExpense.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		result, err := usecase.CreateExpense(ctx, userID, amountIDR, description, receiptURL)
		s.Nil(err)
		s.NotNil(result)
	})

}

func (s *Suite) TestGetExpenseByID() {
	ctx := context.Background()
	userID := 1
	expenseID := 1

	mockExpense := new(mocks.ExpenseRepository)
	mockApproval := new(mocks.ApprovalRepository)

	usecase := NewExpenseUseCase(mockExpense, mockApproval)

	s.T().Run("success GetExpenseByID", func(t *testing.T) {
		currDate := time.Now()
		resp := &domain.Expense{
			ID:               userID,
			UserID:           userID,
			AmountIDR:        10000,
			Description:      "TEST",
			ReceiptURL:       "TEST",
			Status:           "Test",
			SubmittedAt:      currDate,
			ProcessedAt:      &currDate,
			RequiresApproval: false,
			AutoApproved:     false,
		}
		mockExpense.On("FindByID", mock.Anything, mock.Anything).Return(resp,nil).Once()

		result, err := usecase.GetExpenseByID(ctx, userID, expenseID)
		s.Nil(err)
		s.NotNil(result)
	})

}
