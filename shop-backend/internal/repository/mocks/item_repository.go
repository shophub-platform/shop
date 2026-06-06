package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

type ItemRepository struct {
	mock.Mock
}

func (m *ItemRepository) Create(ctx context.Context, item *model.Item) error {
	return m.Called(ctx, item).Error(0)
}

func (m *ItemRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Item), args.Error(1)
}

func (m *ItemRepository) List(ctx context.Context, filter repository.ItemFilter) ([]*model.Item, int64, error) {
	args := m.Called(ctx, filter)
	items, _ := args.Get(0).([]*model.Item)
	return items, args.Get(1).(int64), args.Error(2)
}

func (m *ItemRepository) Update(ctx context.Context, item *model.Item) error {
	return m.Called(ctx, item).Error(0)
}

func (m *ItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *ItemRepository) DecrementStock(ctx context.Context, id uuid.UUID, quantity int) error {
	return m.Called(ctx, id, quantity).Error(0)
}
