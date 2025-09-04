package approval

import (
	"context"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
)

type ApprovalRepository interface {
	Create(ctx context.Context, approval *domain.Approval) error
	FindByExpenseID(ctx context.Context, expenseID int) (*domain.Approval, error)
}
