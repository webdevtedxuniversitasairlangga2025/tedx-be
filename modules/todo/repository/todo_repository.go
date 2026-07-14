package repository

import (
	"context"

	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	TodoRepository interface {
		Create(ctx context.Context, tx *gorm.DB, todo entities.Todo) (entities.Todo, error)
		GetAllByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID, limit, offset int) ([]entities.Todo, int64, error)
		GetByID(ctx context.Context, tx *gorm.DB, id, userID uuid.UUID) (entities.Todo, error)
		Update(ctx context.Context, tx *gorm.DB, todo entities.Todo) (entities.Todo, error)
		Delete(ctx context.Context, tx *gorm.DB, id, userID uuid.UUID) error
	}

	todoRepository struct {
		db *gorm.DB
	}
)

func NewTodoRepository(db *gorm.DB) TodoRepository {
	return &todoRepository{
		db: db,
	}
}

func (r *todoRepository) Create(ctx context.Context, tx *gorm.DB, todo entities.Todo) (entities.Todo, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&todo).Error; err != nil {
		return entities.Todo{}, err
	}

	return todo, nil
}

func (r *todoRepository) GetAllByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID, limit, offset int) ([]entities.Todo, int64, error) {
	if tx == nil {
		tx = r.db
	}

	query := tx.WithContext(ctx).Model(&entities.Todo{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var todos []entities.Todo
	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&todos).Error; err != nil {
		return nil, 0, err
	}

	return todos, total, nil
}

func (r *todoRepository) GetByID(ctx context.Context, tx *gorm.DB, id, userID uuid.UUID) (entities.Todo, error) {
	if tx == nil {
		tx = r.db
	}

	var todo entities.Todo
	if err := tx.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Take(&todo).Error; err != nil {
		return entities.Todo{}, err
	}

	return todo, nil
}

func (r *todoRepository) Update(ctx context.Context, tx *gorm.DB, todo entities.Todo) (entities.Todo, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Save(&todo).Error; err != nil {
		return entities.Todo{}, err
	}

	return todo, nil
}

func (r *todoRepository) Delete(ctx context.Context, tx *gorm.DB, id, userID uuid.UUID) error {
	if tx == nil {
		tx = r.db
	}

	return tx.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&entities.Todo{}).Error
}
