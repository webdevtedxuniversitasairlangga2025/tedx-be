package service

import (
	"context"
	"errors"
	"strings"
	"time"
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
	Stats(ctx context.Context) (dto.AttendanceStatsResponse, error)
	ListAttendees(ctx context.Context, req dto.AttendeeListRequest) (dto.AttendeeListResponse, error)
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

func newAttendanceCount(total, checkedIn int64) dto.AttendanceCount {
	return dto.AttendanceCount{Total: total, CheckedIn: checkedIn, NotCheckedIn: total - checkedIn}
}

// Stats merangkum kehadiran per tiket dan per tier, plus total keseluruhan.
// Urutan tiket/tier mengikuti urutan dari repository (nama).
func (s *checkInService) Stats(ctx context.Context) (dto.AttendanceStatsResponse, error) {
	rows, err := s.repository.CountAttendanceByTier(ctx, s.db)
	if err != nil {
		return dto.AttendanceStatsResponse{}, err
	}

	// Slice kosong (bukan nil) agar JSON menjadi [] dan bukan null.
	tickets := make([]dto.TicketAttendance, 0)
	index := make(map[string]int)
	var grandTotal, grandCheckedIn int64

	for _, row := range rows {
		i, ok := index[row.TicketID]
		if !ok {
			i = len(tickets)
			index[row.TicketID] = i
			tickets = append(tickets, dto.TicketAttendance{
				TicketID:   row.TicketID,
				TicketName: row.TicketName,
				Tiers:      make([]dto.TierAttendance, 0),
			})
		}

		tickets[i].Total += row.Total
		tickets[i].CheckedIn += row.CheckedIn
		tickets[i].Tiers = append(tickets[i].Tiers, dto.TierAttendance{
			TierID:          row.TierID,
			TierName:        row.TierName,
			AttendanceCount: newAttendanceCount(row.Total, row.CheckedIn),
		})

		grandTotal += row.Total
		grandCheckedIn += row.CheckedIn
	}

	for i := range tickets {
		tickets[i].NotCheckedIn = tickets[i].Total - tickets[i].CheckedIn
	}

	return dto.AttendanceStatsResponse{
		AttendanceCount: newAttendanceCount(grandTotal, grandCheckedIn),
		Tickets:         tickets,
		GeneratedAt:     time.Now().UTC(),
	}, nil
}

const (
	defaultAttendeePerPage = 20
	maxAttendeePerPage     = 100
	maxAttendeeSearchRunes = 100
)

// ListAttendees memvalidasi filter lalu mengambil satu halaman tiket peserta.
func (s *checkInService) ListAttendees(ctx context.Context, req dto.AttendeeListRequest) (dto.AttendeeListResponse, error) {
	filter := repository.AttendeeFilter{}

	search := strings.TrimSpace(req.Search)
	if utf8.RuneCountInString(search) > maxAttendeeSearchRunes {
		search = string([]rune(search)[:maxAttendeeSearchRunes])
	}
	filter.Search = search

	if tierID := strings.TrimSpace(req.TierID); tierID != "" {
		parsed, err := uuid.Parse(tierID)
		if err != nil {
			return dto.AttendeeListResponse{}, dto.ErrInvalidTierID
		}
		filter.TierID = &parsed
	}

	switch strings.ToLower(strings.TrimSpace(req.Status)) {
	case "", "all":
	case "used":
		used := true
		filter.Used = &used
	case "unused":
		used := false
		filter.Used = &used
	default:
		return dto.AttendeeListResponse{}, dto.ErrInvalidStatusFilter
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	perPage := req.PerPage
	if perPage <= 0 {
		perPage = defaultAttendeePerPage
	}
	if perPage > maxAttendeePerPage {
		perPage = maxAttendeePerPage
	}

	rows, total, err := s.repository.ListAttendees(ctx, s.db, filter, perPage, (page-1)*perPage)
	if err != nil {
		return dto.AttendeeListResponse{}, err
	}

	data := make([]dto.AttendeeListItem, 0, len(rows))
	for _, row := range rows {
		data = append(data, dto.AttendeeListItem{
			ID:            row.ID,
			TicketCode:    row.TicketCode,
			AttendeeName:  row.AttendeeName,
			AttendeeEmail: row.AttendeeEmail,
			AttendeePhone: row.AttendeePhone,
			TicketName:    row.TicketName,
			TierID:        row.TierID,
			TierName:      row.TierName,
			IsUsed:        row.IsUsed,
			UsedAt:        row.UsedAt,
		})
	}

	maxPage := int((total + int64(perPage) - 1) / int64(perPage))
	if maxPage < 1 {
		maxPage = 1
	}

	return dto.AttendeeListResponse{
		Data: data,
		Meta: dto.AttendeePaginationMeta{Page: page, PerPage: perPage, MaxPage: maxPage, Total: total},
	}, nil
}
