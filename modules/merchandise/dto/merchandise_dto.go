package dto

import (
	"errors"
	"time"
)

const (
	MESSAGE_FAILED_GET_DATA_FROM_BODY       = "failed get data from body"
	MESSAGE_FAILED_CREATE_MERCHANDISE       = "failed create merchandise"
	MESSAGE_FAILED_GET_LIST_MERCHANDISE     = "failed get list merchandise"
	MESSAGE_FAILED_GET_MERCHANDISE          = "failed get merchandise"
	MESSAGE_FAILED_UPDATE_MERCHANDISE       = "failed update merchandise"
	MESSAGE_FAILED_DELETE_MERCHANDISE       = "failed delete merchandise"
	MESSAGE_FAILED_ADD_MERCHANDISE_IMAGE    = "failed add merchandise image"
	MESSAGE_FAILED_DELETE_MERCHANDISE_IMAGE = "failed delete merchandise image"

	MESSAGE_SUCCESS_CREATE_MERCHANDISE       = "success create merchandise"
	MESSAGE_SUCCESS_GET_LIST_MERCHANDISE     = "success get list merchandise"
	MESSAGE_SUCCESS_GET_MERCHANDISE          = "success get merchandise"
	MESSAGE_SUCCESS_UPDATE_MERCHANDISE       = "success update merchandise"
	MESSAGE_SUCCESS_DELETE_MERCHANDISE       = "success delete merchandise"
	MESSAGE_SUCCESS_ADD_MERCHANDISE_IMAGE    = "success add merchandise image"
	MESSAGE_SUCCESS_DELETE_MERCHANDISE_IMAGE = "success delete merchandise image"
)

var (
	ErrInvalidPrice       = errors.New("invalid price format")
	ErrPriceOutOfRange    = errors.New("price must be between 0 and 99999999.99")
	ErrMerchImageNotFound = errors.New("merchandise image not found")
	ErrCategoryNotFound   = errors.New("category not found")
)

type MerchandiseCreateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"required,min=1"`
	Price       string `json:"price" binding:"required"`
	CategoryID  string `json:"category_id" binding:"required,uuid4"`
}

type MerchandiseUpdateRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,min=1"`
	Price       *string `json:"price" binding:"omitempty"`
	CategoryID  *string `json:"category_id" binding:"omitempty,uuid4"`
	IsActive    *bool   `json:"is_active"`
}

type MerchImageRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
}

type MerchandiseFilter struct {
	CategoryID string `form:"category_id"`
	IsActive   *bool   `form:"is_active"`
}

type MerchImageResponse struct {
	ID       string `json:"id"`
	ImageURL string `json:"image_url"`
}

type CategoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type MerchandiseResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Price       string               `json:"price"`
	Category    CategoryResponse     `json:"category"`
	IsActive    bool                 `json:"is_active"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Images      []MerchImageResponse `json:"images"`
}
