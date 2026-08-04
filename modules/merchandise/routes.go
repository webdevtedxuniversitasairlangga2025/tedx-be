package merchandise

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/middlewares"

	authService "github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/handler"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/repository"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/service"
)

func RegisterRoutes(r *gin.RouterGroup, i *do.Injector) {

	do.Provide(i, repository.NewMerchandiseRepository)
	do.Provide(i, service.NewMerchandiseService)
	do.Provide(i, handler.NewMerchandiseHandler)

	merchandiseHandler := do.MustInvoke[*handler.MerchandiseHandler](i)
	jwtSvc := do.MustInvokeNamed[authService.JWTService](i, "JWTService")

	merchandiseGroup := r.Group("/merchandise")

	merchandiseGroup.GET("", merchandiseHandler.GetAll)
	merchandiseGroup.GET("/:id", merchandiseHandler.GetByID)

	adminGroup := merchandiseGroup.Group("")
	adminGroup.Use(middlewares.Authenticate(jwtSvc))
	{
		adminGroup.POST("", merchandiseHandler.Create)
		adminGroup.PATCH("/:id", merchandiseHandler.Update)
		adminGroup.DELETE("/:id", merchandiseHandler.Delete)

		adminGroup.POST("/:id/images", merchandiseHandler.AddImage)
		adminGroup.DELETE("/:id/images/:imageId", merchandiseHandler.DeleteImage)
	}
}
