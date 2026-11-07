package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/webdevtedxuniversitasairlangga/database/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReferralRepository interface {
	Create(ctx context.Context, tx *gorm.DB, code *entities.ReferralCode) error
	GetByCode(ctx context.Context, tx *gorm.DB, code string, forUpdate bool) (entities.ReferralCode, error)
	GetByID(ctx context.Context, id uuid.UUID) (entities.ReferralCode, error)
	List(ctx context.Context) ([]entities.ReferralCode, error)
	SetActive(ctx context.Context, id uuid.UUID, active bool) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetTier(ctx context.Context, id uuid.UUID) (entities.TicketTier, error)
	FindClaimedEmails(ctx context.Context, tx *gorm.DB, codeID uuid.UUID) (map[string]bool, error)
	CountRedeemed(ctx context.Context, tx *gorm.DB, codeID uuid.UUID) (int64, error)
	CreateClaims(ctx context.Context, tx *gorm.DB, claims []entities.ReferralClaim) error
	RedeemClaims(ctx context.Context, tx *gorm.DB, orderID uuid.UUID) error
	ReleaseClaims(ctx context.Context, tx *gorm.DB, codeID uuid.UUID, orderID uuid.UUID) (int64, error)
}

type referralRepository struct {
	db *gorm.DB
}

func NewReferralRepository(db *gorm.DB) ReferralRepository {
	return &referralRepository{db: db}
}

func (r *referralRepository) dbOrTx(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *referralRepository) Create(ctx context.Context, tx *gorm.DB, code *entities.ReferralCode) error {
	return r.dbOrTx(tx).WithContext(ctx).Create(code).Error
}

func (r *referralRepository) GetByCode(ctx context.Context, tx *gorm.DB, code string, forUpdate bool) (entities.ReferralCode, error) {
	db := r.dbOrTx(tx).WithContext(ctx)
	if forUpdate {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var referral entities.ReferralCode
	err := db.Where("code = ?", code).Take(&referral).Error
	return referral, err
}

func (r *referralRepository) GetByID(ctx context.Context, id uuid.UUID) (entities.ReferralCode, error) {
	var referral entities.ReferralCode
	err := r.db.WithContext(ctx).First(&referral, "id = ?", id).Error
	return referral, err
}

func (r *referralRepository) List(ctx context.Context) ([]entities.ReferralCode, error) {
	var codes []entities.ReferralCode
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&codes).Error
	return codes, err
}

func (r *referralRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return r.db.WithContext(ctx).Model(&entities.ReferralCode{}).Where("id = ?", id).Update("is_active", active).Error
}

func (r *referralRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var code entities.ReferralCode
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&code, "id = ?", id).Error; err != nil {
			return err
		}
		return tx.Model(&code).Updates(map[string]any{"is_active": false, "deleted_at": gorm.Expr("CURRENT_TIMESTAMP")}).Error
	})
}

func (r *referralRepository) GetTier(ctx context.Context, id uuid.UUID) (entities.TicketTier, error) {
	var tier entities.TicketTier
	err := r.db.WithContext(ctx).
		Joins("JOIN tickets ON tickets.id = ticket_tiers.ticket_id").
		Where("ticket_tiers.id = ? AND ticket_tiers.is_active = ? AND tickets.is_active = ? AND tickets.deleted_at IS NULL", id, true, true).
		Take(&tier).Error
	return tier, err
}

func (r *referralRepository) FindClaimedEmails(ctx context.Context, tx *gorm.DB, codeID uuid.UUID) (map[string]bool, error) {
	var emails []string
	err := r.dbOrTx(tx).WithContext(ctx).Model(&entities.ReferralClaim{}).
		Where("referral_code_id = ? AND status IN ?", codeID, []string{entities.ReferralClaimHeld, entities.ReferralClaimRedeemed}).
		Pluck("attendee_email", &emails).Error
	if err != nil {
		return nil, err
	}
	claimed := make(map[string]bool, len(emails))
	for _, email := range emails {
		claimed[email] = true
	}
	return claimed, nil
}

func (r *referralRepository) CountRedeemed(ctx context.Context, tx *gorm.DB, codeID uuid.UUID) (int64, error) {
	var count int64
	err := r.dbOrTx(tx).WithContext(ctx).Model(&entities.ReferralClaim{}).
		Where("referral_code_id = ? AND status = ?", codeID, entities.ReferralClaimRedeemed).
		Count(&count).Error
	return count, err
}

func (r *referralRepository) CreateClaims(ctx context.Context, tx *gorm.DB, claims []entities.ReferralClaim) error {
	if len(claims) == 0 {
		return nil
	}
	return r.dbOrTx(tx).WithContext(ctx).Create(&claims).Error
}

func (r *referralRepository) RedeemClaims(ctx context.Context, tx *gorm.DB, orderID uuid.UUID) error {
	return r.dbOrTx(tx).WithContext(ctx).Model(&entities.ReferralClaim{}).
		Where("order_id = ? AND status = ?", orderID, entities.ReferralClaimHeld).
		Update("status", entities.ReferralClaimRedeemed).Error
}

func (r *referralRepository) ReleaseClaims(ctx context.Context, tx *gorm.DB, codeID uuid.UUID, orderID uuid.UUID) (int64, error) {
	db := r.dbOrTx(tx).WithContext(ctx)
	var claims []entities.ReferralClaim
	if err := db.Where("order_id = ? AND referral_code_id = ? AND status = ?", orderID, codeID, entities.ReferralClaimHeld).Find(&claims).Error; err != nil {
		return 0, err
	}
	if len(claims) == 0 {
		return 0, nil
	}
	if err := db.Delete(&claims).Error; err != nil {
		return 0, err
	}
	result := db.Unscoped().Model(&entities.ReferralCode{}).Where("id = ? AND quota_held >= ?", codeID, len(claims)).
		Update("quota_held", gorm.Expr("quota_held - ?", len(claims)))
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected != 1 {
		return 0, gorm.ErrRecordNotFound
	}
	return int64(len(claims)), nil
}
