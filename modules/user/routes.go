package user

import (
	"github.com/webdevtedxuniversitasairlangga/middlewares"
	"github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/webdevtedxuniversitasairlangga/modules/user/handler"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	userController := do.MustInvoke[handler.UserHandler](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	userRoutes := server.Group("/users")
	userRoutes.Use(middlewares.Authenticate(jwtService), middlewares.AuthorizeAdmin(jwtService))
	{
		userRoutes.GET("", userController.GetAll)
		userRoutes.GET("/:id", userController.GetByID)
		userRoutes.PATCH("/:id", userController.Update)
		userRoutes.DELETE("/:id", userController.Delete)
	}
}