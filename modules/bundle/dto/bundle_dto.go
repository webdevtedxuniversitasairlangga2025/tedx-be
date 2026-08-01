package dto

import (
	"errors"
	"time"
)

const (
	// Failed
	MESSAGE_FAILED_GET_DATA_FROM_BODY  = "failed get data from body"
	MESSAGE_FAILED_CREATE_BUNDLE       = "failed create bundle"
	MESSAGE_FAILED_GET_LIST_BUNDLE     = "failed get list bundle"
	MESSAGE_FAILED_GET_BUNDLE          = "failed get bundle"
	MESSAGE_FAILED_UPDATE_BUNDLE       = "failed update bundle"
	MESSAGE_FAILED_DELETE_BUNDLE       = "failed delete bundle"
	MESSAGE_FAILED_ADD_BUNDLE_IMAGE    = "failed add bundle image"
	MESSAGE_FAILED_DELETE_BUNDLE_IMAGE = "failed delete bundle image"

	// Success
	MESSAGE_SUCCESS_CREATE_BUNDLE       = "success create bundle"
	MESSAGE_SUCCESS_GET_LIST_BUNDLE     = "success get list bundle"
	MESSAGE_SUCCESS_GET_BUNDLE          = "success get bundle"
	MESSAGE_SUCCESS_UPDATE_BUNDLE       = "success update bundle"
	MESSAGE_SUCCESS_DELETE_BUNDLE       = "success delete bundle"
	MESSAGE_SUCCESS_ADD_BUNDLE_IMAGE    = "success add bundle image"
	MESSAGE_SUCCESS_DELETE_BUNDLE_IMAGE = "success delete bundle image"
)

var (
	ErrBundleNotFound      = errors.New("bundle not found")
	ErrBundleImageNotFound = errors.New("bundle image not found")
	ErrInvalidPrice        = errors.New("invalid price format")
	ErrPriceOutOfRange     = errors.New("price must be between 0 and 99999999.99")
)

type (
	// BundleCreateRequest sengaja TIDAK menerima is_active: bundle selalu dibuat
	// dalam keadaan aktif, lalu disembunyikan lewat PATCH bila perlu.
	// Alasannya kolom IsActive di entity ber-tag `default:true`, dan GORM
	// mengabaikan nilai zero (false) untuk field bertag default saat INSERT —
	// sehingga "is_active": false tetap akan tersimpan sebagai true.
	// Jangan tambahkan field is_active di sini tanpa mengubah tipe kolomnya
	// menjadi *bool di entity (dan menyamakannya dengan merchandise).
	BundleCreateRequest struct {
		Name        string `json:"name" form:"name" binding:"required,min=1,max=255"`
		Description string `json:"description" form:"description" binding:"required,min=1"`
		// Price dikirim sebagai string (mis. "150000.00") agar presisi uang
		// tidak rusak oleh pembulatan float di sisi client maupun server.
		Price string `json:"price" form:"price" binding:"required"`
	}

	// Field pointer agar bisa membedakan "tidak dikirim" vs "dikirim kosong".
	BundleUpdateRequest struct {
		Name        *string `json:"name" form:"name" binding:"omitempty,min=1,max=255"`
		Description *string `json:"description" form:"description" binding:"omitempty,min=1"`
		Price       *string `json:"price" form:"price" binding:"omitempty"`
		IsActive    *bool   `json:"is_active" form:"is_active"`
	}

	BundleImageCreateRequest struct {
		ImageURL string `json:"image_url" form:"image_url" binding:"required,url,max=255"`
	}

	// IsActive memakai pointer agar bisa membedakan "tidak difilter" vs
	// "difilter dengan nilai false".
	BundleQueryRequest struct {
		Page     int   `form:"page"`
		PerPage  int   `form:"per_page"`
		IsActive *bool `form:"is_active"`
	}

	BundleImageResponse struct {
		ID       string `json:"id"`
		ImageURL string `json:"image_url"`
	}

	BundleResponse struct {
		ID          string    `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Price       string    `json:"price"`
		IsActive    bool      `json:"is_active"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	// BundleDetailResponse dipakai khusus endpoint detail, yang ikut membawa
	// daftar gambar. Field BundleResponse di-embed sehingga JSON-nya tetap rata
	// (tidak bersarang), sama seperti pola embed Timestamp di entity.
	BundleDetailResponse struct {
		BundleResponse
		Images []BundleImageResponse `json:"images"`
	}

	PaginationMeta struct {
		Page    int   `json:"page"`
		PerPage int   `json:"per_page"`
		MaxPage int   `json:"max_page"`
		Total   int64 `json:"total"`
	}

	BundlePaginationResponse struct {
		Data []BundleResponse `json:"data"`
		Meta PaginationMeta   `json:"meta"`
	}
)
