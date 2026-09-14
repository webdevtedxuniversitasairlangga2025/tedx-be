package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"gorm.io/gorm"
)

type CheckInRepository interface {
	MarkAsUsed(ctx context.Context, tx *gorm.DB, ticketCode string, checkedBy uuid.UUID) (entities.AttendeeTicket, bool, error)
	FindByTicketCode(ctx context.Context, tx *gorm.DB, ticketCode string) (entities.AttendeeTicket, error)
}

type checkInRepository struct {
	db *gorm.DB
}

func NewCheckInRepository(db *gorm.DB) CheckInRepository {
	return &checkInRepository{db: db}
}

func (r *checkInRepository) MarkAsUsed(ctx context.Context, tx *gorm.DB, ticketCode string, checkedBy uuid.UUID) (entities.AttendeeTicket, bool, error) {
	if tx == nil {
		tx = r.db
	}

	var ticket entities.AttendeeTicket
	result := tx.WithContext(ctx).Raw(`
		UPDATE attendee_tickets
		SET is_used = true, used_at = CURRENT_TIMESTAMP, checked_by = ?
		WHERE ticket_code = ? AND is_used = false
		RETURNING id, order_id, checked_by, ticket_code, attendee_name, attendee_email,
			attendee_phone, audience_type, institution, is_sent, sent_at, is_used, used_at, created_at
	`, checkedBy, ticketCode).Scan(&ticket)
	if result.Error != nil {
		return entities.AttendeeTicket{}, false, result.Error
	}

	return ticket, result.RowsAffected == 1, nil
}

func (r *checkInRepository) FindByTicketCode(ctx context.Context, tx *gorm.DB, ticketCode string) (entities.AttendeeTicket, error) {
	if tx == nil {
		tx = r.db
	}

	var ticket entities.AttendeeTicket
	if err := tx.WithContext(ctx).Where("ticket_code = ?", ticketCode).Take(&ticket).Error; err != nil {
		return entities.AttendeeTicket{}, err
	}

	return ticket, nil
}
