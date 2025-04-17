package handler

import (
	grpc "github.com/BeksultanSE/Assignment1-api-gateway/internal/adapter/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

type Handler struct {
	Clients *grpc.Clients
}

func NewHandler(clients *grpc.Clients) *Handler {
	return &Handler{Clients: clients}
}

func mapGRPCErrorToHTTP(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	statusErr, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, "internal server error"
	}

	switch statusErr.Code() {
	case codes.InvalidArgument:
		return http.StatusBadRequest, statusErr.Message()
	case codes.NotFound:
		return http.StatusNotFound, statusErr.Message()
	case codes.AlreadyExists:
		return http.StatusConflict, statusErr.Message()
	case codes.Unauthenticated:
		return http.StatusUnauthorized, statusErr.Message()
	case codes.PermissionDenied:
		return http.StatusForbidden, statusErr.Message()
	case codes.Unavailable:
		return http.StatusServiceUnavailable, "target service is down"
	default:
		return http.StatusInternalServerError, statusErr.Message()
	}
}
