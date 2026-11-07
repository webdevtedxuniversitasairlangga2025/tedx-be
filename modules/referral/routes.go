package referral

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/middlewares"
	authService "github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/webdevtedxuniversitasairlangga/modules/referral/handler"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
)

func RegisterRoutes(r *gin.RouterGroup, i *do.Injector) {
	h := do.MustInvoke[handler.ReferralHandler](i)
	jwtSvc := do.MustInvokeNamed[authService.JWTService](i, constants.JWTService)

	userGroup := r.Group("/referrals")
	userGroup.Use(middlewares.Authenticate(jwtSvc))
	userGroup.POST("/preview", h.Preview)

	adminGroup := r.Group("/referrals")
	adminGroup.Use(middlewares.Authenticate(jwtSvc), middlewares.AuthorizeAdmin(jwtSvc))
	adminGroup.POST("", h.Create)
	adminGroup.GET("", h.List)
	adminGroup.PATCH("/:id/active", h.SetActive)
	adminGroup.DELETE("/:id", h.Delete)
}
