package dto

import (
	"errors"
	"time"
)

const (
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"

	MESSAGE_FAILED_CREATE_TICKET   = "failed create ticket"
	MESSAGE_FAILED_GET_LIST_TICKET = "failed get list ticket"
	MESSAGE_FAILED_GET_TICKET      = "failed get ticket"
	MESSAGE_FAILED_UPDATE_TICKET   = "failed update ticket"
	MESSAGE_FAILED_DELETE_TICKET   = "failed delete ticket"

	MESSAGE_SUCCESS_CREATE_TICKET   = "success create ticket"
	MESSAGE_SUCCESS_GET_LIST_TICKET = "success get list ticket"
	MESSAGE_SUCCESS_GET_TICKET      = "success get ticket"
	MESSAGE_SUCCESS_UPDATE_TICKET   = "success update ticket"
	MESSAGE_SUCCESS_DELETE_TICKET   = "success delete ticket"

	MESSAGE_FAILED_CREATE_TICKET_TIER = "failed create ticket tier"
	MESSAGE_FAILED_UPDATE_TICKET_TIER = "failed update ticket tier"
	MESSAGE_FAILED_DELETE_TICKET_TIER = "failed delete ticket tier"

	MESSAGE_SUCCESS_CREATE_TICKET_TIER = "success create ticket tier"
	MESSAGE_SUCCESS_UPDATE_TICKET_TIER = "success update ticket tier"
	MESSAGE_SUCCESS_DELETE_TICKET_TIER = "success delete ticket tier"
)

var (
	ErrInvalidPrice       = errors.New("invalid price format")
	ErrPriceOutOfRange    = errors.New("price must be between 0 and 99999999.99")
	ErrTicketNotFound     = errors.New("ticket not found")
	ErrTicketTierNotFound = errors.New("ticket tier not found")
	ErrInvalidSaleWindow  = errors.New("sale_end must be after sale_start")
	ErrQuotaBelowFilled   = errors.New("quota cannot be lower than quota already filled")
)

// ---------- Ticket ----------

type TicketCreateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"required,min=1"`
}

type TicketUpdateRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,min=1"`
	IsActive    *bool   `json:"is_active"`
}

type TicketFilter struct {
	IsActive *bool `form:"is_active"`
}

type TicketResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	IsActive    bool                 `json:"is_active"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Tiers       []TicketTierResponse `json:"tiers"`
}

// ---------- Ticket Tier ----------

type TicketTierCreateRequest struct {
	Tier      string     `json:"tier" binding:"required,min=1,max=50"`
	Price     string     `json:"price" binding:"required"`
	Quota     int        `json:"quota" binding:"required,min=1"`
	SaleStart *time.Time `json:"sale_start"`
	SaleEnd   *time.Time `json:"sale_end"`
}

type TicketTierUpdateRequest struct {
	Tier      *string    `json:"tier" binding:"omitempty,min=1,max=50"`
	Price     *string    `json:"price" binding:"omitempty"`
	Quota     *int       `json:"quota" binding:"omitempty,min=1"`
	SaleStart *time.Time `json:"sale_start"`
	SaleEnd   *time.Time `json:"sale_end"`
	IsActive  *bool      `json:"is_active"`
}

type TicketTierResponse struct {
	ID          string     `json:"id"`
	TicketID    string     `json:"ticket_id"`
	Tier        string     `json:"tier"`
	Price       string     `json:"price"`
	Quota       int        `json:"quota"`
	QuotaFilled int        `json:"quota_filled"`
	QuotaLeft   int        `json:"quota_left"`
	SaleStart   *time.Time `json:"sale_start"`
	SaleEnd     *time.Time `json:"sale_end"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
