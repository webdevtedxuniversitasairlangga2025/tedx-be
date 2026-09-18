package dto

import "errors"

const (
	MESSAGE_FAILED_PROCESS_WEBHOOK  = "failed process midtrans webhook"
	MESSAGE_SUCCESS_PROCESS_WEBHOOK = "success process midtrans webhook"
)

var (
	ErrInvalidSignature = errors.New("invalid signature key")
	ErrOrderNotFound    = errors.New("order not found")
)

type MidtransNotificationRequest struct {
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
}
