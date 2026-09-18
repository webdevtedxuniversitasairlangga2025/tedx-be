package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/modules/webhook/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/webhook/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
)

type WebhookHandler interface {
	HandleMidtransNotification(ctx *gin.Context)
}

type webhookHandler struct {
	webhookService service.WebhookService
}

func NewWebhookHandler(injector *do.Injector, ws service.WebhookService) WebhookHandler {
	return &webhookHandler{webhookService: ws}
}

func (c *webhookHandler) HandleMidtransNotification(ctx *gin.Context) {
	var req dto.MidtransNotificationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROCESS_WEBHOOK, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	err := c.webhookService.ProcessMidtransNotification(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROCESS_WEBHOOK, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_PROCESS_WEBHOOK, nil)
	ctx.JSON(http.StatusOK, res)
}
