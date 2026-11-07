package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/modules/referral/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/referral/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
	"gorm.io/gorm"
)

type ReferralHandler interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	SetActive(c *gin.Context)
	Delete(c *gin.Context)
	Preview(c *gin.Context)
}

type referralHandler struct {
	service service.ReferralService
}

func NewReferralHandler(svc service.ReferralService) ReferralHandler {
	return &referralHandler{service: svc}
}

func (h *referralHandler) Create(c *gin.Context) {
	var req dto.CreateReferralCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.BuildResponseFailed("failed create referral code", err.Error(), nil))
		return
	}
	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildResponseFailed("failed create referral code", err.Error(), nil))
		return
	}
	c.JSON(http.StatusCreated, utils.BuildResponseSuccess("success create referral code", result))
}

func (h *referralHandler) List(c *gin.Context) {
	result, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed list referral codes", err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponseSuccess("success list referral codes", result))
}

func (h *referralHandler) SetActive(c *gin.Context) {
	if _, err := uuid.Parse(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildResponseFailed("failed update referral code", dto.ErrInvalidReferralID.Error(), nil))
		return
	}
	var req dto.SetReferralActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.BuildResponseFailed("failed update referral code", err.Error(), nil))
		return
	}
	result, err := h.service.SetActive(c.Request.Context(), c.Param("id"), req.IsActive)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildResponseFailed("failed update referral code", err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponseSuccess("success update referral code", result))
}

func (h *referralHandler) Delete(c *gin.Context) {
	if _, err := uuid.Parse(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildResponseFailed("failed delete referral code", dto.ErrInvalidReferralID.Error(), nil))
		return
	}
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		status := http.StatusInternalServerError
		if err == gorm.ErrRecordNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, utils.BuildResponseFailed("failed delete referral code", err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponseSuccess("success delete referral code", nil))
}

func (h *referralHandler) Preview(c *gin.Context) {
	var req dto.ReferralPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, utils.BuildResponseFailed("failed preview referral code", err.Error(), nil))
		return
	}
	result, err := h.service.Preview(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildResponseFailed("failed preview referral code", err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponseSuccess("success preview referral code", result))
}
