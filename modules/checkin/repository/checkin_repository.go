package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
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
		UPDATE attendee_tickets AS attendee_ticket
		SET is_used = true, used_at = CURRENT_TIMESTAMP, checked_by = ?
		FROM orders
		WHERE attendee_ticket.order_id = orders.id
			AND attendee_ticket.ticket_code = ?
			AND attendee_ticket.is_used = false
			AND orders.status = ?
		RETURNING attendee_ticket.id, attendee_ticket.order_id, attendee_ticket.checked_by,
			attendee_ticket.ticket_code, attendee_ticket.attendee_name, attendee_ticket.attendee_email,
			attendee_ticket.attendee_phone, attendee_ticket.audience_type, attendee_ticket.institution,
			attendee_ticket.is_sent, attendee_ticket.sent_at, attendee_ticket.is_used,
			attendee_ticket.used_at, attendee_ticket.created_at
	`, checkedBy, ticketCode, constants.ENUM_ORDER_STATUS_PAID).Scan(&ticket)
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
	if err := tx.WithContext(ctx).Preload("Order").Where("ticket_code = ?", ticketCode).Take(&ticket).Error; err != nil {
		return entities.AttendeeTicket{}, err
	}

	return ticket, nil
}
