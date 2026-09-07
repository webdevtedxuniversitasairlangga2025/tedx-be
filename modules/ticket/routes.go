package ticket

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/middlewares"

	authService "github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/webdevtedxuniversitasairlangga/modules/ticket/handler"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
)

func RegisterRoutes(r *gin.RouterGroup, i *do.Injector) {

	ticketHandler := do.MustInvoke[handler.TicketHandler](i)
	jwtSvc := do.MustInvokeNamed[authService.JWTService](i, constants.JWTService)

	ticketGroup := r.Group("/tickets")

	ticketGroup.GET("", ticketHandler.GetAll)
	ticketGroup.GET("/:id", ticketHandler.GetByID)

	adminGroup := ticketGroup.Group("")
	adminGroup.Use(middlewares.Authenticate(jwtSvc), middlewares.AuthorizeAdmin(jwtSvc))
	{
		adminGroup.POST("", ticketHandler.Create)
		adminGroup.PATCH("/:id", ticketHandler.Update)
		adminGroup.DELETE("/:id", ticketHandler.Delete)

		adminGroup.POST("/:id/tiers", ticketHandler.CreateTier)
		adminGroup.PATCH("/:id/tiers/:tierId", ticketHandler.UpdateTier)
		adminGroup.DELETE("/:id/tiers/:tierId", ticketHandler.DeleteTier)
	}
}
