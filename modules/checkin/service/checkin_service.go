package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/checkin/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/checkin/repository"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"gorm.io/gorm"
)

type CheckInService interface {
	CheckIn(ctx context.Context, checkerID string, req dto.CheckInRequest) (dto.CheckInResponse, error)
	Lookup(ctx context.Context, ticketCode string) (dto.TicketDetailResponse, error)
}

type checkInService struct {
	repository repository.CheckInRepository
	db         *gorm.DB
}

func NewCheckInService(repository repository.CheckInRepository, db *gorm.DB) CheckInService {
	return &checkInService{repository: repository, db: db}
}

func isValidTicketCode(code string) bool {
	return strings.TrimSpace(code) != "" && utf8.RuneCountInString(code) <= 255 && !strings.Contains(code, " ")
}

func toDetailResponse(ticket entities.AttendeeTicket) dto.TicketDetailResponse {
	var checkedBy *string
	if ticket.CheckedBy != nil {
		id := ticket.CheckedBy.String()
		checkedBy = &id
	}

	return dto.TicketDetailResponse{
		ID:            ticket.ID.String(),
		TicketCode:    ticket.TicketCode,
		AttendeeName:  ticket.AttendeeName,
		AttendeeEmail: ticket.AttendeeEmail,
		AttendeePhone: ticket.AttendeePhone,
		TicketName:    ticket.Order.TicketTier.Ticket.Name,
		TierName:      ticket.Order.TicketTier.Tier,
		OrderStatus:   ticket.Order.Status,
		IsUsed:        ticket.IsUsed,
		UsedAt:        ticket.UsedAt,
		CheckedBy:     checkedBy,
	}
}

func toResponse(ticket entities.AttendeeTicket) dto.CheckInResponse {
	var checkedBy *string
	if ticket.CheckedBy != nil {
		id := ticket.CheckedBy.String()
		checkedBy = &id
	}

	return dto.CheckInResponse{
		ID:           ticket.ID.String(),
		AttendeeName: ticket.AttendeeName,
		IsUsed:       ticket.IsUsed,
		UsedAt:       ticket.UsedAt,
		CheckedBy:    checkedBy,
	}
}

func (s *checkInService) CheckIn(ctx context.Context, checkerID string, req dto.CheckInRequest) (dto.CheckInResponse, error) {
	checkerUUID, err := uuid.Parse(checkerID)
	if err != nil || checkerUUID == uuid.Nil {
		return dto.CheckInResponse{}, dto.ErrInvalidChecker
	}
	if !isValidTicketCode(req.TicketCode) {
		return dto.CheckInResponse{}, dto.ErrInvalidTicketCode
	}

	ticket, updated, err := s.repository.MarkAsUsed(ctx, s.db, req.TicketCode, checkerUUID)
	if err != nil {
		return dto.CheckInResponse{}, err
	}
	if updated {
		return toResponse(ticket), nil
	}

	ticket, err = s.repository.FindByTicketCode(ctx, s.db, req.TicketCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.CheckInResponse{}, dto.ErrAttendeeTicketNotFound
		}
		return dto.CheckInResponse{}, err
	}
	if ticket.IsUsed {
		return toResponse(ticket), dto.ErrTicketAlreadyUsed
	}
	if ticket.Order.Status != constants.ENUM_ORDER_STATUS_PAID {
		return dto.CheckInResponse{}, dto.ErrTicketNotPaid
	}

	return dto.CheckInResponse{}, dto.ErrCheckInNotApplied
}

// Lookup mengembalikan detail tiket tanpa menandainya terpakai. Dipakai halaman
// registrasi untuk menampilkan data peserta sebelum admin menekan konfirmasi
// (konfirmasi sendiri tetap lewat CheckIn / POST /checkins/scan).
func (s *checkInService) Lookup(ctx context.Context, ticketCode string) (dto.TicketDetailResponse, error) {
	if !isValidTicketCode(ticketCode) {
		return dto.TicketDetailResponse{}, dto.ErrInvalidTicketCode
	}

	ticket, err := s.repository.FindDetailByTicketCode(ctx, s.db, ticketCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.TicketDetailResponse{}, dto.ErrAttendeeTicketNotFound
		}
		return dto.TicketDetailResponse{}, err
	}

	return toDetailResponse(ticket), nil
}
