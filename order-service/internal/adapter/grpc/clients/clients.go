package clients

import (
	"fmt"
	"github.com/BeksultanSE/Assignment1-order/config"
	proto "github.com/BeksultanSE/Assignment1-order/protos/gen/golang"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

type Clients struct {
	Inventory proto.InventoryServiceClient
	conns     []*grpc.ClientConn
}

func NewClients(cfg *config.Config) (*Clients, error) {
	clients := &Clients{}

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
