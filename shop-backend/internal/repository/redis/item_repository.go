package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

// Key schema:
//   item:{id}         — HASH with all item fields
//   items:index       — ZSET (score=created_at_unix, member=id) for ordered listing
//   items:deleted     — SET of soft-deleted item IDs

const (
	itemsIndexKey   = "items:index"
	itemsDeletedKey = "items:deleted"
)

func itemKey(id uuid.UUID) string { return "item:" + id.String() }

type itemRepository struct {
	rdb *redis.Client
}

func NewItemRepository(rdb *redis.Client) repository.ItemRepository {
	return &itemRepository{rdb: rdb}
}

func (r *itemRepository) Create(ctx context.Context, item *model.Item) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now

	pipe := r.rdb.Pipeline()
	pipe.HSet(ctx, itemKey(item.ID), itemToFields(item))
	pipe.ZAdd(ctx, itemsIndexKey, redis.Z{
		Score:  float64(now.Unix()),
		Member: item.ID.String(),
	})
	_, err := pipe.Exec(ctx)
	return err
}

func (r *itemRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Item, error) {
	deleted, err := r.rdb.SIsMember(ctx, itemsDeletedKey, id.String()).Result()
	if err != nil {
		return nil, err
	}
	if deleted {
		return nil, repository.ErrNotFound
	}
	fields, err := r.rdb.HGetAll(ctx, itemKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, repository.ErrNotFound
	}
	return itemFromFields(fields)
}

func (r *itemRepository) List(ctx context.Context, filter repository.ItemFilter) ([]*model.Item, int64, error) {
	ids, err := r.rdb.ZRevRange(ctx, itemsIndexKey, 0, -1).Result()
	if err != nil {
		return nil, 0, err
	}

	var matched []*model.Item
	for _, idStr := range ids {
		deleted, _ := r.rdb.SIsMember(ctx, itemsDeletedKey, idStr).Result()
		if deleted {
			continue
		}
		fields, err := r.rdb.HGetAll(ctx, "item:"+idStr).Result()
		if err != nil || len(fields) == 0 {
			continue
		}
		item, err := itemFromFields(fields)
		if err != nil {
			continue
		}
		if filter.Search != "" && !containsInsensitive(item.Name, filter.Search) {
			continue
		}
		if filter.MinPrice != nil && item.Price < *filter.MinPrice {
			continue
		}
		if filter.MaxPrice != nil && item.Price > *filter.MaxPrice {
			continue
		}
		if filter.InStock != nil && *filter.InStock && item.Stock <= 0 {
			continue
		}
		matched = append(matched, item)
	}

	total := int64(len(matched))

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(matched) {
		return []*model.Item{}, total, nil
	}
	end := start + pageSize
	if end > len(matched) {
		end = len(matched)
	}
	return matched[start:end], total, nil
}

func (r *itemRepository) Update(ctx context.Context, item *model.Item) error {
	exists, err := r.rdb.Exists(ctx, itemKey(item.ID)).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return repository.ErrNotFound
	}
	item.UpdatedAt = time.Now().UTC()
	return r.rdb.HSet(ctx, itemKey(item.ID), itemToFields(item)).Err()
}

func (r *itemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	exists, err := r.rdb.Exists(ctx, itemKey(id)).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return repository.ErrNotFound
	}
	pipe := r.rdb.Pipeline()
	pipe.SAdd(ctx, itemsDeletedKey, id.String())
	pipe.ZRem(ctx, itemsIndexKey, id.String())
	_, err = pipe.Exec(ctx)
	return err
}

// decrementStockScript atomically checks stock >= quantity and decrements.
// Returns an error reply if the item is not found or stock is insufficient.
var decrementStockScript = redis.NewScript(`
local stock = redis.call('HGET', KEYS[1], 'stock')
if not stock then return redis.error_reply('not found') end
local s = tonumber(stock)
local q = tonumber(ARGV[1])
if s < q then return redis.error_reply('insufficient stock') end
redis.call('HSET', KEYS[1], 'stock', tostring(s - q))
return s - q
`)

func (r *itemRepository) DecrementStock(ctx context.Context, id uuid.UUID, quantity int) error {
	err := decrementStockScript.Run(ctx, r.rdb, []string{itemKey(id)}, quantity).Err()
	if err == nil {
		return nil
	}
	switch err.Error() {
	case "not found":
		return repository.ErrNotFound
	case "insufficient stock":
		return repository.ErrInsufficientStock
	default:
		return err
	}
}

// --- serialization helpers ---

func itemToFields(item *model.Item) []any {
	return []any{
		"id", item.ID.String(),
		"name", item.Name,
		"price", strconv.FormatFloat(item.Price, 'f', 8, 64),
		"stock", strconv.Itoa(item.Stock),
		"description", ptrToStr(item.Description),
		"image_url", ptrToStr(item.ImageURL),
		"created_at", item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at", item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func itemFromFields(f map[string]string) (*model.Item, error) {
	id, err := uuid.Parse(f["id"])
	if err != nil {
		return nil, fmt.Errorf("parse item id: %w", err)
	}
	price, _ := strconv.ParseFloat(f["price"], 64)
	stock, _ := strconv.Atoi(f["stock"])
	createdAt, _ := time.Parse(time.RFC3339, f["created_at"])
	updatedAt, _ := time.Parse(time.RFC3339, f["updated_at"])

	return &model.Item{
		ID:          id,
		Name:        f["name"],
		Price:       price,
		Stock:       stock,
		Description: strToPtr(f["description"]),
		ImageURL:    strToPtr(f["image_url"]),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
