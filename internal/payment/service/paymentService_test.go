package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessPayment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "/v1/payments", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"id":"pay_1","external_id":"ext_1","status":"success"}}`))
		}))
		defer server.Close()

		svc := NewPaymentService(server.URL)
		resp, err := svc.ProcessPayment(context.Background(), 12000, "ext_1")
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, "pay_1", resp.Data.ID)
	})

	t.Run("failed payment response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"insufficient funds"}`))
		}))
		defer server.Close()

		svc := NewPaymentService(server.URL)
		resp, err := svc.ProcessPayment(context.Background(), 12000, "ext_2")
		require.Error(t, err)
		require.Nil(t, resp)
	})

	t.Run("invalid json response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`not-json`))
		}))
		defer server.Close()

		svc := NewPaymentService(server.URL)
		resp, err := svc.ProcessPayment(context.Background(), 12000, "ext_3")
		require.Error(t, err)
		require.Nil(t, resp)
	})
}
