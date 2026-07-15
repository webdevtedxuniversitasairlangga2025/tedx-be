# SOP: Membuat Modul CRUD Baru

Dokumen ini adalah **prosedur baku** untuk membuat modul fitur baru dengan endpoint
CRUD lengkap (GetAll, GetByID, Create, Update, Delete). Ikuti langkah demi langkah.
Modul `todo` adalah **contoh acuan (reference implementation)** — jika ragu, buka file
padanannya di `modules/todo/`.

Sebagai contoh, misalkan kita membuat modul **`event`** dengan field `title` dan `location`.
Ganti `event`/`Event` dengan nama fiturmu di setiap langkah.

---

## Ringkasan langkah

1. [Buat entity](#langkah-1-buat-entity)
2. [Daftarkan entity ke migrations](#langkah-2-daftarkan-entity-ke-migrations)
3. [Buat DTO](#langkah-3-buat-dto)
4. [Buat repository](#langkah-4-buat-repository)
5. [Buat service](#langkah-5-buat-service)
6. [Buat handler](#langkah-6-buat-handler)
7. [Buat routes](#langkah-7-buat-routes)
8. [Wire dependency di provider](#langkah-8-wire-dependency-di-provider)
9. [Daftarkan routes di main](#langkah-9-daftarkan-routes-di-main)
10. [Build, test, dokumentasikan](#langkah-10-build-test-dokumentasikan)

Struktur folder akhir:

```
modules/event/
  dto/event_dto.go
  repository/event_repository.go
  service/event_service.go
  handler/event_handler.go
  routes.go
```

---

## Langkah 1: Buat entity

File: `database/entities/event_entity.go`

Ikuti pola `todo_entity.go`. Semua entity milik-user harus punya `UserID` + relasi
`User` dengan constraint cascade, dan meng-embed `Timestamp`.

```go
package entities

import "github.com/google/uuid"

type Event struct {
	ID       uuid.UUID `gorm:"primaryKey;default:uuid_generate_v4()"`
	UserID   uuid.UUID `gorm:"not null;index"`
	Title    string    `gorm:"size:100;not null"`
	Location string    `gorm:"size:100;not null"`

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE"`

	Timestamp
}
```

> `Timestamp` (di `common.go`) sudah menyediakan `CreatedAt`, `UpdatedAt`, `DeletedAt`.
> Jangan tambahkan field waktu manual.

---

## Langkah 2: Daftarkan entity ke migrations

File: `database/migrations.go`

Tambahkan entity baru ke `AutoMigrate`:

```go
func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entities.User{},
		&entities.RefreshToken{},
		&entities.Todo{},
		&entities.Event{}, // <-- tambahkan
	)
	if err != nil {
		return err
	}
	return nil
}
```

> Tanpa langkah ini, tabel `events` tidak akan dibuat saat aplikasi boot.

---

## Langkah 3: Buat DTO

File: `modules/event/dto/event_dto.go`

DTO memuat: konstanta pesan, error domain, request/response struct, dan struct pagination.
**Salin blok pagination persis seperti di `todo`** agar konsisten.

```go
package dto

import (
	"errors"
	"time"
)

const (
	// Failed
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_FAILED_CREATE_EVENT       = "failed create event"
	MESSAGE_FAILED_GET_LIST_EVENT     = "failed get list event"
	MESSAGE_FAILED_GET_EVENT          = "failed get event"
	MESSAGE_FAILED_UPDATE_EVENT       = "failed update event"
	MESSAGE_FAILED_DELETE_EVENT       = "failed delete event"

	// Success
	MESSAGE_SUCCESS_CREATE_EVENT   = "success create event"
	MESSAGE_SUCCESS_GET_LIST_EVENT = "success get list event"
	MESSAGE_SUCCESS_GET_EVENT      = "success get event"
	MESSAGE_SUCCESS_UPDATE_EVENT   = "success update event"
	MESSAGE_SUCCESS_DELETE_EVENT   = "success delete event"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrInvalidUser   = errors.New("invalid user id")
)

type (
	EventCreateRequest struct {
		Title    string `json:"title" form:"title" binding:"required,min=1,max=100"`
		Location string `json:"location" form:"location" binding:"required,min=1,max=100"`
	}

	// Field pointer agar bisa membedakan "tidak dikirim" vs "dikirim kosong".
	EventUpdateRequest struct {
		Title    *string `json:"title" form:"title" binding:"omitempty,min=1,max=100"`
		Location *string `json:"location" form:"location" binding:"omitempty,min=1,max=100"`
	}

	EventResponse struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Location  string    `json:"location"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	PaginationRequest struct {
		Page    int `form:"page"`
		PerPage int `form:"per_page"`
	}

	PaginationMeta struct {
		Page    int   `json:"page"`
		PerPage int   `json:"per_page"`
		MaxPage int   `json:"max_page"`
		Total   int64 `json:"total"`
	}

	EventPaginationResponse struct {
		Data []EventResponse `json:"data"`
		Meta PaginationMeta  `json:"meta"`
	}
)
```

---

## Langkah 4: Buat repository

File: `modules/event/repository/event_repository.go`

Repository **hanya** berisi query GORM. Perhatikan:
- Setiap method menerima `tx *gorm.DB` dengan fallback `if tx == nil { tx = r.db }`.
- Query milik-user **selalu** difilter `user_id` agar data antar-user terisolasi.
- `GetAllByUserID` mengembalikan total count untuk pagination.

```go
package repository

import (
	"context"

	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	EventRepository interface {
		Create(ctx context.Context, tx *gorm.DB, event entities.Event) (entities.Event, error)
		GetAllByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID, limit, offset int) ([]entities.Event, int64, error)
		GetByID(ctx context.Context, tx *gorm.DB, id, userID uuid.UUID) (entities.Event, error)
		Update(ctx context.Context, tx *gorm.DB, event entities.Event) (entities.Event, error)
		Delete(ctx context.Context, tx *gorm.DB, id, userID uuid.UUID) error
	}

	eventRepository struct {
		db *gorm.DB
	}
)

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(ctx context.Context, tx *gorm.DB, event entities.Event) (entities.Event, error) {
	if tx == nil {
		tx = r.db
	}
	if err := tx.WithContext(ctx).Create(&event).Error; err != nil {
		return entities.Event{}, err
	}
	return event, nil
}

func (r *eventRepository) GetAllByUserID(ctx context.Context, tx *gorm.DB, userID uuid.UUID, limit, offset int) ([]entities.Event, int64, error) {
	if tx == nil {
		tx = r.db
	}

	query := tx.WithContext(ctx).Model(&entities.Event{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var events []entities.Event
	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *eventRepository) GetByID(ctx context.Context, tx *gorm.DB, id, userID uuid.UUID) (entities.Event, error) {
	if tx == nil {
		tx = r.db
	}
	var event entities.Event
	if err := tx.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Take(&event).Error; err != nil {
		return entities.Event{}, err
	}
	return event, nil
}

func (r *eventRepository) Update(ctx context.Context, tx *gorm.DB, event entities.Event) (entities.Event, error) {
	if tx == nil {
		tx = r.db
	}
	if err := tx.WithContext(ctx).Save(&event).Error; err != nil {
		return entities.Event{}, err
	}
	return event, nil
}

func (r *eventRepository) Delete(ctx context.Context, tx *gorm.DB, id, userID uuid.UUID) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&entities.Event{}).Error
}
```

---

## Langkah 5: Buat service

File: `modules/event/service/event_service.go`

Service berisi logika bisnis: parse & validasi `user_id`, hitung pagination, mapping
entity ↔ DTO. Fungsi `toResponse` mengubah entity ke DTO response.

```go
package service

import (
	"context"

	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/event/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/event/repository"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventService interface {
	Create(ctx context.Context, userID string, req dto.EventCreateRequest) (dto.EventResponse, error)
	GetAll(ctx context.Context, userID string, req dto.PaginationRequest) (dto.EventPaginationResponse, error)
	GetByID(ctx context.Context, userID, id string) (dto.EventResponse, error)
	Update(ctx context.Context, userID, id string, req dto.EventUpdateRequest) (dto.EventResponse, error)
	Delete(ctx context.Context, userID, id string) error
}

type eventService struct {
	eventRepository repository.EventRepository
	db              *gorm.DB
}

func NewEventService(eventRepo repository.EventRepository, db *gorm.DB) EventService {
	return &eventService{
		eventRepository: eventRepo,
		db:              db,
	}
}

func toResponse(e entities.Event) dto.EventResponse {
	return dto.EventResponse{
		ID:        e.ID.String(),
		Title:     e.Title,
		Location:  e.Location,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func (s *eventService) Create(ctx context.Context, userID string, req dto.EventCreateRequest) (dto.EventResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.EventResponse{}, dto.ErrInvalidUser
	}

	event := entities.Event{
		ID:       uuid.New(),
		UserID:   uid,
		Title:    req.Title,
		Location: req.Location,
	}

	created, err := s.eventRepository.Create(ctx, s.db, event)
	if err != nil {
		return dto.EventResponse{}, err
	}

	return toResponse(created), nil
}

func (s *eventService) GetAll(ctx context.Context, userID string, req dto.PaginationRequest) (dto.EventPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.EventPaginationResponse{}, dto.ErrInvalidUser
	}

	if req.Page <= 0 {
		req.Page = constants.ENUM_PAGINATION_PAGE
	}
	if req.PerPage <= 0 {
		req.PerPage = constants.ENUM_PAGINATION_PER_PAGE
	}
	offset := (req.Page - 1) * req.PerPage

	events, total, err := s.eventRepository.GetAllByUserID(ctx, s.db, uid, req.PerPage, offset)
	if err != nil {
		return dto.EventPaginationResponse{}, err
	}

	data := make([]dto.EventResponse, 0, len(events))
	for _, e := range events {
		data = append(data, toResponse(e))
	}

	maxPage := int((total + int64(req.PerPage) - 1) / int64(req.PerPage))

	return dto.EventPaginationResponse{
		Data: data,
		Meta: dto.PaginationMeta{
			Page:    req.Page,
			PerPage: req.PerPage,
			MaxPage: maxPage,
			Total:   total,
		},
	}, nil
}

func (s *eventService) GetByID(ctx context.Context, userID, id string) (dto.EventResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.EventResponse{}, dto.ErrInvalidUser
	}

	eid, err := uuid.Parse(id)
	if err != nil {
		return dto.EventResponse{}, dto.ErrEventNotFound
	}

	event, err := s.eventRepository.GetByID(ctx, s.db, eid, uid)
	if err != nil {
		return dto.EventResponse{}, dto.ErrEventNotFound
	}

	return toResponse(event), nil
}

func (s *eventService) Update(ctx context.Context, userID, id string, req dto.EventUpdateRequest) (dto.EventResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.EventResponse{}, dto.ErrInvalidUser
	}

	eid, err := uuid.Parse(id)
	if err != nil {
		return dto.EventResponse{}, dto.ErrEventNotFound
	}

	event, err := s.eventRepository.GetByID(ctx, s.db, eid, uid)
	if err != nil {
		return dto.EventResponse{}, dto.ErrEventNotFound
	}

	// Hanya update field yang dikirim (pola pointer).
	if req.Title != nil {
		event.Title = *req.Title
	}
	if req.Location != nil {
		event.Location = *req.Location
	}

	updated, err := s.eventRepository.Update(ctx, s.db, event)
	if err != nil {
		return dto.EventResponse{}, err
	}

	return toResponse(updated), nil
}

func (s *eventService) Delete(ctx context.Context, userID, id string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.ErrInvalidUser
	}

	eid, err := uuid.Parse(id)
	if err != nil {
		return dto.ErrEventNotFound
	}

	if _, err := s.eventRepository.GetByID(ctx, s.db, eid, uid); err != nil {
		return dto.ErrEventNotFound
	}

	return s.eventRepository.Delete(ctx, s.db, eid, uid)
}
```

---

## Langkah 6: Buat handler

File: `modules/event/handler/event_handler.go`

Handler membaca input, memanggil service, dan membungkus response. **Tidak ada logika
bisnis di sini.** Perhatikan pemetaan error → status HTTP (GetByID pakai `404`, sisanya `400`).

```go
package handler

import (
	"net/http"

	"github.com/webdevtedxuniversitasairlangga/modules/event/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/event/service"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

type (
	EventHandler interface {
		Create(ctx *gin.Context)
		GetAll(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		Update(ctx *gin.Context)
		Delete(ctx *gin.Context)
	}

	eventHandler struct {
		eventService service.EventService
	}
)

func NewEventHandler(injector *do.Injector, es service.EventService) EventHandler {
	return &eventHandler{eventService: es}
}

func (c *eventHandler) Create(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)

	var req dto.EventCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.eventService.Create(ctx.Request.Context(), userID, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_EVENT, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_EVENT, result)
	ctx.JSON(http.StatusCreated, res)
}

func (c *eventHandler) GetAll(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)

	var pagination dto.PaginationRequest
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.eventService.GetAll(ctx.Request.Context(), userID, pagination)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_EVENT, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_EVENT, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *eventHandler) GetByID(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")

	result, err := c.eventService.GetByID(ctx.Request.Context(), userID, id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_EVENT, err.Error(), nil)
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_EVENT, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *eventHandler) Update(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")

	var req dto.EventUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.eventService.Update(ctx.Request.Context(), userID, id, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_EVENT, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_EVENT, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *eventHandler) Delete(ctx *gin.Context) {
	userID := ctx.MustGet("user_id").(string)
	id := ctx.Param("id")

	err := c.eventService.Delete(ctx.Request.Context(), userID, id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_EVENT, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_EVENT, nil)
	ctx.JSON(http.StatusOK, res)
}
```

---

## Langkah 7: Buat routes

File: `modules/event/routes.go`

Semua endpoint di belakang middleware `Authenticate`. Perhatikan verb & path:
`POST ""`, `GET ""`, `GET "/:id"`, `PATCH "/:id"`, `DELETE "/:id"`.

```go
package event

import (
	"github.com/webdevtedxuniversitasairlangga/middlewares"
	"github.com/webdevtedxuniversitasairlangga/modules/auth/service"
	"github.com/webdevtedxuniversitasairlangga/modules/event/handler"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	eventController := do.MustInvoke[handler.EventHandler](injector)
	jwtService := do.MustInvokeNamed[service.JWTService](injector, constants.JWTService)

	eventRoutes := server.Group("/events", middlewares.Authenticate(jwtService))
	{
		eventRoutes.POST("", eventController.Create)
		eventRoutes.GET("", eventController.GetAll)
		eventRoutes.GET("/:id", eventController.GetByID)
		eventRoutes.PATCH("/:id", eventController.Update)
		eventRoutes.DELETE("/:id", eventController.Delete)
	}
}
```

---

## Langkah 8: Wire dependency di provider

File: `providers/core.go`

Tambahkan import dan registrasi. Ikuti pola blok `todo` di akhir `RegisterDependencies`.

Tambahkan import:
```go
eventHandler "github.com/webdevtedxuniversitasairlangga/modules/event/handler"
eventRepo "github.com/webdevtedxuniversitasairlangga/modules/event/repository"
eventService "github.com/webdevtedxuniversitasairlangga/modules/event/service"
```

Di dalam `RegisterDependencies`, tambahkan (setelah blok todo):
```go
eventRepository := eventRepo.NewEventRepository(db)
eventSvc := eventService.NewEventService(eventRepository, db)

do.Provide(
	injector, func(i *do.Injector) (eventHandler.EventHandler, error) {
		return eventHandler.NewEventHandler(i, eventSvc), nil
	},
)
```

---

## Langkah 9: Daftarkan routes di main

File: `cmd/main.go`

Tambahkan import modul dan panggil `RegisterRoutes` di dalam grup `v1`:

```go
import (
	// ...
	"github.com/webdevtedxuniversitasairlangga/modules/event"
)

// di dalam func main, grup v1:
v1 := server.Group("api/v1")
{
	auth.RegisterRoutes(v1, injector)
	todo.RegisterRoutes(v1, injector)
	event.RegisterRoutes(v1, injector) // <-- tambahkan
}
```

---

## Langkah 10: Build, test, dokumentasikan

```bash
go build ./...     # harus sukses
go vet ./...       # harus bersih
go test ./...      # harus lulus
go run ./cmd       # jalankan, cek AutoMigrate membuat tabel baru
```

Lalu:
1. Update `graphify update .` agar peta kode ikut terbarui.
2. Tambahkan endpoint baru ke tabel API di `README.md`.
3. Uji manual via `API_Test/` (OpenCollection) atau tool sejenis.

### Endpoint yang dihasilkan

| Method | Path | Auth | Body / Query |
|--------|------|------|--------------|
| POST | `/api/v1/events` | Bearer | `title`, `location` |
| GET | `/api/v1/events` | Bearer | `page`, `per_page` |
| GET | `/api/v1/events/:id` | Bearer | — |
| PATCH | `/api/v1/events/:id` | Bearer | `title?`, `location?` |
| DELETE | `/api/v1/events/:id` | Bearer | — |

---

## Checklist akhir (wajib sebelum PR)

- [ ] Entity dibuat + didaftarkan di `migrations.go`.
- [ ] 5 file modul dibuat: dto, repository, service, handler, routes.
- [ ] Query milik-user difilter `user_id` di repository.
- [ ] `user_id` diambil dari `ctx.MustGet("user_id")`, bukan dari body.
- [ ] Pesan sukses/gagal sebagai konstanta di DTO.
- [ ] Response memakai `BuildResponseSuccess` / `BuildResponseFailed`.
- [ ] Dependency di-wire di `providers/core.go`.
- [ ] Routes didaftarkan di `cmd/main.go`.
- [ ] `go build`, `go vet`, `go test` lulus.
- [ ] Dokumentasi API di `README.md` diperbarui.
