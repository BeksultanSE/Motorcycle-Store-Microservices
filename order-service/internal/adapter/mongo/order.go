package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/BeksultanSE/Assignment1-order/internal/adapter/mongo/dao"
	"github.com/BeksultanSE/Assignment1-order/internal/domain"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type OrderRepo struct {
	conn       *mongo.Database
	collection string
}

func NewOrderRepo(conn *mongo.Database) *OrderRepo {
	return &OrderRepo{
		conn:       conn,
		collection: CollectionOrders,
	}
}

func (o *OrderRepo) Create(ctx context.Context, order domain.Order) error {
	orderDoc := dao.FromOrder(order)
	_, err := o.conn.Collection(o.collection).InsertOne(ctx, orderDoc)
	if err != nil {
		return fmt.Errorf("failed to create order with ID %d: %w", order.ID, err)
	}

	return nil
}

func (o *OrderRepo) Update(ctx context.Context, filter domain.OrderFilter, update domain.OrderUpdateData) error {
	res, err := o.conn.Collection(o.collection).UpdateOne(
		ctx,
		dao.FromOrderFilter(filter),
		dao.FromOrderUpdateData(update),
	)
	if err != nil {
		return fmt.Errorf("failed to update order with ID: %d, err: %w", filter.ID, err)
	}

	if res.ModifiedCount == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func (o *OrderRepo) GetWithFilter(ctx context.Context, filter domain.OrderFilter) (domain.Order, error) {
	var orderDAO dao.Order
	err := o.conn.Collection(o.collection).FindOne(
		ctx,
		dao.FromOrderFilter(filter),
	).Decode(&orderDAO)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Order{}, domain.ErrOrderNotFound
		}
		return domain.Order{}, fmt.Errorf("failed to update order with ID: %d, err: %w", filter.ID, err)
	}
	return dao.ToOrder(orderDAO), nil
}

func (o *OrderRepo) GetAllWithFilter(ctx context.Context, filter domain.OrderFilter, page, limit int64) ([]domain.Order, int64, error) {
	// Create the filter for the query
	findFilter := dao.FromOrderFilter(filter)

	// Calculate pagination parameters
	skip := (page - 1) * limit

	// Set up find options for pagination
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	// Get total count
	totalCount, err := o.conn.Collection(o.collection).CountDocuments(ctx, findFilter)
	if err != nil {
		return nil, 0, err
	}

	// Execute the query with pagination
	cursor, err := o.conn.Collection(o.collection).Find(ctx, findFilter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find orders: %w", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("failed to close cursor: %v", err)
		}
	}(cursor, ctx)

	var daoOrders []dao.Order
	if err := cursor.All(ctx, &daoOrders); err != nil {
		return nil, 0, fmt.Errorf("failed to decode orders: %w", err)
	}
	orders := dao.ToOrderList(daoOrders)
	return orders, totalCount, nil
}

func (o *OrderRepo) Delete(ctx context.Context, filter domain.OrderFilter) error {
	res, err := o.conn.Collection(o.collection).DeleteOne(
		ctx,
		dao.FromOrderFilter(filter),
	)
	if err != nil {
		return fmt.Errorf("failed to delete order with filter: %v, err: %w", filter, err)
	}

	if res.DeletedCount == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}
