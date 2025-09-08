package worker

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/evrintobing17/expense-management-backend/internal/expense"
	"github.com/evrintobing17/expense-management-backend/internal/payment"
	"github.com/evrintobing17/expense-management-backend/pkg/utils"
)

type PaymentWorker struct {
	expenseRepo    expense.ExpenseRepository
	paymentService payment.PaymentService
	interval       time.Duration
}

func NewPaymentWorker(expenseRepo expense.ExpenseRepository, paymentService payment.PaymentService, interval time.Duration) *PaymentWorker {
	return &PaymentWorker{
		expenseRepo:    expenseRepo,
		paymentService: paymentService,
		interval:       interval,
	}
}

func (w *PaymentWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.processPayments(ctx)
		case <-ctx.Done():
			log.Println("Payment worker stopped")
			return
		}
	}
}

func (w *PaymentWorker) processPayments(ctx context.Context) {
	expenses, err := w.expenseRepo.FindByStatus(ctx, domain.ExpenseStatusApproved, domain.ExpenseStatusAutoApproved)
	if err != nil {
		log.Printf("Error fetching expenses for payment processing: %v", err)
		return
	}

	for _, expense := range expenses {
		err := w.processPayment(ctx, expense)
		if err != nil {
			log.Printf("Error processing payment for expense %d: %v", expense.ID, err)
			w.expenseRepo.UpdateStatus(ctx, expense.ID, domain.ExpenseStatusFailed, nil)
		}
	}
}

func (w *PaymentWorker) processPayment(ctx context.Context, expense *domain.Expense) error {
	idempotencyKey := utils.GenerateID()

	paymentResp, err := w.paymentService.ProcessPayment(ctx, expense.AmountIDR, idempotencyKey)
	if err != nil {
		if isIdempotencyError(err) {
			now := time.Now()
			return w.expenseRepo.UpdateStatus(ctx, expense.ID, domain.ExpenseStatusCompleted, &now)
		}
		return err
	}

	if paymentResp.Data.Status == "success" {
		now := time.Now()
		return w.expenseRepo.UpdateStatus(ctx, expense.ID, domain.ExpenseStatusCompleted, &now)
	}

	return fmt.Errorf("payment processing failed for expense %d", expense.ID)
}

func isIdempotencyError(err error) bool {
	return strings.Contains(err.Error(), "external id already exists")
}
