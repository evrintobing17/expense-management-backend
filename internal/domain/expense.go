package domain

import (
	"time"
)

type ExpenseStatus string

const (
	ExpenseStatusPending          ExpenseStatus = "pending"
	ExpenseStatusAwaitingApproval ExpenseStatus = "awaiting_approval"
	ExpenseStatusApproved         ExpenseStatus = "approved"
	ExpenseStatusRejected         ExpenseStatus = "rejected"
	ExpenseStatusAutoApproved     ExpenseStatus = "auto_approved"
	ExpenseStatusProcessing       ExpenseStatus = "processing"
	ExpenseStatusCompleted        ExpenseStatus = "completed"
	ExpenseStatusFailed           ExpenseStatus = "failed"
)

type Expense struct {
	ID               int           `json:"id"`
	UserID           int           `json:"user_id"`
	AmountIDR        int           `json:"amount_idr"`
	Description      string        `json:"description"`
	ReceiptURL       string        `json:"receipt_url"`
	Status           ExpenseStatus `json:"status"`
	SubmittedAt      time.Time     `json:"submitted_at"`
	ProcessedAt      *time.Time    `json:"processed_at"`
	RequiresApproval bool          `json:"requires_approval"`
	AutoApproved     bool          `json:"auto_approved"`
}
