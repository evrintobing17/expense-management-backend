package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/evrintobing17/expense-management-backend/internal/expense"
)

type expenseRepository struct {
	db *sql.DB
}

func NewExpenseRepository(db *sql.DB) expense.ExpenseRepository {
	return &expenseRepository{db: db}
}

func (r *expenseRepository) Create(ctx context.Context, expense *domain.Expense) error {
	query := `
		INSERT INTO expenses (user_id, amount_idr, description, receipt_url, status, requires_approval, auto_approved)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, submitted_at
	`

	requiresApproval := expense.AmountIDR >= domain.ApprovalThreshold
	autoApproved := !requiresApproval
	initialStatus := domain.ExpenseStatusPending

	if autoApproved {
		initialStatus = domain.ExpenseStatusAutoApproved
	} else if requiresApproval {
		initialStatus = domain.ExpenseStatusAwaitingApproval
	}

	err := r.db.QueryRowContext(ctx, query,
		expense.UserID,
		expense.AmountIDR,
		expense.Description,
		expense.ReceiptURL,
		initialStatus,
		requiresApproval,
		autoApproved,
	).Scan(&expense.ID, &expense.SubmittedAt)

	if err != nil {
		return err
	}

	expense.Status = initialStatus
	expense.RequiresApproval = requiresApproval
	expense.AutoApproved = autoApproved

	return nil
}

func (r *expenseRepository) FindByID(ctx context.Context, id int) (*domain.Expense, error) {
	query := `
		SELECT id, user_id, amount_idr, description, receipt_url, status, submitted_at, processed_at, requires_approval, auto_approved
		FROM expenses
		WHERE id = $1
	`

	expense := &domain.Expense{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&expense.ID,
		&expense.UserID,
		&expense.AmountIDR,
		&expense.Description,
		&expense.ReceiptURL,
		&expense.Status,
		&expense.SubmittedAt,
		&expense.ProcessedAt,
		&expense.RequiresApproval,
		&expense.AutoApproved,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return expense, nil
}

func (r *expenseRepository) FindByUserID(ctx context.Context, userID int, status domain.ExpenseStatus, limit, offset int) ([]*domain.Expense, error) {
	query := `
		SELECT id, user_id, amount_idr, description, receipt_url, status, submitted_at, processed_at, requires_approval, auto_approved
		FROM expenses
		WHERE user_id = $1
	`

	args := []interface{}{userID}
	argIndex := 2

	if status != "" {
		query += " AND status = $" + strconv.Itoa(argIndex)
		args = append(args, status)
		argIndex++
	}

	query += " ORDER BY submitted_at DESC"

	if limit > 0 {
		query += " LIMIT $" + strconv.Itoa(argIndex)
		args = append(args, limit)
		argIndex++
	}

	if offset > 0 {
		query += " OFFSET $" + strconv.Itoa(argIndex)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*domain.Expense
	for rows.Next() {
		expense := &domain.Expense{}
		err := rows.Scan(
			&expense.ID,
			&expense.UserID,
			&expense.AmountIDR,
			&expense.Description,
			&expense.ReceiptURL,
			&expense.Status,
			&expense.SubmittedAt,
			&expense.ProcessedAt,
			&expense.RequiresApproval,
			&expense.AutoApproved,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

func (r *expenseRepository) UpdateStatus(ctx context.Context, id int, status domain.ExpenseStatus, processedAt *time.Time) error {
	query := `
		UPDATE expenses
		SET status = $1, processed_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, status, processedAt, id)
	return err
}

func (r *expenseRepository) FindPendingApproval(ctx context.Context) ([]*domain.Expense, error) {
	query := `
		SELECT id, user_id, amount_idr, description, receipt_url, status, submitted_at, processed_at, requires_approval, auto_approved
		FROM expenses
		WHERE status = $1 AND requires_approval = true
		ORDER BY submitted_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, domain.ExpenseStatusAwaitingApproval)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*domain.Expense
	for rows.Next() {
		expense := &domain.Expense{}
		err := rows.Scan(
			&expense.ID,
			&expense.UserID,
			&expense.AmountIDR,
			&expense.Description,
			&expense.ReceiptURL,
			&expense.Status,
			&expense.SubmittedAt,
			&expense.ProcessedAt,
			&expense.RequiresApproval,
			&expense.AutoApproved,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

func (r *expenseRepository) FindByStatus(ctx context.Context, statuses ...domain.ExpenseStatus) ([]*domain.Expense, error) {
	if len(statuses) == 0 {
		return nil, fmt.Errorf("at least one status must be provided")
	}

	query := `
		SELECT id, user_id, amount_idr, description, receipt_url, status, submitted_at, processed_at, requires_approval, auto_approved
		FROM expenses
		WHERE status IN (`

	// Create placeholders for each status
	placeholders := make([]string, len(statuses))
	args := make([]interface{}, len(statuses))
	for i, status := range statuses {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = status
	}
	query += strings.Join(placeholders, ", ") + ")"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []*domain.Expense
	for rows.Next() {
		expense := &domain.Expense{}
		err := rows.Scan(
			&expense.ID,
			&expense.UserID,
			&expense.AmountIDR,
			&expense.Description,
			&expense.ReceiptURL,
			&expense.Status,
			&expense.SubmittedAt,
			&expense.ProcessedAt,
			&expense.RequiresApproval,
			&expense.AutoApproved,
		)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}
