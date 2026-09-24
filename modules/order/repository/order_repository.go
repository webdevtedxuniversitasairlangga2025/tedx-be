package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository interface {
	Create(ctx context.Context, tx *gorm.DB, order entities.Order) (entities.Order, error)
	GetByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entities.Order, error)
	GetByIDForUpdate(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entities.Order, error)
	GetAllByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID, limit, offset int) ([]entities.Order, int64, error)
	GetAllForExport(ctx context.Context, tx *gorm.DB) ([]entities.Order, error)
	GetAll(ctx context.Context, tx *gorm.DB, status *string, limit, offset int) ([]entities.Order, int64, error)
	Update(ctx context.Context, tx *gorm.DB, order entities.Order) (entities.Order, error)
	SoftDelete(ctx context.Context, tx *gorm.DB, id uuid.UUID) error
	GetTierForUpdate(ctx context.Context, tx *gorm.DB, tierID uuid.UUID) (entities.TicketTier, error)
	UpdateTier(ctx context.Context, tx *gorm.DB, tier entities.TicketTier) error
	CreateAttendeeTickets(ctx context.Context, tx *gorm.DB, tickets []entities.AttendeeTicket) error
	FindExpiredAwaitingApproval(ctx context.Context, tx *gorm.DB) ([]entities.Order, error)
	MarkTicketsSent(ctx context.Context, orderID uuid.UUID) error
	TicketCodeExists(ctx context.Context, tx *gorm.DB, code string) (bool, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) dbOrTx(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *orderRepository) Create(ctx context.Context, tx *gorm.DB, order entities.Order) (entities.Order, error) {
	db := r.dbOrTx(tx)
	if err := db.WithContext(ctx).Create(&order).Error; err != nil {
		return entities.Order{}, err
	}
	return order, nil
}

func (r *orderRepository) GetByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entities.Order, error) {
	db := r.dbOrTx(tx)
	var order entities.Order
	if err := db.WithContext(ctx).Preload("AttendeeTickets").Where("id = ?", id).Take(&order).Error; err != nil {
		return entities.Order{}, err
	}
	return order, nil
}

func (r *orderRepository) GetByIDForUpdate(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entities.Order, error) {
	db := r.dbOrTx(tx)
	var order entities.Order
	if err := db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Preload("AttendeeTickets").Where("id = ?", id).Take(&order).Error; err != nil {
		return entities.Order{}, err
	}
	return order, nil
}

func (r *orderRepository) GetAllByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID, limit, offset int) ([]entities.Order, int64, error) {
	db := r.dbOrTx(tx)
	query := db.WithContext(ctx).Model(&entities.Order{}).Where("user_id = ?", userID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var orders []entities.Order
	if err := query.Preload("AttendeeTickets").Order("created_at desc").Limit(limit).Offset(offset).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *orderRepository) GetAll(ctx context.Context, tx *gorm.DB, status *string, limit, offset int) ([]entities.Order, int64, error) {
	db := r.dbOrTx(tx)
	query := db.WithContext(ctx).Model(&entities.Order{})
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var orders []entities.Order
	if err := query.Preload("AttendeeTickets").Preload("User").Order("created_at desc").Limit(limit).Offset(offset).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// SoftDelete — sembunyikan history dari list (row tetap utk FK attendee).
func (r *orderRepository) SoftDelete(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	return r.dbOrTx(tx).WithContext(ctx).Delete(&entities.Order{}, id).Error
}

func (r *orderRepository) GetAllForExport(ctx context.Context, tx *gorm.DB) ([]entities.Order, error) {
	db := r.dbOrTx(tx)
	var orders []entities.Order
	if err := db.WithContext(ctx).
		Preload("User").
		Preload("ApprovedByUser").
		Preload("TicketTier.Ticket").
		Preload("AttendeeTickets.CheckedByUser").
		Order("created_at asc").
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) Update(ctx context.Context, tx *gorm.DB, order entities.Order) (entities.Order, error) {
	db := r.dbOrTx(tx)
	if err := db.WithContext(ctx).Save(&order).Error; err != nil {
		return entities.Order{}, err
	}
	return order, nil
}

func (r *orderRepository) GetTierForUpdate(ctx context.Context, tx *gorm.DB, tierID uuid.UUID) (entities.TicketTier, error) {
	db := r.dbOrTx(tx)
	var tier entities.TicketTier
	// Unscoped: approve/reject/release tetap temukan tier soft-deleted (ada order FK); Create wajib cek DeletedAt sendiri
	if err := db.WithContext(ctx).Unscoped().Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", tierID).Take(&tier).Error; err != nil {
		return entities.TicketTier{}, err
	}
	return tier, nil
}

func (r *orderRepository) UpdateTier(ctx context.Context, tx *gorm.DB, tier entities.TicketTier) error {
	db := r.dbOrTx(tx)
	return db.WithContext(ctx).Save(&tier).Error
}

func (r *orderRepository) CreateAttendeeTickets(ctx context.Context, tx *gorm.DB, tickets []entities.AttendeeTicket) error {
	db := r.dbOrTx(tx)
	return db.WithContext(ctx).Create(&tickets).Error
}

func (r *orderRepository) FindExpiredAwaitingApproval(ctx context.Context, tx *gorm.DB) ([]entities.Order, error) {
	db := r.dbOrTx(tx)
	var orders []entities.Order
	if err := db.WithContext(ctx).Where("status = ? AND expired_at < ?", "awaiting_approval", time.Now()).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) MarkTicketsSent(ctx context.Context, orderID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entities.AttendeeTicket{}).
		Where("order_id = ?", orderID).
		Updates(map[string]any{"is_sent": true, "sent_at": time.Now()}).Error
}

func (r *orderRepository) TicketCodeExists(ctx context.Context, tx *gorm.DB, code string) (bool, error) {
	db := r.dbOrTx(tx)
	var count int64
	if err := db.WithContext(ctx).Model(&entities.AttendeeTicket{}).
		Where("ticket_code = ?", code).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
