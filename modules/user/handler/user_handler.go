package handler

import (
	"net/http"

	"github.com/webdevtedxuniversitasairlangga/modules/user/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/user/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

type (
	UserHandler interface {
		GetAll(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		Update(ctx *gin.Context)
		Delete(ctx *gin.Context)
	}

	userHandler struct {
		userService service.UserService
	}
)

func NewUserHandler(injector *do.Injector, us service.UserService) UserHandler {
	return &userHandler{
		userService: us,
	}
}

func (h *userHandler) GetAll(c *gin.Context) {
	var filter dto.UserFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.userService.GetAll(c.Request.Context(), filter)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_USER, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_USER, result)
	c.JSON(http.StatusOK, res)
}

func (h *userHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_USER, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_USER, user)
	c.JSON(http.StatusOK, res)
}

func (h *userHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	user, err := h.userService.Update(c.Request.Context(), id, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_USER, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_USER, user)
	c.JSON(http.StatusOK, res)
}

func (h *userHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_USER, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_USER, nil)
	c.JSON(http.StatusOK, res)
}