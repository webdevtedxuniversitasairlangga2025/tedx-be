package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/checkin/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/checkin/repository"
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

	detailTicket entities.AttendeeTicket
	detailErr    error
	detailCalls  int

	statsRows  []repository.AttendanceRow
	statsErr   error
	statsCalls int

	listRows   []repository.AttendeeRow
	listTotal  int64
	listErr    error
	listFilter repository.AttendeeFilter
	listLimit  int
	listOffset int
	listCalls  int
}

func (r *fakeCheckInRepository) MarkAsUsed(_ context.Context, _ *gorm.DB, _ string, _ uuid.UUID) (entities.AttendeeTicket, bool, error) {
	r.markCalls++
	return r.markTicket, r.marked, r.markErr
}

func (r *fakeCheckInRepository) FindByTicketCode(_ context.Context, _ *gorm.DB, _ string) (entities.AttendeeTicket, error) {
	r.findCalls++
	return r.findTicket, r.findErr
}

func (r *fakeCheckInRepository) FindDetailByTicketCode(_ context.Context, _ *gorm.DB, _ string) (entities.AttendeeTicket, error) {
	r.detailCalls++
	return r.detailTicket, r.detailErr
}

func (r *fakeCheckInRepository) CountAttendanceByTier(_ context.Context, _ *gorm.DB) ([]repository.AttendanceRow, error) {
	r.statsCalls++
	return r.statsRows, r.statsErr
}

func (r *fakeCheckInRepository) ListAttendees(_ context.Context, _ *gorm.DB, filter repository.AttendeeFilter, limit, offset int) ([]repository.AttendeeRow, int64, error) {
	r.listCalls++
	r.listFilter, r.listLimit, r.listOffset = filter, limit, offset
	return r.listRows, r.listTotal, r.listErr
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

func TestLookup(t *testing.T) {
	checkerID := uuid.New()
	usedAt := time.Now().UTC()
	phone := "081234567890"
	ticket := entities.AttendeeTicket{
		ID:            uuid.New(),
		TicketCode:    "K7P2QX",
		AttendeeName:  "Attendee",
		AttendeeEmail: "attendee@example.test",
		AttendeePhone: &phone,
		IsUsed:        true,
		UsedAt:        &usedAt,
		CheckedBy:     &checkerID,
		Order: entities.Order{
			Status: "paid",
			TicketTier: entities.TicketTier{
				Tier:   "Early Bird",
				Ticket: entities.Ticket{Name: "Main Event"},
			},
		},
	}
	databaseErr := errors.New("database error")

	tests := []struct {
		name        string
		ticketCode  string
		repository  *fakeCheckInRepository
		wantErr     error
		wantDetails int
	}{
		{
			name:        "returns ticket details",
			ticketCode:  "K7P2QX",
			repository:  &fakeCheckInRepository{detailTicket: ticket},
			wantDetails: 1,
		},
		{
			name:       "rejects an empty ticket code",
			ticketCode: " ",
			repository: &fakeCheckInRepository{},
			wantErr:    dto.ErrInvalidTicketCode,
		},
		{
			name:       "rejects a ticket code containing spaces",
			ticketCode: "K7P 2QX",
			repository: &fakeCheckInRepository{},
			wantErr:    dto.ErrInvalidTicketCode,
		},
		{
			name:        "reports a missing ticket",
			ticketCode:  "MISSING",
			repository:  &fakeCheckInRepository{detailErr: gorm.ErrRecordNotFound},
			wantErr:     dto.ErrAttendeeTicketNotFound,
			wantDetails: 1,
		},
		{
			name:        "returns a lookup error",
			ticketCode:  "ERROR1",
			repository:  &fakeCheckInRepository{detailErr: databaseErr},
			wantErr:     databaseErr,
			wantDetails: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewCheckInService(test.repository, nil)

			response, err := service.Lookup(context.Background(), test.ticketCode)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Lookup() error = %v, want %v", err, test.wantErr)
			}
			if test.repository.detailCalls != test.wantDetails {
				t.Fatalf("detail calls = %d, want %d", test.repository.detailCalls, test.wantDetails)
			}
			// Lookup harus read-only: tidak boleh menyentuh MarkAsUsed / FindByTicketCode.
			if test.repository.markCalls != 0 || test.repository.findCalls != 0 {
				t.Fatalf("Lookup() must not mutate: mark:%d find:%d", test.repository.markCalls, test.repository.findCalls)
			}
			if test.wantErr != nil {
				return
			}
			if response.ID != ticket.ID.String() || response.TicketCode != "K7P2QX" ||
				response.AttendeeName != "Attendee" || response.AttendeeEmail != "attendee@example.test" ||
				response.AttendeePhone == nil || *response.AttendeePhone != phone ||
				response.TicketName != "Main Event" || response.TierName != "Early Bird" ||
				response.OrderStatus != "paid" || !response.IsUsed || response.UsedAt == nil ||
				response.CheckedBy == nil || *response.CheckedBy != checkerID.String() {
				t.Fatalf("Lookup() response = %#v, want full ticket detail", response)
			}
		})
	}
}

func TestStats(t *testing.T) {
	t.Run("aggregates per tier, per ticket and overall", func(t *testing.T) {
		repo := &fakeCheckInRepository{statsRows: []repository.AttendanceRow{
			{TicketID: "t1", TicketName: "Main Event", TierID: "a", TierName: "Early Bird", Total: 10, CheckedIn: 4},
			{TicketID: "t1", TicketName: "Main Event", TierID: "b", TierName: "Regular", Total: 5, CheckedIn: 5},
			{TicketID: "t2", TicketName: "Workshop", TierID: "c", TierName: "Regular", Total: 3, CheckedIn: 0},
		}}
		svc := &checkInService{repository: repo}

		got, err := svc.Stats(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Total != 18 || got.CheckedIn != 9 || got.NotCheckedIn != 9 {
			t.Errorf("overall = %+v", got.AttendanceCount)
		}
		if len(got.Tickets) != 2 {
			t.Fatalf("tickets = %d, want 2", len(got.Tickets))
		}
		main := got.Tickets[0]
		if main.Total != 15 || main.CheckedIn != 9 || main.NotCheckedIn != 6 || len(main.Tiers) != 2 {
			t.Errorf("main ticket = %+v", main)
		}
		if main.Tiers[0].NotCheckedIn != 6 || main.Tiers[1].NotCheckedIn != 0 {
			t.Errorf("tiers = %+v", main.Tiers)
		}
	})

	t.Run("empty data returns empty slices and zero counts", func(t *testing.T) {
		svc := &checkInService{repository: &fakeCheckInRepository{}}
		got, err := svc.Stats(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Tickets == nil || len(got.Tickets) != 0 || got.Total != 0 {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		boom := errors.New("db down")
		svc := &checkInService{repository: &fakeCheckInRepository{statsErr: boom}}
		if _, err := svc.Stats(context.Background()); !errors.Is(err, boom) {
			t.Errorf("err = %v, want %v", err, boom)
		}
	})
}

func TestListAttendees(t *testing.T) {
	tierID := uuid.New()

	t.Run("passes filters and maps rows with pagination meta", func(t *testing.T) {
		repo := &fakeCheckInRepository{
			listRows:  []repository.AttendeeRow{{ID: "1", TicketCode: "K7P2QX", AttendeeName: "Ega", TierName: "Early Bird", IsUsed: true}},
			listTotal: 45,
		}
		svc := &checkInService{repository: repo}

		got, err := svc.ListAttendees(context.Background(), dto.AttendeeListRequest{
			Search: "  ega ", TierID: tierID.String(), Status: "USED", Page: 3, PerPage: 20,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.listFilter.Search != "ega" || repo.listFilter.TierID == nil || *repo.listFilter.TierID != tierID {
			t.Errorf("filter = %+v", repo.listFilter)
		}
		if repo.listFilter.Used == nil || !*repo.listFilter.Used {
			t.Errorf("used filter = %v", repo.listFilter.Used)
		}
		if repo.listLimit != 20 || repo.listOffset != 40 {
			t.Errorf("limit/offset = %d/%d", repo.listLimit, repo.listOffset)
		}
		if got.Meta.Page != 3 || got.Meta.MaxPage != 3 || got.Meta.Total != 45 {
			t.Errorf("meta = %+v", got.Meta)
		}
		if len(got.Data) != 1 || got.Data[0].TicketCode != "K7P2QX" || !got.Data[0].IsUsed {
			t.Errorf("data = %+v", got.Data)
		}
	})

	t.Run("defaults, unused filter and per-page cap", func(t *testing.T) {
		repo := &fakeCheckInRepository{}
		svc := &checkInService{repository: repo}

		got, err := svc.ListAttendees(context.Background(), dto.AttendeeListRequest{Status: "unused", PerPage: 1000})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.listFilter.Used == nil || *repo.listFilter.Used {
			t.Errorf("used filter = %v, want false", repo.listFilter.Used)
		}
		if repo.listLimit != 100 || repo.listOffset != 0 {
			t.Errorf("limit/offset = %d/%d", repo.listLimit, repo.listOffset)
		}
		if got.Data == nil || len(got.Data) != 0 || got.Meta.MaxPage != 1 || got.Meta.Page != 1 {
			t.Errorf("got = %+v", got)
		}
	})

	t.Run("rejects bad tier id and status without hitting the repository", func(t *testing.T) {
		repo := &fakeCheckInRepository{}
		svc := &checkInService{repository: repo}

		if _, err := svc.ListAttendees(context.Background(), dto.AttendeeListRequest{TierID: "nope"}); !errors.Is(err, dto.ErrInvalidTierID) {
			t.Errorf("tier err = %v", err)
		}
		if _, err := svc.ListAttendees(context.Background(), dto.AttendeeListRequest{Status: "maybe"}); !errors.Is(err, dto.ErrInvalidStatusFilter) {
			t.Errorf("status err = %v", err)
		}
		if repo.listCalls != 0 {
			t.Errorf("repository called %d times", repo.listCalls)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		boom := errors.New("db down")
		svc := &checkInService{repository: &fakeCheckInRepository{listErr: boom}}
		if _, err := svc.ListAttendees(context.Background(), dto.AttendeeListRequest{}); !errors.Is(err, boom) {
			t.Errorf("err = %v", err)
		}
	})
}
