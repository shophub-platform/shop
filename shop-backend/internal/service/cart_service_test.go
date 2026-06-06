package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
	"github.com/shophub/shop/internal/repository/mocks"
	"github.com/shophub/shop/internal/service"
)

func newCartSvc(t *testing.T) (*service.CartService, *mocks.CartRepository, *mocks.ItemRepository) {
	t.Helper()
	cartRepo := new(mocks.CartRepository)
	itemRepo := new(mocks.ItemRepository)
	return service.NewCartService(cartRepo, itemRepo), cartRepo, itemRepo
}

func TestCartService_GetCart(t *testing.T) {
	svc, cartRepo, _ := newCartSvc(t)

	userID := "user-1"
	expected := &model.Cart{ID: uuid.New(), UserID: userID}
	cartRepo.On("FindOrCreateByUserID", mock.Anything, userID).Return(expected, nil)

	cart, err := svc.GetCart(context.Background(), userID)

	require.NoError(t, err)
	assert.Equal(t, expected, cart)
	cartRepo.AssertExpectations(t)
}

func TestCartService_AddItem_Success(t *testing.T) {
	svc, cartRepo, itemRepo := newCartSvc(t)

	userID := "user-1"
	itemID := uuid.New()
	cartID := uuid.New()

	itemRepo.On("FindByID", mock.Anything, itemID).Return(&model.Item{ID: itemID}, nil)
	cartRepo.On("FindOrCreateByUserID", mock.Anything, userID).Return(&model.Cart{ID: cartID}, nil)
	cartRepo.On("AddOrUpdateItem", mock.Anything, cartID, itemID, 3).Return(
		&model.Cart{ID: cartID, Items: []model.CartItem{{ItemID: itemID, Quantity: 3}}}, nil,
	)

	cart, err := svc.AddItem(context.Background(), userID, itemID, 3)

	require.NoError(t, err)
	assert.Len(t, cart.Items, 1)
	assert.Equal(t, 3, cart.Items[0].Quantity)
	cartRepo.AssertExpectations(t)
	itemRepo.AssertExpectations(t)
}

func TestCartService_AddItem_ItemNotFound(t *testing.T) {
	svc, cartRepo, itemRepo := newCartSvc(t)

	itemID := uuid.New()
	itemRepo.On("FindByID", mock.Anything, itemID).Return(nil, repository.ErrNotFound)

	_, err := svc.AddItem(context.Background(), "user-1", itemID, 1)

	assert.ErrorIs(t, err, repository.ErrNotFound)
	cartRepo.AssertNotCalled(t, "FindOrCreateByUserID")
}

func TestCartService_SetItemQuantity(t *testing.T) {
	svc, cartRepo, _ := newCartSvc(t)

	userID := "user-1"
	itemID := uuid.New()
	cartID := uuid.New()

	cartRepo.On("FindOrCreateByUserID", mock.Anything, userID).Return(&model.Cart{ID: cartID}, nil)
	cartRepo.On("SetItemQuantity", mock.Anything, cartID, itemID, 7).Return(
		&model.Cart{ID: cartID, Items: []model.CartItem{{ItemID: itemID, Quantity: 7}}}, nil,
	)

	cart, err := svc.SetItemQuantity(context.Background(), userID, itemID, 7)

	require.NoError(t, err)
	assert.Equal(t, 7, cart.Items[0].Quantity)
	cartRepo.AssertExpectations(t)
}

func TestCartService_RemoveItem(t *testing.T) {
	svc, cartRepo, _ := newCartSvc(t)

	userID := "user-1"
	itemID := uuid.New()
	cartID := uuid.New()

	cartRepo.On("FindOrCreateByUserID", mock.Anything, userID).Return(&model.Cart{ID: cartID}, nil)
	cartRepo.On("RemoveItem", mock.Anything, cartID, itemID).Return(
		&model.Cart{ID: cartID, Items: []model.CartItem{}}, nil,
	)

	cart, err := svc.RemoveItem(context.Background(), userID, itemID)

	require.NoError(t, err)
	assert.Empty(t, cart.Items)
	cartRepo.AssertExpectations(t)
}

func TestCartService_ClearCart(t *testing.T) {
	svc, cartRepo, _ := newCartSvc(t)

	userID := "user-1"
	cartID := uuid.New()

	cartRepo.On("FindOrCreateByUserID", mock.Anything, userID).Return(&model.Cart{ID: cartID}, nil)
	cartRepo.On("Clear", mock.Anything, cartID).Return(nil)

	err := svc.ClearCart(context.Background(), userID)

	require.NoError(t, err)
	cartRepo.AssertExpectations(t)
}
