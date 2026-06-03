package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shophub/shop/internal/model"
)

type OrderFilter struct {
	UserID   *string
	Status   *model.OrderStatus
	Page     int
	PageSize int
}

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	List(ctx context.Context, filter OrderFilter) ([]*model.Order, int64, error)
	// UpdateStatus sets the new status and optionally saves the txHash.
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus, txHash *string) error
	FindByTxHash(ctx context.Context, txHash string) (*model.Order, error)
}
