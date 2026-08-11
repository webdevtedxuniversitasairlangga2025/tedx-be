package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/repository"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
)

var maxPrice = decimal.RequireFromString("99999999.99")

type MerchandiseService interface {
	GetAll(ctx context.Context, filter dto.MerchandiseFilter) ([]dto.MerchandiseResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (dto.MerchandiseResponse, error)
	Create(ctx context.Context, req dto.MerchandiseCreateRequest) (dto.MerchandiseResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.MerchandiseUpdateRequest) (dto.MerchandiseResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error

	AddImage(ctx context.Context, merchId uuid.UUID, req dto.MerchImageRequest) error
	DeleteImage(ctx context.Context, merchId, imageId uuid.UUID) error
}

type merchandiseService struct {
	repo repository.MerchandiseRepository
}

func NewMerchandiseService(repo repository.MerchandiseRepository) MerchandiseService {
	return &merchandiseService{
		repo: repo,
	}
}

func toImageResponse(i entities.MerchImage) dto.MerchImageResponse {
	return dto.MerchImageResponse{
		ID:       i.ID.String(),
		ImageURL: i.ImageURL,
	}
}

func toResponse(m entities.Merchandise) dto.MerchandiseResponse {
	images := make([]dto.MerchImageResponse, 0, len(m.MerchImages))
	for _, img := range m.MerchImages {
		images = append(images, toImageResponse(img))
	}

	return dto.MerchandiseResponse{
		ID:          m.ID.String(),
		Name:        m.Name,
		Description: m.Description,
		Price:       m.Price.StringFixed(2),
		Category:    m.Category,
		IsActive:    m.IsActive,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		Images:      images,
	}
}

func (s *merchandiseService) GetAll(ctx context.Context, filter dto.MerchandiseFilter) ([]dto.MerchandiseResponse, error) {
	if filter.IsActive == nil {
		activeOnly := true
		filter.IsActive = &activeOnly
	}

	merchandises, err := s.repo.FindAll(ctx, filter.Category, filter.IsActive)
	if err != nil {
		return nil, err
	}

	data := make([]dto.MerchandiseResponse, 0, len(merchandises))
	for _, m := range merchandises {
		data = append(data, toResponse(m))
	}

	return data, nil
}

func (s *merchandiseService) GetByID(ctx context.Context, id uuid.UUID) (dto.MerchandiseResponse, error) {
	merch, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.MerchandiseResponse{}, err
	}

	return toResponse(*merch), nil
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

func isValidCategory(category string) bool {
	switch category {
	case constants.ENUM_MERCH_CATEGORY_TSHIRT,
		constants.ENUM_MERCH_CATEGORY_CAP,
		constants.ENUM_MERCH_CATEGORY_STICKER,
		constants.ENUM_MERCH_CATEGORY_OTHER:
		return true
	}
	return false
}

func (s *merchandiseService) Create(ctx context.Context, req dto.MerchandiseCreateRequest) (dto.MerchandiseResponse, error) {
	price, err := parsePrice(req.Price)
	if err != nil {
		return dto.MerchandiseResponse{}, err
	}
	if !isValidCategory(req.Category) {
		return dto.MerchandiseResponse{}, dto.ErrInvalidCategory
	}

	merch := &entities.Merchandise{
		Name:        req.Name,
		Description: req.Description,
		Price:       price,
		Category:    req.Category,
		IsActive:    true,
	}

	created, err := s.repo.Create(ctx, merch)
	if err != nil {
		return dto.MerchandiseResponse{}, err
	}

	return toResponse(*created), nil
}

func (s *merchandiseService) Update(ctx context.Context, id uuid.UUID, req dto.MerchandiseUpdateRequest) (dto.MerchandiseResponse, error) {

	merch, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.MerchandiseResponse{}, err
	}

	if req.Name != nil {
		merch.Name = *req.Name
	}
	if req.Description != nil {
		merch.Description = *req.Description
	}
	if req.Price != nil {
		price, err := parsePrice(*req.Price)
		if err != nil {
			return dto.MerchandiseResponse{}, err
		}
		merch.Price = price
	}
	if req.Category != nil {
		if !isValidCategory(*req.Category) {
			return dto.MerchandiseResponse{}, dto.ErrInvalidCategory
		}
		merch.Category = *req.Category
	}
	if req.IsActive != nil {
		merch.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, merch); err != nil {
		return dto.MerchandiseResponse{}, err
	}

	return toResponse(*merch), nil
}

func (s *merchandiseService) Delete(ctx context.Context, id uuid.UUID) error {

	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *merchandiseService) AddImage(ctx context.Context, merchId uuid.UUID, req dto.MerchImageRequest) error {

	if _, err := s.repo.FindByID(ctx, merchId); err != nil {
		return err
	}

	image := &entities.MerchImage{
		MerchandiseID: merchId,
		ImageURL:      req.ImageURL,
	}

	return s.repo.AddImage(ctx, image)
}

func (s *merchandiseService) DeleteImage(ctx context.Context, merchId, imageId uuid.UUID) error {

	if _, err := s.repo.FindByID(ctx, merchId); err != nil {
		return err
	}

	affected, err := s.repo.DeleteImage(ctx, merchId, imageId)
	if err != nil {
		return err
	}
	if affected == 0 {
		return dto.ErrMerchImageNotFound
	}

	return nil
}
