package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/webdevtedxuniversitasairlangga/modules/bundle/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/bundle/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
	"github.com/gin-gonic/gin"
)

// fakeBundleService merekam argumen yang diterima dari handler, supaya bisa
// dipastikan proses binding request berjalan seperti yang diharapkan.
type fakeBundleService struct {
	gotQuery    dto.BundleQueryRequest
	gotCreate   dto.BundleCreateRequest
	gotUpdate   dto.BundleUpdateRequest
	gotImageReq dto.BundleImageCreateRequest
	gotID       string
	gotBundleID string
	gotImageID  string

	err error
}

func (f *fakeBundleService) Create(_ context.Context, req dto.BundleCreateRequest) (dto.BundleResponse, error) {
	f.gotCreate = req
	if f.err != nil {
		return dto.BundleResponse{}, f.err
	}
	return dto.BundleResponse{ID: "generated-id", Name: req.Name, Price: req.Price}, nil
}

func (f *fakeBundleService) GetAll(_ context.Context, req dto.BundleQueryRequest) (dto.BundlePaginationResponse, error) {
	f.gotQuery = req
	if f.err != nil {
		return dto.BundlePaginationResponse{}, f.err
	}
	return dto.BundlePaginationResponse{Data: []dto.BundleResponse{}}, nil
}

func (f *fakeBundleService) GetByID(_ context.Context, id string) (dto.BundleDetailResponse, error) {
	f.gotID = id
	if f.err != nil {
		return dto.BundleDetailResponse{}, f.err
	}
	return dto.BundleDetailResponse{}, nil
}

func (f *fakeBundleService) Update(_ context.Context, id string, req dto.BundleUpdateRequest) (dto.BundleResponse, error) {
	f.gotID = id
	f.gotUpdate = req
	if f.err != nil {
		return dto.BundleResponse{}, f.err
	}
	return dto.BundleResponse{ID: id}, nil
}

func (f *fakeBundleService) Delete(_ context.Context, id string) error {
	f.gotID = id
	return f.err
}

func (f *fakeBundleService) AddImage(_ context.Context, bundleID string, req dto.BundleImageCreateRequest) (dto.BundleImageResponse, error) {
	f.gotBundleID = bundleID
	f.gotImageReq = req
	if f.err != nil {
		return dto.BundleImageResponse{}, f.err
	}
	return dto.BundleImageResponse{ID: "image-id", ImageURL: req.ImageURL}, nil
}

func (f *fakeBundleService) DeleteImage(_ context.Context, bundleID, imageID string) error {
	f.gotBundleID = bundleID
	f.gotImageID = imageID
	return f.err
}

func newTestRouter(svc service.BundleService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	h := NewBundleHandler(nil, svc)
	router := gin.New()
	router.GET("/bundles", h.GetAll)
	router.GET("/bundles/:id", h.GetByID)
	router.POST("/bundles", h.Create)
	router.PATCH("/bundles/:id", h.Update)
	router.DELETE("/bundles/:id", h.Delete)
	router.POST("/bundles/:id/images", h.AddImage)
	router.DELETE("/bundles/:id/images/:imageId", h.DeleteImage)

	return router
}

func doRequest(router *gin.Engine, method, target, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, target, nil)
	} else {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// TestGetAll_BindsIsActiveFilter adalah test terpenting di berkas ini: seluruh
// fitur filter bergantung pada kemampuan Gin mem-binding *bool dari query string.
func TestGetAll_BindsIsActiveFilter(t *testing.T) {
	cases := []struct {
		query string
		want  *bool
	}{
		{query: "/bundles", want: nil},
		{query: "/bundles?is_active=true", want: boolPtr(true)},
		{query: "/bundles?is_active=false", want: boolPtr(false)},
	}

	for _, tc := range cases {
		svc := &fakeBundleService{}
		rec := doRequest(newTestRouter(svc), http.MethodGet, tc.query, "")

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d, seharusnya 200", tc.query, rec.Code)
		}

		got := svc.gotQuery.IsActive
		switch {
		case tc.want == nil && got != nil:
			t.Errorf("%s: is_active seharusnya nil, dapat %v", tc.query, *got)
		case tc.want != nil && got == nil:
			t.Errorf("%s: is_active seharusnya %v, dapat nil", tc.query, *tc.want)
		case tc.want != nil && got != nil && *tc.want != *got:
			t.Errorf("%s: is_active seharusnya %v, dapat %v", tc.query, *tc.want, *got)
		}
	}
}

func TestGetAll_BindsPagination(t *testing.T) {
	svc := &fakeBundleService{}
	rec := doRequest(newTestRouter(svc), http.MethodGet, "/bundles?page=3&per_page=25", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, seharusnya 200", rec.Code)
	}
	if svc.gotQuery.Page != 3 || svc.gotQuery.PerPage != 25 {
		t.Errorf("pagination tidak ter-binding: page=%d per_page=%d", svc.gotQuery.Page, svc.gotQuery.PerPage)
	}
}

func TestCreate_Success(t *testing.T) {
	svc := &fakeBundleService{}
	body := `{"name":"Bundle A","description":"isi paket","price":"150000.00"}`
	rec := doRequest(newTestRouter(svc), http.MethodPost, "/bundles", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d, seharusnya 201. body: %s", rec.Code, rec.Body.String())
	}
	if svc.gotCreate.Price != "150000.00" {
		t.Errorf("price harus ter-binding sebagai string, dapat %q", svc.gotCreate.Price)
	}

	var res utils.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("response bukan JSON yang valid: %v", err)
	}
	if !res.Status || res.Message != dto.MESSAGE_SUCCESS_CREATE_BUNDLE {
		t.Errorf("envelope tidak sesuai: %+v", res)
	}
}

func TestCreate_MissingRequiredField(t *testing.T) {
	svc := &fakeBundleService{}
	// description dan price sengaja tidak dikirim.
	rec := doRequest(newTestRouter(svc), http.MethodPost, "/bundles", `{"name":"Bundle A"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, seharusnya 400", rec.Code)
	}
	if svc.gotCreate.Name != "" {
		t.Error("service tidak boleh dipanggil ketika validasi body gagal")
	}
}

func TestUpdate_PartialBodyKeepsUnsentFieldsNil(t *testing.T) {
	svc := &fakeBundleService{}
	rec := doRequest(newTestRouter(svc), http.MethodPatch, "/bundles/abc", `{"is_active":false}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, seharusnya 200. body: %s", rec.Code, rec.Body.String())
	}
	if svc.gotUpdate.Name != nil || svc.gotUpdate.Price != nil {
		t.Error("field yang tidak dikirim harus tetap nil")
	}
	if svc.gotUpdate.IsActive == nil || *svc.gotUpdate.IsActive {
		t.Error("is_active=false harus ter-binding sebagai pointer bernilai false")
	}
	if svc.gotID != "abc" {
		t.Errorf("path param id salah: %q", svc.gotID)
	}
}

func TestGetByID_NotFoundReturns404(t *testing.T) {
	svc := &fakeBundleService{err: dto.ErrBundleNotFound}
	rec := doRequest(newTestRouter(svc), http.MethodGet, "/bundles/xyz", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d, seharusnya 404", rec.Code)
	}
}

func TestAddImage_RejectsInvalidURL(t *testing.T) {
	svc := &fakeBundleService{}
	rec := doRequest(newTestRouter(svc), http.MethodPost, "/bundles/abc/images", `{"image_url":"bukan-url"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, seharusnya 400", rec.Code)
	}
	if svc.gotImageReq.ImageURL != "" {
		t.Error("service tidak boleh dipanggil ketika image_url tidak valid")
	}
}

func TestAddImage_Success(t *testing.T) {
	svc := &fakeBundleService{}
	body := `{"image_url":"https://example.com/a.png"}`
	rec := doRequest(newTestRouter(svc), http.MethodPost, "/bundles/abc/images", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d, seharusnya 201. body: %s", rec.Code, rec.Body.String())
	}
	if svc.gotBundleID != "abc" {
		t.Errorf("bundle id dari path salah: %q", svc.gotBundleID)
	}
}

func TestDeleteImage_ExtractsBothPathParams(t *testing.T) {
	svc := &fakeBundleService{}
	rec := doRequest(newTestRouter(svc), http.MethodDelete, "/bundles/bundle-1/images/image-9", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, seharusnya 200", rec.Code)
	}
	if svc.gotBundleID != "bundle-1" || svc.gotImageID != "image-9" {
		t.Errorf("path param salah: bundle=%q image=%q", svc.gotBundleID, svc.gotImageID)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
