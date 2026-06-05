package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/shophub/shop/internal/middleware"
	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
	"github.com/shophub/shop/internal/service"
	"github.com/shophub/shop/pkg/response"
)

type OrderHandler struct {
	svc    *service.OrderService
	logger *zap.Logger
}

func NewOrderHandler(svc *service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, logger: logger}
}

// List handles GET /api/v1/orders — admins see all orders, users see only their own.
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := repository.OrderFilter{Page: 1, PageSize: 20}

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
	if s := r.URL.Query().Get("status"); s != "" {
		status := model.OrderStatus(s)
		filter.Status = &status
	}

	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	if role != "admin" {
		userID := middleware.UserIDFromContext(r.Context())
		filter.UserID = &userID
	}

	orders, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("list orders", zap.Error(err))
		response.InternalError(w, "failed to list orders")
		return
	}
	response.Ok(w, map[string]any{
		"orders":   orders,
		"total":    total,
		"page":     filter.Page,
		"pageSize": filter.PageSize,
	})
}

// Create handles POST /api/v1/orders — creates an order from the user's active cart.
// Supports the Idempotency-Key header: sending the same key twice returns the same order.
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	var req service.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	req.IdempotencyKey = r.Header.Get("Idempotency-Key")
	order, err := h.svc.CreateFromCart(r.Context(), userID, req)
	if errors.Is(err, repository.ErrInsufficientStock) {
		response.BadRequest(w, "one or more items are out of stock")
		return
	}
	if err != nil {
		h.logger.Error("create order", zap.Error(err))
		response.InternalError(w, err.Error())
		return
	}
	response.Created(w, order)
}

// GetByID handles GET /api/v1/orders/{id} — users can only fetch their own orders.
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid order id")
		return
	}
	order, err := h.svc.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(w, "order not found")
		return
	}
	if err != nil {
		h.logger.Error("get order", zap.Error(err))
		response.InternalError(w, "failed to get order")
		return
	}
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	if role != "admin" && order.UserID != middleware.UserIDFromContext(r.Context()) {
		response.NotFound(w, "order not found")
		return
	}
	response.Ok(w, order)
}

type confirmPaymentRequest struct {
	TxHash string `json:"txHash"`
}

// ConfirmPayment handles POST /api/v1/orders/{id}/confirm — called by the blockchain listener.
func (h *OrderHandler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid order id")
		return
	}
	var req confirmPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TxHash == "" {
		response.BadRequest(w, "txHash is required")
		return
	}
	order, err := h.svc.ConfirmPayment(r.Context(), id, req.TxHash)
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(w, "order not found")
		return
	}
	if err != nil {
		h.logger.Error("confirm payment", zap.Error(err))
		response.InternalError(w, err.Error())
		return
	}
	response.Ok(w, order)
}

type updateStatusRequest struct {
	Status model.OrderStatus `json:"status"`
}

// UpdateStatus handles PATCH /api/v1/orders/{id}/status — admin only, advances order lifecycle.
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "invalid order id")
		return
	}
	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
		response.BadRequest(w, "status is required")
		return
	}
	order, err := h.svc.UpdateStatus(r.Context(), id, req.Status)
	if errors.Is(err, repository.ErrNotFound) {
		response.NotFound(w, "order not found")
		return
	}
	if err != nil {
		h.logger.Error("update order status", zap.Error(err))
		response.InternalError(w, err.Error())
		return
	}
	response.Ok(w, order)
}
