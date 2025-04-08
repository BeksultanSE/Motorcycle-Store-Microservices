package dto

import (
	"encoding/json"
	"errors"
	"github.com/BeksultanSE/Assignment1-order/internal/domain"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// OrderRequest the request body for creating an order
type OrderRequest struct {
	UserID uint64             `json:"user_id" binding:"required"`
	Items  []OrderItemRequest `json:"items" binding:"required"`
	Status domain.OrderStatus `json:"status" binding:"required,oneof=pending paid shipped delivered cancelled"`
}

// OrderItemRequest an item in the order request
type OrderItemRequest struct {
	ProductID uint64  `json:"product_id" binding:"required"`
	Name      string  `json:"name" binding:"required"`
	Price     float64 `json:"price" binding:"required,min=0"`
	Quantity  uint64  `json:"quantity" binding:"required,min=1"`
}

// OrderUpdateRequest request body for updating an order
type OrderUpdateRequest struct {
	Status *domain.OrderStatus `json:"status" binding:"required,oneof=paid shipped delivered cancelled"`
}

// OrderStatus constants of order status
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusPaid      OrderStatus = "paid"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

func (os *OrderStatus) UnmarshalJSON(b []byte) error {
	var status string
	if err := json.Unmarshal(b, &status); err != nil {
		return err
	}

	switch OrderStatus(status) {
	case StatusPending, StatusPaid, StatusShipped, StatusDelivered, StatusCancelled:
		*os = OrderStatus(status)
		return nil
	default:
		return errors.New("invalid order status")
	}
}

// OrderResponse the response body after creating an order
type OrderResponse struct {
	ID          uint64              `json:"id"`
	UserID      uint64              `json:"user_id"`
	Items       []OrderItemResponse `json:"items"`
	TotalAmount float64             `json:"total_amount"`
	Status      string              `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// OrderItemResponse represents an item in the order response
type OrderItemResponse struct {
	ProductID  uint64  `json:"product_id"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Quantity   uint64  `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
}

// OrderListResponse represents a list of orders
type OrderListResponse struct {
	Items      []OrderResponse `json:"items"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}

// FromOrderRequest converts a Gin request to an Order model
func FromOrderRequest(ctx *gin.Context) (domain.Order, error) {
	var req OrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return domain.Order{}, err
	}

	// Convert OrderRequest to domain.Order
	orderItems := make([]domain.OrderItem, len(req.Items))
	totalAmount := 0.0
	for i, item := range req.Items {
		orderItems[i] = domain.OrderItem{
			ProductID:  item.ProductID,
			Name:       item.Name,
			Price:      item.Price,
			Quantity:   item.Quantity,
			TotalPrice: item.Price * float64(item.Quantity), // Calculate total price
		}
		totalAmount += item.Price
	}

	return domain.Order{
		UserID:      req.UserID,
		Items:       orderItems,
		TotalAmount: totalAmount,
		Status:      domain.OrderStatus(req.Status),
	}, nil
}

// ToOrderResponse converts an Order model to a response DTO
func ToOrderResponse(order domain.Order) OrderResponse {
	orderItemsResponse := make([]OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		orderItemsResponse[i] = OrderItemResponse{
			ProductID:  item.ProductID,
			Name:       item.Name,
			Price:      item.Price,
			Quantity:   item.Quantity,
			TotalPrice: item.TotalPrice,
		}
	}

	return OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		Items:       orderItemsResponse,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}

// ToOrderListResponse converts a list of Order models to a response DTO
func ToOrderListResponse(orders []domain.Order, total, page, limit int64) OrderListResponse {
	orderResponses := make([]OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = ToOrderResponse(order)
	}

	return OrderListResponse{
		Items:      orderResponses,
		Total:      int(total),
		Page:       int(page),
		Limit:      int(limit),
		TotalPages: int((total + limit - 1) / limit), // Calculate total pages
	}
}

func ToOrderUpdate(order OrderUpdateRequest) domain.OrderUpdateData {
	return domain.OrderUpdateData{
		Status: order.Status,
	}
}
