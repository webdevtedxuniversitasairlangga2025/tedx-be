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
	ErrInvalidPrice    = errors.New("invalid price format")
	ErrPriceOutOfRange = errors.New("price must be between 0 and 99999999.99")
)

type MerchandiseCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Price       string `json:"price" binding:"required"`
	Category    string `json:"category" binding:"required"`
}

type MerchandiseUpdateRequest struct {
	Name        *string `json:"name" binding:"omitempty"`
	Description *string `json:"description" binding:"omitempty"`
	Price       *string `json:"price" binding:"omitempty"`
	Category    *string `json:"category" binding:"omitempty"`
	IsActive    *bool   `json:"is_active"`
}

type MerchImageRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
}

type MerchandiseFilter struct {
	Category string `form:"category"`
	IsActive *bool  `form:"is_active"`
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
}

type MerchImageResponse struct {
	ID       string `json:"id"`
	ImageURL string `json:"image_url"`
}

type MerchandiseResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Price       string               `json:"price"`
	Category    string               `json:"category"`
	IsActive    bool                 `json:"is_active"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Images      []MerchImageResponse `json:"images"`
}

type PaginationMeta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	MaxPage int   `json:"max_page"`
	Total   int64 `json:"total"`
}

type MerchandisePaginationResponse struct {
	Data []MerchandiseResponse `json:"data"`
	Meta PaginationMeta        `json:"meta"`
}
