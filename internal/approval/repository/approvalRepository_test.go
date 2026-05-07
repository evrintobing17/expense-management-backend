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

func TestApprovalRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := &approvalRepository{db: db}
	query := regexp.QuoteMeta(`
		INSERT INTO approvals (expense_id, approver_id, status, notes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`)
	createdAt := time.Now()
	approval := &domain.Approval{
		ExpenseID:  1,
		ApproverID: 2,
		Status:     domain.ApprovalStatusApproved,
		Notes:      "looks good",
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "created_at"}).AddRow(88, createdAt)
		mock.ExpectQuery(query).
			WithArgs(approval.ExpenseID, approval.ApproverID, approval.Status, approval.Notes).
			WillReturnRows(rows)

		createErr := repo.Create(context.Background(), approval)
		require.NoError(t, createErr)
		require.Equal(t, 88, approval.ID)
		require.Equal(t, createdAt, approval.CreatedAt)
	})

	t.Run("query error", func(t *testing.T) {
		expectedErr := errors.New("insert failed")
		mock.ExpectQuery(query).
			WithArgs(approval.ExpenseID, approval.ApproverID, approval.Status, approval.Notes).
			WillReturnError(expectedErr)

		createErr := repo.Create(context.Background(), approval)
		require.ErrorIs(t, createErr, expectedErr)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApprovalRepositoryFindByExpenseID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := &approvalRepository{db: db}
	query := regexp.QuoteMeta(`
		SELECT id, expense_id, approver_id, status, notes, created_at
		FROM approvals
		WHERE expense_id = $1
	`)
	createdAt := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "expense_id", "approver_id", "status", "notes", "created_at"}).
			AddRow(1, 2, 3, "approved", "ok", createdAt)
		mock.ExpectQuery(query).WithArgs(2).WillReturnRows(rows)

		result, findErr := repo.FindByExpenseID(context.Background(), 2)
		require.NoError(t, findErr)
		require.NotNil(t, result)
		require.Equal(t, domain.ApprovalStatusApproved, result.Status)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(query).WithArgs(404).WillReturnError(sql.ErrNoRows)

		result, findErr := repo.FindByExpenseID(context.Background(), 404)
		require.NoError(t, findErr)
		require.Nil(t, result)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}
