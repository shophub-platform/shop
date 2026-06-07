package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RoleUser  = "USER"
	RoleAdmin = "ADMIN"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"        json:"id"`
	Email        string         `gorm:"uniqueIndex;not null;size:254" json:"email"`
	PasswordHash string         `gorm:"not null"                    json:"-"`
	Role         string         `gorm:"not null;default:'USER'"     json:"role"`
	CreatedAt    time.Time      `                                   json:"createdAt"`
	UpdatedAt    time.Time      `                                   json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                       json:"-"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
