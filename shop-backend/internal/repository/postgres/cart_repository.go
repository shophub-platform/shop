package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) repository.CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) FindOrCreateByUserID(ctx context.Context, userID string) (*model.Cart, error) {
	var cart model.Cart
	err := r.db.WithContext(ctx).
		Preload("Items.Item").
		Where("user_id = ?", userID).
		First(&cart).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		cart = model.Cart{UserID: userID, Items: []model.CartItem{}}
		if createErr := r.db.WithContext(ctx).Create(&cart).Error; createErr != nil {
			return nil, createErr
		}
		return &cart, nil
	}
	return &cart, err
}

func (r *cartRepository) AddOrUpdateItem(ctx context.Context, cartID uuid.UUID, itemID uuid.UUID, quantity int) (*model.Cart, error) {
	// Upsert: if the item is already in the cart, increment quantity.
	// Otherwise insert a new CartItem.
	cartItem := model.CartItem{
		CartID:   cartID,
		ItemID:   itemID,
		Quantity: quantity,
	}

	err := r.db.WithContext(ctx).
		Where(model.CartItem{CartID: cartID, ItemID: itemID}).
		Assign(model.CartItem{Quantity: quantity}).
		FirstOrCreate(&cartItem).Error
	if err != nil {
		return nil, err
	}

	// If record already existed, increment rather than replace.
	if cartItem.Quantity != quantity {
		err = r.db.WithContext(ctx).
			Model(&cartItem).
			Update("quantity", gorm.Expr("quantity + ?", quantity)).Error
		if err != nil {
			return nil, err
		}
	}

	return r.loadCart(ctx, cartID)
}

func (r *cartRepository) SetItemQuantity(ctx context.Context, cartID uuid.UUID, itemID uuid.UUID, quantity int) (*model.Cart, error) {
	result := r.db.WithContext(ctx).
		Model(&model.CartItem{}).
		Where("cart_id = ? AND item_id = ?", cartID, itemID).
		Update("quantity", quantity)

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, repository.ErrNotFound
	}

	return r.loadCart(ctx, cartID)
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartID uuid.UUID, itemID uuid.UUID) (*model.Cart, error) {
	result := r.db.WithContext(ctx).
		Where("cart_id = ? AND item_id = ?", cartID, itemID).
		Delete(&model.CartItem{})

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, repository.ErrNotFound
	}

	return r.loadCart(ctx, cartID)
}

func (r *cartRepository) Clear(ctx context.Context, cartID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("cart_id = ?", cartID).
		Delete(&model.CartItem{}).Error
}

func (r *cartRepository) loadCart(ctx context.Context, cartID uuid.UUID) (*model.Cart, error) {
	var cart model.Cart
	err := r.db.WithContext(ctx).
		Preload(clause.Associations).
		Preload("Items.Item").
		First(&cart, "id = ?", cartID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	return &cart, err
}
