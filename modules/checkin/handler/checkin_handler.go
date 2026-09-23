package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/modules/checkin/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/checkin/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
)

type CheckInHandler interface {
	CheckIn(ctx *gin.Context)
}

type checkInHandler struct {
	service service.CheckInService
}

func NewCheckInHandler(_ *do.Injector, service service.CheckInService) CheckInHandler {
	return &checkInHandler{service: service}
}

func (h *checkInHandler) CheckIn(ctx *gin.Context) {
	var req dto.CheckInRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := h.service.CheckIn(ctx.Request.Context(), ctx.MustGet("user_id").(string), req)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, dto.ErrAttendeeTicketNotFound) {
			status = http.StatusNotFound
		} else if errors.Is(err, dto.ErrTicketAlreadyUsed) || errors.Is(err, dto.ErrTicketNotPaid) {
			status = http.StatusConflict
		}
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CHECK_IN, err.Error(), nil)
		ctx.JSON(status, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CHECK_IN, result)
	ctx.JSON(http.StatusOK, res)
}
