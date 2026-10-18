// Package storage — upload bukti bayar order ke Cloudinary.
// Kredensial dibaca dari env CLOUDINARY_URL
// (format: cloudinary://<api_key>:<api_secret>@<cloud_name>).
package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
)

// UploadProof — upload bukti bayar (validasi ekstensi/ukuran dilakukan
// handler) ke folder "tedx/proofs" Cloudinary; balik secure URL https.
func UploadProof(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {
	if strings.TrimSpace(os.Getenv("CLOUDINARY_URL")) == "" {
		return "", fmt.Errorf("CLOUDINARY_URL not set")
	}
	cld, err := cloudinary.New()
	if err != nil {
		return "", fmt.Errorf("cloudinary: %w", err)
	}
	resp, err := cld.Upload.Upload(ctx, fileHeader, uploader.UploadParams{
		Folder:   "tedx/proofs",
		PublicID: uuid.New().String(),
	})
	if err != nil {
		return "", fmt.Errorf("cloudinary upload: %w", err)
	}
	if resp.SecureURL == "" {
		return "", fmt.Errorf("cloudinary: empty secure url")
	}
	return resp.SecureURL, nil
}
