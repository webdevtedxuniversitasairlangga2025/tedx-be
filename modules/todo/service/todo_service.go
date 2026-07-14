package service

import (
	"context"

	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/todo/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/todo/repository"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TodoService interface {
	Create(ctx context.Context, userID string, req dto.TodoCreateRequest) (dto.TodoResponse, error)
	GetAll(ctx context.Context, userID string, req dto.PaginationRequest) (dto.TodoPaginationResponse, error)
	GetByID(ctx context.Context, userID, id string) (dto.TodoResponse, error)
	Update(ctx context.Context, userID, id string, req dto.TodoUpdateRequest) (dto.TodoResponse, error)
	Delete(ctx context.Context, userID, id string) error
}

type todoService struct {
	todoRepository repository.TodoRepository
	db             *gorm.DB
}

func NewTodoService(todoRepo repository.TodoRepository, db *gorm.DB) TodoService {
	return &todoService{
		todoRepository: todoRepo,
		db:             db,
	}
}

func toResponse(t entities.Todo) dto.TodoResponse {
	return dto.TodoResponse{
		ID:        t.ID.String(),
		Name:      t.Name,
		Category:  t.Category,
		IsDone:    t.IsDone,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func (s *todoService) Create(ctx context.Context, userID string, req dto.TodoCreateRequest) (dto.TodoResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.TodoResponse{}, dto.ErrInvalidUser
	}

	todo := entities.Todo{
		ID:       uuid.New(),
		UserID:   uid,
		Name:     req.Name,
		Category: req.Category,
		IsDone:   false,
	}

	created, err := s.todoRepository.Create(ctx, s.db, todo)
	if err != nil {
		return dto.TodoResponse{}, err
	}

	return toResponse(created), nil
}

func (s *todoService) GetAll(ctx context.Context, userID string, req dto.PaginationRequest) (dto.TodoPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.TodoPaginationResponse{}, dto.ErrInvalidUser
	}

	if req.Page <= 0 {
		req.Page = constants.ENUM_PAGINATION_PAGE
	}
	if req.PerPage <= 0 {
		req.PerPage = constants.ENUM_PAGINATION_PER_PAGE
	}
	offset := (req.Page - 1) * req.PerPage

	todos, total, err := s.todoRepository.GetAllByUserID(ctx, s.db, uid, req.PerPage, offset)
	if err != nil {
		return dto.TodoPaginationResponse{}, err
	}

	data := make([]dto.TodoResponse, 0, len(todos))
	for _, t := range todos {
		data = append(data, toResponse(t))
	}

	maxPage := int((total + int64(req.PerPage) - 1) / int64(req.PerPage))

	return dto.TodoPaginationResponse{
		Data: data,
		Meta: dto.PaginationMeta{
			Page:    req.Page,
			PerPage: req.PerPage,
			MaxPage: maxPage,
			Total:   total,
		},
	}, nil
}

func (s *todoService) GetByID(ctx context.Context, userID, id string) (dto.TodoResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.TodoResponse{}, dto.ErrInvalidUser
	}

	tid, err := uuid.Parse(id)
	if err != nil {
		return dto.TodoResponse{}, dto.ErrTodoNotFound
	}

	todo, err := s.todoRepository.GetByID(ctx, s.db, tid, uid)
	if err != nil {
		return dto.TodoResponse{}, dto.ErrTodoNotFound
	}

	return toResponse(todo), nil
}

func (s *todoService) Update(ctx context.Context, userID, id string, req dto.TodoUpdateRequest) (dto.TodoResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.TodoResponse{}, dto.ErrInvalidUser
	}

	tid, err := uuid.Parse(id)
	if err != nil {
		return dto.TodoResponse{}, dto.ErrTodoNotFound
	}

	todo, err := s.todoRepository.GetByID(ctx, s.db, tid, uid)
	if err != nil {
		return dto.TodoResponse{}, dto.ErrTodoNotFound
	}

	if req.Name != nil {
		todo.Name = *req.Name
	}
	if req.Category != nil {
		todo.Category = *req.Category
	}
	if req.IsDone != nil {
		todo.IsDone = *req.IsDone
	}

	updated, err := s.todoRepository.Update(ctx, s.db, todo)
	if err != nil {
		return dto.TodoResponse{}, err
	}

	return toResponse(updated), nil
}

func (s *todoService) Delete(ctx context.Context, userID, id string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.ErrInvalidUser
	}

	tid, err := uuid.Parse(id)
	if err != nil {
		return dto.ErrTodoNotFound
	}

	if _, err := s.todoRepository.GetByID(ctx, s.db, tid, uid); err != nil {
		return dto.ErrTodoNotFound
	}

	return s.todoRepository.Delete(ctx, s.db, tid, uid)
}
