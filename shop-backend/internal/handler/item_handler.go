package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/shophub/shop/internal/repository"
	"github.com/shophub/shop/internal/service"
	"github.com/shophub/shop/pkg/response"
)

type ItemHandler struct {
	svc    *service.ItemService
	logger *zap.Logger
}

func NewItemHandler(svc *service.ItemService, logger *zap.Logger) *ItemHandler {
	return &ItemHandler{svc: svc, logger: logger}
}

// List handles GET /api/v1/items — public, supports search/price/stock/pagination query params.
func (h *ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := repository.ItemFilter{Page: 1, PageSize: 20}

	filter.Search = r.URL.Query().Get("search")

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			filter.Page = v
		}
	}
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			filter.PageSize = v
		}
	}
	if s := r.URL.Query().Get("minPrice"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			filter.MinPrice = &v
		}
	}
	if s := r.URL.Query().Get("maxPrice"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			filter.MaxPrice = &v
		}
	}
	if r.URL.Query().Get("inStock") == "true" {
		t := true
		filter.InStock = &t
	}

	items, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("list items", zap.Error(err))
		response.InternalError(w, "failed to list items")
		return
	}
	response.Ok(w, map[string]any{
		"items":    items,
		"total":    total,
		"page":     filter.Page,
		"pageSize": filter.PageSize,
	})
}

// GetByID handles GET /api/v1/items/{id} — public.
func (h *ItemHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid item id")
		return
	}
	item, err := h.svc.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(w, "item not found")
		return
	}
	if err != nil {
		h.logger.Error("get item", zap.Error(err))
		response.InternalError(w, "failed to get item")
		return
	}
	response.Ok(w, item)
}

// Create handles POST /api/v1/items — admin only.
func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Name == "" {
		response.BadRequest(w, "name is required")
		return
	}
	if req.Price <= 0 {
		response.BadRequest(w, "price must be positive")
		return
	}
	if req.Stock < 0 {
		response.BadRequest(w, "stock cannot be negative")
		return
	}
	item, err := h.svc.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("create item", zap.Error(err))
		response.InternalError(w, "failed to create item")
		return
	}
	response.Created(w, item)
}

// Update handles PATCH /api/v1/items/{id} — admin only.
func (h *ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid item id")
		return
	}
	var req service.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	item, err := h.svc.Update(r.Context(), id, req)
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(w, "item not found")
		return
	}
	if err != nil {
		h.logger.Error("update item", zap.Error(err))
		response.InternalError(w, "failed to update item")
		return
	}
	response.Ok(w, item)
}

// Delete handles DELETE /api/v1/items/{id} — admin only.
func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid item id")
		return
	}
	err = h.svc.Delete(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(w, "item not found")
		return
	}
	if err != nil {
		h.logger.Error("delete item", zap.Error(err))
		response.InternalError(w, "failed to delete item")
		return
	}
	response.Ok(w, map[string]string{"message": "item deleted"})
}
