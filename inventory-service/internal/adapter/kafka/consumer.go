package kafka

import (
	"github.com/BeksultanSE/Assignment1-inventory/internal/domain"
	"github.com/BeksultanSE/Assignment1-inventory/internal/usecase"
	events "github.com/BeksultanSE/Assignment1-inventory/protos/gen/golang"
	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
	"log"
)

type Consumer struct {
	usecase *usecase.Product
	Topic   string
}

func NewConsumer(usecase *usecase.Product, topic string) *Consumer {
	return &Consumer{usecase: usecase, Topic: topic}
}

func (h *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var event events.OrderCreatedEvent
		if err := proto.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Failed to unmarshal OrderCreatedEvent: %v", err)
			continue
		}

		for _, item := range event.Items {
			filter := domain.ProductFilter{ID: &item.ProductId}
			//updating the stock
			currentProduct, err := h.usecase.Get(session.Context(), filter)
			if err != nil {
				log.Printf("Failed to get product from consumer: %v", err)
			}
			newStock := currentProduct.Stock - item.Quantity

			update := domain.ProductUpdateData{
				Stock: &newStock,
			}
			log.Printf("ProductUpdateData: %v", newStock)
			if err := h.usecase.Update(session.Context(), filter, update); err != nil {
				log.Printf("Failed to update stock for product %d: %v", item.ProductId, err)
			}
		}
		session.MarkMessage(message, "")
	}
	return nil
}
