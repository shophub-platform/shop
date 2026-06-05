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

func newOrderSvc(t *testing.T) (*service.OrderService, *mocks.OrderRepository, *mocks.CartRepository, *mocks.ItemRepository) {
	t.Helper()
	orderRepo := new(mocks.OrderRepository)
	cartRepo := new(mocks.CartRepository)
	itemRepo := new(mocks.ItemRepository)
	return service.NewOrderService(orderRepo, cartRepo, itemRepo), orderRepo, cartRepo, itemRepo
}

func TestOrderService_CreateFromCart_Success(t *testing.T) {
	svc, orderRepo, cartRepo, itemRepo := newOrderSvc(t)

	userID := "user-1"
	itemID := uuid.New()
	cartID := uuid.New()

	cartRepo.On("FindOrCreateByUserID", mock.Anything, userID).Return(&model.Cart{
		ID:    cartID,
		Items: []model.CartItem{{ID: uuid.New(), ItemID: itemID, Quantity: 2}},
	}, nil)
	itemRepo.On("FindByID", mock.Anything, itemID).Return(&model.Item{
		ID: itemID, Name: "Book", Price: 15.0, Stock: 10,
	}, nil)
	itemRepo.On("DecrementStock", mock.Anything, itemID, 2).Return(nil)
	orderRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Order")).Return(nil)
	cartRepo.On("Clear", mock.Anything, cartID).Return(nil)

	order, err := svc.CreateFromCart(context.Background(), userID, service.CreateOrderRequest{})

	require.NoError(t, err)
	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, 30.0, order.Total) // 15.0 * 2
	assert.Len(t, order.Items, 1)
	assert.Equal(t, "Book", order.Items[0].ItemName)
	assert.Equal(t, 15.0, order.Items[0].UnitPrice)
	assert.Equal(t, model.OrderStatusPendingPayment, order.Status)
	orderRepo.AssertExpectations(t)
	cartRepo.AssertExpectations(t)
	itemRepo.AssertExpectations(t)
}

func TestOrderService_CreateFromCart_WalletFromIsStoredOnOrder(t *testing.T) {
	svc, orderRepo, cartRepo, itemRepo := newOrderSvc(t)

	userID := "user-1"
	itemID := uuid.New()
	cartID := uuid.New()
	wallet := "0xDEADBEEF"

	cartRepo.On("FindOrCreateByUserID", mock.Anything, userID).Return(&model.Cart{
		ID:    cartID,
		Items: []model.CartItem{{ItemID: itemID, Quantity: 1}},
	}, nil)
	itemRepo.On("FindByID", mock.Anything, itemID).Return(&model.Item{ID: itemID, Name: "X", Price: 5.0, Stock: 5}, nil)
	itemRepo.On("DecrementStock", mock.Anything, itemID, 1).Return(nil)
	orderRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Order")).Return(nil)
	cartRepo.On("Clear", mock.Anything, cartID).Return(nil)

	order, err := svc.CreateFromCart(context.Background(), userID, service.CreateOrderRequest{WalletFrom: &wallet})

	require.NoError(t, err)
	require.NotNil(t, order.WalletFrom)
	assert.Equal(t, wallet, *order.WalletFrom)
}

func TestOrderService_CreateFromCart_EmptyCart(t *testing.T) {
	svc, _, cartRepo, _ := newOrderSvc(t)

	cartRepo.On("FindOrCreateByUserID", mock.Anything, "user-1").Return(
		&model.Cart{ID: uuid.New(), Items: []model.CartItem{}}, nil,
	)

	_, err := svc.CreateFromCart(context.Background(), "user-1", service.CreateOrderRequest{})

	assert.EqualError(t, err, "cart is empty")
}

func TestOrderService_CreateFromCart_InsufficientStock(t *testing.T) {
	svc, _, cartRepo, itemRepo := newOrderSvc(t)

	itemID := uuid.New()
	cartRepo.On("FindOrCreateByUserID", mock.Anything, "user-1").Return(&model.Cart{
		ID:    uuid.New(),
		Items: []model.CartItem{{ItemID: itemID, Quantity: 100}},
	}, nil)
	itemRepo.On("FindByID", mock.Anything, itemID).Return(
		&model.Item{ID: itemID, Name: "Rare", Price: 5.0, Stock: 1}, nil,
	)
	itemRepo.On("DecrementStock", mock.Anything, itemID, 100).Return(repository.ErrInsufficientStock)

	_, err := svc.CreateFromCart(context.Background(), "user-1", service.CreateOrderRequest{})

	assert.ErrorIs(t, err, repository.ErrInsufficientStock)
}

func TestOrderService_CreateFromCart_ItemNotFound(t *testing.T) {
	svc, _, cartRepo, itemRepo := newOrderSvc(t)

	itemID := uuid.New()
	cartRepo.On("FindOrCreateByUserID", mock.Anything, "user-1").Return(&model.Cart{
		ID:    uuid.New(),
		Items: []model.CartItem{{ItemID: itemID, Quantity: 1}},
	}, nil)
	itemRepo.On("FindByID", mock.Anything, itemID).Return(nil, repository.ErrNotFound)

	_, err := svc.CreateFromCart(context.Background(), "user-1", service.CreateOrderRequest{})

	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestOrderService_ConfirmPayment_Success(t *testing.T) {
	svc, orderRepo, _, _ := newOrderSvc(t)

	orderID := uuid.New()
	txHash := "0xabc123"
	paid := &model.Order{ID: orderID, Status: model.OrderStatusPaid, TxHash: &txHash}

	orderRepo.On("FindByID", mock.Anything, orderID).Return(
		&model.Order{ID: orderID, Status: model.OrderStatusPendingPayment}, nil,
	).Once()
	orderRepo.On("UpdateStatus", mock.Anything, orderID, model.OrderStatusPaid, &txHash).Return(nil)
	orderRepo.On("FindByID", mock.Anything, orderID).Return(paid, nil).Once()

	result, err := svc.ConfirmPayment(context.Background(), orderID, txHash)

	require.NoError(t, err)
	assert.Equal(t, model.OrderStatusPaid, result.Status)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_ConfirmPayment_WrongStatus(t *testing.T) {
	svc, orderRepo, _, _ := newOrderSvc(t)

	orderID := uuid.New()
	orderRepo.On("FindByID", mock.Anything, orderID).Return(
		&model.Order{ID: orderID, Status: model.OrderStatusPaid}, nil,
	)

	_, err := svc.ConfirmPayment(context.Background(), orderID, "0xabc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not pending payment")
}

func TestOrderService_ConfirmPayment_NotFound(t *testing.T) {
	svc, orderRepo, _, _ := newOrderSvc(t)

	orderID := uuid.New()
	orderRepo.On("FindByID", mock.Anything, orderID).Return(nil, repository.ErrNotFound)

	_, err := svc.ConfirmPayment(context.Background(), orderID, "0xabc")

	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestOrderService_UpdateStatus_ValidTransition(t *testing.T) {
	svc, orderRepo, _, _ := newOrderSvc(t)

	orderID := uuid.New()
	processing := &model.Order{ID: orderID, Status: model.OrderStatusProcessing}

	orderRepo.On("FindByID", mock.Anything, orderID).Return(
		&model.Order{ID: orderID, Status: model.OrderStatusPaid}, nil,
	).Once()
	orderRepo.On("UpdateStatus", mock.Anything, orderID, model.OrderStatusProcessing, (*string)(nil)).Return(nil)
	orderRepo.On("FindByID", mock.Anything, orderID).Return(processing, nil).Once()

	result, err := svc.UpdateStatus(context.Background(), orderID, model.OrderStatusProcessing)

	require.NoError(t, err)
	assert.Equal(t, model.OrderStatusProcessing, result.Status)
	orderRepo.AssertExpectations(t)
}

func TestOrderService_UpdateStatus_InvalidTransition(t *testing.T) {
	svc, orderRepo, _, _ := newOrderSvc(t)

	orderID := uuid.New()
	orderRepo.On("FindByID", mock.Anything, orderID).Return(
		&model.Order{ID: orderID, Status: model.OrderStatusPendingPayment}, nil,
	)

	_, err := svc.UpdateStatus(context.Background(), orderID, model.OrderStatusDelivered)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot transition")
}

func TestOrderService_List(t *testing.T) {
	svc, orderRepo, _, _ := newOrderSvc(t)

	filter := repository.OrderFilter{Page: 1, PageSize: 10}
	expected := []*model.Order{{ID: uuid.New()}, {ID: uuid.New()}}
	orderRepo.On("List", mock.Anything, filter).Return(expected, int64(2), nil)

	orders, total, err := svc.List(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, orders, 2)
	orderRepo.AssertExpectations(t)
}
