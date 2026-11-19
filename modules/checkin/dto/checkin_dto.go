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
	ErrInvalidTierID          = errors.New("invalid tier id")
	ErrInvalidStatusFilter    = errors.New("invalid status filter, use all, used, or unused")
)

const (
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_FAILED_CHECK_IN            = "failed check in ticket"
	MESSAGE_FAILED_INVALID_TICKET_CODE = "invalid ticket code"
	MESSAGE_SUCCESS_CHECK_IN           = "success check in ticket"

	MESSAGE_FAILED_GET_TICKET_DETAIL  = "failed get ticket detail"
	MESSAGE_SUCCESS_GET_TICKET_DETAIL = "success get ticket detail"

	MESSAGE_FAILED_GET_ATTENDANCE_STATS  = "failed get attendance stats"
	MESSAGE_SUCCESS_GET_ATTENDANCE_STATS = "success get attendance stats"

	MESSAGE_FAILED_GET_ATTENDEE_LIST  = "failed get attendee tickets"
	MESSAGE_SUCCESS_GET_ATTENDEE_LIST = "success get attendee tickets"
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

// TicketDetailResponse adalah ringkasan tiket peserta untuk halaman registrasi
// (lookup read-only, tidak mengubah status tiket).
type TicketDetailResponse struct {
	ID            string     `json:"id"`
	TicketCode    string     `json:"ticket_code"`
	AttendeeName  string     `json:"attendee_name"`
	AttendeeEmail string     `json:"attendee_email"`
	AttendeePhone *string    `json:"attendee_phone"`
	TicketName    string     `json:"ticket_name"`
	TierName      string     `json:"tier_name"`
	OrderStatus   string     `json:"order_status"`
	IsUsed        bool       `json:"is_used"`
	UsedAt        *time.Time `json:"used_at"`
	CheckedBy     *string    `json:"checked_by"`
}

// AttendanceCount adalah angka kehadiran. Total hanya menghitung tiket dari
// pesanan berstatus paid (yang memang bisa check-in).
type AttendanceCount struct {
	Total        int64 `json:"total"`
	CheckedIn    int64 `json:"checked_in"`
	NotCheckedIn int64 `json:"not_checked_in"`
}

type TierAttendance struct {
	TierID   string `json:"tier_id"`
	TierName string `json:"tier_name"`
	AttendanceCount
}

type TicketAttendance struct {
	TicketID   string `json:"ticket_id"`
	TicketName string `json:"ticket_name"`
	AttendanceCount
	Tiers []TierAttendance `json:"tiers"`
}

type AttendanceStatsResponse struct {
	AttendanceCount
	Tickets     []TicketAttendance `json:"tickets"`
	GeneratedAt time.Time          `json:"generated_at"`
}

// AttendeeListRequest adalah query GET /checkins/tickets.
// Status: all (default) | used | unused. TierID kosong = semua tier.
type AttendeeListRequest struct {
	Search  string `form:"search"`
	TierID  string `form:"tier_id"`
	Status  string `form:"status"`
	Page    int    `form:"page"`
	PerPage int    `form:"per_page"`
}

type AttendeeListItem struct {
	ID            string     `json:"id"`
	TicketCode    string     `json:"ticket_code"`
	AttendeeName  string     `json:"attendee_name"`
	AttendeeEmail string     `json:"attendee_email"`
	AttendeePhone *string    `json:"attendee_phone"`
	TicketName    string     `json:"ticket_name"`
	TierID        string     `json:"tier_id"`
	TierName      string     `json:"tier_name"`
	IsUsed        bool       `json:"is_used"`
	UsedAt        *time.Time `json:"used_at"`
}

type AttendeePaginationMeta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	MaxPage int   `json:"max_page"`
	Total   int64 `json:"total"`
}

type AttendeeListResponse struct {
	Data []AttendeeListItem     `json:"data"`
	Meta AttendeePaginationMeta `json:"meta"`
}
