package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"gorm.io/gorm"
)

type CheckInRepository interface {
	MarkAsUsed(ctx context.Context, tx *gorm.DB, ticketCode string, checkedBy uuid.UUID) (entities.AttendeeTicket, bool, error)
	FindByTicketCode(ctx context.Context, tx *gorm.DB, ticketCode string) (entities.AttendeeTicket, error)
	FindDetailByTicketCode(ctx context.Context, tx *gorm.DB, ticketCode string) (entities.AttendeeTicket, error)
	CountAttendanceByTier(ctx context.Context, tx *gorm.DB) ([]AttendanceRow, error)
	ListAttendees(ctx context.Context, tx *gorm.DB, filter AttendeeFilter, limit, offset int) ([]AttendeeRow, int64, error)
}

// AttendeeFilter: Used nil = semua status; TierID nil = semua tier.
type AttendeeFilter struct {
	Search string
	TierID *uuid.UUID
	Used   *bool
}

type AttendeeRow struct {
	ID            string     `gorm:"column:id"`
	TicketCode    string     `gorm:"column:ticket_code"`
	AttendeeName  string     `gorm:"column:attendee_name"`
	AttendeeEmail string     `gorm:"column:attendee_email"`
	AttendeePhone *string    `gorm:"column:attendee_phone"`
	TicketName    string     `gorm:"column:ticket_name"`
	TierID        string     `gorm:"column:tier_id"`
	TierName      string     `gorm:"column:tier_name"`
	IsUsed        bool       `gorm:"column:is_used"`
	UsedAt        *time.Time `gorm:"column:used_at"`
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// AttendanceRow adalah hasil agregasi satu tier: jumlah tiket lunas dan berapa
// yang sudah check-in.
type AttendanceRow struct {
	TicketID   string `gorm:"column:ticket_id"`
	TicketName string `gorm:"column:ticket_name"`
	TierID     string `gorm:"column:tier_id"`
	TierName   string `gorm:"column:tier_name"`
	Total      int64  `gorm:"column:total"`
	CheckedIn  int64  `gorm:"column:checked_in"`
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

// FindDetailByTicketCode membaca tiket beserta pesanan, tier, dan nama tiketnya.
// Read-only: tidak mengubah apa pun (berbeda dengan MarkAsUsed).
func (r *checkInRepository) FindDetailByTicketCode(ctx context.Context, tx *gorm.DB, ticketCode string) (entities.AttendeeTicket, error) {
	if tx == nil {
		tx = r.db
	}

	var ticket entities.AttendeeTicket
	if err := tx.WithContext(ctx).Preload("Order.TicketTier.Ticket").Where("ticket_code = ?", ticketCode).Take(&ticket).Error; err != nil {
		return entities.AttendeeTicket{}, err
	}

	return ticket, nil
}

// CountAttendanceByTier menghitung tiket peserta per tier dalam satu query.
// Hanya pesanan paid dan belum dihapus yang dihitung. Tier/ticket yang sudah
// di-soft-delete tetap ikut (join mentah tidak memfilter deleted_at) supaya
// angka kehadiran tidak berkurang diam-diam.
func (r *checkInRepository) CountAttendanceByTier(ctx context.Context, tx *gorm.DB) ([]AttendanceRow, error) {
	if tx == nil {
		tx = r.db
	}

	var rows []AttendanceRow
	err := tx.WithContext(ctx).Raw(`
		SELECT tickets.id::text AS ticket_id, tickets.name AS ticket_name,
			ticket_tiers.id::text AS tier_id, ticket_tiers.tier AS tier_name,
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE attendee_tickets.is_used) AS checked_in
		FROM attendee_tickets
		JOIN orders ON orders.id = attendee_tickets.order_id
		JOIN ticket_tiers ON ticket_tiers.id = orders.ticket_tier_id
		JOIN tickets ON tickets.id = ticket_tiers.ticket_id
		WHERE orders.status = ? AND orders.deleted_at IS NULL
		GROUP BY tickets.id, tickets.name, ticket_tiers.id, ticket_tiers.tier
		ORDER BY tickets.name, ticket_tiers.tier
	`, constants.ENUM_ORDER_STATUS_PAID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ListAttendees mengembalikan satu halaman tiket peserta dari pesanan paid
// (sama dengan definisi di statistik), terbaru dulu. Pencarian dan filter
// dijalankan di database sehingga hasil di halaman manapun ikut tercari.
func (r *checkInRepository) ListAttendees(ctx context.Context, tx *gorm.DB, filter AttendeeFilter, limit, offset int) ([]AttendeeRow, int64, error) {
	if tx == nil {
		tx = r.db
	}

	query := tx.WithContext(ctx).
		Table("attendee_tickets").
		Joins("JOIN orders ON orders.id = attendee_tickets.order_id").
		Joins("JOIN ticket_tiers ON ticket_tiers.id = orders.ticket_tier_id").
		Joins("JOIN tickets ON tickets.id = ticket_tiers.ticket_id").
		Where("orders.status = ? AND orders.deleted_at IS NULL", constants.ENUM_ORDER_STATUS_PAID)

	if search := strings.TrimSpace(filter.Search); search != "" {
		pattern := "%" + likeEscaper.Replace(search) + "%"
		query = query.Where(`(attendee_tickets.ticket_code ILIKE ? ESCAPE '\' OR attendee_tickets.attendee_name ILIKE ? ESCAPE '\'
			OR attendee_tickets.attendee_email ILIKE ? ESCAPE '\' OR attendee_tickets.attendee_phone ILIKE ? ESCAPE '\')`,
			pattern, pattern, pattern, pattern)
	}
	if filter.TierID != nil {
		query = query.Where("ticket_tiers.id = ?", *filter.TierID)
	}
	if filter.Used != nil {
		query = query.Where("attendee_tickets.is_used = ?", *filter.Used)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []AttendeeRow
	err := query.Session(&gorm.Session{}).
		Select(`attendee_tickets.id::text AS id, attendee_tickets.ticket_code, attendee_tickets.attendee_name,
			attendee_tickets.attendee_email, attendee_tickets.attendee_phone, tickets.name AS ticket_name,
			ticket_tiers.id::text AS tier_id, ticket_tiers.tier AS tier_name,
			attendee_tickets.is_used, attendee_tickets.used_at`).
		Order("attendee_tickets.created_at DESC, attendee_tickets.ticket_code").
		Limit(limit).Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}
