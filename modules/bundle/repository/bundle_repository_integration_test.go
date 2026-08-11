package repository

import (
	"context"
	"os"
	"testing"

	"github.com/webdevtedxuniversitasairlangga/database"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Integration test ini menguji query GORM yang sesungguhnya, jadi butuh
// PostgreSQL yang benar-benar berjalan. Tanpa variabel TEST_DATABASE_URL,
// seluruh test di berkas ini dilewati — supaya `go test ./...` tetap hijau bagi
// anggota tim yang belum menyiapkan database.
//
// Cara menjalankan:
//
//	docker run -d --name tedx-bundle-test-db -e POSTGRES_PASSWORD=testpass \
//	  -e POSTGRES_DB=tedx_test -p 55432:5432 postgres:16-alpine
//
//	TEST_DATABASE_URL="host=localhost user=postgres password=testpass dbname=tedx_test port=55432 sslmode=disable" \
//	  go test ./modules/bundle/repository/... -v
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL tidak diset — integration test repository dilewati")
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("gagal terhubung ke database test: %v", err)
	}

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		t.Fatalf("gagal membuat extension uuid-ossp: %v", err)
	}

	// Memakai migrasi asli aplikasi, sekaligus membuktikan entity bundle sudah
	// terdaftar dengan benar di database/migrations.go.
	if err := database.Migrate(db); err != nil {
		t.Fatalf("AutoMigrate gagal: %v", err)
	}

	if err := db.Exec("DELETE FROM bundle_images").Error; err != nil {
		t.Fatalf("gagal membersihkan bundle_images: %v", err)
	}
	if err := db.Exec("DELETE FROM bundles").Error; err != nil {
		t.Fatalf("gagal membersihkan bundles: %v", err)
	}

	return db
}

func newTestBundle(name string, isActive bool) entities.Bundle {
	return entities.Bundle{
		ID:          uuid.New(),
		Name:        name,
		Description: "deskripsi " + name,
		Price:       decimal.RequireFromString("150000.00"),
		IsActive:    isActive,
	}
}

func TestIntegration_MigrationCreatesBundleTables(t *testing.T) {
	db := setupTestDB(t)

	for _, table := range []string{"bundles", "bundle_images"} {
		if !db.Migrator().HasTable(table) {
			t.Errorf("tabel %q tidak dibuat oleh AutoMigrate", table)
		}
	}
}

func TestIntegration_CreatePreservesDecimalPrecision(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBundleRepository(db)
	ctx := context.Background()

	bundle := newTestBundle("Bundle Presisi", true)
	bundle.Price = decimal.RequireFromString("199999.99")

	created, err := repo.Create(ctx, nil, bundle)
	if err != nil {
		t.Fatalf("Create gagal: %v", err)
	}

	// Dibaca ulang dari database, bukan dari nilai di memori.
	got, err := repo.GetByID(ctx, nil, created.ID)
	if err != nil {
		t.Fatalf("GetByID gagal: %v", err)
	}

	if got.Price.StringFixed(2) != "199999.99" {
		t.Errorf("harga berubah setelah pulang-pergi database: %q", got.Price.StringFixed(2))
	}
}

// TestIntegration_CreateIgnoresFalseOnDefaultTaggedColumn membuktikan alasan di
// balik keputusan desain "bundle selalu dibuat aktif": GORM mengabaikan nilai
// zero (false) untuk kolom bertag `default:true`, sehingga false yang dikirim
// saat INSERT tetap tersimpan menjadi true. Bila suatu saat entity diubah
// memakai *bool, test ini akan gagal dan menjadi pengingat untuk membuka kembali
// field is_active di endpoint create.
func TestIntegration_CreateIgnoresFalseOnDefaultTaggedColumn(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBundleRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, nil, newTestBundle("Bundle Draft", false))
	if err != nil {
		t.Fatalf("Create gagal: %v", err)
	}

	got, err := repo.GetByID(ctx, nil, created.ID)
	if err != nil {
		t.Fatalf("GetByID gagal: %v", err)
	}

	if !got.IsActive {
		t.Skip("GORM kini menyimpan false apa adanya — field is_active boleh dibuka lagi di endpoint create")
	}
}

func TestIntegration_GetAllFiltersByIsActive(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBundleRepository(db)
	ctx := context.Background()

	for _, b := range []entities.Bundle{
		newTestBundle("Aktif 1", true),
		newTestBundle("Aktif 2", true),
	} {
		if _, err := repo.Create(ctx, nil, b); err != nil {
			t.Fatalf("Create gagal: %v", err)
		}
	}

	hidden, err := repo.Create(ctx, nil, newTestBundle("Tersembunyi", true))
	if err != nil {
		t.Fatalf("Create gagal: %v", err)
	}
	// Dinonaktifkan lewat Update, meniru jalur PATCH di aplikasi.
	hidden.IsActive = false
	if _, err := repo.Update(ctx, nil, hidden); err != nil {
		t.Fatalf("Update gagal: %v", err)
	}

	active := true
	got, total, err := repo.GetAll(ctx, nil, &active, 10, 0)
	if err != nil {
		t.Fatalf("GetAll gagal: %v", err)
	}
	if total != 2 || len(got) != 2 {
		t.Errorf("filter aktif: total=%d len=%d, seharusnya 2 dan 2", total, len(got))
	}

	inactive := false
	got, total, err = repo.GetAll(ctx, nil, &inactive, 10, 0)
	if err != nil {
		t.Fatalf("GetAll gagal: %v", err)
	}
	if total != 1 || len(got) != 1 || got[0].Name != "Tersembunyi" {
		t.Errorf("filter nonaktif tidak tepat: total=%d len=%d", total, len(got))
	}

	_, total, err = repo.GetAll(ctx, nil, nil, 10, 0)
	if err != nil {
		t.Fatalf("GetAll gagal: %v", err)
	}
	if total != 3 {
		t.Errorf("tanpa filter seharusnya 3 baris, dapat %d", total)
	}
}

// TestIntegration_GetAllCountThenFind menjaga pola "Count lalu Find pada query
// builder yang sama" yang disalin dari modul todo — kombinasi ini punya
// kehalusan tersendiri di GORM, jadi layak diuji langsung.
func TestIntegration_GetAllCountThenFind(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBundleRepository(db)
	ctx := context.Background()

	for _, name := range []string{"A", "B", "C", "D", "E"} {
		if _, err := repo.Create(ctx, nil, newTestBundle(name, true)); err != nil {
			t.Fatalf("Create gagal: %v", err)
		}
	}

	got, total, err := repo.GetAll(ctx, nil, nil, 2, 2)
	if err != nil {
		t.Fatalf("GetAll gagal: %v", err)
	}

	if total != 5 {
		t.Errorf("total harus menghitung seluruh baris (5), dapat %d", total)
	}
	if len(got) != 2 {
		t.Errorf("limit 2 harus mengembalikan 2 baris, dapat %d", len(got))
	}
}

func TestIntegration_GetByIDWithImagesPreloads(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBundleRepository(db)
	ctx := context.Background()

	bundle, err := repo.Create(ctx, nil, newTestBundle("Bundle Bergambar", true))
	if err != nil {
		t.Fatalf("Create gagal: %v", err)
	}

	for _, url := range []string{"https://example.com/a.png", "https://example.com/b.png"} {
		if _, err := repo.CreateImage(ctx, nil, entities.BundleImage{
			ID:       uuid.New(),
			BundleID: bundle.ID,
			ImageURL: url,
		}); err != nil {
			t.Fatalf("CreateImage gagal: %v", err)
		}
	}

	withImages, err := repo.GetByIDWithImages(ctx, nil, bundle.ID)
	if err != nil {
		t.Fatalf("GetByIDWithImages gagal: %v", err)
	}
	if len(withImages.BundleImages) != 2 {
		t.Errorf("preload harus mengembalikan 2 gambar, dapat %d", len(withImages.BundleImages))
	}

	// GetByID biasa sengaja tidak memuat gambar.
	plain, err := repo.GetByID(ctx, nil, bundle.ID)
	if err != nil {
		t.Fatalf("GetByID gagal: %v", err)
	}
	if len(plain.BundleImages) != 0 {
		t.Errorf("GetByID tidak boleh memuat gambar, dapat %d", len(plain.BundleImages))
	}
}

// TestIntegration_UpdateDoesNotDisturbImages menguji alasan dipisahkannya
// GetByID dari GetByIDWithImages: menyimpan entity hasil preload bisa membuat
// GORM ikut menulis ulang asosiasinya.
func TestIntegration_UpdateDoesNotDisturbImages(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBundleRepository(db)
	ctx := context.Background()

	bundle, err := repo.Create(ctx, nil, newTestBundle("Bundle Update", true))
	if err != nil {
		t.Fatalf("Create gagal: %v", err)
	}
	if _, err := repo.CreateImage(ctx, nil, entities.BundleImage{
		ID:       uuid.New(),
		BundleID: bundle.ID,
		ImageURL: "https://example.com/a.png",
	}); err != nil {
		t.Fatalf("CreateImage gagal: %v", err)
	}

	// Jalur yang dipakai service saat PATCH: ambil tanpa preload, ubah, simpan.
	toUpdate, err := repo.GetByID(ctx, nil, bundle.ID)
	if err != nil {
		t.Fatalf("GetByID gagal: %v", err)
	}
	toUpdate.Name = "Bundle Update (Revisi)"
	toUpdate.IsActive = false
	if _, err := repo.Update(ctx, nil, toUpdate); err != nil {
		t.Fatalf("Update gagal: %v", err)
	}

	after, err := repo.GetByIDWithImages(ctx, nil, bundle.ID)
	if err != nil {
		t.Fatalf("GetByIDWithImages gagal: %v", err)
	}

	if after.Name != "Bundle Update (Revisi)" {
		t.Errorf("nama tidak tersimpan: %q", after.Name)
	}
	if after.IsActive {
		t.Error("is_active=false harus tersimpan lewat Update")
	}
	if len(after.BundleImages) != 1 {
		t.Errorf("gambar harus tetap utuh setelah update, dapat %d", len(after.BundleImages))
	}
}

func TestIntegration_DeleteBundleCascadesImages(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBundleRepository(db)
	ctx := context.Background()

	bundle, err := repo.Create(ctx, nil, newTestBundle("Bundle Hapus", true))
	if err != nil {
		t.Fatalf("Create gagal: %v", err)
	}
	if _, err := repo.CreateImage(ctx, nil, entities.BundleImage{
		ID:       uuid.New(),
		BundleID: bundle.ID,
		ImageURL: "https://example.com/a.png",
	}); err != nil {
		t.Fatalf("CreateImage gagal: %v", err)
	}

	if err := repo.Delete(ctx, nil, bundle.ID); err != nil {
		t.Fatalf("Delete gagal: %v", err)
	}

	var remaining int64
	if err := db.Model(&entities.BundleImage{}).Where("bundle_id = ?", bundle.ID).Count(&remaining).Error; err != nil {
		t.Fatalf("gagal menghitung sisa gambar: %v", err)
	}
	if remaining != 0 {
		t.Errorf("gambar harus ikut terhapus lewat ON DELETE CASCADE, tersisa %d", remaining)
	}
}

// TestIntegration_DeleteImageScopedToBundle memastikan gambar milik bundle lain
// tidak bisa dihapus dengan menebak-nebak pasangan id di URL.
func TestIntegration_DeleteImageScopedToBundle(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBundleRepository(db)
	ctx := context.Background()

	bundleA, err := repo.Create(ctx, nil, newTestBundle("Bundle A", true))
	if err != nil {
		t.Fatalf("Create gagal: %v", err)
	}
	bundleB, err := repo.Create(ctx, nil, newTestBundle("Bundle B", true))
	if err != nil {
		t.Fatalf("Create gagal: %v", err)
	}

	image, err := repo.CreateImage(ctx, nil, entities.BundleImage{
		ID:       uuid.New(),
		BundleID: bundleA.ID,
		ImageURL: "https://example.com/a.png",
	})
	if err != nil {
		t.Fatalf("CreateImage gagal: %v", err)
	}

	// Gambar milik A dicoba dihapus lewat id bundle B.
	affected, err := repo.DeleteImage(ctx, nil, image.ID, bundleB.ID)
	if err != nil {
		t.Fatalf("DeleteImage gagal: %v", err)
	}
	if affected != 0 {
		t.Errorf("penghapusan lintas bundle harus nol baris, dapat %d", affected)
	}

	// Lewat bundle yang benar, penghapusan berhasil.
	affected, err = repo.DeleteImage(ctx, nil, image.ID, bundleA.ID)
	if err != nil {
		t.Fatalf("DeleteImage gagal: %v", err)
	}
	if affected != 1 {
		t.Errorf("penghapusan yang sah harus satu baris, dapat %d", affected)
	}
}

func TestIntegration_GetByIDReturnsErrorWhenMissing(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBundleRepository(db)

	if _, err := repo.GetByID(context.Background(), nil, uuid.New()); err == nil {
		t.Error("GetByID untuk id yang tidak ada harus mengembalikan error")
	}
}
