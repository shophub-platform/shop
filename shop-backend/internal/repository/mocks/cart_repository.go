package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/shophub/shop/internal/model"
)

type CartRepository struct {
	mock.Mock
}

func (m *CartRepository) FindOrCreateByUserID(ctx context.Context, userID string) (*model.Cart, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cart), args.Error(1)
}

func (m *CartRepository) AddOrUpdateItem(ctx context.Context, cartID, itemID uuid.UUID, quantity int) (*model.Cart, error) {
	args := m.Called(ctx, cartID, itemID, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cart), args.Error(1)
}

func (m *CartRepository) SetItemQuantity(ctx context.Context, cartID, itemID uuid.UUID, quantity int) (*model.Cart, error) {
	args := m.Called(ctx, cartID, itemID, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cart), args.Error(1)
}

func (m *CartRepository) RemoveItem(ctx context.Context, cartID, itemID uuid.UUID) (*model.Cart, error) {
	args := m.Called(ctx, cartID, itemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cart), args.Error(1)
}

func (m *CartRepository) Clear(ctx context.Context, cartID uuid.UUID) error {
	return m.Called(ctx, cartID).Error(0)
}
