package dto

import (
	"github.com/BeksultanSE/Assignment1-order/internal/domain"
	order "github.com/BeksultanSE/Assignment1-order/protos/gen/golang"
	"time"
)

type CreateOrderRequestDTO struct {
	UserID uint64
	Items  []OrderItemDTO
}

type OrderItemDTO struct {
	ProductID  uint64
	Name       string
	Price      float64
	Quantity   uint64
	TotalPrice float64
}

type OrderResponseDTO struct {
	ID          uint64
	UserID      uint64
	Items       []OrderItemDTO
	TotalAmount float64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GetOrderRequestDTO struct {
	OrderID uint64
}

type UpdateOrderRequestDTO struct {
	OrderID uint64
	Status  string
}

type ListOrdersRequestDTO struct {
	UserID uint64
	Page   int64
	Limit  int64
}

func FromCreateOrderRequestProto(req *order.CreateOrderRequest) *CreateOrderRequestDTO {
	items := make([]OrderItemDTO, len(req.Items))
	for i, item := range req.Items {
		items[i] = OrderItemDTO{
			ProductID:  item.ProductId,
			Name:       item.Name,
			Price:      item.Price,
			Quantity:   item.Quantity,
			TotalPrice: item.TotalPrice,
		}
	}
	return &CreateOrderRequestDTO{
		UserID: req.UserId,
		Items:  items,
	}
}

func (d *CreateOrderRequestDTO) ToDomainOrder() domain.Order {
	items := make([]domain.OrderItem, len(d.Items))
	for i, item := range d.Items {
		items[i] = domain.OrderItem{
			ProductID:  item.ProductID,
			Name:       item.Name,
			Price:      item.Price,
			Quantity:   item.Quantity,
			TotalPrice: item.TotalPrice,
		}
	}
	return domain.Order{
		UserID: d.UserID,
		Items:  items,
		Status: domain.StatusPending,
	}
}

func FromDomainOrder(order domain.Order) *OrderResponseDTO {
	items := make([]OrderItemDTO, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemDTO{
			ProductID:  item.ProductID,
			Name:       item.Name,
			Price:      item.Price,
			Quantity:   item.Quantity,
			TotalPrice: item.TotalPrice,
		}
	}
	return &OrderResponseDTO{
		ID:          order.ID,
		UserID:      order.UserID,
		Items:       items,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}

func (d *OrderResponseDTO) ToProtoOrderResponse() *order.OrderResponse {
	items := make([]*order.OrderItem, len(d.Items))
	for i, item := range d.Items {
		items[i] = &order.OrderItem{
			ProductId:  item.ProductID,
			Name:       item.Name,
			Price:      item.Price,
			Quantity:   item.Quantity,
			TotalPrice: item.TotalPrice,
		}
	}
	return &order.OrderResponse{
		OrderId:     d.ID,
		UserId:      d.UserID,
		Items:       items,
		TotalAmount: d.TotalAmount,
		Status:      d.Status,
		CreatedAt:   d.CreatedAt.String(),
		UpdatedAt:   d.UpdatedAt.String(),
	}
}

func FromGetOrderRequestProto(req *order.GetOrderRequest) *GetOrderRequestDTO {
	return &GetOrderRequestDTO{
		OrderID: req.OrderId,
	}
}

func (d *GetOrderRequestDTO) ToDomainFilter() domain.OrderFilter {
	return domain.OrderFilter{
		ID: &d.OrderID,
	}
}

func FromUpdateOrderRequestProto(req *order.UpdateOrderRequest) *UpdateOrderRequestDTO {
	return &UpdateOrderRequestDTO{
		OrderID: req.OrderId,
		Status:  req.Status,
	}
}

func (d *UpdateOrderRequestDTO) ToDomainFilterAndUpdate() (domain.OrderFilter, domain.OrderUpdateData) {
	filter := domain.OrderFilter{
		ID: &d.OrderID,
	}
	update := domain.OrderUpdateData{
		Status: (*domain.OrderStatus)(&d.Status),
	}
	return filter, update
}

func FromListOrdersRequestProto(req *order.ListOrdersRequest) *ListOrdersRequestDTO {
	return &ListOrdersRequestDTO{
		UserID: req.UserId,
		Page:   req.Page,
		Limit:  req.Limit,
	}
}
