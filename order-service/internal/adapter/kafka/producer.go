package kafka

import (
	"context"
	"github.com/BeksultanSE/Assignment1-order/internal/domain"
	events "github.com/BeksultanSE/Assignment1-order/protos/gen/golang"
	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &Producer{producer: producer, topic: topic}, nil
}

func (p *Producer) PublishOrderCreated(ctx context.Context, order domain.Order) error {
	// Map domain.Order to events.OrderCreatedEvent
	orderProto := &events.OrderCreatedEvent{
		OrderId: order.ID,
		UserId:  order.UserID,
		Items:   make([]*events.OrderItemEvent, 0, len(order.Items)),
	}

	for _, item := range order.Items {
		orderProto.Items = append(orderProto.Items, &events.OrderItemEvent{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	// Marshal to protobuf
	eventBytes, err := proto.Marshal(orderProto)
	if err != nil {
		return err
	}

	// Produce Kafka message
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(eventBytes),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err

}

func (p *Producer) Close() error {
	return p.producer.Close()
}
