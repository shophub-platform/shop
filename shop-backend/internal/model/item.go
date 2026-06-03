package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Item struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"           json:"id"`
	Name        string         `gorm:"not null;size:200"               json:"name"`
	Description *string        `gorm:"size:2000"                       json:"description,omitempty"`
	Price       float64        `gorm:"type:numeric(18,8);not null"     json:"price"`
	Stock       int            `gorm:"not null;check:stock >= 0"       json:"stock"`
	ImageURL    *string        `gorm:"column:image_url"                json:"imageUrl,omitempty"`
	CreatedAt   time.Time      `                                       json:"createdAt"`
	UpdatedAt   time.Time      `                                       json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                           json:"-"`
}

func (i *Item) BeforeCreate(_ *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
