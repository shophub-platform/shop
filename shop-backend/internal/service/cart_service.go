package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

type CartService struct {
	cartRepo repository.CartRepository
	itemRepo repository.ItemRepository
}

func NewCartService(cartRepo repository.CartRepository, itemRepo repository.ItemRepository) *CartService {
	return &CartService{cartRepo: cartRepo, itemRepo: itemRepo}
}

func (s *CartService) GetCart(ctx context.Context, userID string) (*model.Cart, error) {
	cart, err := s.cartRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.enrichCart(ctx, cart), nil
}

func (s *CartService) AddItem(ctx context.Context, userID string, itemID uuid.UUID, quantity int) (*model.Cart, error) {
	if _, err := s.itemRepo.FindByID(ctx, itemID); err != nil {
		return nil, err
	}
	cart, err := s.cartRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	cart, err = s.cartRepo.AddOrUpdateItem(ctx, cart.ID, itemID, quantity)
	if err != nil {
		return nil, err
	}
	return s.enrichCart(ctx, cart), nil
}

func (s *CartService) SetItemQuantity(ctx context.Context, userID string, itemID uuid.UUID, quantity int) (*model.Cart, error) {
	cart, err := s.cartRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	cart, err = s.cartRepo.SetItemQuantity(ctx, cart.ID, itemID, quantity)
	if err != nil {
		return nil, err
	}
	return s.enrichCart(ctx, cart), nil
}

func (s *CartService) RemoveItem(ctx context.Context, userID string, itemID uuid.UUID) (*model.Cart, error) {
	cart, err := s.cartRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	cart, err = s.cartRepo.RemoveItem(ctx, cart.ID, itemID)
	if err != nil {
		return nil, err
	}
	return s.enrichCart(ctx, cart), nil
}

func (s *CartService) enrichCart(ctx context.Context, cart *model.Cart) *model.Cart {
	for i, ci := range cart.Items {
		item, err := s.itemRepo.FindByID(ctx, ci.ItemID)
		if err != nil {
			continue
		}
		cart.Items[i].Item = *item
	}
	return cart
}

func (s *CartService) ClearCart(ctx context.Context, userID string) error {
	cart, err := s.cartRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.cartRepo.Clear(ctx, cart.ID)
}
