package handler

import (
	"net/http"

	"github.com/webdevtedxuniversitasairlangga/modules/todo/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/todo/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

type (
	TodoHandler interface {
		Create(ctx *gin.Context)
		GetAll(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		Update(ctx *gin.Context)
		Delete(ctx *gin.Context)
	}

	todoHandler struct {
		todoService service.TodoService
	}
)

func NewTodoHandler(injector *do.Injector, ts service.TodoService) TodoHandler {
	return &todoHandler{
		todoService: ts,
	}
}

func (c *todoHandler) Create(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)

	var req dto.TodoCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.todoService.Create(ctx.Request.Context(), userID, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_TODO, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_TODO, result)
	ctx.JSON(http.StatusCreated, res)
}

func (c *todoHandler) GetAll(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)

	var pagination dto.PaginationRequest
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.todoService.GetAll(ctx.Request.Context(), userID, pagination)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_TODO, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_TODO, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *todoHandler) GetByID(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")

	result, err := c.todoService.GetByID(ctx.Request.Context(), userID, id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_TODO, err.Error(), nil)
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_TODO, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *todoHandler) Update(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")

	var req dto.TodoUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.todoService.Update(ctx.Request.Context(), userID, id, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_TODO, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_TODO, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *todoHandler) Delete(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")

	err := c.todoService.Delete(ctx.Request.Context(), userID, id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_TODO, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_TODO, nil)
	ctx.JSON(http.StatusOK, res)
}
