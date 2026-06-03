package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

type itemRepository struct {
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) repository.ItemRepository {
	return &itemRepository{db: db}
}

func (r *itemRepository) Create(ctx context.Context, item *model.Item) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *itemRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Item, error) {
	var item model.Item
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	return &item, err
}

func (r *itemRepository) List(ctx context.Context, filter repository.ItemFilter) ([]*model.Item, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Item{})

	if filter.Search != "" {
		q = q.Where("name ILIKE ?", "%"+filter.Search+"%")
	}
	if filter.MinPrice != nil {
		q = q.Where("price >= ?", *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		q = q.Where("price <= ?", *filter.MaxPrice)
	}
	if filter.InStock != nil && *filter.InStock {
		q = q.Where("stock > 0")
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

	var items []*model.Item
	err := q.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error

	return items, total, err
}

func (r *itemRepository) Update(ctx context.Context, item *model.Item) error {
	result := r.db.WithContext(ctx).Save(item)
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return result.Error
}

func (r *itemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.Item{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *itemRepository) DecrementStock(ctx context.Context, id uuid.UUID, quantity int) error {
	result := r.db.WithContext(ctx).
		Model(&model.Item{}).
		Where("id = ? AND stock >= ?", id, quantity).
		Update("stock", gorm.Expr("stock - ?", quantity))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrInsufficientStock
	}
	return nil
}
