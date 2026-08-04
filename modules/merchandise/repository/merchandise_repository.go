package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/dto"
	"gorm.io/gorm"
)

type MerchandiseRepository interface {
	FindAll(ctx context.Context, filter dto.MerchandiseFilter) ([]entities.Merchandise, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Merchandise, error)
	Create(ctx context.Context, merch *entities.Merchandise) error
	Update(ctx context.Context, merch *entities.Merchandise) error
	Delete(ctx context.Context, id uuid.UUID) error

	AddImage(ctx context.Context, image *entities.MerchImage) error
	DeleteImage(ctx context.Context, imageId uuid.UUID) error
}

type merchandiseRepositoryImpl struct {
	db *gorm.DB
}

func NewMerchandiseRepository(i *do.Injector) (MerchandiseRepository, error) {
	db := do.MustInvokeNamed[*gorm.DB](i, "db") // Tambahkan 'Named' dan parameter "db"
	return &merchandiseRepositoryImpl{
		db: db,
	}, nil
}

func (r *merchandiseRepositoryImpl) FindAll(ctx context.Context, filter dto.MerchandiseFilter) ([]entities.Merchandise, int64, error) {
	var merchandises []entities.Merchandise
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.Merchandise{}).Preload("MerchImages")

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	if err := query.Offset(offset).Limit(filter.Limit).Find(&merchandises).Error; err != nil {
		return nil, 0, err
	}

	return merchandises, total, nil
}

func (r *merchandiseRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entities.Merchandise, error) {
	var merch entities.Merchandise

	if err := r.db.WithContext(ctx).Preload("MerchImages").First(&merch, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &merch, nil
}

func (r *merchandiseRepositoryImpl) Create(ctx context.Context, merch *entities.Merchandise) error {
	return r.db.WithContext(ctx).Create(merch).Error
}

func (r *merchandiseRepositoryImpl) Update(ctx context.Context, merch *entities.Merchandise) error {
	return r.db.WithContext(ctx).Save(merch).Error
}

func (r *merchandiseRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {

	return r.db.WithContext(ctx).Delete(&entities.Merchandise{}, "id = ?", id).Error
}

func (r *merchandiseRepositoryImpl) AddImage(ctx context.Context, image *entities.MerchImage) error {
	return r.db.WithContext(ctx).Create(image).Error
}

func (r *merchandiseRepositoryImpl) DeleteImage(ctx context.Context, imageId uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.MerchImage{}, "id = ?", imageId).Error
}
