package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestExpenseRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &expenseRepository{db: db}

	query := regexp.QuoteMeta(`
		INSERT INTO expenses (user_id, amount_idr, description, receipt_url, status, requires_approval, auto_approved)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, submitted_at
	`)
	submittedAt := time.Now()

	t.Run("auto approved below threshold", func(t *testing.T) {
		exp := &domain.Expense{UserID: 1, AmountIDR: 100_000, Description: "taxi", ReceiptURL: "url"}
		rows := sqlmock.NewRows([]string{"id", "submitted_at"}).AddRow(5, submittedAt)
		mock.ExpectQuery(query).
			WithArgs(1, 100_000, "taxi", "url", domain.ExpenseStatusAutoApproved, false, true).
			WillReturnRows(rows)

		createErr := repo.Create(context.Background(), exp)
		require.NoError(t, createErr)
		require.Equal(t, domain.ExpenseStatusAutoApproved, exp.Status)
		require.False(t, exp.RequiresApproval)
		require.True(t, exp.AutoApproved)
	})

	t.Run("awaiting approval at threshold", func(t *testing.T) {
		exp := &domain.Expense{UserID: 1, AmountIDR: domain.ApprovalThreshold, Description: "hotel", ReceiptURL: "url"}
		rows := sqlmock.NewRows([]string{"id", "submitted_at"}).AddRow(6, submittedAt)
		mock.ExpectQuery(query).
			WithArgs(1, domain.ApprovalThreshold, "hotel", "url", domain.ExpenseStatusAwaitingApproval, true, false).
			WillReturnRows(rows)

		createErr := repo.Create(context.Background(), exp)
		require.NoError(t, createErr)
		require.Equal(t, domain.ExpenseStatusAwaitingApproval, exp.Status)
		require.True(t, exp.RequiresApproval)
		require.False(t, exp.AutoApproved)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepositoryFindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &expenseRepository{db: db}
	query := regexp.QuoteMeta(`
		SELECT id, user_id, amount_idr, description, receipt_url, status, submitted_at, processed_at, requires_approval, auto_approved
		FROM expenses
		WHERE id = $1
	`)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "amount_idr", "description", "receipt_url", "status", "submitted_at", "processed_at", "requires_approval", "auto_approved"}).
		AddRow(1, 2, 30000, "meal", "url", "pending", now, now, false, false)
	mock.ExpectQuery(query).WithArgs(1).WillReturnRows(rows)
	result, findErr := repo.FindByID(context.Background(), 1)
	require.NoError(t, findErr)
	require.NotNil(t, result)
	require.Equal(t, 2, result.UserID)

	mock.ExpectQuery(query).WithArgs(999).WillReturnError(sql.ErrNoRows)
	result, findErr = repo.FindByID(context.Background(), 999)
	require.NoError(t, findErr)
	require.Nil(t, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepositoryFindByStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &expenseRepository{db: db}

	_, findErr := repo.FindByStatus(context.Background())
	require.Error(t, findErr)

	query := regexp.QuoteMeta(`
		SELECT id, user_id, amount_idr, description, receipt_url, status, submitted_at, processed_at, requires_approval, auto_approved
		FROM expenses
		WHERE status IN ($1, $2)`)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "amount_idr", "description", "receipt_url", "status", "submitted_at", "processed_at", "requires_approval", "auto_approved"}).
		AddRow(1, 2, 40000, "flight", "url", "approved", now, now, true, false)
	mock.ExpectQuery(query).WithArgs(domain.ExpenseStatusApproved, domain.ExpenseStatusRejected).WillReturnRows(rows)

	result, findErr := repo.FindByStatus(context.Background(), domain.ExpenseStatusApproved, domain.ExpenseStatusRejected)
	require.NoError(t, findErr)
	require.Len(t, result, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepositoryUpdateStatusAndPending(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &expenseRepository{db: db}
	now := time.Now()

	updateQuery := regexp.QuoteMeta(`
		UPDATE expenses
		SET status = $1, processed_at = $2
		WHERE id = $3
	`)
	mock.ExpectExec(updateQuery).
		WithArgs(domain.ExpenseStatusApproved, &now, 10).
		WillReturnResult(sqlmock.NewResult(0, 1))
	updateErr := repo.UpdateStatus(context.Background(), 10, domain.ExpenseStatusApproved, &now)
	require.NoError(t, updateErr)

	pendingQuery := regexp.QuoteMeta(`
		SELECT id, user_id, amount_idr, description, receipt_url, status, submitted_at, processed_at, requires_approval, auto_approved
		FROM expenses
		WHERE status = $1 AND requires_approval = true
		ORDER BY submitted_at ASC
	`)
	rows := sqlmock.NewRows([]string{"id", "user_id", "amount_idr", "description", "receipt_url", "status", "submitted_at", "processed_at", "requires_approval", "auto_approved"}).
		AddRow(1, 2, 3000000, "conference", "url", "awaiting_approval", now, nil, true, false)
	mock.ExpectQuery(pendingQuery).WithArgs(domain.ExpenseStatusAwaitingApproval).WillReturnRows(rows)
	pending, findErr := repo.FindPendingApproval(context.Background())
	require.NoError(t, findErr)
	require.Len(t, pending, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepositoryFindByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := &expenseRepository{db: db}
	now := time.Now()

	queryWithFilter := regexp.QuoteMeta(`
		SELECT id, user_id, amount_idr, description, receipt_url, status, submitted_at, processed_at, requires_approval, auto_approved
		FROM expenses
		WHERE user_id = $1
	 AND status = $2 ORDER BY submitted_at DESC LIMIT $3`)
	rows := sqlmock.NewRows([]string{"id", "user_id", "amount_idr", "description", "receipt_url", "status", "submitted_at", "processed_at", "requires_approval", "auto_approved"}).
		AddRow(1, 4, 50000, "parking", "url", "approved", now, nil, false, true)
	mock.ExpectQuery(queryWithFilter).WithArgs(4, domain.ExpenseStatusApproved, 10).WillReturnRows(rows)

	result, findErr := repo.FindByUserID(context.Background(), 4, domain.ExpenseStatusApproved, 10, 0)
	require.NoError(t, findErr)
	require.Len(t, result, 1)

	queryNoFilter := regexp.QuoteMeta(`
		SELECT id, user_id, amount_idr, description, receipt_url, status, submitted_at, processed_at, requires_approval, auto_approved
		FROM expenses
		WHERE user_id = $1
	 ORDER BY submitted_at DESC`)
	mock.ExpectQuery(queryNoFilter).WithArgs(4).WillReturnError(errors.New("query failed"))
	_, findErr = repo.FindByUserID(context.Background(), 4, "", 0, 0)
	require.Error(t, findErr)

	require.NoError(t, mock.ExpectationsWereMet())
}
