package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/evrintobing17/expense-management-backend/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthHandlerLogin(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		mockUC := new(mocks.AuthUseCase)
		h := NewAuthHandler(mockUC)
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader("{"))
		rr := httptest.NewRecorder()

		h.Login(rr, req)
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		mockUC := new(mocks.AuthUseCase)
		h := NewAuthHandler(mockUC)
		body := `{"email":"user@example.com","password":"wrong"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
		rr := httptest.NewRecorder()
		mockUC.On("Login", mock.Anything, "user@example.com", "wrong").Return("", (*domain.UserResponse)(nil), errors.New("invalid credentials")).Once()

		h.Login(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("success", func(t *testing.T) {
		mockUC := new(mocks.AuthUseCase)
		h := NewAuthHandler(mockUC)
		body := `{"email":"user@example.com","password":"secret"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
		rr := httptest.NewRecorder()
		user := &domain.UserResponse{ID: 1, Email: "user@example.com", Name: "User", Role: domain.RoleEmployee}
		mockUC.On("Login", mock.Anything, "user@example.com", "secret").Return("token-123", user, nil).Once()

		h.Login(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		require.Contains(t, rr.Body.String(), `"token":"token-123"`)
	})
}
