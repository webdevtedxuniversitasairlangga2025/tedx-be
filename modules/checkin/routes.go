package checkin

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/middlewares"
	authService "github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/webdevtedxuniversitasairlangga/modules/checkin/handler"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
)

func RegisterRoutes(r *gin.RouterGroup, i *do.Injector) {
	checkInHandler := do.MustInvoke[handler.CheckInHandler](i)
	jwtSvc := do.MustInvokeNamed[authService.JWTService](i, constants.JWTService)

	checkInRoutes := r.Group("/checkins", middlewares.Authenticate(jwtSvc), middlewares.AuthorizeAdmin(jwtSvc))
	checkInRoutes.POST("/scan", checkInHandler.CheckIn)
}
