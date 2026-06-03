package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shophub/shop/internal/model"
)

type ItemFilter struct {
	Search   string
	MinPrice *float64
	MaxPrice *float64
	InStock  *bool
	Page     int
	PageSize int
}

type ItemRepository interface {
	Create(ctx context.Context, item *model.Item) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Item, error)
	List(ctx context.Context, filter ItemFilter) ([]*model.Item, int64, error)
	Update(ctx context.Context, item *model.Item) error
	Delete(ctx context.Context, id uuid.UUID) error
	// DecrementStock atomically reduces stock by quantity.
	// Returns ErrInsufficientStock if stock < quantity.
	DecrementStock(ctx context.Context, id uuid.UUID, quantity int) error
}
