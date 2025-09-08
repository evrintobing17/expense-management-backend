package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/evrintobing17/expense-management-backend/internal/expense"
	"github.com/evrintobing17/expense-management-backend/internal/middleware"
	"github.com/gorilla/mux"
)

type ExpenseHandler struct {
	expenseUseCase expense.ExpenseUseCase
}

func NewExpenseHandler(expenseUseCase expense.ExpenseUseCase) *ExpenseHandler {
	return &ExpenseHandler{expenseUseCase: expenseUseCase}
}

func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		AmountIDR   int    `json:"amount_idr"`
		Description string `json:"description"`
		ReceiptURL  string `json:"receipt_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	expense, err := h.expenseUseCase.CreateExpense(ctx, userID, req.AmountIDR, req.Description, req.ReceiptURL)
	if err != nil {
		switch err {
		case domain.ErrInvalidAmount, domain.ErrMissingDescription:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(expense)
}

func (h *ExpenseHandler) GetExpenses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get query parameters
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = 10
	}

	expenses, err := h.expenseUseCase.GetUserExpenses(ctx, userID, domain.ExpenseStatus(status), page, limit)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}

func (h *ExpenseHandler) GetExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	expense, err := h.expenseUseCase.GetExpenseByID(ctx, id, userID)
	if err != nil {
		switch err {
		case domain.ErrExpenseNotFound:
			http.Error(w, "Expense not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expense)
}

func (h *ExpenseHandler) ApproveExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	approverID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.expenseUseCase.ApproveExpense(ctx, id, approverID, req.Notes)
	if err != nil {
		switch err {
		case domain.ErrExpenseNotFound:
			http.Error(w, "Expense not found", http.StatusNotFound)
		case domain.ErrInvalidExpenseStatus:
			http.Error(w, "Expense cannot be approved", http.StatusBadRequest)
		case domain.ErrUnauthorizedAction:
			http.Error(w, "Only managers can approve expenses", http.StatusForbidden)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ExpenseHandler) RejectExpense(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	approverID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.expenseUseCase.RejectExpense(ctx, id, approverID, req.Notes)
	if err != nil {
		switch err {
		case domain.ErrExpenseNotFound:
			http.Error(w, "Expense not found", http.StatusNotFound)
		case domain.ErrInvalidExpenseStatus:
			http.Error(w, "Expense cannot be rejected", http.StatusBadRequest)
		case domain.ErrUnauthorizedAction:
			http.Error(w, "Only managers can reject expenses", http.StatusForbidden)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ExpenseHandler) GetPendingApproval(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	expenses, err := h.expenseUseCase.GetPendingApproval(ctx)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}
