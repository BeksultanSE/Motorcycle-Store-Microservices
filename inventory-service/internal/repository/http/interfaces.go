package http

import (
	"github.com/BeksultanSE/Assignment1-inventory/internal/repository/http/handler"
)

type ProductHandler interface {
	handler.ProductUseCase
}
