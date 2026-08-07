package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/repository"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
)

type MerchandiseService interface {
	GetAll(ctx context.Context, filter dto.MerchandiseFilter) (dto.MerchandisePaginationResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (dto.MerchandiseResponse, error)
	Create(ctx context.Context, req dto.MerchandiseRequest) error
	Update(ctx context.Context, id uuid.UUID, req dto.MerchandiseRequest) error
	Delete(ctx context.Context, id uuid.UUID) error

	AddImage(ctx context.Context, merchId uuid.UUID, req dto.MerchImageRequest) error
	DeleteImage(ctx context.Context, imageId uuid.UUID) error
}

type merchandiseServiceImpl struct {
	repo repository.MerchandiseRepository
}

func NewMerchandiseService(i *do.Injector) (MerchandiseService, error) {
	repo := do.MustInvoke[repository.MerchandiseRepository](i)
	return &merchandiseServiceImpl{
		repo: repo,
	}, nil
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

func (s *merchandiseServiceImpl) GetAll(ctx context.Context, filter dto.MerchandiseFilter) (dto.MerchandisePaginationResponse, error) {
	if filter.Page <= 0 {
		filter.Page = constants.ENUM_PAGINATION_PAGE
	}
	if filter.Limit <= 0 {
		filter.Limit = constants.ENUM_PAGINATION_PER_PAGE
	}

	merchandises, total, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return dto.MerchandisePaginationResponse{}, err
	}

	data := make([]dto.MerchandiseResponse, 0, len(merchandises))
	for _, m := range merchandises {
		data = append(data, toResponse(m))
	}

	maxPage := int((total + int64(filter.Limit) - 1) / int64(filter.Limit))

	return dto.MerchandisePaginationResponse{
		Data: data,
		Meta: dto.PaginationMeta{
			Page:    filter.Page,
			PerPage: filter.Limit,
			MaxPage: maxPage,
			Total:   total,
		},
	}, nil
}

func (s *merchandiseServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (dto.MerchandiseResponse, error) {
	merch, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.MerchandiseResponse{}, err
	}

	return toResponse(*merch), nil
}

func (s *merchandiseServiceImpl) Create(ctx context.Context, req dto.MerchandiseRequest) error {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	merch := &entities.Merchandise{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		IsActive:    isActive,
	}

	return s.repo.Create(ctx, merch)
}

func (s *merchandiseServiceImpl) Update(ctx context.Context, id uuid.UUID, req dto.MerchandiseRequest) error {

	merch, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	isActive := merch.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	merch.Name = req.Name
	merch.Description = req.Description
	merch.Price = req.Price
	merch.Category = req.Category
	merch.IsActive = isActive

	return s.repo.Update(ctx, merch)
}

func (s *merchandiseServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {

	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *merchandiseServiceImpl) AddImage(ctx context.Context, merchId uuid.UUID, req dto.MerchImageRequest) error {

	if _, err := s.repo.FindByID(ctx, merchId); err != nil {
		return err
	}

	image := &entities.MerchImage{
		MerchandiseID: merchId,
		ImageURL:      req.ImageURL,
	}

	return s.repo.AddImage(ctx, image)
}

func (s *merchandiseServiceImpl) DeleteImage(ctx context.Context, imageId uuid.UUID) error {

	return s.repo.DeleteImage(ctx, imageId)
}
