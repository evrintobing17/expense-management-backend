package domain

import (
	"time"
)

type ApprovalStatus string

const (
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

type Approval struct {
	ID         int            `json:"id"`
	ExpenseID  int            `json:"expense_id"`
	ApproverID int            `json:"approver_id"`
	Status     ApprovalStatus `json:"status"`
	Notes      string         `json:"notes"`
	CreatedAt  time.Time      `json:"created_at"`
}
