package usecase

import (
	"context"
	"time"

	"github.com/evrintobing17/expense-management-backend/internal/approval"
	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/evrintobing17/expense-management-backend/internal/expense"
)

type expenseUseCase struct {
	expenseRepo  expense.ExpenseRepository
	approvalRepo approval.ApprovalRepository
}

func NewExpenseUseCase(expenseRepo expense.ExpenseRepository, approvalRepo approval.ApprovalRepository) expense.ExpenseUseCase {
	return &expenseUseCase{
		expenseRepo:  expenseRepo,
		approvalRepo: approvalRepo,
	}
}

func (uc *expenseUseCase) CreateExpense(ctx context.Context, userID int, amountIDR int, description, receiptURL string) (*domain.Expense, error) {
	if amountIDR < domain.MinExpenseAmount || amountIDR > domain.MaxExpenseAmount {
		return nil, domain.ErrInvalidAmount
	}

	if description == "" {
		return nil, domain.ErrMissingDescription
	}

	expense := &domain.Expense{
		UserID:      userID,
		AmountIDR:   amountIDR,
		Description: description,
		ReceiptURL:  receiptURL,
	}

	err := uc.expenseRepo.Create(ctx, expense)
	if err != nil {
		return nil, err
	}

	return expense, nil
}

func (uc *expenseUseCase) GetExpenseByID(ctx context.Context, id int, userID int) (*domain.Expense, error) {
	expense, err := uc.expenseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if expense == nil {
		return nil, domain.ErrExpenseNotFound
	}

	return expense, nil
}

func (uc *expenseUseCase) GetUserExpenses(ctx context.Context, userID int, status domain.ExpenseStatus, page, limit int) ([]*domain.Expense, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	return uc.expenseRepo.FindByUserID(ctx, userID, status, limit, offset)
}

func (uc *expenseUseCase) ApproveExpense(ctx context.Context, expenseID int, approverID int, notes string) error {
	expense, err := uc.expenseRepo.FindByID(ctx, expenseID)
	if err != nil {
		return err
	}

	if expense == nil {
		return domain.ErrExpenseNotFound
	}

	if expense.Status != domain.ExpenseStatusAwaitingApproval {
		return domain.ErrInvalidExpenseStatus
	}

	// Create approval record
	approval := &domain.Approval{
		ExpenseID:  expenseID,
		ApproverID: approverID,
		Status:     domain.ApprovalStatusApproved,
		Notes:      notes,
	}

	err = uc.approvalRepo.Create(ctx, approval)
	if err != nil {
		return err
	}

	now := time.Now()
	return uc.expenseRepo.UpdateStatus(ctx, expenseID, domain.ExpenseStatusApproved, &now)
}

func (uc *expenseUseCase) RejectExpense(ctx context.Context, expenseID int, approverID int, notes string) error {
	expense, err := uc.expenseRepo.FindByID(ctx, expenseID)
	if err != nil {
		return err
	}

	if expense == nil {
		return domain.ErrExpenseNotFound
	}

	if expense.Status != domain.ExpenseStatusAwaitingApproval {
		return domain.ErrInvalidExpenseStatus
	}

	// Create approval record
	approval := &domain.Approval{
		ExpenseID:  expenseID,
		ApproverID: approverID,
		Status:     domain.ApprovalStatusRejected,
		Notes:      notes,
	}

	err = uc.approvalRepo.Create(ctx, approval)
	if err != nil {
		return err
	}

	now := time.Now()
	return uc.expenseRepo.UpdateStatus(ctx, expenseID, domain.ExpenseStatusRejected, &now)
}

func (uc *expenseUseCase) GetPendingApproval(ctx context.Context) ([]*domain.Expense, error) {
	return uc.expenseRepo.FindPendingApproval(ctx)
}
