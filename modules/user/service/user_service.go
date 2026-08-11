package service

import (
	"context"
	"errors"

	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/user/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/user/repository"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"gorm.io/gorm"
)

type UserService interface {
	GetAll(ctx context.Context, filter dto.UserFilter) (dto.UserPaginationResponse, error)
	GetByID(ctx context.Context, id string) (dto.UserResponse, error)
	Update(ctx context.Context, id string, req dto.UserUpdateRequest) (dto.UserResponse, error)
	Delete(ctx context.Context, id string) error
}

type userService struct {
	userRepository repository.UserRepository
	db             *gorm.DB
}

func NewUserService(userRepo repository.UserRepository, db *gorm.DB) UserService {
	return &userService{
		userRepository: userRepo,
		db:             db,
	}
}

func toResponse(u entities.User) dto.UserResponse {
	return dto.UserResponse{
		ID:         u.ID.String(),
		Name:       u.Name,
		Email:      u.Email,
		TelpNumber: u.TelpNumber,
		Role:       u.Role,
		IsVerified: u.IsVerified,
	}
}

func (s *userService) GetAll(ctx context.Context, filter dto.UserFilter) (dto.UserPaginationResponse, error) {
	if filter.Page <= 0 {
		filter.Page = constants.ENUM_PAGINATION_PAGE
	}
	if filter.PerPage <= 0 {
		filter.PerPage = constants.ENUM_PAGINATION_PER_PAGE
	}

	offset := (filter.Page - 1) * filter.PerPage
	users, total, err := s.userRepository.GetAll(ctx, s.db, filter.Search, filter.Role, filter.PerPage, offset)
	if err != nil {
		return dto.UserPaginationResponse{}, err
	}

	data := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		data = append(data, toResponse(u))
	}

	maxPage := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return dto.UserPaginationResponse{
		Data: data,
		Meta: dto.PaginationMeta{
			Page:    filter.Page,
			PerPage: filter.PerPage,
			MaxPage: maxPage,
			Total:   total,
		},
	}, nil
}

func (s *userService) GetByID(ctx context.Context, id string) (dto.UserResponse, error) {
	user, err := s.userRepository.GetUserById(ctx, s.db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.UserResponse{}, dto.ErrUserNotFound
		}
		return dto.UserResponse{}, err
	}

	return toResponse(user), nil
}

func (s *userService) Update(ctx context.Context, id string, req dto.UserUpdateRequest) (dto.UserResponse, error) {
	user, err := s.userRepository.GetUserById(ctx, s.db, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.UserResponse{}, dto.ErrUserNotFound
		}
		return dto.UserResponse{}, err
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.TelpNumber != nil {
		user.TelpNumber = req.TelpNumber
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Role != nil {
		if *req.Role != constants.ENUM_ROLE_ADMIN && *req.Role != constants.ENUM_ROLE_USER {
			return dto.UserResponse{}, dto.ErrInvalidRole
		}
		user.Role = *req.Role
	}

	updated, err := s.userRepository.Update(ctx, s.db, user)
	if err != nil {
		return dto.UserResponse{}, err
	}

	return toResponse(updated), nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	if _, err := s.userRepository.GetUserById(ctx, s.db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.ErrUserNotFound
		}
		return err
	}

	return s.userRepository.Delete(ctx, s.db, id)
}