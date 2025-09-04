package repository

import (
	"context"
	"database/sql"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/evrintobing17/expense-management-backend/internal/user"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) user.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	query := `
		SELECT id, email, name, role, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, name, role, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
