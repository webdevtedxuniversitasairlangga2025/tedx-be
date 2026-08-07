package service

import (
	"context"

	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/bundle/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/bundle/repository"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var maxPrice = decimal.RequireFromString("99999999.99")

type BundleService interface {
	Create(ctx context.Context, req dto.BundleCreateRequest) (dto.BundleResponse, error)
	GetAll(ctx context.Context, req dto.BundleQueryRequest) (dto.BundlePaginationResponse, error)
	GetByID(ctx context.Context, id string) (dto.BundleDetailResponse, error)
	Update(ctx context.Context, id string, req dto.BundleUpdateRequest) (dto.BundleResponse, error)
	Delete(ctx context.Context, id string) error
	AddImage(ctx context.Context, bundleID string, req dto.BundleImageCreateRequest) (dto.BundleImageResponse, error)
	DeleteImage(ctx context.Context, bundleID, imageID string) error
}

type bundleService struct {
	bundleRepository repository.BundleRepository
	db               *gorm.DB
}

func NewBundleService(bundleRepo repository.BundleRepository, db *gorm.DB) BundleService {
	return &bundleService{
		bundleRepository: bundleRepo,
		db:               db,
	}
}

func toResponse(b entities.Bundle) dto.BundleResponse {
	return dto.BundleResponse{
		ID:          b.ID.String(),
		Name:        b.Name,
		Description: b.Description,
		Price:       b.Price.StringFixed(2),
		IsActive:    b.IsActive,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

func toImageResponse(i entities.BundleImage) dto.BundleImageResponse {
	return dto.BundleImageResponse{
		ID:       i.ID.String(),
		ImageURL: i.ImageURL,
	}
}

func toDetailResponse(b entities.Bundle) dto.BundleDetailResponse {
	images := make([]dto.BundleImageResponse, 0, len(b.BundleImages))
	for _, img := range b.BundleImages {
		images = append(images, toImageResponse(img))
	}

	return dto.BundleDetailResponse{
		BundleResponse: toResponse(b),
		Images:         images,
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

func (s *bundleService) Create(ctx context.Context, req dto.BundleCreateRequest) (dto.BundleResponse, error) {
	price, err := parsePrice(req.Price)
	if err != nil {
		return dto.BundleResponse{}, err
	}

	bundle := entities.Bundle{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Price:       price,
		IsActive:    true,
	}

	created, err := s.bundleRepository.Create(ctx, s.db, bundle)
	if err != nil {
		return dto.BundleResponse{}, err
	}

	return toResponse(created), nil
}

func (s *bundleService) GetAll(ctx context.Context, req dto.BundleQueryRequest) (dto.BundlePaginationResponse, error) {
	if req.Page <= 0 {
		req.Page = constants.ENUM_PAGINATION_PAGE
	}
	if req.PerPage <= 0 {
		req.PerPage = constants.ENUM_PAGINATION_PER_PAGE
	}
	offset := (req.Page - 1) * req.PerPage

	isActive := req.IsActive
	if isActive == nil {
		activeOnly := true
		isActive = &activeOnly
	}

	bundles, total, err := s.bundleRepository.GetAll(ctx, s.db, isActive, req.PerPage, offset)
	if err != nil {
		return dto.BundlePaginationResponse{}, err
	}

	data := make([]dto.BundleResponse, 0, len(bundles))
	for _, b := range bundles {
		data = append(data, toResponse(b))
	}

	maxPage := int((total + int64(req.PerPage) - 1) / int64(req.PerPage))

	return dto.BundlePaginationResponse{
		Data: data,
		Meta: dto.PaginationMeta{
			Page:    req.Page,
			PerPage: req.PerPage,
			MaxPage: maxPage,
			Total:   total,
		},
	}, nil
}

func (s *bundleService) GetByID(ctx context.Context, id string) (dto.BundleDetailResponse, error) {
	bid, err := uuid.Parse(id)
	if err != nil {
		return dto.BundleDetailResponse{}, dto.ErrBundleNotFound
	}

	bundle, err := s.bundleRepository.GetByIDWithImages(ctx, s.db, bid)
	if err != nil {
		return dto.BundleDetailResponse{}, dto.ErrBundleNotFound
	}

	return toDetailResponse(bundle), nil
}

func (s *bundleService) Update(ctx context.Context, id string, req dto.BundleUpdateRequest) (dto.BundleResponse, error) {
	bid, err := uuid.Parse(id)
	if err != nil {
		return dto.BundleResponse{}, dto.ErrBundleNotFound
	}

	bundle, err := s.bundleRepository.GetByID(ctx, s.db, bid)
	if err != nil {
		return dto.BundleResponse{}, dto.ErrBundleNotFound
	}

	if req.Name != nil {
		bundle.Name = *req.Name
	}
	if req.Description != nil {
		bundle.Description = *req.Description
	}
	if req.Price != nil {
		price, err := parsePrice(*req.Price)
		if err != nil {
			return dto.BundleResponse{}, err
		}
		bundle.Price = price
	}
	if req.IsActive != nil {
		bundle.IsActive = *req.IsActive
	}

	updated, err := s.bundleRepository.Update(ctx, s.db, bundle)
	if err != nil {
		return dto.BundleResponse{}, err
	}

	return toResponse(updated), nil
}

func (s *bundleService) Delete(ctx context.Context, id string) error {
	bid, err := uuid.Parse(id)
	if err != nil {
		return dto.ErrBundleNotFound
	}

	if _, err := s.bundleRepository.GetByID(ctx, s.db, bid); err != nil {
		return dto.ErrBundleNotFound
	}

	return s.bundleRepository.Delete(ctx, s.db, bid)
}

func (s *bundleService) AddImage(ctx context.Context, bundleID string, req dto.BundleImageCreateRequest) (dto.BundleImageResponse, error) {
	bid, err := uuid.Parse(bundleID)
	if err != nil {
		return dto.BundleImageResponse{}, dto.ErrBundleNotFound
	}

	if _, err := s.bundleRepository.GetByID(ctx, s.db, bid); err != nil {
		return dto.BundleImageResponse{}, dto.ErrBundleNotFound
	}

	image := entities.BundleImage{
		ID:       uuid.New(),
		BundleID: bid,
		ImageURL: req.ImageURL,
	}

	created, err := s.bundleRepository.CreateImage(ctx, s.db, image)
	if err != nil {
		return dto.BundleImageResponse{}, err
	}

	return toImageResponse(created), nil
}

func (s *bundleService) DeleteImage(ctx context.Context, bundleID, imageID string) error {
	bid, err := uuid.Parse(bundleID)
	if err != nil {
		return dto.ErrBundleNotFound
	}

	iid, err := uuid.Parse(imageID)
	if err != nil {
		return dto.ErrBundleImageNotFound
	}

	if _, err := s.bundleRepository.GetByID(ctx, s.db, bid); err != nil {
		return dto.ErrBundleNotFound
	}

	affected, err := s.bundleRepository.DeleteImage(ctx, s.db, iid, bid)
	if err != nil {
		return err
	}

	if affected == 0 {
		return dto.ErrBundleImageNotFound
	}

	return nil
}
