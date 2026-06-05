package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
	gormpg "github.com/shophub/shop/internal/repository/postgres"
)

func TestItemRepository_Create_FindByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo := gormpg.NewItemRepository(testDB)
	ctx := context.Background()

	item := &model.Item{Name: "Integration Laptop", Price: 1299.99, Stock: 3}
	require.NoError(t, repo.Create(ctx, item))
	assert.NotEqual(t, uuid.Nil, item.ID)

	found, err := repo.FindByID(ctx, item.ID)
	require.NoError(t, err)
	assert.Equal(t, item.Name, found.Name)
	assert.Equal(t, item.Price, found.Price)
	assert.Equal(t, item.Stock, found.Stock)
}

func TestItemRepository_FindByID_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo := gormpg.NewItemRepository(testDB)

	_, err := repo.FindByID(context.Background(), uuid.New())

	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestItemRepository_List_WithFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo := gormpg.NewItemRepository(testDB)
	ctx := context.Background()

	// Unique prefix prevents interference with other tests.
	prefix := "listtest-" + uuid.New().String()[:8]
	seeds := []*model.Item{
		{Name: prefix + " Cheap", Price: 5.0, Stock: 100},
		{Name: prefix + " Mid", Price: 50.0, Stock: 10},
		{Name: prefix + " Pricey", Price: 500.0, Stock: 0},
	}
	for _, s := range seeds {
		require.NoError(t, repo.Create(ctx, s))
	}

	t.Run("search by prefix", func(t *testing.T) {
		items, total, err := repo.List(ctx, repository.ItemFilter{Search: prefix, Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, items, 3)
	})

	t.Run("in stock only", func(t *testing.T) {
		inStock := true
		items, _, err := repo.List(ctx, repository.ItemFilter{Search: prefix, InStock: &inStock, Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Len(t, items, 2)
	})

	t.Run("max price filter", func(t *testing.T) {
		maxPrice := 60.0
		items, _, err := repo.List(ctx, repository.ItemFilter{Search: prefix, MaxPrice: &maxPrice, Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Len(t, items, 2)
	})

	t.Run("pagination", func(t *testing.T) {
		items, total, err := repo.List(ctx, repository.ItemFilter{Search: prefix, Page: 1, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, items, 2)
	})
}

func TestItemRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo := gormpg.NewItemRepository(testDB)
	ctx := context.Background()

	item := &model.Item{Name: "Original Name", Price: 10.0, Stock: 5}
	require.NoError(t, repo.Create(ctx, item))

	item.Name = "Updated Name"
	item.Price = 25.0
	require.NoError(t, repo.Update(ctx, item))

	found, err := repo.FindByID(ctx, item.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, 25.0, found.Price)
	assert.Equal(t, 5, found.Stock) // unchanged
}

func TestItemRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo := gormpg.NewItemRepository(testDB)
	ctx := context.Background()

	item := &model.Item{Name: "To Be Deleted", Price: 1.0, Stock: 1}
	require.NoError(t, repo.Create(ctx, item))
	require.NoError(t, repo.Delete(ctx, item.ID))

	_, err := repo.FindByID(ctx, item.ID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestItemRepository_Delete_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo := gormpg.NewItemRepository(testDB)

	err := repo.Delete(context.Background(), uuid.New())

	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestItemRepository_DecrementStock_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo := gormpg.NewItemRepository(testDB)
	ctx := context.Background()

	item := &model.Item{Name: "Stock Item", Price: 5.0, Stock: 10}
	require.NoError(t, repo.Create(ctx, item))

	require.NoError(t, repo.DecrementStock(ctx, item.ID, 4))

	found, err := repo.FindByID(ctx, item.ID)
	require.NoError(t, err)
	assert.Equal(t, 6, found.Stock)
}

func TestItemRepository_DecrementStock_Insufficient(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo := gormpg.NewItemRepository(testDB)
	ctx := context.Background()

	item := &model.Item{Name: "Low Stock", Price: 5.0, Stock: 2}
	require.NoError(t, repo.Create(ctx, item))

	err := repo.DecrementStock(ctx, item.ID, 100)

	assert.ErrorIs(t, err, repository.ErrInsufficientStock)

	// Stock must be unchanged after the failed decrement.
	found, _ := repo.FindByID(ctx, item.ID)
	assert.Equal(t, 2, found.Stock)
}
