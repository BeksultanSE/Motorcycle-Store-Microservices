package dao

import (
	"github.com/BeksultanSE/Assignment1-order/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type Order struct {
	ID          uint64      `bson:"_id"`
	UserID      uint64      `bson:"userId"`
	Items       []OrderItem `bson:"items"`
	TotalAmount float64     `bson:"totalAmount"`
	Status      string      `bson:"status"`
	CreatedAt   time.Time   `bson:"createdAt"`
	UpdatedAt   time.Time   `bson:"updatedAt"`
}

type OrderItem struct {
	ProductID  uint64  `bson:"productId"`
	Quantity   uint64  `bson:"quantity"`
	TotalPrice float64 `bson:"totalPrice"`
}

func ToOrderList(daoOrders []Order) []domain.Order {
	orders := make([]domain.Order, len(daoOrders))
	for i, o := range daoOrders {
		orders[i] = domain.Order{
			ID:          o.ID,
			UserID:      o.UserID,
			Items:       ToOrderItemList(o.Items),
			TotalAmount: o.TotalAmount,
			Status:      domain.OrderStatus(o.Status),
			CreatedAt:   o.CreatedAt,
			UpdatedAt:   o.UpdatedAt,
		}
	}
	return orders
}

func ToOrder(daoOrder Order) domain.Order {
	return domain.Order{
		ID:          daoOrder.ID,
		UserID:      daoOrder.UserID,
		Items:       ToOrderItemList(daoOrder.Items),
		TotalAmount: daoOrder.TotalAmount,
		Status:      domain.OrderStatus(daoOrder.Status),
		CreatedAt:   daoOrder.CreatedAt,
		UpdatedAt:   daoOrder.UpdatedAt,
	}
}

func FromOrder(order domain.Order) Order {
	return Order{
		ID:          order.ID,
		UserID:      order.UserID,
		Items:       FromOrderItemList(order.Items),
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}

func ToOrderItemList(daoItems []OrderItem) []domain.OrderItem {
	items := make([]domain.OrderItem, len(daoItems))
	for i, item := range daoItems {
		items[i] = domain.OrderItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			TotalPrice: item.TotalPrice,
		}
	}
	return items
}

func FromOrderItemList(items []domain.OrderItem) []OrderItem {
	daoItems := make([]OrderItem, len(items))
	for i, item := range items {
		daoItems[i] = OrderItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			TotalPrice: item.TotalPrice,
		}
	}
	return daoItems
}

func FromOrderFilter(filter domain.OrderFilter) bson.M {
	query := bson.M{}

	if filter.ID != nil {
		query["_id"] = *filter.ID
	}

	if filter.UserID != nil {
		query["userId"] = *filter.UserID
	}

	if filter.Status != nil {
		query["status"] = string(*filter.Status)
	}

	if filter.CreatedAt != nil {
		query["createdAt"] = *filter.CreatedAt
	}

	if filter.UpdatedAt != nil {
		query["updatedAt"] = *filter.UpdatedAt
	}

	return query
}

func FromOrderUpdateData(updateData domain.OrderUpdateData) bson.M {
	query := bson.M{}

	if updateData.Status != nil {
		query["status"] = string(*updateData.Status)
	}

	if updateData.UpdatedAt != nil {
		query["updatedAt"] = updateData.UpdatedAt
	}

	return bson.M{"$set": query}
}
