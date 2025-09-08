package domain

type PaymentRequest struct {
	Amount     int    `json:"amount"`
	ExternalID string `json:"external_id"`
}

type PaymentResponse struct {
	Data struct {
		ID         string `json:"id"`
		ExternalID string `json:"external_id"`
		Status     string `json:"status"`
	} `json:"data"`
	Message string `json:"message,omitempty"`
}
