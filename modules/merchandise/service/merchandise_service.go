package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/repository"
)


type MerchandiseService interface {
	GetAll(ctx context.Context, filter dto.MerchandiseFilter) ([]entities.Merchandise, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Merchandise, error)
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

func (s *merchandiseServiceImpl) GetAll(ctx context.Context, filter dto.MerchandiseFilter) ([]entities.Merchandise, int64, error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *merchandiseServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Merchandise, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *merchandiseServiceImpl) Create(ctx context.Context, req dto.MerchandiseRequest) error {
	isActive := true // Default true
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