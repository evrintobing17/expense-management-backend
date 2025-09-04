package expense

import (
	"context"
	"time"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
)

type ExpenseRepository interface {
	Create(ctx context.Context, expense *domain.Expense) error
	FindByID(ctx context.Context, id int) (*domain.Expense, error)
	FindByUserID(ctx context.Context, userID int, status domain.ExpenseStatus, limit, offset int) ([]*domain.Expense, error)
	UpdateStatus(ctx context.Context, id int, status domain.ExpenseStatus, processedAt *time.Time) error
	FindPendingApproval(ctx context.Context) ([]*domain.Expense, error)
}
