package dto

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	ErrReferralNotFound  = errors.New("referral code not found")
	ErrInvalidReferral   = errors.New("invalid referral code configuration")
	ErrInvalidReferralID = errors.New("invalid referral code id")
)

const (
	ReasonNotFound        = "not_found"
	ReasonInactive        = "inactive"
	ReasonAlreadyUsed     = "already_used"
	ReasonQuotaExhausted  = "quota_exhausted"
	ReasonPriceIneligible = "price_ineligible"
)

type CreateReferralCodeRequest struct {
	Code              string `json:"code" binding:"required,min=1,max=100"`
	DiscountPerTicket string `json:"discount_per_ticket" binding:"required"`
	Quota             int    `json:"quota" binding:"required,gt=0"`
}

type SetReferralActiveRequest struct {
	IsActive bool `json:"is_active"`
}

type ReferralAttendeeInput struct {
	Email     string
	UnitPrice decimal.Decimal
}

type ReferralCodeResponse struct {
	ID                string `json:"id"`
	Code              string `json:"code"`
	DiscountPerTicket string `json:"discount_per_ticket"`
	Quota             int    `json:"quota"`
	QuotaHeld         int    `json:"quota_held"`
	QuotaUsed         int    `json:"quota_used"`
	QuotaRemaining    int    `json:"quota_remaining"`
	IsActive          bool   `json:"is_active"`
}

type ReferralPreviewRequest struct {
	TicketTierID string   `json:"ticket_tier_id" binding:"required"`
	Code         string   `json:"code" binding:"required,min=1,max=100"`
	Emails       []string `json:"attendee_emails" binding:"required,min=1,max=5,dive,email"`
}

type ReferralPreviewResponse struct {
	Applicable              bool                 `json:"applicable"`
	Reason                  string               `json:"reason,omitempty"`
	ReasonCode              string               `json:"reason_code,omitempty"`
	ReferralCode            string               `json:"referral_code,omitempty"`
	DiscountPerTicket       string               `json:"discount_per_ticket"`
	DiscountedAttendeeCount int                  `json:"discounted_attendee_count"`
	TotalDiscount           string               `json:"total_discount"`
	Allocations             []ReferralAllocation `json:"allocations"`
}

type ReferralAllocation struct {
	AttendeePosition int
	DiscountAmount   decimal.Decimal
}
