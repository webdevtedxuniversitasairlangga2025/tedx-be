package dto

import (
	"errors"
	"time"
)

const (
	// Failed
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_FAILED_CREATE_TODO        = "failed create todo"
	MESSAGE_FAILED_GET_LIST_TODO      = "failed get list todo"
	MESSAGE_FAILED_GET_TODO           = "failed get todo"
	MESSAGE_FAILED_UPDATE_TODO        = "failed update todo"
	MESSAGE_FAILED_DELETE_TODO        = "failed delete todo"

	// Success
	MESSAGE_SUCCESS_CREATE_TODO   = "success create todo"
	MESSAGE_SUCCESS_GET_LIST_TODO = "success get list todo"
	MESSAGE_SUCCESS_GET_TODO      = "success get todo"
	MESSAGE_SUCCESS_UPDATE_TODO   = "success update todo"
	MESSAGE_SUCCESS_DELETE_TODO   = "success delete todo"
)

var (
	ErrTodoNotFound = errors.New("todo not found")
	ErrInvalidUser  = errors.New("invalid user id")
)

type (
	TodoCreateRequest struct {
		Name     string `json:"name" form:"name" binding:"required,min=1,max=100"`
		Category string `json:"category" form:"category" binding:"required,min=1,max=100"`
	}

	TodoUpdateRequest struct {
		Name     *string `json:"name" form:"name" binding:"omitempty,min=1,max=100"`
		Category *string `json:"category" form:"category" binding:"omitempty,min=1,max=100"`
		IsDone   *bool   `json:"is_done" form:"is_done"`
	}

	TodoResponse struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Category  string    `json:"category"`
		IsDone    bool      `json:"is_done"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	PaginationRequest struct {
		Page    int `form:"page"`
		PerPage int `form:"per_page"`
	}

	PaginationMeta struct {
		Page    int   `json:"page"`
		PerPage int   `json:"per_page"`
		MaxPage int   `json:"max_page"`
		Total   int64 `json:"total"`
	}

	TodoPaginationResponse struct {
		Data []TodoResponse `json:"data"`
		Meta PaginationMeta `json:"meta"`
	}
)
