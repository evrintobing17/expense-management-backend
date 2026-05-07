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

func TestUserRepositoryFindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := &userRepository{db: db}
	query := regexp.QuoteMeta(`
		SELECT id, email, name, role, password_hash, created_at
		FROM users
		WHERE id = $1
	`)
	createdAt := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "name", "role", "password_hash", "created_at"}).
			AddRow(1, "user@example.com", "User", "employee", "hash", createdAt)
		mock.ExpectQuery(query).WithArgs(1).WillReturnRows(rows)

		user, findErr := repo.FindByID(context.Background(), 1)
		require.NoError(t, findErr)
		require.Equal(t, &domain.User{
			ID:           1,
			Email:        "user@example.com",
			Name:         "User",
			Role:         domain.RoleEmployee,
			PasswordHash: "hash",
			CreatedAt:    createdAt,
		}, user)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(query).WithArgs(100).WillReturnError(sql.ErrNoRows)

		user, findErr := repo.FindByID(context.Background(), 100)
		require.NoError(t, findErr)
		require.Nil(t, user)
	})

	t.Run("query error", func(t *testing.T) {
		expectedErr := errors.New("db error")
		mock.ExpectQuery(query).WithArgs(2).WillReturnError(expectedErr)

		user, findErr := repo.FindByID(context.Background(), 2)
		require.ErrorIs(t, findErr, expectedErr)
		require.Nil(t, user)
	})

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepositoryFindByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := &userRepository{db: db}
	query := regexp.QuoteMeta(`
		SELECT id, email, name, role, password_hash, created_at
		FROM users
		WHERE email = $1
	`)
	createdAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "email", "name", "role", "password_hash", "created_at"}).
		AddRow(1, "user@example.com", "User", "employee", "hash", createdAt)
	mock.ExpectQuery(query).WithArgs("user@example.com").WillReturnRows(rows)

	user, findErr := repo.FindByEmail(context.Background(), "user@example.com")
	require.NoError(t, findErr)
	require.NotNil(t, user)
	require.Equal(t, "user@example.com", user.Email)
	require.NoError(t, mock.ExpectationsWereMet())
}
