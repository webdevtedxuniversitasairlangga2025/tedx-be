package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"gorm.io/gorm"
)

type MerchandiseRepository interface {
	FindAll(ctx context.Context, category string, isActive *bool, limit, offset int) ([]entities.Merchandise, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Merchandise, error)
	Create(ctx context.Context, merch *entities.Merchandise) (*entities.Merchandise, error)
	Update(ctx context.Context, merch *entities.Merchandise) error
	Delete(ctx context.Context, id uuid.UUID) error

	AddImage(ctx context.Context, image *entities.MerchImage) error
	DeleteImage(ctx context.Context, merchId, imageId uuid.UUID) (int64, error)
}

type merchandiseRepositoryImpl struct {
	db *gorm.DB
}

func NewMerchandiseRepository(db *gorm.DB) MerchandiseRepository {
	return &merchandiseRepositoryImpl{
		db: db,
	}
}

func (r *merchandiseRepositoryImpl) FindAll(ctx context.Context, category string, isActive *bool, limit, offset int) ([]entities.Merchandise, int64, error) {
	var merchandises []entities.Merchandise
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.Merchandise{}).Preload("MerchImages")

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&merchandises).Error; err != nil {
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

func (r *merchandiseRepositoryImpl) Create(ctx context.Context, merch *entities.Merchandise) (*entities.Merchandise, error) {
	if err := r.db.WithContext(ctx).Create(merch).Error; err != nil {
		return nil, err
	}
	return merch, nil
}

func (r *merchandiseRepositoryImpl) Update(ctx context.Context, merch *entities.Merchandise) error {
	return r.db.WithContext(ctx).Save(merch).Error
}

func (r *merchandiseRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(t *gorm.DB) error {
		if err := t.Where("merchandise_id = ?", id).Delete(&entities.MerchImage{}).Error; err != nil {
			return err
		}

		return t.Where("id = ?", id).Delete(&entities.Merchandise{}).Error
	})
}

func (r *merchandiseRepositoryImpl) AddImage(ctx context.Context, image *entities.MerchImage) error {
	return r.db.WithContext(ctx).Create(image).Error
}

func (r *merchandiseRepositoryImpl) DeleteImage(ctx context.Context, merchId, imageId uuid.UUID) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("id = ? AND merchandise_id = ?", imageId, merchId).
		Delete(&entities.MerchImage{})

	return result.RowsAffected, result.Error
}
