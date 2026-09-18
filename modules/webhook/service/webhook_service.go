package service

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"github.com/webdevtedxuniversitasairlangga/modules/webhook/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/webhook/repository"
	"gorm.io/gorm"
)

type WebhookService interface {
	ProcessMidtransNotification(ctx context.Context, req dto.MidtransNotificationRequest) error
}

type webhookService struct {
	webhookRepo repository.WebhookRepository
	serverKey   string
}

func NewWebhookService(webhookRepo repository.WebhookRepository, serverKey string) WebhookService {
	return &webhookService{
		webhookRepo: webhookRepo,
		serverKey:   serverKey,
	}
}

func (s *webhookService) verifySignature(req dto.MidtransNotificationRequest) bool {

	payload := req.OrderID + req.StatusCode + req.GrossAmount + s.serverKey
	hash := sha512.New()
	hash.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(hash.Sum(nil))

	return req.SignatureKey == expectedSignature
}

func (s *webhookService) ProcessMidtransNotification(ctx context.Context, req dto.MidtransNotificationRequest) error {

	if !s.verifySignature(req) {
		return dto.ErrInvalidSignature
	}

	return s.webhookRepo.Transaction(ctx, func(tx *gorm.DB) error {

		order, err := s.webhookRepo.GetOrder(ctx, tx, req.OrderID)
		if err != nil {
			return dto.ErrOrderNotFound
		}

		if order.Status == "paid" || order.Status == "failed" {
			return nil
		}

		if req.TransactionStatus == "settlement" || req.TransactionStatus == "capture" {

			tier, err := s.webhookRepo.GetTierWithLock(ctx, tx, order.TicketTierID.String())
			if err != nil {
				return err
			}

			if tier.QuotaFilled+order.Quantity > tier.Quota {
				return errors.New("kuota tiket tidak mencukupi, oversell terdeteksi")
			}

			tier.QuotaFilled += order.Quantity
			if err := s.webhookRepo.UpdateTier(ctx, tx, tier); err != nil {
				return err
			}

			var attendeeTickets []entities.AttendeeTicket
			for i := 0; i < order.Quantity; i++ {
				ticketCode := uuid.New().String()
				attendeeTickets = append(attendeeTickets, entities.AttendeeTicket{
					ID:         uuid.New(),
					OrderID:    order.ID,
					TicketCode: ticketCode,
					IsUsed:     false,
				})
			}

			if err := s.webhookRepo.CreateAttendeeTickets(ctx, tx, attendeeTickets); err != nil {
				return err
			}

			order.Status = "paid"
			if err := s.webhookRepo.UpdateOrder(ctx, tx, order); err != nil {
				return err
			}

		} else if req.TransactionStatus == "cancel" || req.TransactionStatus == "expire" || req.TransactionStatus == "deny" {
			order.Status = "failed"
			if err := s.webhookRepo.UpdateOrder(ctx, tx, order); err != nil {
				return err
			}
		}

		return nil
	})
}
