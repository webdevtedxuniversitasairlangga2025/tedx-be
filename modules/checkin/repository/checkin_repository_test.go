package repository

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/webdevtedxuniversitasairlangga/database"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func integrationDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		t.Fatalf("create uuid-ossp extension: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	return db
}

func createAttendeeTicket(t *testing.T, db *gorm.DB, ticketCode string) (entities.AttendeeTicket, []uuid.UUID) {
	t.Helper()

	buyer := entities.User{
		ID:       uuid.New(),
		Name:     "Buyer",
		Email:    uuid.NewString() + "@example.test",
		Password: "password",
		Role:     "user",
	}
	checkerOne := entities.User{
		ID:       uuid.New(),
		Name:     "Checker One",
		Email:    uuid.NewString() + "@example.test",
		Password: "password",
		Role:     "admin",
	}
	checkerTwo := entities.User{
		ID:       uuid.New(),
		Name:     "Checker Two",
		Email:    uuid.NewString() + "@example.test",
		Password: "password",
		Role:     "admin",
	}
	ticket := entities.Ticket{
		ID:          uuid.New(),
		Name:        "Integration Ticket",
		Description: "Ticket used by check-in repository tests",
		IsActive:    true,
	}
	tier := entities.TicketTier{
		ID:       uuid.New(),
		TicketID: ticket.ID,
		Tier:     "regular",
		Price:    decimal.NewFromInt(100000),
		Quota:    1,
		IsActive: true,
	}
	order := entities.Order{
		ID:           uuid.New(),
		UserID:       buyer.ID,
		TicketTierID: tier.ID,
		OrderNumber:  uuid.NewString(),
		Quantity:     1,
		UnitPrice:    decimal.NewFromInt(100000),
		TotalAmount:  decimal.NewFromInt(100000),
		Status:       "paid",
		ExpiredAt:    time.Now().Add(time.Hour),
	}
	attendeeTicket := entities.AttendeeTicket{
		ID:            uuid.New(),
		OrderID:       order.ID,
		TicketCode:    ticketCode,
		AttendeeName:  "Attendee",
		AttendeeEmail: uuid.NewString() + "@example.test",
		AudienceType:  "umum",
	}

	for _, entity := range []any{&buyer, &checkerOne, &checkerTwo, &ticket, &tier, &order, &attendeeTicket} {
		if err := db.Create(entity).Error; err != nil {
			t.Fatalf("create test fixture: %v", err)
		}
	}

	t.Cleanup(func() {
		db.Where("id = ?", attendeeTicket.ID).Delete(&entities.AttendeeTicket{})
		db.Where("id = ?", order.ID).Delete(&entities.Order{})
		db.Where("id = ?", tier.ID).Delete(&entities.TicketTier{})
		db.Where("id = ?", ticket.ID).Delete(&entities.Ticket{})
		db.Where("id IN ?", []uuid.UUID{buyer.ID, checkerOne.ID, checkerTwo.ID}).Delete(&entities.User{})
	})

	return attendeeTicket, []uuid.UUID{checkerOne.ID, checkerTwo.ID}
}

func TestCheckInRepositoryMarkAsUsed(t *testing.T) {
	db := integrationDatabase(t)
	repo := NewCheckInRepository(db)
	fixture, checkers := createAttendeeTicket(t, db, uuid.NewString())

	ticket, updated, err := repo.MarkAsUsed(context.Background(), nil, fixture.TicketCode, checkers[0])
	if err != nil {
		t.Fatalf("MarkAsUsed() error = %v", err)
	}
	if !updated {
		t.Fatal("MarkAsUsed() updated = false, want true")
	}
	if !ticket.IsUsed || ticket.UsedAt == nil || ticket.CheckedBy == nil || *ticket.CheckedBy != checkers[0] {
		t.Fatalf("MarkAsUsed() ticket = %#v, want used ticket checked by %s", ticket, checkers[0])
	}

	_, updated, err = repo.MarkAsUsed(context.Background(), nil, fixture.TicketCode, checkers[1])
	if err != nil {
		t.Fatalf("second MarkAsUsed() error = %v", err)
	}
	if updated {
		t.Fatal("second MarkAsUsed() updated = true, want false")
	}

	stored, err := repo.FindByTicketCode(context.Background(), nil, fixture.TicketCode)
	if err != nil {
		t.Fatalf("FindByTicketCode() error = %v", err)
	}
	if stored.CheckedBy == nil || *stored.CheckedBy != checkers[0] {
		t.Fatalf("stored ticket checker = %v, want %s", stored.CheckedBy, checkers[0])
	}
}

func TestCheckInRepositoryConcurrentMarkAsUsed(t *testing.T) {
	db := integrationDatabase(t)
	repo := NewCheckInRepository(db)
	fixture, checkers := createAttendeeTicket(t, db, uuid.NewString())

	type result struct {
		ticket  entities.AttendeeTicket
		updated bool
		err     error
	}
	results := make(chan result, len(checkers))
	start := make(chan struct{})
	var group sync.WaitGroup

	for _, checkerID := range checkers {
		group.Add(1)
		go func(checkerID uuid.UUID) {
			defer group.Done()
			<-start
			ticket, updated, err := repo.MarkAsUsed(context.Background(), nil, fixture.TicketCode, checkerID)
			results <- result{ticket: ticket, updated: updated, err: err}
		}(checkerID)
	}

	close(start)
	group.Wait()
	close(results)

	successes := 0
	var successfulTicket entities.AttendeeTicket
	for result := range results {
		if result.err != nil {
			t.Fatalf("MarkAsUsed() error = %v", result.err)
		}
		if result.updated {
			successes++
			successfulTicket = result.ticket
		}
	}
	if successes != 1 {
		t.Fatalf("successful concurrent check-ins = %d, want 1", successes)
	}

	stored, err := repo.FindByTicketCode(context.Background(), nil, fixture.TicketCode)
	if err != nil {
		t.Fatalf("FindByTicketCode() error = %v", err)
	}
	if stored.CheckedBy == nil || successfulTicket.CheckedBy == nil || *stored.CheckedBy != *successfulTicket.CheckedBy {
		t.Fatalf("stored ticket checker = %v, want %v", stored.CheckedBy, successfulTicket.CheckedBy)
	}
}
