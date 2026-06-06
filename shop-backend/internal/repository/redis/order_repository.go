package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

// Key schema:
//   order:{id}                    — HASH with order fields
//   order:{id}:items              — STRING (JSON array of OrderItem)
//   orders:user:{userID}          — ZSET (score=created_at_unix, member=order_id)
//   orders:all                    — ZSET (score=created_at_unix, member=order_id)
//   orders:txhash:{hash}          — STRING (order_id)

const ordersAllKey = "orders:all"

func orderKey(id uuid.UUID) string          { return "order:" + id.String() }
func orderItemsKey(id uuid.UUID) string     { return "order:" + id.String() + ":items" }
func ordersByUserKey(userID string) string  { return "orders:user:" + userID }
func orderByTxHashKey(txHash string) string { return "orders:txhash:" + txHash }

type orderRepository struct {
	rdb *redis.Client
}

func NewOrderRepository(rdb *redis.Client) repository.OrderRepository {
	return &orderRepository{rdb: rdb}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	now := time.Now().UTC()
	order.CreatedAt = now
	order.UpdatedAt = now
	if order.Status == "" {
		order.Status = model.OrderStatusPendingPayment
	}

	// Ensure each OrderItem has an ID.
	for i := range order.Items {
		if order.Items[i].ID == uuid.Nil {
			order.Items[i].ID = uuid.New()
		}
	}

	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return err
	}

	score := float64(now.Unix())
	idStr := order.ID.String()

	pipe := r.rdb.Pipeline()
	pipe.HSet(ctx, orderKey(order.ID), orderToFields(order))
	pipe.Set(ctx, orderItemsKey(order.ID), string(itemsJSON), 0)
	pipe.ZAdd(ctx, ordersByUserKey(order.UserID), redis.Z{Score: score, Member: idStr})
	pipe.ZAdd(ctx, ordersAllKey, redis.Z{Score: score, Member: idStr})
	if order.IdempotencyKey != nil && *order.IdempotencyKey != "" {
		pipe.Set(ctx, "idem:"+*order.IdempotencyKey, idStr, 0)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (r *orderRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	fields, err := r.rdb.HGetAll(ctx, orderKey(id)).Result()
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, repository.ErrNotFound
	}

	order, err := orderFromFields(fields)
	if err != nil {
		return nil, err
	}

	itemsJSON, err := r.rdb.Get(ctx, orderItemsKey(id)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	if itemsJSON != "" {
		if err := json.Unmarshal([]byte(itemsJSON), &order.Items); err != nil {
			return nil, err
		}
	}

	return order, nil
}

func (r *orderRepository) List(ctx context.Context, filter repository.OrderFilter) ([]*model.Order, int64, error) {
	var indexKey string
	if filter.UserID != nil {
		indexKey = ordersByUserKey(*filter.UserID)
	} else {
		indexKey = ordersAllKey
	}

	ids, err := r.rdb.ZRevRange(ctx, indexKey, 0, -1).Result()
	if err != nil {
		return nil, 0, err
	}

	var matched []*model.Order
	for _, idStr := range ids {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		order, err := r.FindByID(ctx, id)
		if err != nil {
			continue
		}
		if filter.Status != nil && order.Status != *filter.Status {
			continue
		}
		matched = append(matched, order)
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
		return []*model.Order{}, total, nil
	}
	end := start + pageSize
	if end > len(matched) {
		end = len(matched)
	}
	return matched[start:end], total, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus, txHash *string) error {
	exists, err := r.rdb.Exists(ctx, orderKey(id)).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return repository.ErrNotFound
	}

	fields := []any{
		"status", string(status),
		"updated_at", time.Now().UTC().Format(time.RFC3339),
	}
	if txHash != nil {
		fields = append(fields, "tx_hash", *txHash)
		if err := r.rdb.Set(ctx, orderByTxHashKey(*txHash), id.String(), 0).Err(); err != nil {
			return err
		}
	}

	return r.rdb.HSet(ctx, orderKey(id), fields...).Err()
}

func (r *orderRepository) FindByTxHash(ctx context.Context, txHash string) (*model.Order, error) {
	idStr, err := r.rdb.Get(ctx, orderByTxHashKey(txHash)).Result()
	if err == redis.Nil {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *orderRepository) FindByIdempotencyKey(ctx context.Context, key string) (*model.Order, error) {
	idStr, err := r.rdb.Get(ctx, "idem:"+key).Result()
	if err == redis.Nil {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

// --- serialization helpers ---

func orderToFields(o *model.Order) []any {
	return []any{
		"id", o.ID.String(),
		"user_id", o.UserID,
		"total", strconv.FormatFloat(o.Total, 'f', 8, 64),
		"status", string(o.Status),
		"wallet_from", ptrToStr(o.WalletFrom),
		"tx_hash", ptrToStr(o.TxHash),
		"idempotency_key", ptrToStr(o.IdempotencyKey),
		"created_at", o.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at", o.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func orderFromFields(f map[string]string) (*model.Order, error) {
	id, err := uuid.Parse(f["id"])
	if err != nil {
		return nil, fmt.Errorf("parse order id: %w", err)
	}
	total, _ := strconv.ParseFloat(f["total"], 64)
	createdAt, _ := time.Parse(time.RFC3339, f["created_at"])
	updatedAt, _ := time.Parse(time.RFC3339, f["updated_at"])

	return &model.Order{
		ID:             id,
		UserID:         f["user_id"],
		Total:          total,
		Status:         model.OrderStatus(f["status"]),
		WalletFrom:     strToPtr(f["wallet_from"]),
		TxHash:         strToPtr(f["tx_hash"]),
		IdempotencyKey: strToPtr(f["idempotency_key"]),
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}, nil
}
