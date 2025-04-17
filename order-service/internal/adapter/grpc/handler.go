package grpc

import (
	"context"
	"errors"
	"github.com/BeksultanSE/Assignment1-order/internal/adapter/grpc/dto"
	"github.com/BeksultanSE/Assignment1-order/internal/domain"
	"github.com/BeksultanSE/Assignment1-order/internal/usecase"
	proto "github.com/BeksultanSE/Assignment1-order/protos/gen/golang"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGRPCServer struct {
	proto.UnimplementedOrderServiceServer
	orderUsecase *usecase.Order
}

func NewOrderGRPCServer(orderUsecase *usecase.Order) *OrderGRPCServer {
	return &OrderGRPCServer{
		orderUsecase: orderUsecase,
	}
}

func (s *OrderGRPCServer) CreateOrder(ctx context.Context, req *proto.CreateOrderRequest) (*proto.OrderResponse, error) {
	requestDTO := dto.FromCreateOrderRequestProto(req)

	domainOrder := requestDTO.ToDomainOrder()
	createdOrder, err := s.orderUsecase.Create(ctx, domainOrder)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	responseDTO := dto.FromDomainOrder(createdOrder)
	return responseDTO.ToProtoOrderResponse(), nil
}

func (s *OrderGRPCServer) GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.OrderResponse, error) {
	requestDTO := dto.FromGetOrderRequestProto(req)
	filter := requestDTO.ToDomainFilter()

	order, err := s.orderUsecase.Get(ctx, filter)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	responseDTO := dto.FromDomainOrder(order)
	return responseDTO.ToProtoOrderResponse(), nil
}

func (s *OrderGRPCServer) UpdateOrder(ctx context.Context, req *proto.UpdateOrderRequest) (*proto.OrderResponse, error) {
	requestDTO := dto.FromUpdateOrderRequestProto(req)
	filter, update := requestDTO.ToDomainFilterAndUpdate()

	err := s.orderUsecase.Update(ctx, filter, update)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	updatedOrder, err := s.orderUsecase.Get(ctx, filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	responseDTO := dto.FromDomainOrder(updatedOrder)
	return responseDTO.ToProtoOrderResponse(), nil
}

func (s *OrderGRPCServer) ListOrders(ctx context.Context, req *proto.ListOrdersRequest) (*proto.ListOrdersResponse, error) {
	requestDTO := dto.FromListOrdersRequestProto(req)
	filter := domain.OrderFilter{UserID: &requestDTO.UserID}

	orders, total, err := s.orderUsecase.GetAll(ctx, filter, requestDTO.Page, requestDTO.Limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &proto.ListOrdersResponse{
		Orders: make([]*proto.OrderResponse, len(orders)),
		Total:  total,
	}
	for i, ord := range orders {
		responseDTO := dto.FromDomainOrder(ord)
		response.Orders[i] = responseDTO.ToProtoOrderResponse()
	}

	return response, nil
}
