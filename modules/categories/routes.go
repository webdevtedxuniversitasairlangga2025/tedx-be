package categories

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/middlewares"
	authService "github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/webdevtedxuniversitasairlangga/modules/categories/handler"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	categoryController := do.MustInvoke[handler.CategoryHandler](injector)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	categoryRoutes := server.Group("/categories")
	{
		categoryRoutes.GET("", categoryController.GetAll)
		categoryRoutes.GET("/:id", categoryController.GetByID)

		adminRoutes := categoryRoutes.Group("", middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService))
		{
			adminRoutes.POST("", categoryController.Create)
			adminRoutes.PATCH("/:id", categoryController.Update)
			adminRoutes.DELETE("/:id", categoryController.Delete)
		}
	}
}
