package service

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/webdevtedxuniversitasairlangga/modules/referral/dto"
)

func TestAllocateDiscount(t *testing.T) {
	tests := []struct {
		name          string
		attendees     []dto.ReferralAttendeeInput
		priorClaims   map[string]bool
		remaining     int
		discount      decimal.Decimal
		unitPrice     decimal.Decimal
		wantPositions []int
	}{
		{
			name: "allocates top down until quota exhausted",
			attendees: []dto.ReferralAttendeeInput{
				{Email: "first@example.com", UnitPrice: decimal.NewFromInt(100)},
				{Email: "second@example.com", UnitPrice: decimal.NewFromInt(100)},
			},
			remaining:     1,
			discount:      decimal.NewFromInt(10),
			unitPrice:     decimal.NewFromInt(100),
			wantPositions: []int{0},
		},
		{
			name: "skips prior and duplicate normalized emails",
			attendees: []dto.ReferralAttendeeInput{
				{Email: "USED@example.com", UnitPrice: decimal.NewFromInt(100)},
				{Email: " New@example.com ", UnitPrice: decimal.NewFromInt(100)},
				{Email: "new@example.com", UnitPrice: decimal.NewFromInt(100)},
				{Email: "last@example.com", UnitPrice: decimal.NewFromInt(100)},
			},
			priorClaims:   map[string]bool{"used@example.com": true},
			remaining:     2,
			discount:      decimal.NewFromInt(10),
			unitPrice:     decimal.NewFromInt(100),
			wantPositions: []int{1, 3},
		},
		{
			name: "skips tier when discount is equal to or greater than price",
			attendees: []dto.ReferralAttendeeInput{
				{Email: "equal@example.com", UnitPrice: decimal.NewFromInt(10)},
				{Email: "lower@example.com", UnitPrice: decimal.NewFromInt(9)},
				{Email: "eligible@example.com", UnitPrice: decimal.NewFromInt(11)},
			},
			remaining:     3,
			discount:      decimal.NewFromInt(10),
			wantPositions: []int{2},
		},
		{
			name:          "no discount when quota exhausted",
			attendees:     []dto.ReferralAttendeeInput{{Email: "a@example.com", UnitPrice: decimal.NewFromInt(100)}},
			discount:      decimal.NewFromInt(10),
			wantPositions: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AllocateDiscount(tt.attendees, tt.priorClaims, tt.remaining, tt.discount)
			if len(got) != len(tt.wantPositions) {
				t.Fatalf("got %d allocations, want %d", len(got), len(tt.wantPositions))
			}
			for i, position := range tt.wantPositions {
				if got[i].AttendeePosition != position {
					t.Errorf("allocation %d position = %d, want %d", i, got[i].AttendeePosition, position)
				}
			}
		})
	}
}

func TestNormalizeReferralValues(t *testing.T) {
	if got := NormalizeReferralCode("  friend10 "); got != "FRIEND10" {
		t.Fatalf("normalizeReferralCode() = %q, want FRIEND10", got)
	}
	if got := NormalizeAttendeeEmail("  USER@Example.COM "); got != "user@example.com" {
		t.Fatalf("normalizeAttendeeEmail() = %q, want user@example.com", got)
	}
}
