// Package storage — klien S3 minimal berbasis stdlib untuk MinIO.
// Mendukung PUT object, GET object, dan ensure bucket dengan auth SigV4.
// Sengaja tanpa dependensi baru (lingkungan build offline).
package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Client — koneksi ke 1 bucket S3-compatible (path-style, cocok untuk MinIO).
type Client struct {
	Endpoint   string // host:port tanpa skema, mis. "minio:9000"
	AccessKey  string
	SecretKey  string
	Bucket     string
	UseSSL     bool
	Region     string
	HTTPClient *http.Client
}

// NewFromEnv — baca MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY,
// MINIO_BUCKET (default tedx-assets), MINIO_REGION (default us-east-1),
// MINIO_USE_SSL=true bila https.
// ok=false bila endpoint/key kosong — caller pakai fallback lain (mis. ImageKit).
func NewFromEnv() (c *Client, ok bool) {
	endpoint := strings.TrimSpace(os.Getenv("MINIO_ENDPOINT"))
	access := strings.TrimSpace(os.Getenv("MINIO_ACCESS_KEY"))
	secret := os.Getenv("MINIO_SECRET_KEY")
	if endpoint == "" || access == "" || secret == "" {
		return nil, false
	}
	bucket := strings.TrimSpace(os.Getenv("MINIO_BUCKET"))
	if bucket == "" {
		bucket = "tedx-assets"
	}
	region := strings.TrimSpace(os.Getenv("MINIO_REGION"))
	if region == "" {
		region = "us-east-1"
	}
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
	return &Client{
		Endpoint:   endpoint,
		AccessKey:  access,
		SecretKey:  secret,
		Bucket:     bucket,
		UseSSL:     strings.EqualFold(strings.TrimSpace(os.Getenv("MINIO_USE_SSL")), "true"),
		Region:     region,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}, true
}

func (c *Client) scheme() string {
	if c.UseSSL {
		return "https"
	}
	return "http"
}

// encodePath — encode tiap segmen path ala S3 (pertahankan '/').
func encodePath(p string) string {
	parts := strings.Split(strings.TrimPrefix(p, "/"), "/")
	for i, s := range parts {
		parts[i] = url.PathEscape(s)
	}
	return "/" + strings.Join(parts, "/")
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// sign — tempel header Auth SigV4. canonicalURI = path yang sudah di-encode,
// payloadHash = hex sha256 body (body kosong untuk GET/PUT tanpa body).
func (c *Client) sign(req *http.Request, canonicalURI, payloadHash string) {
	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	headers := map[string]string{
		"host":                 req.URL.Host,
		"x-amz-content-sha256": payloadHash,
		"x-amz-date":           amzDate,
	}
	if ct := strings.TrimSpace(req.Header.Get("Content-Type")); ct != "" {
		headers["content-type"] = ct
	}
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var canonHeaders strings.Builder
	for _, k := range keys {
		canonHeaders.WriteString(k + ":" + headers[k] + "\n")
	}
	signedHeaders := strings.Join(keys, ";")

	canonical := req.Method + "\n" + canonicalURI + "\n\n" +
		canonHeaders.String() + "\n" + signedHeaders + "\n" + payloadHash
	scope := dateStamp + "/" + c.Region + "/s3/aws4_request"
	toSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + sha256Hex([]byte(canonical))

	kDate := hmacSHA256([]byte("AWS4"+c.SecretKey), dateStamp)
	kRegion := hmacSHA256(kDate, c.Region)
	kService := hmacSHA256(kRegion, "s3")
	kSign := hmacSHA256(kService, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(kSign, toSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		c.AccessKey, scope, signedHeaders, signature))
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
}

// request — bangun, sign, dan kirim request. path sudah termasuk "/bucket/..." ter-encode.
func (c *Client) request(ctx context.Context, method, path string, body []byte, contentType string) (*http.Response, error) {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader([]byte{})
	} else {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.scheme()+"://"+c.Endpoint+path, reader)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	payloadHash := sha256Hex(body)
	c.sign(req, path, payloadHash)
	return c.HTTPClient.Do(req)
}

func drainAndClose(resp *http.Response) string {
	defer resp.Body.Close()
	snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return strings.TrimSpace(string(snippet))
}

// PublicObjectURL — URL publik untuk key (butuh MINIO_PUBLIC_URL, mis.
// "https://img.tedxunair.com"). ok=false bila env kosong.
func PublicObjectURL(key string) (string, bool) {
	base := strings.TrimSuffix(strings.TrimSpace(os.Getenv("MINIO_PUBLIC_URL")), "/")
	if base == "" {
		return "", false
	}
	bucket := strings.TrimSpace(os.Getenv("MINIO_BUCKET"))
	if bucket == "" {
		bucket = "tedx-assets"
	}
	return base + "/" + bucket + encodePath(key), true
}

// imageContentType — mapping ekstensi yang diizinkan ke content-type.
// ok=false untuk ekstensi lain.
func imageContentType(ext string) (string, bool) {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg", true
	case ".png":
		return "image/png", true
	case ".webp":
		return "image/webp", true
	}
	return "", false
}

// UploadPublicImage — validasi + simpan file upload ke "public/<prefix>/..."
// dan kembalikan URL publiknya. Butuh MINIO_* lengkap + MINIO_PUBLIC_URL.
func UploadPublicImage(ctx context.Context, fileHeader *multipart.FileHeader, prefix string) (string, error) {
	if fileHeader.Size > maxPublicImageBytes {
		return "", fmt.Errorf("file too large, max 5MB")
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	contentType, ok := imageContentType(ext)
	if !ok {
		return "", fmt.Errorf("only jpg, jpeg, png, webp allowed")
	}
	sc, ok := NewFromEnv()
	if !ok {
		return "", fmt.Errorf("storage not configured")
	}
	if strings.TrimSpace(os.Getenv("MINIO_PUBLIC_URL")) == "" {
		return "", fmt.Errorf("MINIO_PUBLIC_URL not set")
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
	if err := sc.EnsureBucket(ctx); err != nil {
		return "", err
	}
	if err := sc.Put(ctx, key, data, contentType); err != nil {
		return "", err
	}
	publicURL, _ := PublicObjectURL(key)
	return publicURL, nil
}

// EnsureBucket — buat bucket bila belum ada. 200/409 = sudah ada (milik sendiri).
func (c *Client) EnsureBucket(ctx context.Context) error {
	resp, err := c.request(ctx, http.MethodPut, "/"+c.Bucket, nil, "")
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return nil
	}
	return fmt.Errorf("minio ensure bucket %s: %s", c.Bucket, drainAndClose(resp))
}

// Put — simpan object (timpa bila key sama).
func (c *Client) Put(ctx context.Context, key string, data []byte, contentType string) error {
	resp, err := c.request(ctx, http.MethodPut, "/"+c.Bucket+encodePath(key), data, contentType)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("minio put %s: %s", key, drainAndClose(resp))
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return nil
}

// Get — ambil object beserta content-type simpanan.
func (c *Client) Get(ctx context.Context, key string) (data []byte, contentType string, err error) {
	resp, err := c.request(ctx, http.MethodGet, "/"+c.Bucket+encodePath(key), nil, "")
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, "", fmt.Errorf("minio get %s: %s", key, strings.TrimSpace(string(snippet)))
	}
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return data, resp.Header.Get("Content-Type"), nil
}
