package webhook

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/modules/webhook/handler"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	webhookController := do.MustInvoke[handler.WebhookHandler](injector)

	webhookRoutes := server.Group("/webhook")
	{
		webhookRoutes.POST("/midtrans", webhookController.HandleMidtransNotification)
	}
}
