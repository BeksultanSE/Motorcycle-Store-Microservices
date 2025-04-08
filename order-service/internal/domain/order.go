package domain

import "time"

// Order represents the core order entity
type Order struct {
	ID          uint64
	UserID      uint64
	Items       []OrderItem
	TotalAmount float64
	Status      OrderStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// OrderStatus represents the current state of an order
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusPaid      OrderStatus = "paid"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

// OrderItem represents a product in an order with its quantity
type OrderItem struct {
	ProductID  uint64
	Name       string
	Price      float64
	Quantity   uint64
	TotalPrice float64
}

// OrderFilter represents the criteria for filtering orders
type OrderFilter struct {
	ID        *uint64
	UserID    *uint64
	Status    *OrderStatus
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// OrderUpdateData represents the data needed to update an order
type OrderUpdateData struct {
	Status    *OrderStatus
	UpdatedAt *time.Time
}
