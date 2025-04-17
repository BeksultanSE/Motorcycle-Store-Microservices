package handler

import (
	"errors"
	"github.com/BeksultanSE/Assignment1-order/internal/adapter/http/handler/dto"
	"github.com/BeksultanSE/Assignment1-order/internal/domain"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

type OrderHandler struct {
	uc OrderUsecase
}

func NewOrderHandler(usecase OrderUsecase) *OrderHandler {
	return &OrderHandler{uc: usecase}
}

func (h *OrderHandler) Create(ctx *gin.Context) {
	order, err := dto.FromOrderRequest(ctx)
	if err != nil {
		return // Error response is already handled in FromOrderRequest
	}

	createdOrder, err := h.uc.Create(ctx.Request.Context(), order)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	ctx.JSON(http.StatusCreated, dto.ToOrderResponse(createdOrder))
}

func (h *OrderHandler) GetByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	log.Println("Getting order with id:", idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	filter := domain.OrderFilter{ID: &id}
	order, err := h.uc.Get(ctx.Request.Context(), filter)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	ctx.JSON(http.StatusOK, dto.ToOrderResponse(order))
}

func (h *OrderHandler) Update(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var updateData dto.OrderUpdateRequest
	if err := ctx.ShouldBindJSON(&updateData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := domain.OrderFilter{ID: &id}
	update := dto.ToOrderUpdate(updateData)
	err = h.uc.Update(ctx.Request.Context(), filter, update)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	ctx.Status(http.StatusOK)
}

// GetAll retrieves a list of orders for a user
func (h *OrderHandler) GetAll(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "10")
	log.Println("userIDStr:", userIDStr)
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		log.Println(userID)
		return
	}

	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10 // Default limit
	}

	filter := domain.OrderFilter{UserID: &userID}

	orders, total, err := h.uc.GetAll(ctx, filter, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
		return
	}

	response := dto.ToOrderListResponse(orders, total, page, limit)
	ctx.JSON(http.StatusOK, response)
}
