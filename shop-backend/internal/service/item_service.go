package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

type CreateItemRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	ImageURL    *string `json:"imageUrl"`
}

type UpdateItemRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Stock       *int     `json:"stock"`
	ImageURL    *string  `json:"imageUrl"`
}

type ItemService struct {
	repo repository.ItemRepository
}

func NewItemService(repo repository.ItemRepository) *ItemService {
	return &ItemService{repo: repo}
}

func (s *ItemService) Create(ctx context.Context, req CreateItemRequest) (*model.Item, error) {
	item := &model.Item{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ItemService) GetByID(ctx context.Context, id uuid.UUID) (*model.Item, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ItemService) List(ctx context.Context, filter repository.ItemFilter) ([]*model.Item, int64, error) {
	return s.repo.List(ctx, filter)
}

func (s *ItemService) Update(ctx context.Context, id uuid.UUID, req UpdateItemRequest) (*model.Item, error) {
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = req.Description
	}
	if req.Price != nil {
		item.Price = *req.Price
	}
	if req.Stock != nil {
		item.Stock = *req.Stock
	}
	if req.ImageURL != nil {
		item.ImageURL = req.ImageURL
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ItemService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
