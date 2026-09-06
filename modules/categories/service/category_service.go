package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/categories/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/categories/repository"
	"gorm.io/gorm"
)

type CategoryService interface {
	GetAll(ctx context.Context) ([]dto.CategoryResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (dto.CategoryResponse, error)
	Create(ctx context.Context, req dto.CategoryCreateRequest) (dto.CategoryResponse, error)
	Update(ctx context.Context, id uuid.UUID, req dto.CategoryUpdateRequest) (dto.CategoryResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{
		repo: repo,
	}
}

func toResponse(c entities.Category) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:        c.ID.String(),
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func (s *categoryService) GetAll(ctx context.Context) ([]dto.CategoryResponse, error) {
	categories, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	data := make([]dto.CategoryResponse, 0, len(categories))
	for _, c := range categories {
		data = append(data, toResponse(c))
	}

	return data, nil
}

func (s *categoryService) GetByID(ctx context.Context, id uuid.UUID) (dto.CategoryResponse, error) {
	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.CategoryResponse{}, dto.ErrCategoryNotFound
	}

	return toResponse(*category), nil
}

func (s *categoryService) Create(ctx context.Context, req dto.CategoryCreateRequest) (dto.CategoryResponse, error) {
	name := strings.TrimSpace(req.Name)

	if _, err := s.repo.FindByName(ctx, name); err == nil {
		return dto.CategoryResponse{}, dto.ErrCategoryExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.CategoryResponse{}, err
	}

	category := &entities.Category{
		Name: name,
	}

	created, err := s.repo.Create(ctx, category)
	if err != nil {
		return dto.CategoryResponse{}, err
	}

	return toResponse(*created), nil
}

func (s *categoryService) Update(ctx context.Context, id uuid.UUID, req dto.CategoryUpdateRequest) (dto.CategoryResponse, error) {
	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.CategoryResponse{}, dto.ErrCategoryNotFound
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if existing, err := s.repo.FindByName(ctx, name); err == nil && existing.ID != category.ID {
			return dto.CategoryResponse{}, dto.ErrCategoryExists
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.CategoryResponse{}, err
		}
		category.Name = name
	}

	if err := s.repo.Update(ctx, category); err != nil {
		return dto.CategoryResponse{}, err
	}

	return toResponse(*category), nil
}

func (s *categoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return dto.ErrCategoryNotFound
	}

	used, err := s.repo.CountMerchandise(ctx, id)
	if err != nil {
		return err
	}
	if used > 0 {
		return dto.ErrCategoryInUse
	}

	return s.repo.Delete(ctx, id)
}
