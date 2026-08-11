package bundle

import (
	"github.com/webdevtedxuniversitasairlangga/middlewares"
	"github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/webdevtedxuniversitasairlangga/modules/bundle/handler"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	bundleController := do.MustInvoke[handler.BundleHandler](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	bundleRoutes := server.Group("/bundles")
	{
		bundleRoutes.GET("", bundleController.GetAll)
		bundleRoutes.GET("/:id", bundleController.GetByID)

		adminRoutes := bundleRoutes.Group("", middlewares.Authenticate(jwtService))
		{
			adminRoutes.POST("", bundleController.Create)
			adminRoutes.PATCH("/:id", bundleController.Update)
			adminRoutes.DELETE("/:id", bundleController.Delete)
			adminRoutes.POST("/:id/images", bundleController.AddImage)
			adminRoutes.DELETE("/:id/images/:imageId", bundleController.DeleteImage)
		}
	}
}
