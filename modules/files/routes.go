package files

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/webdevtedxuniversitasairlangga/pkg/storage"
)

// RegisterRoutes — file publik katalog (bundle/merch) TANPA auth.
// Hanya key prefix "public/" yang dilayani; sisanya 404 (bukti bayar
// tetap lewat GET /orders/:id/proof yang butuh auth).
func RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/files/*filepath", func(ctx *gin.Context) {
		key := strings.TrimPrefix(ctx.Param("filepath"), "/")
		if !strings.HasPrefix(key, "public/") {
			ctx.JSON(http.StatusNotFound, gin.H{"status": false, "message": "not found"})
			return
		}
		data, err := storage.ReadFile(key)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"status": false, "message": "not found"})
			return
		}
		contentType := http.DetectContentType(data)
		ctx.Header("Cache-Control", "public, max-age=86400")
		ctx.Data(http.StatusOK, contentType, data)
	})
}
