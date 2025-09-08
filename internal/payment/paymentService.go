package payment

import (
	"context"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
)

type PaymentService interface {
	ProcessPayment(ctx context.Context, amount int, externalID string) (*domain.PaymentResponse, error)
}
