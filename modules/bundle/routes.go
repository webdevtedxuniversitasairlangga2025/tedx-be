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
		// Publik — katalog bundle bisa dilihat tanpa login.
		bundleRoutes.GET("", bundleController.GetAll)
		bundleRoutes.GET("/:id", bundleController.GetByID)

		// TODO(BE-2): pasang middlewares.AuthorizeAdmin() pada grup ini setelah
		// task "Setup admin authorization middleware" (Manager) selesai.
		// Untuk saat ini grup di bawah baru menuntut login dan BELUM membatasi
		// role admin, sehingga user biasa masih bisa mengubah katalog.
		// Checklist "endpoint admin diproteksi middleware admin" belum terpenuhi
		// sampai baris itu ditambahkan.
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
