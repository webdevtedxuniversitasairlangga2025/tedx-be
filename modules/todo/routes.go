package todo

import (
	"github.com/webdevtedxuniversitasairlangga/middlewares"
	"github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/webdevtedxuniversitasairlangga/modules/todo/handler"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	todoController := do.MustInvoke[handler.TodoHandler](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	todoRoutes := server.Group("/todos", middlewares.Authenticate(jwtService))
	{
		todoRoutes.POST("", todoController.Create)
		todoRoutes.GET("", todoController.GetAll)
		todoRoutes.GET("/:id", todoController.GetByID)
		todoRoutes.PATCH("/:id", todoController.Update)
		todoRoutes.DELETE("/:id", todoController.Delete)
	}
}
