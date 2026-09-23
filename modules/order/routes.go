package order

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/middlewares"
	authService "github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/webdevtedxuniversitasairlangga/modules/order/handler"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
)

func RegisterRoutes(r *gin.RouterGroup, i *do.Injector) {
	orderHandler := do.MustInvoke[handler.OrderHandler](i)
	jwtSvc := do.MustInvokeNamed[authService.JWTService](i, constants.JWTService)

	userGroup := r.Group("/orders")
	userGroup.Use(middlewares.Authenticate(jwtSvc))
	{
		userGroup.POST("", orderHandler.Create)
		userGroup.GET("", orderHandler.GetMyOrders)
		userGroup.GET("/:id", orderHandler.GetByID)
		userGroup.PATCH("/:id/proof", orderHandler.UploadProof)
	}

	adminGroup := r.Group("/orders")
	adminGroup.Use(middlewares.Authenticate(jwtSvc), middlewares.AuthorizeAdmin(jwtSvc))
	{
		adminGroup.GET("/admin/all", orderHandler.GetAll)
		adminGroup.PATCH("/:id/approve", orderHandler.Approve)
		adminGroup.PATCH("/:id/reject", orderHandler.Reject)
		adminGroup.POST("/:id/resend-email", orderHandler.ResendEmail)
	}
}
