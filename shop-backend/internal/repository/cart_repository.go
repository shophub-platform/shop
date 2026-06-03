package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shophub/shop/internal/model"
)

type CartRepository interface {
	// FindOrCreateByUserID returns the user's cart, creating it if it doesn't exist.
	FindOrCreateByUserID(ctx context.Context, userID string) (*model.Cart, error)
	// AddOrUpdateItem adds the item to the cart. If the item is already present,
	// its quantity is incremented by the given amount.
	AddOrUpdateItem(ctx context.Context, cartID uuid.UUID, itemID uuid.UUID, quantity int) (*model.Cart, error)
	// SetItemQuantity replaces the quantity for an existing cart item.
	SetItemQuantity(ctx context.Context, cartID uuid.UUID, itemID uuid.UUID, quantity int) (*model.Cart, error)
	// RemoveItem removes a single item from the cart.
	RemoveItem(ctx context.Context, cartID uuid.UUID, itemID uuid.UUID) (*model.Cart, error)
	// Clear removes all items from the cart (cart record itself remains).
	Clear(ctx context.Context, cartID uuid.UUID) error
}
