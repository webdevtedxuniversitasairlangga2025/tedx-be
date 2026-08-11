package middlewares

import (
	"net/http"

	"github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/webdevtedxuniversitasairlangga/modules/user/dto"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
	"github.com/gin-gonic/gin"
)

func AuthorizeAdmin(jwtService service.JWTService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetString("token")

		role, err := jwtService.GetRoleByToken(token)
		if err != nil {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_TOKEN_NOT_VALID, nil)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		if role != constants.ENUM_ROLE_ADMIN {
			response := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, dto.MESSAGE_FAILED_DENIED_ACCESS, nil)
			ctx.AbortWithStatusJSON(http.StatusForbidden, response)
			return
		}

		ctx.Next()
	}
}
