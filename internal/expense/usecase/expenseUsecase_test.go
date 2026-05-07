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

func TestCreateExpense(t *testing.T) {
	ctx := context.Background()
	userID := 1
	amountIDR := 10000
	description := "test"
	receiptURL := "test"

	t.Run("success", func(t *testing.T) {
		mockExpense := new(mocks.ExpenseRepository)
		mockApproval := new(mocks.ApprovalRepository)
		uc := NewExpenseUseCase(mockExpense, mockApproval)

		mockExpense.On("Create", mock.Anything, mock.AnythingOfType("*domain.Expense")).Return(nil).Once()

		result, err := uc.CreateExpense(ctx, userID, amountIDR, description, receiptURL)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, userID, result.UserID)
	})

	t.Run("invalid amount", func(t *testing.T) {
		mockExpense := new(mocks.ExpenseRepository)
		mockApproval := new(mocks.ApprovalRepository)
		uc := NewExpenseUseCase(mockExpense, mockApproval)

		result, err := uc.CreateExpense(ctx, userID, domain.MinExpenseAmount-1, description, receiptURL)
		require.ErrorIs(t, err, domain.ErrInvalidAmount)
		require.Nil(t, result)
		mockExpense.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("missing description", func(t *testing.T) {
		mockExpense := new(mocks.ExpenseRepository)
		mockApproval := new(mocks.ApprovalRepository)
		uc := NewExpenseUseCase(mockExpense, mockApproval)

		result, err := uc.CreateExpense(ctx, userID, amountIDR, "", receiptURL)
		require.ErrorIs(t, err, domain.ErrMissingDescription)
		require.Nil(t, result)
		mockExpense.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("repository error", func(t *testing.T) {
		mockExpense := new(mocks.ExpenseRepository)
		mockApproval := new(mocks.ApprovalRepository)
		uc := NewExpenseUseCase(mockExpense, mockApproval)
		expectedErr := errors.New("db failed")
		mockExpense.On("Create", mock.Anything, mock.AnythingOfType("*domain.Expense")).Return(expectedErr).Once()

		result, err := uc.CreateExpense(ctx, userID, amountIDR, description, receiptURL)
		require.ErrorIs(t, err, expectedErr)
		require.Nil(t, result)
	})
}

func TestGetExpenseByID(t *testing.T) {
	ctx := context.Background()
	userID := 1
	expenseID := 1

	t.Run("success", func(t *testing.T) {
		mockExpense := new(mocks.ExpenseRepository)
		mockApproval := new(mocks.ApprovalRepository)
		uc := NewExpenseUseCase(mockExpense, mockApproval)
		resp := &domain.Expense{
			ID:               expenseID,
			UserID:           userID,
			AmountIDR:        10000,
			Description:      "TEST",
			ReceiptURL:       "TEST",
			Status:           domain.ExpenseStatusPending,
			RequiresApproval: false,
			AutoApproved:     false,
		}
		mockExpense.On("FindByID", mock.Anything, expenseID).Return(resp, nil).Once()

		result, err := uc.GetExpenseByID(ctx, expenseID, userID)
		require.NoError(t, err)
		require.Equal(t, resp, result)
	})

	t.Run("not found", func(t *testing.T) {
		mockExpense := new(mocks.ExpenseRepository)
		mockApproval := new(mocks.ApprovalRepository)
		uc := NewExpenseUseCase(mockExpense, mockApproval)
		mockExpense.On("FindByID", mock.Anything, expenseID).Return((*domain.Expense)(nil), nil).Once()

		result, err := uc.GetExpenseByID(ctx, expenseID, userID)
		require.ErrorIs(t, err, domain.ErrExpenseNotFound)
		require.Nil(t, result)
	})

	t.Run("unauthorized", func(t *testing.T) {
		mockExpense := new(mocks.ExpenseRepository)
		mockApproval := new(mocks.ApprovalRepository)
		uc := NewExpenseUseCase(mockExpense, mockApproval)
		resp := &domain.Expense{
			ID:               expenseID,
			UserID:           2,
			AmountIDR:        10000,
			Description:      "TEST",
			ReceiptURL:       "TEST",
			Status:           domain.ExpenseStatusPending,
			RequiresApproval: false,
			AutoApproved:     false,
		}
		mockExpense.On("FindByID", mock.Anything, expenseID).Return(resp, nil).Once()

		result, err := uc.GetExpenseByID(ctx, expenseID, userID)
		require.ErrorIs(t, err, domain.ErrUnauthorizedAction)
		require.Nil(t, result)
	})

	t.Run("repository error", func(t *testing.T) {
		mockExpense := new(mocks.ExpenseRepository)
		mockApproval := new(mocks.ApprovalRepository)
		uc := NewExpenseUseCase(mockExpense, mockApproval)
		expectedErr := errors.New("db failed")
		mockExpense.On("FindByID", mock.Anything, expenseID).Return((*domain.Expense)(nil), expectedErr).Once()

		result, err := uc.GetExpenseByID(ctx, expenseID, userID)
		require.ErrorIs(t, err, expectedErr)
		require.Nil(t, result)
	})
}

func TestGetUserExpenses(t *testing.T) {
	ctx := context.Background()
	mockExpense := new(mocks.ExpenseRepository)
	mockApproval := new(mocks.ApprovalRepository)
	uc := NewExpenseUseCase(mockExpense, mockApproval)
	userID := 10
	status := domain.ExpenseStatusApproved
	expected := []*domain.Expense{{ID: 1, UserID: userID, Status: status}}

	mockExpense.On("FindByUserID", mock.Anything, userID, status, 10, 0).Return(expected, nil).Once()

	result, err := uc.GetUserExpenses(ctx, userID, status, 0, 0)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestApproveExpense(t *testing.T) {
	testProcessApprovalFlow(t, true)
}

func TestRejectExpense(t *testing.T) {
	testProcessApprovalFlow(t, false)
}

func testProcessApprovalFlow(t *testing.T, approve bool) {
	ctx := context.Background()
	expenseID := 9
	approverID := 3
	notes := "ok"
	expenseStatus := domain.ExpenseStatusApproved
	approvalStatus := domain.ApprovalStatusApproved
	if !approve {
		expenseStatus = domain.ExpenseStatusRejected
		approvalStatus = domain.ApprovalStatusRejected
	}

	t.Run("success", func(t *testing.T) {
		mockExpense := new(mocks.ExpenseRepository)
		mockApproval := new(mocks.ApprovalRepository)
		uc := NewExpenseUseCase(mockExpense, mockApproval)
		mockExpense.On("FindByID", mock.Anything, expenseID).
			Return(&domain.Expense{ID: expenseID, Status: domain.ExpenseStatusAwaitingApproval}, nil).Once()
		mockApproval.On("Create", mock.Anything, mock.MatchedBy(func(a *domain.Approval) bool {
			return a.ExpenseID == expenseID && a.ApproverID == approverID && a.Status == approvalStatus
		})).Return(nil).Once()
		mockExpense.On("UpdateStatus", mock.Anything, expenseID, expenseStatus, mock.AnythingOfType("*time.Time")).Return(nil).Once()

		var err error
		if approve {
			err = uc.ApproveExpense(ctx, expenseID, approverID, notes)
		} else {
			err = uc.RejectExpense(ctx, expenseID, approverID, notes)
		}
		require.NoError(t, err)
	})

	t.Run("expense not found", func(t *testing.T) {
		mockExpense := new(mocks.ExpenseRepository)
		mockApproval := new(mocks.ApprovalRepository)
		uc := NewExpenseUseCase(mockExpense, mockApproval)
		mockExpense.On("FindByID", mock.Anything, expenseID).Return((*domain.Expense)(nil), nil).Once()

		var err error
		if approve {
			err = uc.ApproveExpense(ctx, expenseID, approverID, notes)
		} else {
			err = uc.RejectExpense(ctx, expenseID, approverID, notes)
		}
		require.ErrorIs(t, err, domain.ErrExpenseNotFound)
		mockApproval.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("invalid expense status", func(t *testing.T) {
		mockExpense := new(mocks.ExpenseRepository)
		mockApproval := new(mocks.ApprovalRepository)
		uc := NewExpenseUseCase(mockExpense, mockApproval)
		mockExpense.On("FindByID", mock.Anything, expenseID).
			Return(&domain.Expense{ID: expenseID, Status: domain.ExpenseStatusApproved}, nil).Once()

		var err error
		if approve {
			err = uc.ApproveExpense(ctx, expenseID, approverID, notes)
		} else {
			err = uc.RejectExpense(ctx, expenseID, approverID, notes)
		}
		require.ErrorIs(t, err, domain.ErrInvalidExpenseStatus)
		mockApproval.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("approval repository error", func(t *testing.T) {
		mockExpense := new(mocks.ExpenseRepository)
		mockApproval := new(mocks.ApprovalRepository)
		uc := NewExpenseUseCase(mockExpense, mockApproval)
		expectedErr := errors.New("approval create failed")
		mockExpense.On("FindByID", mock.Anything, expenseID).
			Return(&domain.Expense{ID: expenseID, Status: domain.ExpenseStatusAwaitingApproval}, nil).Once()
		mockApproval.On("Create", mock.Anything, mock.AnythingOfType("*domain.Approval")).Return(expectedErr).Once()

		var err error
		if approve {
			err = uc.ApproveExpense(ctx, expenseID, approverID, notes)
		} else {
			err = uc.RejectExpense(ctx, expenseID, approverID, notes)
		}
		require.ErrorIs(t, err, expectedErr)
		mockExpense.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestGetPendingApproval(t *testing.T) {
	ctx := context.Background()
	mockExpense := new(mocks.ExpenseRepository)
	mockApproval := new(mocks.ApprovalRepository)
	uc := NewExpenseUseCase(mockExpense, mockApproval)
	expected := []*domain.Expense{{ID: 1, Status: domain.ExpenseStatusAwaitingApproval}}
	mockExpense.On("FindPendingApproval", mock.Anything).Return(expected, nil).Once()

	result, err := uc.GetPendingApproval(ctx)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}
