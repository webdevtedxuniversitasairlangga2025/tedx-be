package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/modules/order/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/order/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
)

type OrderHandler interface {
	Create(ctx *gin.Context)
	GetByID(ctx *gin.Context)
	GetMyOrders(ctx *gin.Context)
	GetAll(ctx *gin.Context)
	Approve(ctx *gin.Context)
	Reject(ctx *gin.Context)
}

type orderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(injector *do.Injector, s service.OrderService) OrderHandler {
	return &orderHandler{orderService: s}
}

func (h *orderHandler) Create(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	var req dto.OrderCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	result, err := h.orderService.Create(ctx.Request.Context(), userID, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_ORDER, result)
	ctx.JSON(http.StatusCreated, res)
}

func (h *orderHandler) GetByID(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")
	result, err := h.orderService.GetByID(ctx.Request.Context(), userID, id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusNotFound, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_ORDER, result)
	ctx.JSON(http.StatusOK, res)
}

func (h *orderHandler) GetMyOrders(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	var pagination dto.PaginationRequest
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	result, err := h.orderService.GetMyOrders(ctx.Request.Context(), userID, pagination)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_ORDER, result)
	ctx.JSON(http.StatusOK, res)
}

func (h *orderHandler) GetAll(ctx *gin.Context) {
	var pagination dto.PaginationRequest
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	var filter dto.OrderFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	result, err := h.orderService.GetAll(ctx.Request.Context(), filter, pagination)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_ORDER, result)
	ctx.JSON(http.StatusOK, res)
}

func (h *orderHandler) Approve(ctx *gin.Context) {
	adminID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")
	result, err := h.orderService.Approve(ctx.Request.Context(), adminID, id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_APPROVE_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_APPROVE_ORDER, result)
	ctx.JSON(http.StatusOK, res)
}

func (h *orderHandler) Reject(ctx *gin.Context) {
	adminID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")
	var req dto.OrderRejectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	result, err := h.orderService.Reject(ctx.Request.Context(), adminID, id, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_REJECT_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_REJECT_ORDER, result)
	ctx.JSON(http.StatusOK, res)
}
