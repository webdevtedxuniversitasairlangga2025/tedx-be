package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/gin-gonic/gin"
)

func newAdminTestRouter(jwtSvc service.JWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/admin", Authenticate(jwtSvc), AuthorizeAdmin(jwtSvc), func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return router
}

func adminTestRequest(router *gin.Engine, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAuthorizeAdmin_AllowsAdminRole(t *testing.T) {
	jwtSvc := service.NewJWTService()
	router := newAdminTestRouter(jwtSvc)

	token := jwtSvc.GenerateAccessToken("user-1", "admin")
	rec := adminTestRequest(router, token)

	if rec.Code != http.StatusOK {
		t.Fatalf("role admin harus lolos, dapat status %d", rec.Code)
	}
}

func TestAuthorizeAdmin_RejectsUserRole(t *testing.T) {
	jwtSvc := service.NewJWTService()
	router := newAdminTestRouter(jwtSvc)

	token := jwtSvc.GenerateAccessToken("user-1", "user")
	rec := adminTestRequest(router, token)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("role user harus ditolak 403, dapat status %d", rec.Code)
	}
}

func TestAuthorizeAdmin_RejectsNoToken(t *testing.T) {
	jwtSvc := service.NewJWTService()
	router := newAdminTestRouter(jwtSvc)

	rec := adminTestRequest(router, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("tanpa token harus ditolak 401, dapat status %d", rec.Code)
	}
}
