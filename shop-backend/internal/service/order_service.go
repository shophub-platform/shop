package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shophub/shop/internal/metrics"
	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

type OrderService struct {
	orderRepo repository.OrderRepository
	cartRepo  repository.CartRepository
	itemRepo  repository.ItemRepository
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	itemRepo repository.ItemRepository,
) *OrderService {
	return &OrderService{orderRepo: orderRepo, cartRepo: cartRepo, itemRepo: itemRepo}
}

type CreateOrderRequest struct {
	WalletFrom     *string `json:"walletFrom"`
	IdempotencyKey string  `json:"-"` // set from Idempotency-Key header, not request body
}

// CreateFromCart builds an order from the user's active cart:
// validates stock, decrements it atomically, snapshots prices, then clears the cart.
// If IdempotencyKey is set and an order with that key already exists, it is returned as-is.
func (s *OrderService) CreateFromCart(ctx context.Context, userID string, req CreateOrderRequest) (*model.Order, error) {
	if req.IdempotencyKey != "" {
		if existing, err := s.orderRepo.FindByIdempotencyKey(ctx, req.IdempotencyKey); err == nil {
			return existing, nil
		}
	}

	cart, err := s.cartRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	var total float64
	orderItems := make([]model.OrderItem, 0, len(cart.Items))

	for _, ci := range cart.Items {
		item, err := s.itemRepo.FindByID(ctx, ci.ItemID)
		if err != nil {
			return nil, err
		}
		if err := s.itemRepo.DecrementStock(ctx, item.ID, ci.Quantity); err != nil {
			if errors.Is(err, repository.ErrInsufficientStock) {
				return nil, fmt.Errorf("item %q: %w", item.Name, repository.ErrInsufficientStock)
			}
			return nil, err
		}
		total += item.Price * float64(ci.Quantity)
		orderItems = append(orderItems, model.OrderItem{
			ItemID:    item.ID,
			ItemName:  item.Name,
			Quantity:  ci.Quantity,
			UnitPrice: item.Price,
		})
	}

	var idemKey *string
	if req.IdempotencyKey != "" {
		idemKey = &req.IdempotencyKey
	}
	order := &model.Order{
		UserID:         userID,
		Items:          orderItems,
		Total:          total,
		WalletFrom:     req.WalletFrom,
		Status:         model.OrderStatusPendingPayment,
		IdempotencyKey: idemKey,
	}
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}
	if err := s.cartRepo.Clear(ctx, cart.ID); err != nil {
		return nil, err
	}
	metrics.OrdersCreatedTotal.Inc()
	return order, nil
}

func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	return s.orderRepo.FindByID(ctx, id)
}

func (s *OrderService) List(ctx context.Context, filter repository.OrderFilter) ([]*model.Order, int64, error) {
	return s.orderRepo.List(ctx, filter)
}

// ConfirmPayment is called by the blockchain listener after detecting a matching transaction.
func (s *OrderService) ConfirmPayment(ctx context.Context, orderID uuid.UUID, txHash string) (*model.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.Status != model.OrderStatusPendingPayment {
		return nil, fmt.Errorf("order is not pending payment (current status: %s)", order.Status)
	}
	if err := s.orderRepo.UpdateStatus(ctx, orderID, model.OrderStatusPaid, &txHash); err != nil {
		return nil, err
	}
	metrics.PaymentProcessingDuration.Observe(time.Since(order.CreatedAt).Seconds())
	return s.orderRepo.FindByID(ctx, orderID)
}

// UpdateStatus allows admins to advance an order through its lifecycle.
func (s *OrderService) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus model.OrderStatus) (*model.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !order.Status.CanTransitionTo(newStatus) {
		return nil, fmt.Errorf("cannot transition from %s to %s", order.Status, newStatus)
	}
	if err := s.orderRepo.UpdateStatus(ctx, id, newStatus, nil); err != nil {
		return nil, err
	}
	return s.orderRepo.FindByID(ctx, id)
}
