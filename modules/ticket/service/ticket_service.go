package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/ticket/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/ticket/repository"
)

var maxPrice = decimal.RequireFromString("99999999.99")

type TicketService interface {
	GetAll(ctx context.Context, filter dto.TicketFilter) ([]dto.TicketResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (dto.TicketResponse, error)
	Create(ctx context.Context, req dto.TicketCreateRequest) (dto.TicketResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.TicketUpdateRequest) (dto.TicketResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error

	CreateTier(ctx context.Context, ticketId uuid.UUID, req dto.TicketTierCreateRequest) (dto.TicketTierResponse, error)
	UpdateTier(ctx context.Context, ticketId, tierId uuid.UUID, req dto.TicketTierUpdateRequest) (dto.TicketTierResponse, error)
	DeleteTier(ctx context.Context, ticketId, tierId uuid.UUID) error
}

type ticketService struct {
	repo repository.TicketRepository
}

func NewTicketService(repo repository.TicketRepository) TicketService {
	return &ticketService{
		repo: repo,
	}
}

func parsePrice(raw string) (decimal.Decimal, error) {
	price, err := decimal.NewFromString(raw)
	if err != nil {
		return decimal.Decimal{}, dto.ErrInvalidPrice
	}

	if price.IsNegative() || price.GreaterThan(maxPrice) {
		return decimal.Decimal{}, dto.ErrPriceOutOfRange
	}

	return price, nil
}

func toTierResponse(t entities.TicketTier) dto.TicketTierResponse {
	return dto.TicketTierResponse{
		ID:          t.ID.String(),
		TicketID:    t.TicketID.String(),
		Tier:        t.Tier,
		Price:       t.Price.StringFixed(2),
		Quota:       t.Quota,
		QuotaFilled: t.QuotaFilled,
		QuotaLeft:   t.Quota - t.QuotaFilled,
		SaleStart:   t.SaleStart,
		SaleEnd:     t.SaleEnd,
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func toTicketResponse(t entities.Ticket) dto.TicketResponse {
	tiers := make([]dto.TicketTierResponse, 0, len(t.TicketTiers))
	for _, tier := range t.TicketTiers {
		tiers = append(tiers, toTierResponse(tier))
	}

	return dto.TicketResponse{
		ID:          t.ID.String(),
		Name:        t.Name,
		Description: t.Description,
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		Tiers:       tiers,
	}
}

func (s *ticketService) GetAll(ctx context.Context, filter dto.TicketFilter) ([]dto.TicketResponse, error) {
	if filter.IsActive == nil {
		activeOnly := true
		filter.IsActive = &activeOnly
	}

	tickets, err := s.repo.FindAll(ctx, filter.IsActive)
	if err != nil {
		return nil, err
	}

	data := make([]dto.TicketResponse, 0, len(tickets))
	for _, t := range tickets {
		data = append(data, toTicketResponse(t))
	}

	return data, nil
}

func (s *ticketService) GetByID(ctx context.Context, id uuid.UUID) (dto.TicketResponse, error) {
	ticket, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.TicketResponse{}, dto.ErrTicketNotFound
	}

	return toTicketResponse(*ticket), nil
}

func (s *ticketService) Create(ctx context.Context, req dto.TicketCreateRequest) (dto.TicketResponse, error) {
	ticket := &entities.Ticket{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
	}

	created, err := s.repo.Create(ctx, ticket)
	if err != nil {
		return dto.TicketResponse{}, err
	}

	return toTicketResponse(*created), nil
}

func (s *ticketService) Update(ctx context.Context, id uuid.UUID, req dto.TicketUpdateRequest) (dto.TicketResponse, error) {
	ticket, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.TicketResponse{}, dto.ErrTicketNotFound
	}

	if req.Name != nil {
		ticket.Name = *req.Name
	}
	if req.Description != nil {
		ticket.Description = *req.Description
	}
	if req.IsActive != nil {
		ticket.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, ticket); err != nil {
		return dto.TicketResponse{}, err
	}

	return toTicketResponse(*ticket), nil
}

func (s *ticketService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return dto.ErrTicketNotFound
	}
	return s.repo.Delete(ctx, id)
}

func (s *ticketService) CreateTier(ctx context.Context, ticketId uuid.UUID, req dto.TicketTierCreateRequest) (dto.TicketTierResponse, error) {
	if _, err := s.repo.FindByID(ctx, ticketId); err != nil {
		return dto.TicketTierResponse{}, dto.ErrTicketNotFound
	}

	price, err := parsePrice(req.Price)
	if err != nil {
		return dto.TicketTierResponse{}, err
	}

	if req.SaleStart != nil && req.SaleEnd != nil && !req.SaleEnd.After(*req.SaleStart) {
		return dto.TicketTierResponse{}, dto.ErrInvalidSaleWindow
	}

	tier := &entities.TicketTier{
		TicketID:  ticketId,
		Tier:      req.Tier,
		Price:     price,
		Quota:     req.Quota,
		SaleStart: req.SaleStart,
		SaleEnd:   req.SaleEnd,
		IsActive:  true,
	}

	if err := s.repo.CreateTier(ctx, tier); err != nil {
		return dto.TicketTierResponse{}, err
	}

	return toTierResponse(*tier), nil
}

func (s *ticketService) UpdateTier(ctx context.Context, ticketId, tierId uuid.UUID, req dto.TicketTierUpdateRequest) (dto.TicketTierResponse, error) {
	tier, err := s.repo.FindTierByID(ctx, tierId)
	if err != nil || tier.TicketID != ticketId {
		return dto.TicketTierResponse{}, dto.ErrTicketTierNotFound
	}

	if req.Tier != nil {
		tier.Tier = *req.Tier
	}
	if req.Price != nil {
		price, err := parsePrice(*req.Price)
		if err != nil {
			return dto.TicketTierResponse{}, err
		}
		tier.Price = price
	}
	if req.Quota != nil {
		if *req.Quota < tier.QuotaFilled {
			return dto.TicketTierResponse{}, dto.ErrQuotaBelowFilled
		}
		tier.Quota = *req.Quota
	}
	if req.SaleStart != nil {
		tier.SaleStart = req.SaleStart
	}
	if req.SaleEnd != nil {
		tier.SaleEnd = req.SaleEnd
	}
	if tier.SaleStart != nil && tier.SaleEnd != nil && !tier.SaleEnd.After(*tier.SaleStart) {
		return dto.TicketTierResponse{}, dto.ErrInvalidSaleWindow
	}
	if req.IsActive != nil {
		tier.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateTier(ctx, tier); err != nil {
		return dto.TicketTierResponse{}, err
	}

	return toTierResponse(*tier), nil
}

func (s *ticketService) DeleteTier(ctx context.Context, ticketId, tierId uuid.UUID) error {
	affected, err := s.repo.DeleteTier(ctx, ticketId, tierId)
	if err != nil {
		return err
	}
	if affected == 0 {
		return dto.ErrTicketTierNotFound
	}
	return nil
}
