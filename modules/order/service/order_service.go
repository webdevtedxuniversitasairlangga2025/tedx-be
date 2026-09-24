package service

import (
	"context"
	"crypto/rand"

	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/skip2/go-qrcode"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/order/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/order/repository"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"github.com/webdevtedxuniversitasairlangga/pkg/utils"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type OrderService interface {
	Create(ctx context.Context, userID string, req dto.OrderCreateRequest) (dto.OrderResponse, error)
	GetByID(ctx context.Context, userID string, id string) (dto.OrderResponse, error)
	GetMyOrders(ctx context.Context, userID string, req dto.PaginationRequest) (dto.OrderPaginationResponse, error)
	GetAll(ctx context.Context, filter dto.OrderFilter, req dto.PaginationRequest) (dto.OrderPaginationResponse, error)
	Approve(ctx context.Context, adminID string, id string) (dto.OrderResponse, error)
	Reject(ctx context.Context, adminID string, id string, req dto.OrderRejectRequest) (dto.OrderResponse, error)
	UploadProof(ctx context.Context, userID string, id string, req dto.OrderUploadProofRequest) (dto.OrderResponse, error)
	ReleaseExpiredHolds(ctx context.Context) (int64, error)
	ResendEmail(ctx context.Context, adminID string, id string) error
	GetProofURL(ctx context.Context, id string) (string, error)
	ExportExcel(ctx context.Context) ([]byte, error)
	Delete(ctx context.Context, id string) error
	Cancel(ctx context.Context, userID string, id string) (dto.OrderResponse, error)
}

type orderService struct {
	repo repository.OrderRepository
	db   *gorm.DB
}

func NewOrderService(repo repository.OrderRepository, db *gorm.DB) OrderService {
	return &orderService{repo: repo, db: db}
}

func generateOrderNumber() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("ORD-%s-%s", time.Now().Format("20060102"), hex.EncodeToString(b))
}

const ticketCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randomTicketCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	code := make([]byte, 6)
	for i, v := range b {
		code[i] = ticketCodeCharset[int(v)%len(ticketCodeCharset)]
	}
	return string(code)
}

// generateUniqueTicketCode generate kode 6 karakter dan pastikan belum
// dipakai di DB maupun di batch order yang sama (reserved).
func (s *orderService) generateUniqueTicketCode(ctx context.Context, tx *gorm.DB, reserved map[string]bool) (string, error) {
	for i := 0; i < 10; i++ {
		code := randomTicketCode()
		if reserved[code] {
			continue
		}
		exists, err := s.repo.TicketCodeExists(ctx, tx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			reserved[code] = true
			return code, nil
		}
	}
	return "", fmt.Errorf("gagal generate kode tiket unik, coba lagi")
}

func toAttendeeResponse(a entities.AttendeeTicket) dto.AttendeeTicketResponse {
	return dto.AttendeeTicketResponse{
		ID:            a.ID.String(),
		OrderID:       a.OrderID.String(),
		TicketCode:    a.TicketCode,
		AttendeeName:  a.AttendeeName,
		AttendeeEmail: a.AttendeeEmail,
		AttendeePhone: a.AttendeePhone,
		AudienceType:  a.AudienceType,
		Institution:   a.Institution,
		IsSent:        a.IsSent,
		SentAt:        a.SentAt,
		IsUsed:        a.IsUsed,
		UsedAt:        a.UsedAt,
		CreatedAt:     a.CreatedAt,
	}
}

func toOrderResponse(o entities.Order) dto.OrderResponse {
	tickets := make([]dto.AttendeeTicketResponse, 0, len(o.AttendeeTickets))
	for _, t := range o.AttendeeTickets {
		tickets = append(tickets, toAttendeeResponse(t))
	}
	var approvedBy *string
	if o.ApprovedBy != nil {
		s := o.ApprovedBy.String()
		approvedBy = &s
	}
	return dto.OrderResponse{
		ID:              o.ID.String(),
		UserID:          o.UserID.String(),
		BuyerName:       o.User.Name,
		BuyerPhone:      o.User.TelpNumber,
		BuyerEmail:      o.User.Email,
		TicketTierID:    o.TicketTierID.String(),
		OrderNumber:     o.OrderNumber,
		Quantity:        o.Quantity,
		UnitPrice:       o.UnitPrice.StringFixed(2),
		TotalAmount:     o.TotalAmount.StringFixed(2),
		Status:          o.Status,
		ExpiredAt:       o.ExpiredAt,
		PaidAt:          o.PaidAt,
		ApprovedBy:      approvedBy,
		ApprovedAt:      o.ApprovedAt,
		RejectedReason:  o.RejectedReason,
		PaymentProofURL: o.PaymentProofURL,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
		AttendeeTickets: tickets,
	}
}

func (s *orderService) Create(ctx context.Context, userID string, req dto.OrderCreateRequest) (dto.OrderResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrInvalidUser
	}
	tierID, err := uuid.Parse(req.TicketTierID)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrTicketTierNotFound
	}
	if req.Quantity < 1 || req.Quantity > 5 {
		return dto.OrderResponse{}, dto.ErrQuantityOutOfRange
	}
	if len(req.Attendees) != 0 && len(req.Attendees) != req.Quantity {
		return dto.OrderResponse{}, dto.ErrAttendeesCountMismatch
	}
	var created dto.OrderResponse
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tier, err := s.repo.GetTierForUpdate(ctx, tx, tierID)
		if err != nil {
			return dto.ErrTicketTierNotFound
		}
		// Unscoped di repo → Create harus tolak tier soft-deleted + parent nonaktif/hapus
		if tier.DeletedAt.Valid {
			return dto.ErrTicketTierNotFound
		}
		if !tier.IsActive {
			return dto.ErrTicketTierInactive
		}
		var parent entities.Ticket
		if err := tx.WithContext(ctx).First(&parent, "id = ?", tier.TicketID).Error; err != nil {
			return dto.ErrTicketTierNotFound
		}
		if !parent.IsActive {
			return dto.ErrTicketInactive
		}
		now := time.Now()
		if tier.SaleStart != nil && now.Before(*tier.SaleStart) {
			return dto.ErrTicketSaleNotStarted
		}
		if tier.SaleEnd != nil && now.After(*tier.SaleEnd) {
			return dto.ErrTicketSaleEnded
		}
		available := tier.Quota - tier.QuotaFilled - tier.QuotaHeld
		if available < req.Quantity {
			return dto.ErrQuotaExceeded
		}
		tier.QuotaHeld += req.Quantity
		if err := s.repo.UpdateTier(ctx, tx, tier); err != nil {
			return err
		}
		order := entities.Order{
			ID:           uuid.New(),
			UserID:       uid,
			TicketTierID: tierID,
			OrderNumber:  generateOrderNumber(),
			Quantity:     req.Quantity,
			UnitPrice:    tier.Price,
			TotalAmount:  tier.Price.Mul(decimal.NewFromInt(int64(req.Quantity))),
			Status:       constants.ENUM_ORDER_STATUS_AWAITING_APPROVAL,
			ExpiredAt:    now.Add(15 * time.Minute),
		}
		saved, err := s.repo.Create(ctx, tx, order)
		if err != nil {
			return err
		}
		tickets := make([]entities.AttendeeTicket, 0, req.Quantity)
		reserved := make(map[string]bool)
		if len(req.Attendees) > 0 {
			for _, a := range req.Attendees {
				code, err := s.generateUniqueTicketCode(ctx, tx, reserved)
				if err != nil {
					return err
				}
				tickets = append(tickets, entities.AttendeeTicket{
					ID:            uuid.New(),
					OrderID:       saved.ID,
					TicketCode:    code,
					AttendeeName:  a.AttendeeName,
					AttendeeEmail: a.AttendeeEmail,
					AttendeePhone: a.AttendeePhone,
					AudienceType:  a.AudienceType,
					Institution:   a.Institution,
				})
			}
		} else {
			var buyer entities.User
			if err := tx.WithContext(ctx).Where("id = ?", uid).Take(&buyer).Error; err != nil {
				return err
			}
			// ponytail: nama+phone dari form IdentifyStepper → profile buyer
			// (buyer_name koreksi users.name utk kolom admin; email akun tidak diubah)
			nameChanged := req.BuyerName != nil && *req.BuyerName != "" && *req.BuyerName != buyer.Name
			phoneChanged := req.BuyerPhone != nil && *req.BuyerPhone != "" && buyer.TelpNumber == nil
			if nameChanged {
				buyer.Name = *req.BuyerName
			}
			if phoneChanged {
				buyer.TelpNumber = req.BuyerPhone
			}
			if nameChanged || phoneChanged {
				if err := tx.WithContext(ctx).Save(&buyer).Error; err != nil {
					return err
				}
			}
			phone := buyer.TelpNumber
			for i := 0; i < req.Quantity; i++ {
				code, err := s.generateUniqueTicketCode(ctx, tx, reserved)
				if err != nil {
					return err
				}
				tickets = append(tickets, entities.AttendeeTicket{
					ID:            uuid.New(),
					OrderID:       saved.ID,
					TicketCode:    code,
					AttendeeName:  buyer.Name,
					AttendeeEmail: buyer.Email,
					AttendeePhone: phone,
					AudienceType:  constants.ENUM_AUDIENCE_TYPE_UMUM,
				})
			}
		}
		if err := s.repo.CreateAttendeeTickets(ctx, tx, tickets); err != nil {
			return err
		}
		created = toOrderResponse(saved)
		return nil
	})
	if err != nil {
		return dto.OrderResponse{}, err
	}
	full, err := s.repo.GetByID(ctx, nil, uuid.MustParse(created.ID))
	if err == nil {
		created = toOrderResponse(full)
	}
	_ = full
	return created, nil
}

func (s *orderService) GetByID(ctx context.Context, userID string, id string) (dto.OrderResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrInvalidUser
	}
	oid, err := uuid.Parse(id)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrOrderNotFound
	}
	order, err := s.repo.GetByID(ctx, nil, oid)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrOrderNotFound
	}
	if order.UserID != uid {
		return dto.OrderResponse{}, dto.ErrOrderNotFound
	}
	return toOrderResponse(order), nil
}

func (s *orderService) GetMyOrders(ctx context.Context, userID string, req dto.PaginationRequest) (dto.OrderPaginationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.OrderPaginationResponse{}, dto.ErrInvalidUser
	}
	if req.Page <= 0 {
		req.Page = constants.ENUM_PAGINATION_PAGE
	}
	if req.PerPage <= 0 {
		req.PerPage = constants.ENUM_PAGINATION_PER_PAGE
	}
	if req.PerPage > 100 {
		req.PerPage = 100 // ponytail: cap — cegah per_page=100000
	}
	offset := (req.Page - 1) * req.PerPage
	orders, total, err := s.repo.GetAllByUserID(ctx, nil, uid, req.PerPage, offset)
	if err != nil {
		return dto.OrderPaginationResponse{}, err
	}
	data := make([]dto.OrderResponse, 0, len(orders))
	for _, o := range orders {
		data = append(data, toOrderResponse(o))
	}
	maxPage := int((total + int64(req.PerPage) - 1) / int64(req.PerPage))
	return dto.OrderPaginationResponse{
		Data: data,
		Meta: dto.PaginationMeta{Page: req.Page, PerPage: req.PerPage, MaxPage: maxPage, Total: total},
	}, nil
}

func (s *orderService) GetAll(ctx context.Context, filter dto.OrderFilter, req dto.PaginationRequest) (dto.OrderPaginationResponse, error) {
	if req.Page <= 0 {
		req.Page = constants.ENUM_PAGINATION_PAGE
	}
	if req.PerPage <= 0 {
		req.PerPage = constants.ENUM_PAGINATION_PER_PAGE
	}
	if req.PerPage > 100 {
		req.PerPage = 100
	}
	offset := (req.Page - 1) * req.PerPage
	orders, total, err := s.repo.GetAll(ctx, nil, filter.Status, req.PerPage, offset)
	if err != nil {
		return dto.OrderPaginationResponse{}, err
	}
	data := make([]dto.OrderResponse, 0, len(orders))
	for _, o := range orders {
		data = append(data, toOrderResponse(o))
	}
	maxPage := int((total + int64(req.PerPage) - 1) / int64(req.PerPage))
	return dto.OrderPaginationResponse{
		Data: data,
		Meta: dto.PaginationMeta{Page: req.Page, PerPage: req.PerPage, MaxPage: maxPage, Total: total},
	}, nil
}

func (s *orderService) Approve(ctx context.Context, adminID string, id string) (dto.OrderResponse, error) {
	adminUID, err := uuid.Parse(adminID)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrInvalidUser
	}
	oid, err := uuid.Parse(id)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrOrderNotFound
	}
	var result dto.OrderResponse
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := s.repo.GetByIDForUpdate(ctx, tx, oid)
		if err != nil {
			return dto.ErrOrderNotFound
		}
		if order.Status != constants.ENUM_ORDER_STATUS_AWAITING_APPROVAL {
			return dto.ErrOrderNotAwaitingApproval
		}
		if time.Now().After(order.ExpiredAt) {
			return dto.ErrOrderExpired
		}
		// wajib bukti bayar sebelum approve (audit)
		if order.PaymentProofURL == nil || *order.PaymentProofURL == "" {
			return dto.ErrOrderProofRequired
		}
		tier, err := s.repo.GetTierForUpdate(ctx, tx, order.TicketTierID)
		if err != nil {
			return dto.ErrTicketTierNotFound
		}
		if tier.QuotaHeld < order.Quantity {
			tier.QuotaHeld = 0
		} else {
			tier.QuotaHeld -= order.Quantity
		}
		// setelah lepas hold order ini: sisa kuota harus muat quantity (bukan cuma >= 0)
		available := tier.Quota - tier.QuotaFilled - tier.QuotaHeld
		if available < order.Quantity {
			return dto.ErrQuotaExceeded
		}
		tier.QuotaFilled += order.Quantity
		if err := s.repo.UpdateTier(ctx, tx, tier); err != nil {
			return err
		}
		now := time.Now()
		order.Status = constants.ENUM_ORDER_STATUS_PAID
		order.PaidAt = &now
		order.ApprovedBy = &adminUID
		order.ApprovedAt = &now
		updated, err := s.repo.Update(ctx, tx, order)
		if err != nil {
			return err
		}
		full, err := s.repo.GetByID(ctx, tx, oid)
		if err != nil {
			return err
		}
		result = toOrderResponse(full)
		_ = updated
		return nil
	})
	if err != nil {
		return dto.OrderResponse{}, err
	}
	var buyerEmail string
	var buyerName string
	var buyer entities.User
	if err := s.db.WithContext(ctx).Where("id = ?", result.UserID).Take(&buyer).Error; err == nil {
		buyerEmail = buyer.Email
		buyerName = buyer.Name
	} else if len(result.AttendeeTickets) > 0 {
		buyerEmail = result.AttendeeTickets[0].AttendeeEmail
		buyerName = result.AttendeeTickets[0].AttendeeName
	}
	if buyerEmail != "" {
		var ticketsHTML string
		embeds := make(map[string][]byte)
		for _, t := range result.AttendeeTickets {

			png, err := qrcode.Encode(t.TicketCode, qrcode.Medium, 256)
			if err != nil {
				log.Printf("gagal generate qrcode untuk tiket %s: %v", t.TicketCode, err)
				continue
			}

			filename := fmt.Sprintf("qr_%s.png", t.TicketCode)
			embeds[filename] = png
			imgTag := fmt.Sprintf(`<img src="cid:%s" width="256" height="256" alt="QR Code">`, filename)

			ticketsHTML += fmt.Sprintf(`
			<tr>
				<td style="text-align: center;">%s</td>
				<td style="text-align: center;">%s</td>
				<td style="text-align: center;">%s</td>
			</tr>`, template.HTMLEscapeString(t.AttendeeName), t.TicketCode, imgTag)
		}

		emailData := map[string]any{
			"BuyerName":   buyerName,
			"OrderNumber": result.OrderNumber,
			"Quantity":    fmt.Sprintf("%d", result.Quantity),
			"TotalAmount": result.TotalAmount,
			"TicketsHTML": template.HTML(ticketsHTML),
		}

		if body, err := utils.RenderEmailTemplate("ticket_approved.html", emailData); err != nil {
			log.Printf("gagal render email template: %v", err)
		} else if err := utils.SendMailWithEmbeds(buyerEmail, "TEDx Ticket Approved - "+result.OrderNumber, body, embeds); err != nil {
			log.Printf("gagal kirim email tiket ke %s: %v", buyerEmail, err)
		} else {
			_ = s.repo.MarkTicketsSent(ctx, uuid.MustParse(result.ID))
		}
	}
	return result, nil
}

func (s *orderService) Reject(ctx context.Context, adminID string, id string, req dto.OrderRejectRequest) (dto.OrderResponse, error) {
	adminUID, err := uuid.Parse(adminID)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrInvalidUser
	}
	oid, err := uuid.Parse(id)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrOrderNotFound
	}
	var result dto.OrderResponse
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := s.repo.GetByIDForUpdate(ctx, tx, oid)
		if err != nil {
			return dto.ErrOrderNotFound
		}
		if order.Status != constants.ENUM_ORDER_STATUS_AWAITING_APPROVAL {
			return dto.ErrOrderNotAwaitingApproval
		}
		tier, err := s.repo.GetTierForUpdate(ctx, tx, order.TicketTierID)
		if err != nil {
			return dto.ErrTicketTierNotFound
		}
		if tier.QuotaHeld >= order.Quantity {
			tier.QuotaHeld -= order.Quantity
		} else {
			tier.QuotaHeld = 0
		}
		if err := s.repo.UpdateTier(ctx, tx, tier); err != nil {
			return err
		}
		now := time.Now()
		order.Status = constants.ENUM_ORDER_STATUS_REJECTED
		order.RejectedReason = &req.Reason
		order.ApprovedBy = &adminUID
		order.ApprovedAt = &now
		updated, err := s.repo.Update(ctx, tx, order)
		if err != nil {
			return err
		}
		result = toOrderResponse(updated)
		return nil
	})
	if err != nil {
		return dto.OrderResponse{}, err
	}
	return result, nil
}

func (s *orderService) UploadProof(ctx context.Context, userID string, id string, req dto.OrderUploadProofRequest) (dto.OrderResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrInvalidUser
	}
	oid, err := uuid.Parse(id)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrOrderNotFound
	}
	order, err := s.repo.GetByID(ctx, nil, oid)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrOrderNotFound
	}
	if order.UserID != uid {
		return dto.OrderResponse{}, dto.ErrOrderNotFound
	}
	if order.Status != constants.ENUM_ORDER_STATUS_AWAITING_APPROVAL {
		return dto.OrderResponse{}, dto.ErrOrderNotAwaitingApproval
	}
	if time.Now().After(order.ExpiredAt) {
		return dto.OrderResponse{}, dto.ErrOrderExpired
	}
	order.PaymentProofURL = &req.PaymentProofURL
	order.ExpiredAt = time.Now().Add(24 * time.Hour)
	updated, err := s.repo.Update(ctx, nil, order)
	if err != nil {
		return dto.OrderResponse{}, err
	}
	return toOrderResponse(updated), nil
}

// Cancel — batalkan order awaiting milik sendiri dan lepas hold kuota.
// Idempoten untuk status cancelled agar retry CONFIRM tidak menumpuk order.
func (s *orderService) Cancel(ctx context.Context, userID string, id string) (dto.OrderResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrInvalidUser
	}
	oid, err := uuid.Parse(id)
	if err != nil {
		return dto.OrderResponse{}, dto.ErrOrderNotFound
	}
	var result dto.OrderResponse
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := s.repo.GetByIDForUpdate(ctx, tx, oid)
		if err != nil {
			return dto.ErrOrderNotFound
		}
		if order.UserID != uid {
			return dto.ErrOrderNotFound
		}
		if order.Status == constants.ENUM_ORDER_STATUS_CANCELLED {
			result = toOrderResponse(order)
			return nil
		}
		if order.Status != constants.ENUM_ORDER_STATUS_AWAITING_APPROVAL {
			return dto.ErrOrderNotAwaitingApproval
		}
		tier, err := s.repo.GetTierForUpdate(ctx, tx, order.TicketTierID)
		if err != nil {
			return dto.ErrTicketTierNotFound
		}
		if tier.QuotaHeld >= order.Quantity {
			tier.QuotaHeld -= order.Quantity
		} else {
			tier.QuotaHeld = 0
		}
		if err := s.repo.UpdateTier(ctx, tx, tier); err != nil {
			return err
		}
		order.Status = constants.ENUM_ORDER_STATUS_CANCELLED
		updated, err := s.repo.Update(ctx, tx, order)
		if err != nil {
			return err
		}
		result = toOrderResponse(updated)
		return nil
	})
	if err != nil {
		return dto.OrderResponse{}, err
	}
	return result, nil
}

func (s *orderService) ReleaseExpiredHolds(ctx context.Context) (int64, error) {
	orders, err := s.repo.FindExpiredAwaitingApproval(ctx, nil)
	if err != nil {
		return 0, err
	}
	var count int64
	for _, o := range orders {
		err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			fresh, err := s.repo.GetByIDForUpdate(ctx, tx, o.ID)
			if err != nil {
				return nil
			}
			if fresh.Status != constants.ENUM_ORDER_STATUS_AWAITING_APPROVAL {
				return nil
			}
			if time.Now().Before(fresh.ExpiredAt) {
				return nil
			}
			tier, err := s.repo.GetTierForUpdate(ctx, tx, fresh.TicketTierID)
			if err != nil {
				return nil
			}
			if tier.QuotaHeld >= fresh.Quantity {
				tier.QuotaHeld -= fresh.Quantity
			} else {
				tier.QuotaHeld = 0
			}
			if err := s.repo.UpdateTier(ctx, tx, tier); err != nil {
				return err
			}
			fresh.Status = constants.ENUM_ORDER_STATUS_EXPIRED
			_, err = s.repo.Update(ctx, tx, fresh)
			return err
		})
		if err == nil {
			count++
		}
	}
	return count, nil
}
func (s *orderService) GetProofURL(ctx context.Context, id string) (string, error) {
	oid, err := uuid.Parse(id)
	if err != nil {
		return "", dto.ErrOrderNotFound
	}
	order, err := s.repo.GetByID(ctx, nil, oid)
	if err != nil {
		return "", dto.ErrOrderNotFound
	}
	if order.PaymentProofURL == nil {
		return "", dto.ErrOrderProofRequired
	}
	return *order.PaymentProofURL, nil
}

// Delete — soft-delete history pembayaran (row tetap utk FK attendee)
func (s *orderService) Delete(ctx context.Context, id string) error {
	oid, err := uuid.Parse(id)
	if err != nil {
		return dto.ErrOrderNotFound
	}
	if _, err := s.repo.GetByID(ctx, nil, oid); err != nil {
		return dto.ErrOrderNotFound
	}
	return s.repo.SoftDelete(ctx, nil, oid)
}

func (s *orderService) ResendEmail(ctx context.Context, adminID string, id string) error {
	_, err := uuid.Parse(adminID)
	if err != nil {
		return dto.ErrInvalidUser
	}
	oid, err := uuid.Parse(id)
	if err != nil {
		return dto.ErrOrderNotFound
	}

	order, err := s.repo.GetByID(ctx, nil, oid)
	if err != nil {
		return dto.ErrOrderNotFound
	}

	if order.Status != constants.ENUM_ORDER_STATUS_PAID {
		return fmt.Errorf("hanya pesanan dengan status paid yang bisa dikirim ulang")
	}

	result := toOrderResponse(order)

	var buyerEmail string
	var buyerName string
	var buyer entities.User
	if err := s.db.WithContext(ctx).Where("id = ?", result.UserID).Take(&buyer).Error; err == nil {
		buyerEmail = buyer.Email
		buyerName = buyer.Name
	} else if len(result.AttendeeTickets) > 0 {
		buyerEmail = result.AttendeeTickets[0].AttendeeEmail
		buyerName = result.AttendeeTickets[0].AttendeeName
	}

	if buyerEmail == "" {
		return fmt.Errorf("email pembeli tidak ditemukan")
	}

	var ticketsHTML string
	embeds := make(map[string][]byte)
	for _, t := range result.AttendeeTickets {

		png, err := qrcode.Encode(t.TicketCode, qrcode.Medium, 256)
		if err != nil {
			log.Printf("gagal generate qrcode untuk tiket %s: %v", t.TicketCode, err)
			continue
		}

		filename := fmt.Sprintf("qr_%s.png", t.TicketCode)
		embeds[filename] = png
		imgTag := fmt.Sprintf(`<img src="cid:%s" width="256" height="256" alt="QR Code">`, filename)

		ticketsHTML += fmt.Sprintf(`
			<tr>
				<td style="text-align: center;">%s</td>
				<td style="text-align: center;">%s</td>
				<td style="text-align: center;">%s</td>
			</tr>`, template.HTMLEscapeString(t.AttendeeName), t.TicketCode, imgTag)
	}

	emailData := map[string]any{
		"BuyerName":   buyerName,
		"OrderNumber": result.OrderNumber,
		"Quantity":    fmt.Sprintf("%d", result.Quantity),
		"TotalAmount": result.TotalAmount,
		"TicketsHTML": template.HTML(ticketsHTML),
	}

	body, err := utils.RenderEmailTemplate("ticket_approved.html", emailData)
	if err != nil {
		return fmt.Errorf("gagal render email template: %v", err)
	}

	if err := utils.SendMailWithEmbeds(buyerEmail, "TEDx Ticket Approved - "+result.OrderNumber, body, embeds); err != nil {
		return fmt.Errorf("gagal kirim email tiket ke %s: %v", buyerEmail, err)
	}
	_ = s.repo.MarkTicketsSent(ctx, oid)

	return nil
}

func (s *orderService) ExportExcel(ctx context.Context) ([]byte, error) {
	orders, err := s.repo.GetAllForExport(ctx, nil)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	sheet := "Orders"
	f.SetSheetName(f.GetSheetName(0), sheet)

	headers := []string{
		"Order ID", "Order Number", "User Name", "User Email",
		"Ticket Name", "Tier", "Quantity", "Unit Price", "Total Amount",
		"Status", "Expired At", "Paid At", "Approved By", "Approved At",
		"Rejected Reason", "Payment Proof URL", "Order Created At", "Order Updated At",
		"Ticket ID", "Ticket Code", "Attendee Name", "Attendee Email", "Attendee Phone",
		"Audience Type", "Institution", "Is Sent", "Sent At",
		"Is Used", "Used At", "Checked By", "Ticket Created At",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	const layout = "2006-01-02 15:04:05"
	row := 2
	for _, o := range orders {
		approvedBy, approvedAt, paidAt, rejectedReason, proofURL := "", "", "", "", ""
		if o.ApprovedByUser != nil {
			approvedBy = o.ApprovedByUser.Name
		}
		if o.ApprovedAt != nil {
			approvedAt = o.ApprovedAt.Format(layout)
		}
		if o.PaidAt != nil {
			paidAt = o.PaidAt.Format(layout)
		}
		if o.RejectedReason != nil {
			rejectedReason = *o.RejectedReason
		}
		if o.PaymentProofURL != nil {
			proofURL = *o.PaymentProofURL
		}

		tickets := o.AttendeeTickets
		if len(tickets) == 0 {
			tickets = []entities.AttendeeTicket{{}}
		}

		for _, t := range tickets {
			phone, institution, sentAt, usedAt, checkedBy := "", "", "", "", ""
			if t.AttendeePhone != nil {
				phone = *t.AttendeePhone
			}
			if t.Institution != nil {
				institution = *t.Institution
			}
			if t.SentAt != nil {
				sentAt = t.SentAt.Format(layout)
			}
			if t.UsedAt != nil {
				usedAt = t.UsedAt.Format(layout)
			}
			if t.CheckedByUser != nil {
				checkedBy = t.CheckedByUser.Name
			}
			ticketID, ticketCode, ticketCreatedAt := "", "", ""
			if t.ID != uuid.Nil {
				ticketID = t.ID.String()
				ticketCode = t.TicketCode
				ticketCreatedAt = t.CreatedAt.Format(layout)
			}
			values := []any{
				o.ID.String(), o.OrderNumber, o.User.Name, o.User.Email,
				o.TicketTier.Ticket.Name, o.TicketTier.Tier, o.Quantity,
				o.UnitPrice.String(), o.TotalAmount.String(), o.Status,
				o.ExpiredAt.Format(layout), paidAt, approvedBy, approvedAt,
				rejectedReason, proofURL, o.CreatedAt.Format(layout), o.UpdatedAt.Format(layout),
				ticketID, ticketCode, t.AttendeeName, t.AttendeeEmail, phone,
				t.AudienceType, institution, t.IsSent, sentAt,
				t.IsUsed, usedAt, checkedBy, ticketCreatedAt,
			}
			for i, v := range values {
				cell, _ := excelize.CoordinatesToCellName(i+1, row)
				f.SetCellValue(sheet, cell, v)
			}
			row++
		}
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
