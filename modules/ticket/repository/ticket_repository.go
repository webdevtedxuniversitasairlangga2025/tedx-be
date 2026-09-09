package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"gorm.io/gorm"
)

type TicketRepository interface {
	FindAll(ctx context.Context, isActive *bool) ([]entities.Ticket, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Ticket, error)
	Create(ctx context.Context, ticket *entities.Ticket) (*entities.Ticket, error)
	Update(ctx context.Context, ticket *entities.Ticket) error
	Delete(ctx context.Context, id uuid.UUID) error

	FindTierByID(ctx context.Context, id uuid.UUID) (*entities.TicketTier, error)
	CreateTier(ctx context.Context, tier *entities.TicketTier) error
	UpdateTier(ctx context.Context, tier *entities.TicketTier) error
	DeleteTier(ctx context.Context, ticketId, tierId uuid.UUID) (int64, error)
}

type ticketRepositoryImpl struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepositoryImpl{
		db: db,
	}
}

func (r *ticketRepositoryImpl) FindAll(ctx context.Context, isActive *bool) ([]entities.Ticket, error) {
	var tickets []entities.Ticket

	query := r.db.WithContext(ctx).Model(&entities.Ticket{}).Preload("TicketTiers")

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Find(&tickets).Error; err != nil {
		return nil, err
	}

	return tickets, nil
}

func (r *ticketRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entities.Ticket, error) {
	var ticket entities.Ticket

	if err := r.db.WithContext(ctx).Preload("TicketTiers").First(&ticket, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepositoryImpl) Create(ctx context.Context, ticket *entities.Ticket) (*entities.Ticket, error) {
	if err := r.db.WithContext(ctx).Create(ticket).Error; err != nil {
		return nil, err
	}
	return ticket, nil
}

func (r *ticketRepositoryImpl) Update(ctx context.Context, ticket *entities.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *ticketRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(t *gorm.DB) error {
		if err := t.Where("ticket_id = ?", id).Delete(&entities.TicketTier{}).Error; err != nil {
			return err
		}

		return t.Where("id = ?", id).Delete(&entities.Ticket{}).Error
	})
}

func (r *ticketRepositoryImpl) FindTierByID(ctx context.Context, id uuid.UUID) (*entities.TicketTier, error) {
	var tier entities.TicketTier

	if err := r.db.WithContext(ctx).First(&tier, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tier, nil
}

func (r *ticketRepositoryImpl) CreateTier(ctx context.Context, tier *entities.TicketTier) error {
	return r.db.WithContext(ctx).Create(tier).Error
}

func (r *ticketRepositoryImpl) UpdateTier(ctx context.Context, tier *entities.TicketTier) error {
	return r.db.WithContext(ctx).Save(tier).Error
}

func (r *ticketRepositoryImpl) DeleteTier(ctx context.Context, ticketId, tierId uuid.UUID) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND ticket_id = ?", tierId, ticketId).
		Delete(&entities.TicketTier{})

	return result.RowsAffected, result.Error
}
