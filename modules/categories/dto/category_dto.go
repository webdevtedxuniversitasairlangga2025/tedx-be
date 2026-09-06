package dto

import (
	"errors"
	"time"
)

const (
	MESSAGE_FAILED_GET_DATA_FROM_BODY  = "failed get data from body"
	MESSAGE_FAILED_CREATE_CATEGORY     = "failed create category"
	MESSAGE_FAILED_GET_LIST_CATEGORIES = "failed get list categories"
	MESSAGE_FAILED_GET_CATEGORY        = "failed get category"
	MESSAGE_FAILED_UPDATE_CATEGORY     = "failed update category"
	MESSAGE_FAILED_DELETE_CATEGORY     = "failed delete category"

	MESSAGE_SUCCESS_CREATE_CATEGORY     = "success create category"
	MESSAGE_SUCCESS_GET_LIST_CATEGORIES = "success get list categories"
	MESSAGE_SUCCESS_GET_CATEGORY        = "success get category"
	MESSAGE_SUCCESS_UPDATE_CATEGORY     = "success update category"
	MESSAGE_SUCCESS_DELETE_CATEGORY     = "success delete category"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryExists   = errors.New("category already exists")
	ErrCategoryInUse    = errors.New("category is still used by merchandise")
)

type CategoryCreateRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

type CategoryUpdateRequest struct {
	Name *string `json:"name" binding:"omitempty,min=1,max=50"`
}

type CategoryResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
