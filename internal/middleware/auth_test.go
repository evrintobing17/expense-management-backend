package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/evrintobing17/expense-management-backend/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("missing authorization header", func(t *testing.T) {
		mockAuth := new(mocks.AuthService)
		handler := AuthMiddleware(mockAuth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("invalid header format", func(t *testing.T) {
		mockAuth := new(mocks.AuthService)
		handler := AuthMiddleware(mockAuth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("success stores context values", func(t *testing.T) {
		mockAuth := new(mocks.AuthService)
		mockAuth.On("ValidateToken", mock.Anything, "valid-token").Return(123, domain.RoleManager, nil).Once()
		var gotUserID int
		var gotRole domain.Role
		handler := AuthMiddleware(mockAuth)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var ok bool
			gotUserID, ok = GetUserIDFromContext(r.Context())
			require.True(t, ok)
			gotRole, ok = GetUserRoleFromContext(r.Context())
			require.True(t, ok)
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, 123, gotUserID)
		require.Equal(t, domain.RoleManager, gotRole)
	})
}

func TestManagerOnlyMiddleware(t *testing.T) {
	nextCalled := false
	handler := ManagerOnlyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), userRoleKey, domain.RoleEmployee))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusForbidden, rr.Code)
	require.False(t, nextCalled)

	nextCalled = false
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2 = req2.WithContext(context.WithValue(req2.Context(), userRoleKey, domain.RoleManager))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	require.Equal(t, http.StatusOK, rr2.Code)
	require.True(t, nextCalled)
}

func TestCORS(t *testing.T) {
	handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	require.Equal(t, http.StatusCreated, rr2.Code)
}
