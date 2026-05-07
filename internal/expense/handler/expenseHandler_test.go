package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/evrintobing17/expense-management-backend/internal/middleware"
	"github.com/evrintobing17/expense-management-backend/mocks"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func withUserID(req *http.Request, userID int) *http.Request {
	// Reuse auth middleware to set the same context keys used by handlers.
	mockAuth := new(mocks.AuthService)
	mockAuth.On("ValidateToken", mock.Anything, "t").Return(userID, domain.RoleEmployee, nil).Maybe()
	rr := httptest.NewRecorder()
	next := middleware.AuthMiddleware(mockAuth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = r
	}))
	req.Header.Set("Authorization", "Bearer t")
	next.ServeHTTP(rr, req)
	return req
}

func TestExpenseHandlerCreateExpense(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		mockUC := new(mocks.ExpenseUseCase)
		h := NewExpenseHandler(mockUC)
		req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(`{"amount_idr":10000,"description":"meal","receipt_url":"u"}`))
		rr := httptest.NewRecorder()
		h.CreateExpense(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("domain bad request", func(t *testing.T) {
		mockUC := new(mocks.ExpenseUseCase)
		h := NewExpenseHandler(mockUC)
		req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(`{"amount_idr":1,"description":"meal","receipt_url":"u"}`))
		req = withUserID(req, 1)
		rr := httptest.NewRecorder()
		mockUC.On("CreateExpense", mock.Anything, 1, 1, "meal", "u").Return((*domain.Expense)(nil), domain.ErrInvalidAmount).Once()

		h.CreateExpense(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("success", func(t *testing.T) {
		mockUC := new(mocks.ExpenseUseCase)
		h := NewExpenseHandler(mockUC)
		req := httptest.NewRequest(http.MethodPost, "/expenses", strings.NewReader(`{"amount_idr":10000,"description":"meal","receipt_url":"u"}`))
		req = withUserID(req, 1)
		rr := httptest.NewRecorder()
		exp := &domain.Expense{ID: 1, UserID: 1, AmountIDR: 10000, Description: "meal", ReceiptURL: "u"}
		mockUC.On("CreateExpense", mock.Anything, 1, 10000, "meal", "u").Return(exp, nil).Once()

		h.CreateExpense(rr, req)
		require.Equal(t, http.StatusCreated, rr.Code)
		require.Contains(t, rr.Body.String(), `"id":1`)
	})
}

func TestExpenseHandlerReadEndpoints(t *testing.T) {
	t.Run("get expenses success", func(t *testing.T) {
		mockUC := new(mocks.ExpenseUseCase)
		h := NewExpenseHandler(mockUC)
		req := httptest.NewRequest(http.MethodGet, "/expenses?status=approved&page=2&limit=5", nil)
		req = withUserID(req, 7)
		rr := httptest.NewRecorder()
		mockUC.On("GetUserExpenses", mock.Anything, 7, domain.ExpenseStatus("approved"), 2, 5).
			Return([]*domain.Expense{{ID: 1}}, nil).Once()

		h.GetExpenses(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("get expense not found", func(t *testing.T) {
		mockUC := new(mocks.ExpenseUseCase)
		h := NewExpenseHandler(mockUC)
		req := httptest.NewRequest(http.MethodGet, "/expenses/9", nil)
		req = withUserID(req, 7)
		req = mux.SetURLVars(req, map[string]string{"id": "9"})
		rr := httptest.NewRecorder()
		mockUC.On("GetExpenseByID", mock.Anything, 9, 7).Return((*domain.Expense)(nil), domain.ErrExpenseNotFound).Once()

		h.GetExpense(rr, req)
		require.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestExpenseHandlerApprovalEndpoints(t *testing.T) {
	t.Run("approve invalid id", func(t *testing.T) {
		mockUC := new(mocks.ExpenseUseCase)
		h := NewExpenseHandler(mockUC)
		req := httptest.NewRequest(http.MethodPost, "/expenses/x/approve", strings.NewReader(`{"notes":"ok"}`))
		req = withUserID(req, 3)
		req = mux.SetURLVars(req, map[string]string{"id": "x"})
		rr := httptest.NewRecorder()

		h.ApproveExpense(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("approve forbidden", func(t *testing.T) {
		mockUC := new(mocks.ExpenseUseCase)
		h := NewExpenseHandler(mockUC)
		req := httptest.NewRequest(http.MethodPost, "/expenses/10/approve", strings.NewReader(`{"notes":"ok"}`))
		req = withUserID(req, 3)
		req = mux.SetURLVars(req, map[string]string{"id": "10"})
		rr := httptest.NewRecorder()
		mockUC.On("ApproveExpense", mock.Anything, 10, 3, "ok").Return(domain.ErrUnauthorizedAction).Once()

		h.ApproveExpense(rr, req)
		require.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("reject success", func(t *testing.T) {
		mockUC := new(mocks.ExpenseUseCase)
		h := NewExpenseHandler(mockUC)
		req := httptest.NewRequest(http.MethodPost, "/expenses/10/reject", strings.NewReader(`{"notes":"reject"}`))
		req = withUserID(req, 3)
		req = mux.SetURLVars(req, map[string]string{"id": "10"})
		rr := httptest.NewRecorder()
		mockUC.On("RejectExpense", mock.Anything, 10, 3, "reject").Return(nil).Once()

		h.RejectExpense(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestExpenseHandlerGetPendingApproval(t *testing.T) {
	mockUC := new(mocks.ExpenseUseCase)
	h := NewExpenseHandler(mockUC)
	req := httptest.NewRequest(http.MethodGet, "/expenses/pending-approval", nil).WithContext(context.Background())
	rr := httptest.NewRecorder()
	mockUC.On("GetPendingApproval", mock.Anything).Return(([]*domain.Expense)(nil), errors.New("db error")).Once()

	h.GetPendingApproval(rr, req)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
}
