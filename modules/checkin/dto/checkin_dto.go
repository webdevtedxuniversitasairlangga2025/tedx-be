package dto

import (
	"errors"
	"time"
)

var (
	ErrInvalidTicketCode      = errors.New("invalid ticket code")
	ErrInvalidChecker         = errors.New("invalid checker id")
	ErrAttendeeTicketNotFound = errors.New("attendee ticket not found")
	ErrTicketAlreadyUsed      = errors.New("ticket already used")
	ErrTicketNotPaid          = errors.New("ticket order is not paid")
	ErrCheckInNotApplied      = errors.New("ticket check-in was not applied")
)

const (
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_FAILED_CHECK_IN            = "failed check in ticket"
	MESSAGE_FAILED_INVALID_TICKET_CODE = "invalid ticket code"
	MESSAGE_SUCCESS_CHECK_IN           = "success check in ticket"
)

type CheckInRequest struct {
	TicketCode string `json:"ticket_code" form:"ticket_code" binding:"required,max=255"`
}

type CheckInResponse struct {
	ID           string     `json:"id"`
	AttendeeName string     `json:"attendee_name"`
	IsUsed       bool       `json:"is_used"`
	UsedAt       *time.Time `json:"used_at"`
	CheckedBy    *string    `json:"checked_by"`
}
