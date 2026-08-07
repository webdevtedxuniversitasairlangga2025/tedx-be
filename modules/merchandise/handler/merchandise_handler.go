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
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.service.GetAll(c.Request.Context(), filter)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_MERCHANDISE, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_MERCHANDISE, result)
	c.JSON(http.StatusOK, res)
}

func (h *MerchandiseHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_MERCHANDISE, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	data, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_MERCHANDISE, err.Error(), nil)
		c.JSON(http.StatusNotFound, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_MERCHANDISE, data)
	c.JSON(http.StatusOK, res)
}

func (h *MerchandiseHandler) Create(c *gin.Context) {
	var req dto.MerchandiseCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_MERCHANDISE, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_MERCHANDISE, result)
	c.JSON(http.StatusCreated, res)
}

func (h *MerchandiseHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_MERCHANDISE, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	var req dto.MerchandiseUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_MERCHANDISE, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_MERCHANDISE, result)
	c.JSON(http.StatusOK, res)
}

func (h *MerchandiseHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_MERCHANDISE, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_MERCHANDISE, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_MERCHANDISE, nil)
	c.JSON(http.StatusOK, res)
}

func (h *MerchandiseHandler) AddImage(c *gin.Context) {
	merchId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_ADD_MERCHANDISE_IMAGE, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	var req dto.MerchImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := h.service.AddImage(c.Request.Context(), merchId, req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_ADD_MERCHANDISE_IMAGE, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_ADD_MERCHANDISE_IMAGE, nil)
	c.JSON(http.StatusCreated, res)
}

func (h *MerchandiseHandler) DeleteImage(c *gin.Context) {
	merchId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_MERCHANDISE_IMAGE, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	imageId, err := uuid.Parse(c.Param("imageId"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_MERCHANDISE_IMAGE, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	if err := h.service.DeleteImage(c.Request.Context(), merchId, imageId); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_MERCHANDISE_IMAGE, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_MERCHANDISE_IMAGE, nil)
	c.JSON(http.StatusOK, res)
}
