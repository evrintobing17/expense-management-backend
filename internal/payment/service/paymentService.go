package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/evrintobing17/expense-management-backend/internal/domain"
	"github.com/evrintobing17/expense-management-backend/internal/payment"
)

type paymentService struct {
	baseURL string
	client  *http.Client
}

func NewPaymentService(baseURL string) payment.PaymentService {
	return &paymentService{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *paymentService) ProcessPayment(ctx context.Context, amount int, externalID string) (*domain.PaymentResponse, error) {
	url := s.baseURL + "/v1/payments"

	reqBody := domain.PaymentRequest{
		Amount:     amount,
		ExternalID: externalID,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var paymentResp domain.PaymentResponse
	err = json.Unmarshal(body, &paymentResp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment failed: %s", paymentResp.Message)
	}

	return &paymentResp, nil
}
