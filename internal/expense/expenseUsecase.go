package expense

import (
	"context"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
)

type ExpenseUseCase interface {
	CreateExpense(ctx context.Context, userID int, amountIDR int, description, receiptURL string) (*domain.Expense, error)
	GetExpenseByID(ctx context.Context, id int, userID int) (*domain.Expense, error)
	GetUserExpenses(ctx context.Context, userID int, status domain.ExpenseStatus, page, limit int) ([]*domain.Expense, error)
	ApproveExpense(ctx context.Context, expenseID int, approverID int, notes string) error
	RejectExpense(ctx context.Context, expenseID int, approverID int, notes string) error
	GetPendingApproval(ctx context.Context) ([]*domain.Expense, error)
}
