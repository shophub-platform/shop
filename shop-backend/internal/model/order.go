package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	OrderStatusPaid           OrderStatus = "PAID"
	OrderStatusProcessing     OrderStatus = "PROCESSING"
	OrderStatusShipped        OrderStatus = "SHIPPED"
	OrderStatusDelivered      OrderStatus = "DELIVERED"
	OrderStatusCancelled      OrderStatus = "CANCELLED"
)

// ValidTransitions defines which status changes are allowed.
var ValidTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusPaid:       {OrderStatusProcessing, OrderStatusCancelled},
	OrderStatusProcessing: {OrderStatusShipped, OrderStatusCancelled},
	OrderStatusShipped:    {OrderStatusDelivered},
}

func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	allowed, ok := ValidTransitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == next {
			return true
		}
	}
	return false
}

type Order struct {
	ID         uuid.UUID   `gorm:"type:uuid;primaryKey"                        json:"id"`
	UserID     string      `gorm:"not null;index"                              json:"userId"`
	Items      []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
	Total      float64     `gorm:"type:numeric(18,8);not null"                 json:"total"`
	Status     OrderStatus `gorm:"type:varchar(20);not null"                   json:"status"`
	WalletFrom *string     `gorm:"column:wallet_from"                          json:"walletFrom,omitempty"`
	TxHash     *string     `gorm:"uniqueIndex"                                 json:"txHash,omitempty"`
	CreatedAt  time.Time   `                                                   json:"createdAt"`
	UpdatedAt  time.Time   `                                                   json:"updatedAt"`
}

// OrderItem stores a price/name snapshot so historical orders stay accurate
// even if the original item is later edited or deleted.
type OrderItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"        json:"id"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null;index"    json:"-"`
	ItemID    uuid.UUID `gorm:"type:uuid;not null"          json:"itemId"`
	ItemName  string    `gorm:"not null"                    json:"itemName"`
	Quantity  int       `gorm:"not null;check:quantity > 0" json:"quantity"`
	UnitPrice float64   `gorm:"type:numeric(18,8);not null" json:"unitPrice"`
}

func (o *Order) BeforeCreate(_ *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	if o.Status == "" {
		o.Status = OrderStatusPendingPayment
	}
	return nil
}

func (oi *OrderItem) BeforeCreate(_ *gorm.DB) error {
	if oi.ID == uuid.Nil {
		oi.ID = uuid.New()
	}
	return nil
}
