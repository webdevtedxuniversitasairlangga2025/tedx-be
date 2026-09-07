package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/modules/ticket/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/ticket/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
)

type (
	TicketHandler interface {
		GetAll(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		Create(ctx *gin.Context)
		Update(ctx *gin.Context)
		Delete(ctx *gin.Context)

		CreateTier(ctx *gin.Context)
		UpdateTier(ctx *gin.Context)
		DeleteTier(ctx *gin.Context)
	}

	ticketHandler struct {
		service service.TicketService
	}
)

func NewTicketHandler(injector *do.Injector, ts service.TicketService) TicketHandler {
	return &ticketHandler{service: ts}
}

func (h *ticketHandler) GetAll(c *gin.Context) {
	var filter dto.TicketFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.service.GetAll(c.Request.Context(), filter)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_TICKET, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_TICKET, result)
	c.JSON(http.StatusOK, res)
}

func (h *ticketHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_TICKET, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	data, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_TICKET, err.Error(), nil)
		c.JSON(http.StatusNotFound, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_TICKET, data)
	c.JSON(http.StatusOK, res)
}

func (h *ticketHandler) Create(c *gin.Context) {
	var req dto.TicketCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_TICKET, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_TICKET, result)
	c.JSON(http.StatusCreated, res)
}

func (h *ticketHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_TICKET, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	var req dto.TicketUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_TICKET, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_TICKET, result)
	c.JSON(http.StatusOK, res)
}

func (h *ticketHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_TICKET, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_TICKET, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_TICKET, nil)
	c.JSON(http.StatusOK, res)
}

func (h *ticketHandler) CreateTier(c *gin.Context) {
	ticketId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_TICKET_TIER, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	var req dto.TicketTierCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.service.CreateTier(c.Request.Context(), ticketId, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_TICKET_TIER, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_TICKET_TIER, result)
	c.JSON(http.StatusCreated, res)
}

func (h *ticketHandler) UpdateTier(c *gin.Context) {
	ticketId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_TICKET_TIER, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	tierId, err := uuid.Parse(c.Param("tierId"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_TICKET_TIER, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	var req dto.TicketTierUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		c.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.service.UpdateTier(c.Request.Context(), ticketId, tierId, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_TICKET_TIER, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_TICKET_TIER, result)
	c.JSON(http.StatusOK, res)
}

func (h *ticketHandler) DeleteTier(c *gin.Context) {
	ticketId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_TICKET_TIER, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	tierId, err := uuid.Parse(c.Param("tierId"))
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_TICKET_TIER, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	if err := h.service.DeleteTier(c.Request.Context(), ticketId, tierId); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_TICKET_TIER, err.Error(), nil)
		c.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_TICKET_TIER, nil)
	c.JSON(http.StatusOK, res)
}
