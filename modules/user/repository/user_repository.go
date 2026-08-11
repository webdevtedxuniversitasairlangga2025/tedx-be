package repository

import (
	"context"
	"errors"

	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"gorm.io/gorm"
)

type (
	UserRepository interface {
		Register(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error)
		GetAll(ctx context.Context, tx *gorm.DB, search, role string, limit, offset int) ([]entities.User, int64, error)
		GetUserById(ctx context.Context, tx *gorm.DB, userId string) (entities.User, error)
		GetUserByEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, error)
		CheckEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, bool, error)
		Update(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error)
		Delete(ctx context.Context, tx *gorm.DB, userId string) error
	}

	userRepository struct {
		db *gorm.DB
	}
)

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GetAll(ctx context.Context, tx *gorm.DB, search, role string, limit, offset int) ([]entities.User, int64, error) {
	if tx == nil {
		tx = r.db
	}

	query := tx.WithContext(ctx).Model(&entities.User{})
	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", pattern, pattern)
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []entities.User
	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) Register(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetUserById(ctx context.Context, tx *gorm.DB, userId string) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("id = ?", userId).Take(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("email = ?", email).Take(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) CheckEmail(ctx context.Context, tx *gorm.DB, email string) (entities.User, bool, error) {
	if tx == nil {
		tx = r.db
	}

	var user entities.User
	if err := tx.WithContext(ctx).Where("email = ?", email).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entities.User{}, false, nil
		}
		return entities.User{}, false, err
	}

	return user, true, nil
}

func (r *userRepository) Update(ctx context.Context, tx *gorm.DB, user entities.User) (entities.User, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Updates(&user).Error; err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (r *userRepository) Delete(ctx context.Context, tx *gorm.DB, userId string) error {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Delete(&entities.User{}, "id = ?", userId).Error; err != nil {
		return err
	}

	return nil
}