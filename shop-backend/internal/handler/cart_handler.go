package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/shophub/shop/internal/middleware"
	"github.com/shophub/shop/internal/repository"
	"github.com/shophub/shop/internal/service"
	"github.com/shophub/shop/pkg/response"
)

type CartHandler struct {
	svc    *service.CartService
	logger *zap.Logger
}

func NewCartHandler(svc *service.CartService, logger *zap.Logger) *CartHandler {
	return &CartHandler{svc: svc, logger: logger}
}

// Get handles GET /api/v1/cart — returns the authenticated user's cart.
func (h *CartHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	cart, err := h.svc.GetCart(r.Context(), userID)
	if err != nil {
		h.logger.Error("get cart", zap.Error(err))
		response.InternalError(w, "failed to get cart")
		return
	}
	response.Ok(w, cart)
}

// Clear handles DELETE /api/v1/cart — removes all items from the user's cart.
func (h *CartHandler) Clear(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if err := h.svc.ClearCart(r.Context(), userID); err != nil {
		h.logger.Error("clear cart", zap.Error(err))
		response.InternalError(w, "failed to clear cart")
		return
	}
	response.Ok(w, map[string]string{"message": "cart cleared"})
}

type addItemRequest struct {
	ItemID   uuid.UUID `json:"itemId"`
	Quantity int       `json:"quantity"`
}

// AddItem handles POST /api/v1/cart/items — adds an item (or increments quantity).
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	var req addItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Quantity <= 0 {
		response.BadRequest(w, "quantity must be positive")
		return
	}
	cart, err := h.svc.AddItem(r.Context(), userID, req.ItemID, req.Quantity)
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(w, "item not found")
		return
	}
	if err != nil {
		h.logger.Error("add item to cart", zap.Error(err))
		response.InternalError(w, "failed to add item to cart")
		return
	}
	response.Ok(w, cart)
}

type setQuantityRequest struct {
	Quantity int `json:"quantity"`
}

// SetItemQuantity handles PUT /api/v1/cart/items/{itemId} — replaces the item's quantity.
func (h *CartHandler) SetItemQuantity(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		response.BadRequest(w, "invalid item id")
		return
	}
	var req setQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if req.Quantity <= 0 {
		response.BadRequest(w, "quantity must be positive")
		return
	}
	cart, err := h.svc.SetItemQuantity(r.Context(), userID, itemID, req.Quantity)
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(w, "item not found in cart")
		return
	}
	if err != nil {
		h.logger.Error("set cart item quantity", zap.Error(err))
		response.InternalError(w, "failed to update cart")
		return
	}
	response.Ok(w, cart)
}

// RemoveItem handles DELETE /api/v1/cart/items/{itemId} — removes a single item from the cart.
func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		response.BadRequest(w, "invalid item id")
		return
	}
	cart, err := h.svc.RemoveItem(r.Context(), userID, itemID)
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(w, "item not found in cart")
		return
	}
	if err != nil {
		h.logger.Error("remove cart item", zap.Error(err))
		response.InternalError(w, "failed to remove item from cart")
		return
	}
	response.Ok(w, cart)
}
