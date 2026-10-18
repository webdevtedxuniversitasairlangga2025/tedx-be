package storage

import (
	"context"
	"mime/multipart"
	"strings"
	"testing"
)

// UploadProof tanpa kredensial harus gagal dengan pesan jelas, bukan panic.
func TestUploadProofWithoutCredentials(t *testing.T) {
	t.Setenv("CLOUDINARY_URL", "")
	_, err := UploadProof(context.Background(), &multipart.FileHeader{Filename: "bukti.jpg"})
	if err == nil || !strings.Contains(err.Error(), "CLOUDINARY_URL") {
		t.Fatalf("err = %v, want CLOUDINARY_URL not set", err)
	}
}
