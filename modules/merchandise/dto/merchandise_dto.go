package dto

import (
	"github.com/shopspring/decimal"
)

type MerchandiseRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description" binding:"required"`
	Price       decimal.Decimal `json:"price" binding:"required"`
	Category    string          `json:"category" binding:"required"`
	IsActive    *bool           `json:"is_active" binding:"required"`
}

type MerchImageRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
}

type MerchandiseFilter struct {
	Category string `form:"category"`
	IsActive *bool  `form:"is_active"`
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=10"`
}
