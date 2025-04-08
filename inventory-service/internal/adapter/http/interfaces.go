package http

import (
	"github.com/BeksultanSE/Assignment1-inventory/internal/adapter/http/handler"
)

type ProductHandler interface {
	handler.ProductUseCase
}
