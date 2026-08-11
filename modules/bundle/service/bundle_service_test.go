package service

import (
	"context"
	"errors"
	"testing"

	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/bundle/dto"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// fakeBundleRepository adalah pengganti repository asli supaya logika service
// bisa diuji tanpa koneksi database.
type fakeBundleRepository struct {
	// nilai yang dikembalikan
	bundle      entities.Bundle
	bundleErr   error
	bundles     []entities.Bundle
	total       int64
	getAllErr   error
	rowsDeleted int64

	// rekaman argumen yang diterima, untuk diperiksa di test
	gotIsActive      *bool
	gotLimit         int
	gotOffset        int
	gotCreated       entities.Bundle
	gotCreatedImage  entities.BundleImage
	gotUpdated       entities.Bundle
	gotDeletedID     uuid.UUID
	gotDeletedImgID  uuid.UUID
	gotDeletedBundle uuid.UUID

	updateCalled bool
}

func (f *fakeBundleRepository) Create(_ context.Context, _ *gorm.DB, bundle entities.Bundle) (entities.Bundle, error) {
	f.gotCreated = bundle
	return bundle, nil
}

func (f *fakeBundleRepository) GetAll(_ context.Context, _ *gorm.DB, isActive *bool, limit, offset int) ([]entities.Bundle, int64, error) {
	f.gotIsActive = isActive
	f.gotLimit = limit
	f.gotOffset = offset
	if f.getAllErr != nil {
		return nil, 0, f.getAllErr
	}
	return f.bundles, f.total, nil
}

func (f *fakeBundleRepository) GetByID(_ context.Context, _ *gorm.DB, _ uuid.UUID) (entities.Bundle, error) {
	if f.bundleErr != nil {
		return entities.Bundle{}, f.bundleErr
	}
	return f.bundle, nil
}

func (f *fakeBundleRepository) GetByIDWithImages(_ context.Context, _ *gorm.DB, _ uuid.UUID) (entities.Bundle, error) {
	if f.bundleErr != nil {
		return entities.Bundle{}, f.bundleErr
	}
	return f.bundle, nil
}

func (f *fakeBundleRepository) Update(_ context.Context, _ *gorm.DB, bundle entities.Bundle) (entities.Bundle, error) {
	f.updateCalled = true
	f.gotUpdated = bundle
	return bundle, nil
}

func (f *fakeBundleRepository) Delete(_ context.Context, _ *gorm.DB, id uuid.UUID) error {
	f.gotDeletedID = id
	return nil
}

func (f *fakeBundleRepository) CreateImage(_ context.Context, _ *gorm.DB, image entities.BundleImage) (entities.BundleImage, error) {
	f.gotCreatedImage = image
	return image, nil
}

func (f *fakeBundleRepository) DeleteImage(_ context.Context, _ *gorm.DB, imageID, bundleID uuid.UUID) (int64, error) {
	f.gotDeletedImgID = imageID
	f.gotDeletedBundle = bundleID
	return f.rowsDeleted, nil
}

func newServiceWithFake(repo *fakeBundleRepository) BundleService {
	return NewBundleService(repo, nil)
}

func sampleBundle() entities.Bundle {
	return entities.Bundle{
		ID:          uuid.New(),
		Name:        "Bundle Hemat",
		Description: "Kaos + totebag",
		Price:       decimal.RequireFromString("150000"),
		IsActive:    true,
	}
}

func TestCreate_Success(t *testing.T) {
	repo := &fakeBundleRepository{}
	svc := newServiceWithFake(repo)

	result, err := svc.Create(context.Background(), dto.BundleCreateRequest{
		Name:        "Bundle Hemat",
		Description: "Kaos + totebag",
		Price:       "150000.5",
	})
	if err != nil {
		t.Fatalf("Create mengembalikan error tak terduga: %v", err)
	}

	if result.Price != "150000.50" {
		t.Errorf("harga harus diformat dua desimal, dapat %q", result.Price)
	}
	if !result.IsActive {
		t.Error("bundle baru harus aktif secara default")
	}
	if repo.gotCreated.ID == uuid.Nil {
		t.Error("service harus menetapkan ID sebelum menyimpan")
	}
}

func TestCreate_InvalidPrice(t *testing.T) {
	svc := newServiceWithFake(&fakeBundleRepository{})

	_, err := svc.Create(context.Background(), dto.BundleCreateRequest{
		Name:        "Bundle Hemat",
		Description: "Kaos + totebag",
		Price:       "seratus ribu",
	})
	if !errors.Is(err, dto.ErrInvalidPrice) {
		t.Fatalf("harus ErrInvalidPrice, dapat %v", err)
	}
}

func TestCreate_PriceOutOfRange(t *testing.T) {
	svc := newServiceWithFake(&fakeBundleRepository{})

	for _, price := range []string{"-1", "100000000"} {
		_, err := svc.Create(context.Background(), dto.BundleCreateRequest{
			Name:        "Bundle Hemat",
			Description: "Kaos + totebag",
			Price:       price,
		})
		if !errors.Is(err, dto.ErrPriceOutOfRange) {
			t.Errorf("harga %q harus ErrPriceOutOfRange, dapat %v", price, err)
		}
	}
}

func TestGetAll_DefaultsToActiveOnly(t *testing.T) {
	repo := &fakeBundleRepository{}
	svc := newServiceWithFake(repo)

	// Filter tidak dikirim: endpoint publik hanya boleh menampilkan bundle aktif.
	if _, err := svc.GetAll(context.Background(), dto.BundleQueryRequest{}); err != nil {
		t.Fatalf("GetAll mengembalikan error tak terduga: %v", err)
	}

	if repo.gotIsActive == nil || !*repo.gotIsActive {
		t.Error("tanpa parameter is_active, repository harus dipanggil dengan filter aktif")
	}
	if repo.gotLimit != 10 || repo.gotOffset != 0 {
		t.Errorf("pagination default salah: limit=%d offset=%d", repo.gotLimit, repo.gotOffset)
	}
}

func TestGetAll_ExplicitInactiveFilter(t *testing.T) {
	repo := &fakeBundleRepository{}
	svc := newServiceWithFake(repo)

	inactive := false
	if _, err := svc.GetAll(context.Background(), dto.BundleQueryRequest{IsActive: &inactive}); err != nil {
		t.Fatalf("GetAll mengembalikan error tak terduga: %v", err)
	}

	if repo.gotIsActive == nil || *repo.gotIsActive {
		t.Error("is_active=false harus diteruskan apa adanya ke repository")
	}
}

func TestGetAll_PaginationMeta(t *testing.T) {
	repo := &fakeBundleRepository{
		bundles: []entities.Bundle{sampleBundle()},
		total:   25,
	}
	svc := newServiceWithFake(repo)

	result, err := svc.GetAll(context.Background(), dto.BundleQueryRequest{Page: 2, PerPage: 10})
	if err != nil {
		t.Fatalf("GetAll mengembalikan error tak terduga: %v", err)
	}

	if result.Meta.MaxPage != 3 {
		t.Errorf("max_page untuk 25 data @10 harus 3, dapat %d", result.Meta.MaxPage)
	}
	if result.Meta.Total != 25 {
		t.Errorf("total harus 25, dapat %d", result.Meta.Total)
	}
	if repo.gotOffset != 10 {
		t.Errorf("offset halaman 2 harus 10, dapat %d", repo.gotOffset)
	}
}

func TestGetByID_InvalidUUID(t *testing.T) {
	svc := newServiceWithFake(&fakeBundleRepository{})

	_, err := svc.GetByID(context.Background(), "bukan-uuid")
	if !errors.Is(err, dto.ErrBundleNotFound) {
		t.Fatalf("harus ErrBundleNotFound, dapat %v", err)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := newServiceWithFake(&fakeBundleRepository{bundleErr: gorm.ErrRecordNotFound})

	_, err := svc.GetByID(context.Background(), uuid.New().String())
	if !errors.Is(err, dto.ErrBundleNotFound) {
		t.Fatalf("harus ErrBundleNotFound, dapat %v", err)
	}
}

func TestGetByID_IncludesImages(t *testing.T) {
	bundle := sampleBundle()
	bundle.BundleImages = []entities.BundleImage{
		{ID: uuid.New(), BundleID: bundle.ID, ImageURL: "https://example.com/a.jpg"},
	}
	svc := newServiceWithFake(&fakeBundleRepository{bundle: bundle})

	result, err := svc.GetByID(context.Background(), bundle.ID.String())
	if err != nil {
		t.Fatalf("GetByID mengembalikan error tak terduga: %v", err)
	}

	if len(result.Images) != 1 || result.Images[0].ImageURL != "https://example.com/a.jpg" {
		t.Errorf("detail harus membawa gambarnya, dapat %+v", result.Images)
	}
}

func TestUpdate_OnlyChangesFieldsSent(t *testing.T) {
	repo := &fakeBundleRepository{bundle: sampleBundle()}
	svc := newServiceWithFake(repo)

	newName := "Bundle Spesial"
	result, err := svc.Update(context.Background(), repo.bundle.ID.String(), dto.BundleUpdateRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("Update mengembalikan error tak terduga: %v", err)
	}

	if result.Name != newName {
		t.Errorf("nama harus berubah jadi %q, dapat %q", newName, result.Name)
	}
	if result.Description != "Kaos + totebag" {
		t.Errorf("deskripsi yang tidak dikirim tidak boleh berubah, dapat %q", result.Description)
	}
	if result.Price != "150000.00" {
		t.Errorf("harga yang tidak dikirim tidak boleh berubah, dapat %q", result.Price)
	}
}

func TestUpdate_CanDeactivate(t *testing.T) {
	repo := &fakeBundleRepository{bundle: sampleBundle()}
	svc := newServiceWithFake(repo)

	inactive := false
	result, err := svc.Update(context.Background(), repo.bundle.ID.String(), dto.BundleUpdateRequest{
		IsActive: &inactive,
	})
	if err != nil {
		t.Fatalf("Update mengembalikan error tak terduga: %v", err)
	}

	if result.IsActive {
		t.Error("is_active=false harus tersimpan (jalur menonaktifkan bundle)")
	}
}

func TestUpdate_InvalidPriceTidakMenyentuhRepository(t *testing.T) {
	repo := &fakeBundleRepository{bundle: sampleBundle()}
	svc := newServiceWithFake(repo)

	badPrice := "gratis"
	_, err := svc.Update(context.Background(), repo.bundle.ID.String(), dto.BundleUpdateRequest{
		Price: &badPrice,
	})
	if !errors.Is(err, dto.ErrInvalidPrice) {
		t.Fatalf("harus ErrInvalidPrice, dapat %v", err)
	}
	if repo.updateCalled {
		t.Error("repository.Update tidak boleh dipanggil saat validasi harga gagal")
	}
}

func TestDelete_NotFound(t *testing.T) {
	svc := newServiceWithFake(&fakeBundleRepository{bundleErr: gorm.ErrRecordNotFound})

	err := svc.Delete(context.Background(), uuid.New().String())
	if !errors.Is(err, dto.ErrBundleNotFound) {
		t.Fatalf("harus ErrBundleNotFound, dapat %v", err)
	}
}

func TestAddImage_BundleNotFound(t *testing.T) {
	svc := newServiceWithFake(&fakeBundleRepository{bundleErr: gorm.ErrRecordNotFound})

	_, err := svc.AddImage(context.Background(), uuid.New().String(), dto.BundleImageCreateRequest{
		ImageURL: "https://example.com/a.jpg",
	})
	if !errors.Is(err, dto.ErrBundleNotFound) {
		t.Fatalf("harus ErrBundleNotFound, dapat %v", err)
	}
}

func TestAddImage_Success(t *testing.T) {
	bundle := sampleBundle()
	repo := &fakeBundleRepository{bundle: bundle}
	svc := newServiceWithFake(repo)

	result, err := svc.AddImage(context.Background(), bundle.ID.String(), dto.BundleImageCreateRequest{
		ImageURL: "https://example.com/a.jpg",
	})
	if err != nil {
		t.Fatalf("AddImage mengembalikan error tak terduga: %v", err)
	}

	if result.ImageURL != "https://example.com/a.jpg" {
		t.Errorf("image_url tidak sesuai, dapat %q", result.ImageURL)
	}
	if repo.gotCreatedImage.BundleID != bundle.ID {
		t.Error("gambar harus terikat ke bundle dari parameter path")
	}
}

func TestDeleteImage_NotFound(t *testing.T) {
	bundle := sampleBundle()
	// rowsDeleted 0 berarti tidak ada baris yang cocok (id gambar salah, atau
	// gambar itu milik bundle lain).
	svc := newServiceWithFake(&fakeBundleRepository{bundle: bundle, rowsDeleted: 0})

	err := svc.DeleteImage(context.Background(), bundle.ID.String(), uuid.New().String())
	if !errors.Is(err, dto.ErrBundleImageNotFound) {
		t.Fatalf("harus ErrBundleImageNotFound, dapat %v", err)
	}
}

func TestDeleteImage_Success(t *testing.T) {
	bundle := sampleBundle()
	repo := &fakeBundleRepository{bundle: bundle, rowsDeleted: 1}
	svc := newServiceWithFake(repo)

	imageID := uuid.New()
	if err := svc.DeleteImage(context.Background(), bundle.ID.String(), imageID.String()); err != nil {
		t.Fatalf("DeleteImage mengembalikan error tak terduga: %v", err)
	}

	if repo.gotDeletedImgID != imageID || repo.gotDeletedBundle != bundle.ID {
		t.Error("penghapusan harus dibatasi pada pasangan image_id + bundle_id")
	}
}
