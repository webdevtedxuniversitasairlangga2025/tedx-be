package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/checkin/dto"
	"gorm.io/gorm"
)

type fakeCheckInRepository struct {
	markTicket entities.AttendeeTicket
	marked     bool
	markErr    error
	findTicket entities.AttendeeTicket
	findErr    error
	markCalls  int
	findCalls  int
}

func (r *fakeCheckInRepository) MarkAsUsed(_ context.Context, _ *gorm.DB, _ string, _ uuid.UUID) (entities.AttendeeTicket, bool, error) {
	r.markCalls++
	return r.markTicket, r.marked, r.markErr
}

func (r *fakeCheckInRepository) FindByTicketCode(_ context.Context, _ *gorm.DB, _ string) (entities.AttendeeTicket, error) {
	r.findCalls++
	return r.findTicket, r.findErr
}

func TestCheckIn(t *testing.T) {
	checkerID := uuid.New()
	usedAt := time.Now().UTC()
	ticket := entities.AttendeeTicket{
		ID:           uuid.New(),
		AttendeeName: "Attendee",
		IsUsed:       true,
		UsedAt:       &usedAt,
		CheckedBy:    &checkerID,
	}
	databaseErr := errors.New("database error")

	tests := []struct {
		name          string
		checkerID     string
		request       dto.CheckInRequest
		repository    *fakeCheckInRepository
		wantErr       error
		wantMarkCalls int
		wantFindCalls int
	}{
		{
			name:          "checks in an unused ticket",
			checkerID:     checkerID.String(),
			request:       dto.CheckInRequest{TicketCode: "opaque-ticket-code"},
			repository:    &fakeCheckInRepository{markTicket: ticket, marked: true},
			wantMarkCalls: 1,
		},
		{
			name:       "rejects an invalid checker",
			checkerID:  "invalid",
			request:    dto.CheckInRequest{TicketCode: "opaque-ticket-code"},
			repository: &fakeCheckInRepository{},
			wantErr:    dto.ErrInvalidChecker,
		},
		{
			name:       "rejects an empty ticket code",
			checkerID:  checkerID.String(),
			request:    dto.CheckInRequest{TicketCode: " \t "},
			repository: &fakeCheckInRepository{},
			wantErr:    dto.ErrInvalidTicketCode,
		},
		{
			name:       "rejects a ticket code longer than the entity limit",
			checkerID:  checkerID.String(),
			request:    dto.CheckInRequest{TicketCode: string(make([]rune, 256))},
			repository: &fakeCheckInRepository{},
			wantErr:    dto.ErrInvalidTicketCode,
		},
		{
			name:          "reports a missing ticket",
			checkerID:     checkerID.String(),
			request:       dto.CheckInRequest{TicketCode: "missing"},
			repository:    &fakeCheckInRepository{findErr: gorm.ErrRecordNotFound},
			wantErr:       dto.ErrAttendeeTicketNotFound,
			wantMarkCalls: 1,
			wantFindCalls: 1,
		},
		{
			name:          "reports an already used ticket",
			checkerID:     checkerID.String(),
			request:       dto.CheckInRequest{TicketCode: "used"},
			repository:    &fakeCheckInRepository{findTicket: ticket},
			wantErr:       dto.ErrTicketAlreadyUsed,
			wantMarkCalls: 1,
			wantFindCalls: 1,
		},
		{
			name:          "returns the update error",
			checkerID:     checkerID.String(),
			request:       dto.CheckInRequest{TicketCode: "error"},
			repository:    &fakeCheckInRepository{markErr: databaseErr},
			wantErr:       databaseErr,
			wantMarkCalls: 1,
		},
		{
			name:          "returns a lookup error",
			checkerID:     checkerID.String(),
			request:       dto.CheckInRequest{TicketCode: "error"},
			repository:    &fakeCheckInRepository{findErr: databaseErr},
			wantErr:       databaseErr,
			wantMarkCalls: 1,
			wantFindCalls: 1,
		},
		{
			name:          "reports an unapplied check-in when the ticket remains unused",
			checkerID:     checkerID.String(),
			request:       dto.CheckInRequest{TicketCode: "unapplied"},
			repository:    &fakeCheckInRepository{findTicket: entities.AttendeeTicket{IsUsed: false, Order: entities.Order{Status: "paid"}}},
			wantErr:       dto.ErrCheckInNotApplied,
			wantMarkCalls: 1,
			wantFindCalls: 1,
		},
		{
			name:          "rejects a ticket whose order is not paid",
			checkerID:     checkerID.String(),
			request:       dto.CheckInRequest{TicketCode: "unpaid"},
			repository:    &fakeCheckInRepository{findTicket: entities.AttendeeTicket{Order: entities.Order{Status: "awaiting_approval"}}},
			wantErr:       dto.ErrTicketNotPaid,
			wantMarkCalls: 1,
			wantFindCalls: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewCheckInService(test.repository, nil)

			response, err := service.CheckIn(context.Background(), test.checkerID, test.request)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("CheckIn() error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr == nil && (response.ID != ticket.ID.String() || response.CheckedBy == nil || *response.CheckedBy != checkerID.String()) {
				t.Fatalf("CheckIn() response = %#v, want checked-in ticket", response)
			}
			if test.repository.markCalls != test.wantMarkCalls || test.repository.findCalls != test.wantFindCalls {
				t.Fatalf("repository calls = mark:%d find:%d, want mark:%d find:%d", test.repository.markCalls, test.repository.findCalls, test.wantMarkCalls, test.wantFindCalls)
			}
		})
	}
}
