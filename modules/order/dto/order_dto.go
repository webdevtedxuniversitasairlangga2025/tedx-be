package dto

import (
	"errors"
	"time"
)

const (
	MESSAGE_FAILED_GET_DATA_FROM_BODY   = "failed get data from body"
	MESSAGE_FAILED_CREATE_ORDER         = "failed create order"
	MESSAGE_FAILED_GET_LIST_ORDER       = "failed get list order"
	MESSAGE_FAILED_GET_ORDER            = "failed get order"
	MESSAGE_FAILED_APPROVE_ORDER        = "failed approve order"
	MESSAGE_FAILED_REJECT_ORDER         = "failed reject order"
	MESSAGE_FAILED_UPLOAD_PROOF         = "failed upload payment proof"
	MESSAGE_FAILED_RELEASE_EXPIRED      = "failed release expired holds"
	MESSAGE_FAILED_DELETE_ORDER         = "failed delete order"
	MESSAGE_FAILED_CANCEL_ORDER         = "failed cancel order"

	MESSAGE_SUCCESS_CREATE_ORDER    = "success create order"
	MESSAGE_SUCCESS_GET_LIST_ORDER  = "success get list order"
	MESSAGE_SUCCESS_GET_ORDER       = "success get order"
	MESSAGE_SUCCESS_APPROVE_ORDER   = "success approve order"
	MESSAGE_SUCCESS_REJECT_ORDER    = "success reject order"
	MESSAGE_SUCCESS_UPLOAD_PROOF    = "success upload payment proof"
	MESSAGE_SUCCESS_RELEASE_EXPIRED = "success release expired holds"
	MESSAGE_SUCCESS_DELETE_ORDER    = "success delete order"
	MESSAGE_SUCCESS_CANCEL_ORDER    = "success cancel order"
)

var (
	ErrOrderNotFound            = errors.New("order not found")
	ErrTicketTierNotFound       = errors.New("ticket tier not found")
	ErrTicketTierInactive       = errors.New("ticket tier inactive")
	ErrTicketInactive           = errors.New("ticket inactive")
	ErrTicketSaleNotStarted     = errors.New("ticket sale not started")
	ErrTicketSaleEnded          = errors.New("ticket sale ended")
	ErrQuotaExceeded            = errors.New("quota exceeded")
	ErrQuantityOutOfRange       = errors.New("quantity must be between 1 and 5")
	ErrOrderNotAwaitingApproval = errors.New("order not awaiting approval")
	ErrOrderExpired             = errors.New("order expired")
	ErrOrderProofRequired       = errors.New("payment proof required before approve")
	ErrAttendeesCountMismatch   = errors.New("attendees count must match quantity")
	ErrInvalidUser              = errors.New("invalid user id")
)

type AttendeeCreateRequest struct {
	AttendeeName  string  `json:"attendee_name" binding:"required,min=1,max=255"`
	AttendeeEmail string  `json:"attendee_email" binding:"required,email,max=255"`
	AttendeePhone *string `json:"attendee_phone" binding:"omitempty,max=20"`
	AudienceType  string  `json:"audience_type" binding:"required,oneof=unair umum"`
	Institution   *string `json:"institution" binding:"omitempty,max=255"`
}

type OrderCreateRequest struct {
	TicketTierID string                  `json:"ticket_tier_id" binding:"required"`
	Quantity     int                     `json:"quantity" binding:"required,min=1,max=5"`
	Attendees    []AttendeeCreateRequest `json:"attendees" binding:"omitempty,dive"`
	// buyer_name dari form IdentifyStepper — koreksi users.name (tampil di admin)
	// (email akun tidak diubah: identitas login + risiko konflik unique)
	BuyerName *string `json:"buyer_name" binding:"omitempty,max=150"`
	// buyer_email dari form IdentifyStepper — snapshot di orders.buyer_email (tujuan e-ticket)
	BuyerEmail *string `json:"buyer_email" binding:"omitempty,email,max=225"`
	// buyer_phone dari form IdentifyStepper — simpan ke users.telp_number bila kosong
	// (buyer_phone di response = users.telp_number, bukan field order)
	BuyerPhone *string `json:"buyer_phone" binding:"omitempty,max=20"`
}

type OrderRejectRequest struct {
	Reason string `json:"reason" binding:"required,min=1,max=1000"`
}

type OrderUploadProofRequest struct {
	PaymentProofURL string `json:"payment_proof_url" binding:"required,url,max=500"`
}

type AttendeeTicketResponse struct {
	ID            string     `json:"id"`
	OrderID       string     `json:"order_id"`
	TicketCode    string     `json:"ticket_code"`
	AttendeeName  string     `json:"attendee_name"`
	AttendeeEmail string     `json:"attendee_email"`
	AttendeePhone *string    `json:"attendee_phone"`
	AudienceType  string     `json:"audience_type"`
	Institution   *string    `json:"institution"`
	IsSent        bool       `json:"is_sent"`
	SentAt        *time.Time `json:"sent_at"`
	IsUsed        bool       `json:"is_used"`
	UsedAt        *time.Time `json:"used_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type OrderResponse struct {
	ID              string                   `json:"id"`
	UserID          string                   `json:"user_id"`
	BuyerName       string                   `json:"buyer_name"`
	BuyerPhone      *string                  `json:"buyer_phone"`
	BuyerEmail      string                   `json:"buyer_email"`
	TicketTierID    string                   `json:"ticket_tier_id"`
	OrderNumber     string                   `json:"order_number"`
	Quantity        int                      `json:"quantity"`
	UnitPrice       string                   `json:"unit_price"`
	TotalAmount     string                   `json:"total_amount"`
	Status          string                   `json:"status"`
	ExpiredAt       time.Time                `json:"expired_at"`
	PaidAt          *time.Time               `json:"paid_at"`
	ApprovedBy      *string                  `json:"approved_by"`
	ApprovedAt      *time.Time               `json:"approved_at"`
	RejectedReason  *string                  `json:"rejected_reason"`
	PaymentProofURL *string                  `json:"payment_proof_url"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
	AttendeeTickets []AttendeeTicketResponse `json:"attendee_tickets"`
}

type PaginationRequest struct {
	Page    int `form:"page"`
	PerPage int `form:"per_page"`
}

type PaginationMeta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	MaxPage int   `json:"max_page"`
	Total   int64 `json:"total"`
}

type OrderPaginationResponse struct {
	Data []OrderResponse `json:"data"`
	Meta PaginationMeta  `json:"meta"`
}

type OrderFilter struct {
	Status *string `form:"status"`
}
