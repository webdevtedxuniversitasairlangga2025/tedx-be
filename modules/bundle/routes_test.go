package bundle

import (
	"testing"

	"github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/webdevtedxuniversitasairlangga/modules/bundle/handler"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

// stubBundleHandler hanya memenuhi interface; isi handler diuji di lapisan lain.
type stubBundleHandler struct{}

func (stubBundleHandler) Create(*gin.Context)      {}
func (stubBundleHandler) GetAll(*gin.Context)      {}
func (stubBundleHandler) GetByID(*gin.Context)     {}
func (stubBundleHandler) Update(*gin.Context)      {}
func (stubBundleHandler) Delete(*gin.Context)      {}
func (stubBundleHandler) AddImage(*gin.Context)    {}
func (stubBundleHandler) DeleteImage(*gin.Context) {}

// TestRegisterRoutes memastikan seluruh endpoint terdaftar dengan path yang benar
// dan tidak ada bentrok pola wildcard di router Gin — bentrok seperti itu baru
// ketahuan sebagai panic saat aplikasi boot, bukan saat compile.
func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	injector := do.New()
	do.Provide(injector, func(i *do.Injector) (handler.BundleHandler, error) {
		return stubBundleHandler{}, nil
	})
	do.ProvideNamed(injector, constants.JWTService, func(i *do.Injector) (service.JWTService, error) {
		return service.NewJWTService(), nil
	})

	router := gin.New()
	RegisterRoutes(router.Group("api/v1"), injector)

	registered := make(map[string]bool)
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	expected := []string{
		"GET /api/v1/bundles",
		"GET /api/v1/bundles/:id",
		"POST /api/v1/bundles",
		"PATCH /api/v1/bundles/:id",
		"DELETE /api/v1/bundles/:id",
		"POST /api/v1/bundles/:id/images",
		"DELETE /api/v1/bundles/:id/images/:imageId",
	}

	for _, route := range expected {
		if !registered[route] {
			t.Errorf("route %q tidak terdaftar", route)
		}
	}

	if len(router.Routes()) != len(expected) {
		t.Errorf("jumlah route terdaftar %d, seharusnya %d", len(router.Routes()), len(expected))
	}
}
