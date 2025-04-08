package handler

import (
	"context"
	"github.com/BeksultanSE/Assignment1-order/internal/domain"
)

type OrderUsecase interface {
	Create(ctx context.Context, order domain.Order) (domain.Order, error)
	Get(ctx context.Context, filter domain.OrderFilter) (domain.Order, error)
	GetAll(ctx context.Context, filter domain.OrderFilter, page, limit int64) ([]domain.Order, int64, error)
	Update(ctx context.Context, filter domain.OrderFilter, updated domain.OrderUpdateData) error
	Delete(ctx context.Context, filter domain.OrderFilter) error
}
