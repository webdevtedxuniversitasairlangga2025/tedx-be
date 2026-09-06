package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	FindAll(ctx context.Context) ([]entities.Category, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Category, error)
	FindByName(ctx context.Context, name string) (*entities.Category, error)
	Create(ctx context.Context, category *entities.Category) (*entities.Category, error)
	Update(ctx context.Context, category *entities.Category) error
	Delete(ctx context.Context, id uuid.UUID) error

	CountMerchandise(ctx context.Context, id uuid.UUID) (int64, error)
}

type categoryRepositoryImpl struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepositoryImpl{
		db: db,
	}
}

func (r *categoryRepositoryImpl) FindAll(ctx context.Context) ([]entities.Category, error) {
	var categories []entities.Category

	if err := r.db.WithContext(ctx).Order("name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entities.Category, error) {
	var category entities.Category

	if err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepositoryImpl) FindByName(ctx context.Context, name string) (*entities.Category, error) {
	var category entities.Category

	if err := r.db.WithContext(ctx).First(&category, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepositoryImpl) Create(ctx context.Context, category *entities.Category) (*entities.Category, error) {
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		return nil, err
	}
	return category, nil
}

func (r *categoryRepositoryImpl) Update(ctx context.Context, category *entities.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entities.Category{}).Error
}

func (r *categoryRepositoryImpl) CountMerchandise(ctx context.Context, id uuid.UUID) (int64, error) {
	var count int64

	if err := r.db.WithContext(ctx).Model(&entities.Merchandise{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
