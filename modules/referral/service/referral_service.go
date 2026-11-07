package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/referral/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/referral/repository"
	"gorm.io/gorm"
)

type ReferralService interface {
	Create(ctx context.Context, req dto.CreateReferralCodeRequest) (dto.ReferralCodeResponse, error)
	List(ctx context.Context) ([]dto.ReferralCodeResponse, error)
	SetActive(ctx context.Context, id string, active bool) (dto.ReferralCodeResponse, error)
	Delete(ctx context.Context, id string) error
	Preview(ctx context.Context, req dto.ReferralPreviewRequest) (dto.ReferralPreviewResponse, error)
}

type referralService struct {
	repo repository.ReferralRepository
	db   *gorm.DB
}

func NewReferralService(repo repository.ReferralRepository, db *gorm.DB) ReferralService {
	return &referralService{repo: repo, db: db}
}

func referralCodeResponse(code entities.ReferralCode) dto.ReferralCodeResponse {
	return dto.ReferralCodeResponse{
		ID:                code.ID.String(),
		Code:              code.Code,
		DiscountPerTicket: code.DiscountPerTicket.StringFixed(2),
		Quota:             code.Quota,
		QuotaHeld:         code.QuotaHeld,
		QuotaUsed:         code.QuotaUsed,
		QuotaRemaining:    code.Quota - code.QuotaHeld - code.QuotaUsed,
		IsActive:          code.IsActive,
	}
}

func (s *referralService) Create(ctx context.Context, req dto.CreateReferralCodeRequest) (dto.ReferralCodeResponse, error) {
	codeText := NormalizeReferralCode(req.Code)
	discount, err := decimal.NewFromString(req.DiscountPerTicket)
	if err != nil || codeText == "" || discount.Sign() <= 0 || req.Quota <= 0 {
		return dto.ReferralCodeResponse{}, dto.ErrInvalidReferral
	}
	code := &entities.ReferralCode{Code: codeText, DiscountPerTicket: discount, Quota: req.Quota, IsActive: true}
	if err := s.repo.Create(ctx, nil, code); err != nil {
		return dto.ReferralCodeResponse{}, err
	}
	return referralCodeResponse(*code), nil
}

func (s *referralService) List(ctx context.Context) ([]dto.ReferralCodeResponse, error) {
	codes, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ReferralCodeResponse, 0, len(codes))
	for _, code := range codes {
		result = append(result, referralCodeResponse(code))
	}
	return result, nil
}

func (s *referralService) SetActive(ctx context.Context, id string, active bool) (dto.ReferralCodeResponse, error) {
	codeID, err := uuid.Parse(id)
	if err != nil {
		return dto.ReferralCodeResponse{}, dto.ErrInvalidReferralID
	}
	if err := s.repo.SetActive(ctx, codeID, active); err != nil {
		return dto.ReferralCodeResponse{}, err
	}
	code, err := s.repo.GetByID(ctx, codeID)
	if err != nil {
		return dto.ReferralCodeResponse{}, err
	}
	return referralCodeResponse(code), nil
}

func (s *referralService) Delete(ctx context.Context, id string) error {
	codeID, err := uuid.Parse(id)
	if err != nil {
		return dto.ErrInvalidReferralID
	}
	return s.repo.Delete(ctx, codeID)
}

func (s *referralService) Preview(ctx context.Context, req dto.ReferralPreviewRequest) (dto.ReferralPreviewResponse, error) {
	tierID, err := uuid.Parse(req.TicketTierID)
	if err != nil {
		return dto.ReferralPreviewResponse{}, errors.New("invalid ticket tier id")
	}
	response := dto.ReferralPreviewResponse{Allocations: []dto.ReferralAllocation{}}
	tier, err := s.repo.GetTier(ctx, tierID)
	if err != nil {
		return response, nil
	}
	now := time.Now()
	if tier.SaleStart != nil && now.Before(*tier.SaleStart) || tier.SaleEnd != nil && now.After(*tier.SaleEnd) {
		response.Reason = "Tier tiket belum tersedia untuk dibeli."
		return response, nil
	}
	code, err := s.repo.GetByCode(ctx, nil, NormalizeReferralCode(req.Code), false)
	if err != nil || !code.IsActive {
		response.Reason = "Kode referral tidak tersedia. Order dapat dilanjutkan dengan harga normal."
		return response, nil
	}
	claimed, err := s.repo.FindClaimedEmails(ctx, nil, code.ID)
	if err != nil {
		return dto.ReferralPreviewResponse{}, err
	}
	attendees := make([]dto.ReferralAttendeeInput, len(req.Emails))
	for i, email := range req.Emails {
		attendees[i] = dto.ReferralAttendeeInput{Email: email, UnitPrice: tier.Price}
	}
	remaining := code.Quota - code.QuotaHeld - code.QuotaUsed
	if remaining < 0 {
		remaining = 0
	}
	response.Allocations = AllocateDiscount(attendees, claimed, remaining, code.DiscountPerTicket)
	response.Applicable = len(response.Allocations) > 0
	response.ReferralCode = code.Code
	response.DiscountPerTicket = code.DiscountPerTicket.StringFixed(2)
	response.DiscountedAttendeeCount = len(response.Allocations)
	response.TotalDiscount = code.DiscountPerTicket.Mul(decimal.NewFromInt(int64(len(response.Allocations)))).StringFixed(2)
	if !response.Applicable {
		response.Reason = "Kode referral tidak dapat diterapkan. Order dapat dilanjutkan dengan harga normal."
	}
	return response, nil
}

func NormalizeReferralCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func NormalizeAttendeeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func AllocateDiscount(attendees []dto.ReferralAttendeeInput, priorClaims map[string]bool, remainingQuota int, discount decimal.Decimal) []dto.ReferralAllocation {
	allocations := make([]dto.ReferralAllocation, 0)
	seen := make(map[string]bool, len(attendees))
	for position, attendee := range attendees {
		if len(allocations) >= remainingQuota {
			break
		}
		email := NormalizeAttendeeEmail(attendee.Email)
		if email == "" || seen[email] || priorClaims[email] || !discount.LessThan(attendee.UnitPrice) {
			continue
		}
		seen[email] = true
		allocations = append(allocations, dto.ReferralAllocation{
			AttendeePosition: position,
			DiscountAmount:   discount,
		})
	}
	return allocations
}
