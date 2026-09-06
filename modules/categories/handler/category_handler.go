package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/modules/categories/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/categories/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
)

type (
	CategoryHandler interface {
		GetAll(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		Create(ctx *gin.Context)
		Update(ctx *gin.Context)
		Delete(ctx *gin.Context)
	}

	categoryHandler struct {
		service service.CategoryService
	}
)

func NewCategoryHandler(injector *do.Injector, cs service.CategoryService) CategoryHandler {
	return &categoryHandler{service: cs}
}

func (h *categoryHandler) GetAll(c *gin.Context) {
	result, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_CATEGORIES, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_CATEGORIES, result)
	c.JSON(http.StatusOK, res)
}

func (h *categoryHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_CATEGORY, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	data, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_CATEGORY, err.Error(), nil)
		c.JSON(http.StatusNotFound, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_CATEGORY, data)
	c.JSON(http.StatusOK, res)
}

func (h *categoryHandler) Create(c *gin.Context) {
	var req dto.CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_CATEGORY, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_CATEGORY, result)
	c.JSON(http.StatusCreated, res)
}

func (h *categoryHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_CATEGORY, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	var req dto.CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_CATEGORY, err.Error(), nil)
		status := http.StatusBadRequest
		if errors.Is(err, dto.ErrCategoryNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_CATEGORY, result)
	c.JSON(http.StatusOK, res)
}

func (h *categoryHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_CATEGORY, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_CATEGORY, err.Error(), nil)
		status := http.StatusBadRequest
		if errors.Is(err, dto.ErrCategoryNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_CATEGORY, nil)
	c.JSON(http.StatusOK, res)
}
