package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) repository.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&order, "id = ?", id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	return &order, err
}

func (r *orderRepository) List(ctx context.Context, filter repository.OrderFilter) ([]*model.Order, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Order{})

	if filter.UserID != nil {
		q = q.Where("user_id = ?", *filter.UserID)
	}
	if filter.Status != nil {
		q = q.Where("status = ?", *filter.Status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	var orders []*model.Order
	err := q.Preload("Items").
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus, txHash *string) error {
	updates := map[string]any{"status": status}
	if txHash != nil {
		updates["tx_hash"] = *txHash
	}

	result := r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *orderRepository) FindByTxHash(ctx context.Context, txHash string) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("tx_hash = ?", txHash).
		First(&order).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	return &order, err
}
