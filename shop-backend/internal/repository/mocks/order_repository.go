package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

type OrderRepository struct {
	mock.Mock
}

func (m *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	return m.Called(ctx, order).Error(0)
}

func (m *OrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *OrderRepository) List(ctx context.Context, filter repository.OrderFilter) ([]*model.Order, int64, error) {
	args := m.Called(ctx, filter)
	orders, _ := args.Get(0).([]*model.Order)
	return orders, args.Get(1).(int64), args.Error(2)
}

func (m *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus, txHash *string) error {
	return m.Called(ctx, id, status, txHash).Error(0)
}

func (m *OrderRepository) FindByTxHash(ctx context.Context, txHash string) (*model.Order, error) {
	args := m.Called(ctx, txHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *OrderRepository) FindByIdempotencyKey(ctx context.Context, key string) (*model.Order, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}
