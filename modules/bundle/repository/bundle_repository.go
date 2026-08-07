package repository

import (
	"context"

	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	BundleRepository interface {
		Create(ctx context.Context, tx *gorm.DB, bundle entities.Bundle) (entities.Bundle, error)
		GetAll(ctx context.Context, tx *gorm.DB, isActive *bool, limit, offset int) ([]entities.Bundle, int64, error)
		GetByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entities.Bundle, error)
		GetByIDWithImages(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entities.Bundle, error)
		Update(ctx context.Context, tx *gorm.DB, bundle entities.Bundle) (entities.Bundle, error)
		Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID) error

		CreateImage(ctx context.Context, tx *gorm.DB, image entities.BundleImage) (entities.BundleImage, error)
		DeleteImage(ctx context.Context, tx *gorm.DB, imageID, bundleID uuid.UUID) (int64, error)
	}

	bundleRepository struct {
		db *gorm.DB
	}
)

func NewBundleRepository(db *gorm.DB) BundleRepository {
	return &bundleRepository{
		db: db,
	}
}

func (r *bundleRepository) Create(ctx context.Context, tx *gorm.DB, bundle entities.Bundle) (entities.Bundle, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&bundle).Error; err != nil {
		return entities.Bundle{}, err
	}

	return bundle, nil
}

func (r *bundleRepository) GetAll(ctx context.Context, tx *gorm.DB, isActive *bool, limit, offset int) ([]entities.Bundle, int64, error) {
	if tx == nil {
		tx = r.db
	}

	query := tx.WithContext(ctx).Model(&entities.Bundle{})
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var bundles []entities.Bundle
	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&bundles).Error; err != nil {
		return nil, 0, err
	}

	return bundles, total, nil
}

func (r *bundleRepository) GetByID(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entities.Bundle, error) {
	if tx == nil {
		tx = r.db
	}

	var bundle entities.Bundle
	if err := tx.WithContext(ctx).Where("id = ?", id).Take(&bundle).Error; err != nil {
		return entities.Bundle{}, err
	}

	return bundle, nil
}

func (r *bundleRepository) GetByIDWithImages(ctx context.Context, tx *gorm.DB, id uuid.UUID) (entities.Bundle, error) {
	if tx == nil {
		tx = r.db
	}

	var bundle entities.Bundle
	if err := tx.WithContext(ctx).Preload("BundleImages").Where("id = ?", id).Take(&bundle).Error; err != nil {
		return entities.Bundle{}, err
	}

	return bundle, nil
}

func (r *bundleRepository) Update(ctx context.Context, tx *gorm.DB, bundle entities.Bundle) (entities.Bundle, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Save(&bundle).Error; err != nil {
		return entities.Bundle{}, err
	}

	return bundle, nil
}

func (r *bundleRepository) Delete(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	if tx == nil {
		tx = r.db
	}

	return tx.WithContext(ctx).Transaction(func(t *gorm.DB) error {
		if err := t.Where("bundle_id = ?", id).Delete(&entities.BundleImage{}).Error; err != nil {
			return err
		}

		return t.Where("id = ?", id).Delete(&entities.Bundle{}).Error
	})
}

func (r *bundleRepository) CreateImage(ctx context.Context, tx *gorm.DB, image entities.BundleImage) (entities.BundleImage, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Omit("Bundle").Create(&image).Error; err != nil {
		return entities.BundleImage{}, err
	}

	return image, nil
}

func (r *bundleRepository) DeleteImage(ctx context.Context, tx *gorm.DB, imageID, bundleID uuid.UUID) (int64, error) {
	if tx == nil {
		tx = r.db
	}

	result := tx.WithContext(ctx).
		Where("id = ? AND bundle_id = ?", imageID, bundleID).
		Delete(&entities.BundleImage{})

	return result.RowsAffected, result.Error
}
