package repository

import (
	"context"

	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WebhookRepository interface {
	GetOrder(ctx context.Context, tx *gorm.DB, orderID string) (entities.Order, error)
	GetTierWithLock(ctx context.Context, tx *gorm.DB, tierID string) (entities.TicketTier, error)
	UpdateOrder(ctx context.Context, tx *gorm.DB, order entities.Order) error
	UpdateTier(ctx context.Context, tx *gorm.DB, tier entities.TicketTier) error
	CreateAttendeeTickets(ctx context.Context, tx *gorm.DB, tickets []entities.AttendeeTicket) error

	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type webhookRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) WebhookRepository {
	return &webhookRepository{db: db}
}

func (r *webhookRepository) GetOrder(ctx context.Context, tx *gorm.DB, orderID string) (entities.Order, error) {
	if tx == nil {
		tx = r.db
	}
	var order entities.Order
	if err := tx.WithContext(ctx).Where("id = ?", orderID).Take(&order).Error; err != nil {
		return entities.Order{}, err
	}
	return order, nil
}

func (r *webhookRepository) GetTierWithLock(ctx context.Context, tx *gorm.DB, tierID string) (entities.TicketTier, error) {
	if tx == nil {
		tx = r.db
	}
	var tier entities.TicketTier

	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", tierID).Take(&tier).Error; err != nil {
		return entities.TicketTier{}, err
	}
	return tier, nil
}

func (r *webhookRepository) UpdateOrder(ctx context.Context, tx *gorm.DB, order entities.Order) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Save(&order).Error
}

func (r *webhookRepository) UpdateTier(ctx context.Context, tx *gorm.DB, tier entities.TicketTier) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Save(&tier).Error
}

func (r *webhookRepository) CreateAttendeeTickets(ctx context.Context, tx *gorm.DB, tickets []entities.AttendeeTicket) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Create(&tickets).Error
}

func (r *webhookRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
