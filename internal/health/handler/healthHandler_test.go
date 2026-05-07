package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestHealthHandlerCheck(t *testing.T) {
	t.Run("database ping failed", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		h := NewHealthHandler(db)
		mock.ExpectPing().WillReturnError(errors.New("down"))

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()
		h.Check(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		h := NewHealthHandler(db)
		mock.ExpectPing()

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()
		h.Check(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		require.Contains(t, rr.Body.String(), `"status":"ok"`)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
