package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

// Key schema:
//   cart:user:{userID}        — STRING holding the cart UUID for that user
//   cart:{cartID}             — HASH with cart metadata (id, user_id, updated_at)
//   cart:{cartID}:items       — HASH (key=itemID, value=JSON(CartItem))

func cartKey(cartID uuid.UUID) string      { return "cart:" + cartID.String() }
func cartItemsKey(cartID uuid.UUID) string { return "cart:" + cartID.String() + ":items" }
func userCartKey(userID string) string     { return "cart:user:" + userID }

type cartRepository struct {
	rdb *redis.Client
}

func NewCartRepository(rdb *redis.Client) repository.CartRepository {
	return &cartRepository{rdb: rdb}
}

func (r *cartRepository) FindOrCreateByUserID(ctx context.Context, userID string) (*model.Cart, error) {
	cartIDStr, err := r.rdb.Get(ctx, userCartKey(userID)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	var cartID uuid.UUID
	if cartIDStr != "" {
		cartID, err = uuid.Parse(cartIDStr)
		if err != nil {
			return nil, err
		}
	} else {
		cartID = uuid.New()
		now := time.Now().UTC().Format(time.RFC3339)
		pipe := r.rdb.Pipeline()
		pipe.Set(ctx, userCartKey(userID), cartID.String(), 0)
		pipe.HSet(ctx, cartKey(cartID), "id", cartID.String(), "user_id", userID, "updated_at", now)
		if _, err := pipe.Exec(ctx); err != nil {
			return nil, err
		}
	}

	return r.loadCart(ctx, cartID, userID)
}

func (r *cartRepository) AddOrUpdateItem(ctx context.Context, cartID, itemID uuid.UUID, quantity int) (*model.Cart, error) {
	key := cartItemsKey(cartID)

	existing, err := r.rdb.HGet(ctx, key, itemID.String()).Result()

	var ci model.CartItem
	if err == redis.Nil {
		ci = model.CartItem{ID: uuid.New(), CartID: cartID, ItemID: itemID, Quantity: quantity}
	} else if err != nil {
		return nil, err
	} else {
		if err := json.Unmarshal([]byte(existing), &ci); err != nil {
			return nil, err
		}
		ci.Quantity += quantity
	}

	data, err := json.Marshal(ci)
	if err != nil {
		return nil, err
	}

	pipe := r.rdb.Pipeline()
	pipe.HSet(ctx, key, itemID.String(), string(data))
	pipe.HSet(ctx, cartKey(cartID), "updated_at", time.Now().UTC().Format(time.RFC3339))
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	return r.loadCart(ctx, cartID, r.userIDForCart(ctx, cartID))
}

func (r *cartRepository) SetItemQuantity(ctx context.Context, cartID, itemID uuid.UUID, quantity int) (*model.Cart, error) {
	key := cartItemsKey(cartID)

	existing, err := r.rdb.HGet(ctx, key, itemID.String()).Result()
	if err == redis.Nil {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var ci model.CartItem
	if err := json.Unmarshal([]byte(existing), &ci); err != nil {
		return nil, err
	}
	ci.Quantity = quantity

	data, err := json.Marshal(ci)
	if err != nil {
		return nil, err
	}

	if err := r.rdb.HSet(ctx, key, itemID.String(), string(data)).Err(); err != nil {
		return nil, err
	}

	return r.loadCart(ctx, cartID, r.userIDForCart(ctx, cartID))
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartID, itemID uuid.UUID) (*model.Cart, error) {
	removed, err := r.rdb.HDel(ctx, cartItemsKey(cartID), itemID.String()).Result()
	if err != nil {
		return nil, err
	}
	if removed == 0 {
		return nil, repository.ErrNotFound
	}
	return r.loadCart(ctx, cartID, r.userIDForCart(ctx, cartID))
}

func (r *cartRepository) Clear(ctx context.Context, cartID uuid.UUID) error {
	return r.rdb.Del(ctx, cartItemsKey(cartID)).Err()
}

// --- helpers ---

func (r *cartRepository) userIDForCart(ctx context.Context, cartID uuid.UUID) string {
	userID, _ := r.rdb.HGet(ctx, cartKey(cartID), "user_id").Result()
	return userID
}

func (r *cartRepository) loadCart(ctx context.Context, cartID uuid.UUID, userID string) (*model.Cart, error) {
	itemsMap, err := r.rdb.HGetAll(ctx, cartItemsKey(cartID)).Result()
	if err != nil {
		return nil, err
	}

	cart := &model.Cart{ID: cartID, UserID: userID, Items: []model.CartItem{}}
	for _, jsonData := range itemsMap {
		var ci model.CartItem
		if err := json.Unmarshal([]byte(jsonData), &ci); err != nil {
			continue
		}
		cart.Items = append(cart.Items, ci)
	}
	return cart, nil
}
