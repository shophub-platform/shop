package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Cart struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"                         json:"id"`
	UserID    string     `gorm:"not null;uniqueIndex"                         json:"userId"`
	Items     []CartItem `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE" json:"items"`
	UpdatedAt time.Time  `                                                    json:"updatedAt"`
}

// CartItem has a composite unique constraint on (cart_id, item_id)
// so the same item can't appear twice in the same cart.
type CartItem struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey"                      json:"id"`
	CartID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uidx_cart_item" json:"-"`
	ItemID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:uidx_cart_item" json:"itemId"`
	Item     Item      `gorm:"foreignKey:ItemID"                         json:"item"`
	Quantity int       `gorm:"not null;check:quantity > 0"               json:"quantity"`
}

func (c *Cart) BeforeCreate(_ *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (ci *CartItem) BeforeCreate(_ *gorm.DB) error {
	if ci.ID == uuid.Nil {
		ci.ID = uuid.New()
	}
	return nil
}
