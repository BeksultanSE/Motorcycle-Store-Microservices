package grpc

import (
	"fmt"
	"github.com/BeksultanSE/Assignment1-api-gateway/config"
	proto "github.com/BeksultanSE/Assignment1-api-gateway/pkg/protos/gen/golang"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

type Clients struct {
	User      proto.AuthClient
	Inventory proto.InventoryServiceClient
	Order     proto.OrderServiceClient
	conns     []*grpc.ClientConn
}

func NewClients(cfg *config.Config) (*Clients, error) {
	clients := &Clients{}

	// User Service Client
	userTarget := fmt.Sprintf("%s:%d", cfg.Services.UserService.Host, cfg.Services.UserService.Port)
	userConn, err := grpc.NewClient(
		userTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}
	clients.User = proto.NewAuthClient(userConn)
	clients.conns = append(clients.conns, userConn)

	// Inventory Service Client
	inventoryTarget := fmt.Sprintf("%s:%d", cfg.Services.InventoryService.Host, cfg.Services.InventoryService.Port)
	inventoryConn, err := grpc.NewClient(
		inventoryTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to inventory service: %w", err)
	}
	clients.Inventory = proto.NewInventoryServiceClient(inventoryConn)
	clients.conns = append(clients.conns, inventoryConn)

	// Order Service Client
	orderTarget := fmt.Sprintf("%s:%d", cfg.Services.OrderService.Host, cfg.Services.OrderService.Port)
	orderConn, err := grpc.NewClient(
		orderTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}
	clients.Order = proto.NewOrderServiceClient(orderConn)
	clients.conns = append(clients.conns, orderConn)

	log.Println("Successfully initialized gRPC clients for all services")
	return clients, nil
}

func (c *Clients) Close() {
	for _, conn := range c.conns {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close gRPC connection: %v", err)
		}
	}
}
