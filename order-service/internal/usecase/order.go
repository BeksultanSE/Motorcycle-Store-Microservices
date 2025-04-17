package usecase

import (
	"context"
	"github.com/BeksultanSE/Assignment1-order/internal/adapter/mongo"
	"github.com/BeksultanSE/Assignment1-order/internal/domain"
)

type Order struct {
	aiRepo AutoIncRepo
	repo   OrderRepository
}

func NewOrder(aiRepo AutoIncRepo, repo OrderRepository) *Order {
	return &Order{
		aiRepo: aiRepo,
		repo:   repo,
	}
}

func (o *Order) Create(ctx context.Context, order domain.Order) (domain.Order, error) {
	id, err := o.aiRepo.Next(ctx, mongo.CollectionOrders)
	if err != nil {
		return domain.Order{}, err
	}
	order.ID = id
	err = o.repo.Create(ctx, order)
	if err != nil {
		return domain.Order{}, err
	}
	return domain.Order{
		ID:     id,
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
	return order, nil
}

func (o *Order) GetAll(ctx context.Context, filter domain.OrderFilter, page, limit int64) ([]domain.Order, int64, error) {
	orders, total, err := o.repo.GetAllWithFilter(ctx, filter, page, limit)
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (o *Order) Update(ctx context.Context, filter domain.OrderFilter, updated domain.OrderUpdateData) error {
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
