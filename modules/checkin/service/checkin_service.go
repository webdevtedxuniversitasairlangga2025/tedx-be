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
}

type checkInService struct {
	repository repository.CheckInRepository
	db         *gorm.DB
}

func NewCheckInService(repository repository.CheckInRepository, db *gorm.DB) CheckInService {
	return &checkInService{repository: repository, db: db}
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
	if strings.TrimSpace(req.TicketCode) == "" || utf8.RuneCountInString(req.TicketCode) > 255 {
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
		return dto.CheckInResponse{}, dto.ErrTicketAlreadyUsed
	}
	if ticket.Order.Status != constants.ENUM_ORDER_STATUS_PAID {
		return dto.CheckInResponse{}, dto.ErrTicketNotPaid
	}

	return dto.CheckInResponse{}, dto.ErrCheckInNotApplied
}
