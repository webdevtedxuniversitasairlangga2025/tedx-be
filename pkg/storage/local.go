package storage

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Dir — direktori root penyimpanan lokal (env UPLOAD_DIR, default /app/uploads).
func Dir() string {
	if d := strings.TrimSpace(os.Getenv("UPLOAD_DIR")); d != "" {
		return d
	}
	return "/app/uploads"
}

// FilePublicBase — base URL publik untuk file (env FILE_PUBLIC_URL,
// mis. "http://localhost:8888/api/v1"). ok=false bila kosong.
func FilePublicBase() (string, bool) {
	base := strings.TrimSuffix(strings.TrimSpace(os.Getenv("FILE_PUBLIC_URL")), "/")
	if base == "" {
		return "", false
	}
	return base, true
}

// cleanKey — tolak path traversal/absolut; kembalikan key relatif bersih.
func cleanKey(key string) (string, error) {
	k := filepath.Clean("/" + strings.TrimSpace(key))
	k = strings.TrimPrefix(k, "/")
	if k == "" || k == "." || strings.HasPrefix(k, "..") || strings.Contains(k, "../") {
		return "", fmt.Errorf("invalid storage key")
	}
	return k, nil
}

// SaveFile — tulis data ke <dir>/<key> (buat folder bila perlu).
func SaveFile(key string, data []byte) error {
	k, err := cleanKey(key)
	if err != nil {
		return err
	}
	full := filepath.Join(Dir(), k)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}

// ReadFile — baca <dir>/<key>.
func ReadFile(key string) ([]byte, error) {
	k, err := cleanKey(key)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(Dir(), k))
}

// maxPublicImageBytes — batas upload gambar katalog/galeri.
const maxPublicImageBytes = 5 << 20

// SavePublicFile — validasi (jpg/png/webp ≤5MB) + simpan ke "public/<prefix>/..."
// di disk lokal. Balik key untuk dibentuk URL publik "…/files/<key>".
func SavePublicFile(fileHeader *multipart.FileHeader, prefix string) (string, error) {
	if fileHeader.Size > maxPublicImageBytes {
		return "", fmt.Errorf("file too large, max 5MB")
	}
	var ext string
	switch strings.ToLower(filepath.Ext(fileHeader.Filename)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		ext = strings.ToLower(filepath.Ext(fileHeader.Filename))
	default:
		return "", fmt.Errorf("only jpg, jpeg, png, webp allowed")
	}
	f, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxPublicImageBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(data)) > maxPublicImageBytes {
		return "", fmt.Errorf("file too large, max 5MB")
	}
	key := fmt.Sprintf("public/%s/%s%s", strings.Trim(prefix, "/"), uuid.New().String(), ext)
	if err := SaveFile(key, data); err != nil {
		return "", err
	}
	return key, nil
}

// DeleteFile — hapus <dir>/<key>; tidak ada = sukses (idempoten).
func DeleteFile(key string) error {
	k, err := cleanKey(key)
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(Dir(), k)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// KeyFromPublicURL — ambil key dari URL publik ".../files/<key>".
// ok=false untuk URL luar (ImageKit/unsplash/dll) — jangan dihapus.
func KeyFromPublicURL(rawURL string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Path == "" {
		return "", false
	}
	idx := strings.Index(u.Path, "/files/")
	if idx < 0 {
		return "", false
	}
	key := strings.TrimPrefix(u.Path[idx:], "/files/")
	if !strings.HasPrefix(key, "public/") {
		return "", false
	}
	if _, err := cleanKey(key); err != nil {
		return "", false
	}
	return key, true
}

// RemovePublicURL — hapus file untuk URL publik, best-effort (log saja bila gagal).
// URL luar/ilegal diabaikan.
func RemovePublicURL(rawURL string) {
	key, ok := KeyFromPublicURL(rawURL)
	if !ok {
		return
	}
	if err := DeleteFile(key); err != nil {
		log.Printf("storage: gagal hapus file %s: %v", key, err)
	}
}
