package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/webdevtedxuniversitasairlangga/pkg/constants"
	"gorm.io/gorm"
)

type Order struct {
	ID                    uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID                uuid.UUID       `gorm:"type:uuid;not null;index"`
	TicketTierID          uuid.UUID       `gorm:"type:uuid;not null;index"`
	OrderNumber           string          `gorm:"size:255;uniqueIndex;not null"`
	Quantity              int             `gorm:"not null;default:1"`
	UnitPrice             decimal.Decimal `gorm:"type:numeric(10,2);not null"`
	TotalAmount           decimal.Decimal `gorm:"type:numeric(10,2);not null"`
	Status                string          `gorm:"size:50;not null;default:'pending'"`
	MidtransTransactionID *string         `gorm:"size:255"`
	SnapToken             *string         `gorm:"size:255"`
	SnapRedirectURL       *string         `gorm:"size:500"`
	PaymentType           *string         `gorm:"size:100"`
	ExpiredAt             time.Time       `gorm:"type:timestamp with time zone;not null"`
	PaidAt                *time.Time      `gorm:"type:timestamp with time zone"`
	User                  User            `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	TicketTier            TicketTier      `gorm:"foreignKey:TicketTierID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	Timestamp
}

func (o *Order) BeforeCreate(_ *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}

	if o.Status == "" {
		o.Status = constants.ENUM_ORDER_STATUS_PENDING
	}
	
	return nil
}
