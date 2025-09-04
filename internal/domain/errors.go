package domain

import "errors"

var (
	ErrExpenseNotFound      = errors.New("expense not found")
	ErrInvalidExpenseStatus = errors.New("invalid expense status for this operation")
	ErrUnauthorizedAction   = errors.New("unauthorized action")
	ErrInvalidAmount        = errors.New("amount must be between 10,000 and 50,000,000 IDR")
	ErrMissingDescription   = errors.New("description is required")
)
