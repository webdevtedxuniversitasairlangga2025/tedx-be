package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
)

type MerchandiseHandler struct {
	service service.MerchandiseService
}

func NewMerchandiseHandler(i *do.Injector) (*MerchandiseHandler, error) {
	svc := do.MustInvoke[service.MerchandiseService](i)
	return &MerchandiseHandler{
		service: svc,
	}, nil
}

func (h *MerchandiseHandler) GetAll(c *gin.Context) {
	var filter dto.MerchandiseFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		res := utils.BuildResponseFailed("failed get data from body", err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	data, total, err := h.service.GetAll(c.Request.Context(), filter)
	if err != nil {
		res := utils.BuildResponseFailed("failed get list merchandise", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.Response{
		Status:  true,
		Message: "success get list merchandise",
		Data:    data,
		Meta: gin.H{
			"total": total,
			"page":  filter.Page,
			"limit": filter.Limit,
		},
	}
	c.JSON(http.StatusOK, res)
}

func (h *MerchandiseHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed("failed get merchandise", "Invalid UUID", nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	data, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed("failed get merchandise", err.Error(), nil)
		c.JSON(http.StatusNotFound, res)
		return
	}

	res := utils.BuildResponseSuccess("success get merchandise", data)
	c.JSON(http.StatusOK, res)
}

func (h *MerchandiseHandler) Create(c *gin.Context) {
	var req dto.MerchandiseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed("failed get data from body", err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	if err := h.service.Create(c.Request.Context(), req); err != nil {
		res := utils.BuildResponseFailed("failed create merchandise", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess("success create merchandise", nil)
	c.JSON(http.StatusCreated, res)
}

func (h *MerchandiseHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed("failed update merchandise", "Invalid UUID", nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	var req dto.MerchandiseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed("failed get data from body", err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	if err := h.service.Update(c.Request.Context(), id, req); err != nil {
		res := utils.BuildResponseFailed("failed update merchandise", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess("success update merchandise", nil)
	c.JSON(http.StatusOK, res)
}

func (h *MerchandiseHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed("failed delete merchandise", "Invalid UUID", nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		res := utils.BuildResponseFailed("failed delete merchandise", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess("success delete merchandise", nil)
	c.JSON(http.StatusOK, res)
}

func (h *MerchandiseHandler) AddImage(c *gin.Context) {
	merchId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed("failed add merchandise image", "Invalid Merchandise UUID", nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	var req dto.MerchImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed("failed get data from body", err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	if err := h.service.AddImage(c.Request.Context(), merchId, req); err != nil {
		res := utils.BuildResponseFailed("failed add merchandise image", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess("success add merchandise image", nil)
	c.JSON(http.StatusCreated, res)
}

func (h *MerchandiseHandler) DeleteImage(c *gin.Context) {
	imageId, err := uuid.Parse(c.Param("imageId"))
	if err != nil {
		res := utils.BuildResponseFailed("failed delete merchandise image", "Invalid Image UUID", nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	if err := h.service.DeleteImage(c.Request.Context(), imageId); err != nil {
		res := utils.BuildResponseFailed("failed delete merchandise image", err.Error(), nil)
		c.JSON(http.StatusInternalServerError, res)
		return
	}

	res := utils.BuildResponseSuccess("success delete merchandise image", nil)
	c.JSON(http.StatusOK, res)
}
