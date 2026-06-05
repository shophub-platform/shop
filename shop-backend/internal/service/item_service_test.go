package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
	"github.com/shophub/shop/internal/repository/mocks"
	"github.com/shophub/shop/internal/service"
)

func TestItemService_Create_Success(t *testing.T) {
	repo := new(mocks.ItemRepository)
	svc := service.NewItemService(repo)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*model.Item")).Return(nil)

	item, err := svc.Create(context.Background(), service.CreateItemRequest{
		Name:  "Laptop",
		Price: 999.99,
		Stock: 5,
	})

	require.NoError(t, err)
	assert.Equal(t, "Laptop", item.Name)
	assert.Equal(t, 999.99, item.Price)
	assert.Equal(t, 5, item.Stock)
	repo.AssertExpectations(t)
}

func TestItemService_Create_RepositoryError(t *testing.T) {
	repo := new(mocks.ItemRepository)
	svc := service.NewItemService(repo)

	dbErr := errors.New("connection refused")
	repo.On("Create", mock.Anything, mock.AnythingOfType("*model.Item")).Return(dbErr)

	_, err := svc.Create(context.Background(), service.CreateItemRequest{Name: "X", Price: 1, Stock: 1})

	assert.ErrorIs(t, err, dbErr)
	repo.AssertExpectations(t)
}

func TestItemService_GetByID_Found(t *testing.T) {
	repo := new(mocks.ItemRepository)
	svc := service.NewItemService(repo)

	id := uuid.New()
	expected := &model.Item{ID: id, Name: "Phone", Price: 499.0, Stock: 10}
	repo.On("FindByID", mock.Anything, id).Return(expected, nil)

	item, err := svc.GetByID(context.Background(), id)

	require.NoError(t, err)
	assert.Equal(t, expected, item)
	repo.AssertExpectations(t)
}

func TestItemService_GetByID_NotFound(t *testing.T) {
	repo := new(mocks.ItemRepository)
	svc := service.NewItemService(repo)

	id := uuid.New()
	repo.On("FindByID", mock.Anything, id).Return(nil, repository.ErrNotFound)

	item, err := svc.GetByID(context.Background(), id)

	assert.Nil(t, item)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestItemService_List(t *testing.T) {
	repo := new(mocks.ItemRepository)
	svc := service.NewItemService(repo)

	filter := repository.ItemFilter{Page: 1, PageSize: 10}
	expected := []*model.Item{{Name: "A"}, {Name: "B"}}
	repo.On("List", mock.Anything, filter).Return(expected, int64(2), nil)

	items, total, err := svc.List(context.Background(), filter)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	repo.AssertExpectations(t)
}

func TestItemService_Update_Success(t *testing.T) {
	repo := new(mocks.ItemRepository)
	svc := service.NewItemService(repo)

	id := uuid.New()
	existing := &model.Item{ID: id, Name: "Old", Price: 10.0, Stock: 5}
	newName := "New"
	newPrice := 20.0

	repo.On("FindByID", mock.Anything, id).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*model.Item")).Return(nil)

	updated, err := svc.Update(context.Background(), id, service.UpdateItemRequest{
		Name:  &newName,
		Price: &newPrice,
	})

	require.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	assert.Equal(t, 20.0, updated.Price)
	assert.Equal(t, 5, updated.Stock) // unchanged field
	repo.AssertExpectations(t)
}

func TestItemService_Update_NotFound(t *testing.T) {
	repo := new(mocks.ItemRepository)
	svc := service.NewItemService(repo)

	id := uuid.New()
	repo.On("FindByID", mock.Anything, id).Return(nil, repository.ErrNotFound)

	_, err := svc.Update(context.Background(), id, service.UpdateItemRequest{})

	assert.ErrorIs(t, err, repository.ErrNotFound)
	repo.AssertExpectations(t)
}

func TestItemService_Delete_Success(t *testing.T) {
	repo := new(mocks.ItemRepository)
	svc := service.NewItemService(repo)

	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(nil)

	err := svc.Delete(context.Background(), id)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestItemService_Delete_NotFound(t *testing.T) {
	repo := new(mocks.ItemRepository)
	svc := service.NewItemService(repo)

	id := uuid.New()
	repo.On("Delete", mock.Anything, id).Return(repository.ErrNotFound)

	err := svc.Delete(context.Background(), id)

	assert.ErrorIs(t, err, repository.ErrNotFound)
	repo.AssertExpectations(t)
}
