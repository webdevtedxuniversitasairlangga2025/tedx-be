package service

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/webdevtedxuniversitasairlangga/modules/referral/dto"
)

func TestClassifyPreviewFailure(t *testing.T) {
	discount := decimal.NewFromInt(10)
	price := decimal.NewFromInt(100)

	tests := []struct {
		name      string
		emails    []string
		claimed   map[string]bool
		remaining int
		discount  decimal.Decimal
		price     decimal.Decimal
		want      string
	}{
		{
			name:      "all emails already claimed means code already used",
			emails:    []string{"a@example.com", "b@example.com"},
			claimed:   map[string]bool{"a@example.com": true, "b@example.com": true},
			remaining: 5,
			discount:  discount,
			price:     price,
			want:      dto.ReasonAlreadyUsed,
		},
		{
			name:      "no quota left means exhausted",
			emails:    []string{"new@example.com"},
			claimed:   map[string]bool{},
			remaining: 0,
			discount:  discount,
			price:     price,
			want:      dto.ReasonQuotaExhausted,
		},
		{
			name:      "discount not below price means ineligible tier",
			emails:    []string{"new@example.com"},
			claimed:   map[string]bool{},
			remaining: 5,
			discount:  price,
			price:     price,
			want:      dto.ReasonPriceIneligible,
		},
		{
			name:      "partially claimed still usable for another ticket",
			emails:    []string{"used@example.com", "fresh@example.com"},
			claimed:   map[string]bool{"used@example.com": true},
			remaining: 5,
			discount:  discount,
			price:     price,
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attendees := make([]dto.ReferralAttendeeInput, len(tt.emails))
			for i, email := range tt.emails {
				attendees[i] = dto.ReferralAttendeeInput{Email: email, UnitPrice: tt.price}
			}
			got := ClassifyPreviewFailure(attendees, tt.claimed, tt.remaining, tt.discount)
			if got != tt.want {
				t.Fatalf("ClassifyPreviewFailure() = %q, want %q", got, tt.want)
			}
		})
	}
}
