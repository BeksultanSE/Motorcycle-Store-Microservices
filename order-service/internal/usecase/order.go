package usecase

import (
	"context"
	"errors"
	"github.com/BeksultanSE/Assignment1-order/internal/adapter/mongo"
	"github.com/BeksultanSE/Assignment1-order/internal/domain"
	"time"
)

type Order struct {
	aiRepo          AutoIncRepo
	repo            OrderRepository
	inventoryClient InventoryClient
	eventPublisher  EventPublisher
}

func NewOrder(aiRepo AutoIncRepo, repo OrderRepository, client InventoryClient, publisher EventPublisher) *Order {
	return &Order{
		aiRepo:          aiRepo,
		repo:            repo,
		inventoryClient: client,
		eventPublisher:  publisher,
	}
}

func (o *Order) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	totalAmount := 0.0
	for i, item := range order.Items {
		product, err := o.inventoryClient.GetProduct(ctx, item.ProductID)
		if err != nil {
			return domain.Order{}, err
		}

		if product.Stock < item.Quantity {
			return domain.Order{}, errors.New("insufficient stock for product: " + product.Name)
		}

		if item.Quantity <= 0 {
			return domain.Order{}, errors.New("invalid quantity for product, must be at positive amount: " + product.Name)
		}

		order.Items[i].Name = product.Name
		order.Items[i].Price = product.Price
		order.Items[i].TotalPrice = product.Price * float64(item.Quantity)
		totalAmount += order.Items[i].TotalPrice
	}

	order.TotalAmount = totalAmount

	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	order.Status = domain.StatusPending

	id, err := o.aiRepo.Next(ctx, mongo.CollectionOrders)
	if err != nil {
		return domain.Order{}, err
	}
	order.ID = id

	err = o.repo.Create(ctx, order)
	if err != nil {
		return domain.Order{}, err
	}

	if err := o.eventPublisher.PublishOrderCreated(ctx, order); err != nil {
		return domain.Order{}, err
	}

	return domain.Order{
		ID:     order.ID,
		UserID: order.UserID,
		Status: order.Status,
		// Include other fields as necessary
	}, nil
}

func (o *Order) Get(ctx context.Context, filter domain.OrderFilter) (domain.Order, error) {
	order, err := o.repo.GetWithFilter(ctx, filter)
	if err != nil {
		return domain.Order{}, err
	}

	// Fetch product details for each item
	for i, item := range order.Items {
		product, err := o.inventoryClient.GetProduct(ctx, item.ProductID)
		if err != nil {
			return domain.Order{}, err
		}
		order.Items[i].Name = product.Name
		order.Items[i].Price = product.Price
		order.Items[i].TotalPrice = float64(item.Quantity) * product.Price
	}

	order.TotalAmount = 0
	for _, item := range order.Items {
		order.TotalAmount += item.TotalPrice
	}

	return order, nil
}

func (o *Order) GetAll(ctx context.Context, filter domain.OrderFilter, page, limit int64) ([]domain.Order, int64, error) {
	orders, total, err := o.repo.GetAllWithFilter(ctx, filter, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// Fetch product details for each item in each order
	for i, order := range orders {
		for j, item := range order.Items {
			product, err := o.inventoryClient.GetProduct(ctx, item.ProductID)
			if err != nil {
				return nil, 0, err
			}
			orders[i].Items[j].Name = product.Name
			orders[i].Items[j].Price = product.Price
			orders[i].Items[j].TotalPrice = float64(item.Quantity) * product.Price
		}

		orders[i].TotalAmount = 0
		for _, item := range order.Items {
			orders[i].TotalAmount += item.TotalPrice
		}
	}

	return orders, total, nil
}

func (o *Order) Update(ctx context.Context, filter domain.OrderFilter, updated domain.OrderUpdateData) error {
	curTime := time.Now()
	updated.UpdatedAt = &curTime

	err := o.repo.Update(ctx, filter, updated)
	if err != nil {
		return err
	}
	return nil
}

func (o *Order) Delete(ctx context.Context, filter domain.OrderFilter) error {
	err := o.repo.Delete(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}
