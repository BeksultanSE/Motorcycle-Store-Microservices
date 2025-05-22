package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/BeksultanSE/Assignment1-order/internal/adapter/mongo"
	"github.com/BeksultanSE/Assignment1-order/internal/domain"
	"github.com/BeksultanSE/Assignment1-order/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OrderServiceIntegrationTestSuite for testing the order service with real dependencies
type OrderServiceIntegrationTestSuite struct {
	suite.Suite
	ctx             context.Context
	mongoClient     *mongodriver.Client
	orderRepo       *mongo.OrderRepo
	autoIncRepo     *mongo.AutoInc
	inventoryClient *RealInventoryClient
	eventPublisher  *RealEventPublisher
	orderService    *usecase.Order
	testCollections []string
}

// RealInventoryClient implements the InventoryClient interface but can be configured for testing
type RealInventoryClient struct {
	inventoryServiceURL string
	// For integration tests, we'll mock the behavior locally
	mockProducts map[uint64]domain.Product
}

func NewRealInventoryClient() *RealInventoryClient {
	// In a real scenario, we would configure this with the actual service URL
	client := &RealInventoryClient{
		inventoryServiceURL: os.Getenv("INVENTORY_SERVICE_URL"),
		mockProducts:        make(map[uint64]domain.Product),
	}

	// Add some mock products for testing
	client.mockProducts[1] = domain.Product{
		ID:       1,
		Name:     "Motorcycle Helmet",
		Category: "Safety",
		Price:    150.0,
		Stock:    20,
	}

	client.mockProducts[2] = domain.Product{
		ID:       2,
		Name:     "Motorcycle Jacket",
		Category: "Apparel",
		Price:    250.0,
		Stock:    15,
	}

	client.mockProducts[3] = domain.Product{
		ID:       3,
		Name:     "Motorcycle Gloves",
		Category: "Accessories",
		Price:    50.0,
		Stock:    30,
	}

	return client
}

func (c *RealInventoryClient) GetProduct(ctx context.Context, productID uint64) (domain.Product, error) {
	// In a real implementation, this would make an HTTP or gRPC call to the inventory service
	// For this integration test, we'll use our mock data
	product, exists := c.mockProducts[productID]
	if !exists {
		return domain.Product{}, fmt.Errorf("product not found: %d", productID)
	}
	return product, nil
}

// RealEventPublisher implements the EventPublisher interface but logs events for testing
type RealEventPublisher struct {
	events []domain.Order
}

func NewRealEventPublisher() *RealEventPublisher {
	return &RealEventPublisher{
		events: make([]domain.Order, 0),
	}
}

func (p *RealEventPublisher) PublishOrderCreated(ctx context.Context, event domain.Order) error {
	// In a real implementation, this would publish to a message broker
	// For this integration test, we'll just log and store the event
	p.events = append(p.events, event)
	log.Printf("Published order created event: %+v", event)
	return nil
}

// Get the events published so far (for test verification)
func (p *RealEventPublisher) GetPublishedEvents() []domain.Order {
	return p.events
}

// Setup test suite
func (s *OrderServiceIntegrationTestSuite) SetupSuite() {
	// Connect to MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	s.mongoClient, err = mongodriver.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping database to verify connection
	err = s.mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// Set test database name
	dbName := "order_service_test_db"

	// Create repositories
	s.autoIncRepo = mongo.NewAutoInc(s.mongoClient.Database(dbName))
	s.orderRepo = mongo.NewOrderRepo(s.mongoClient.Database(dbName))

	// Create "real" clients (still mocked for integration testing)
	s.inventoryClient = NewRealInventoryClient()
	s.eventPublisher = NewRealEventPublisher()

	// Create order service with real repositories and dependencies
	s.orderService = usecase.NewOrder(s.autoIncRepo, s.orderRepo, s.inventoryClient, s.eventPublisher)

	// Store collection names for cleanup
	s.testCollections = []string{mongo.CollectionOrders, mongo.CollectionAutoInc}

	// Store context
	s.ctx = context.Background()
}

// Cleanup after each test
func (s *OrderServiceIntegrationTestSuite) TearDownTest() {
	// Clean up test collections after each test
	for _, coll := range s.testCollections {
		_, err := s.mongoClient.Database("order_service_test_db").Collection(coll).DeleteMany(s.ctx, bson.M{})
		if err != nil {
			s.T().Logf("Failed to clean up collection %s: %v", coll, err)
		}
	}

	// Reset event publisher
	s.eventPublisher.events = make([]domain.Order, 0)
}

// Cleanup after all tests
func (s *OrderServiceIntegrationTestSuite) TearDownSuite() {
	// Close MongoDB connection
	if s.mongoClient != nil {
		err := s.mongoClient.Disconnect(s.ctx)
		if err != nil {
			s.T().Logf("Failed to disconnect from MongoDB: %v", err)
		}
	}
}

// Test complete order flow - create, get, update status
func (s *OrderServiceIntegrationTestSuite) TestCompleteOrderFlow() {
	// 1. Create an order with multiple items
	testUserID := uint64(456)

	orderItems := []domain.OrderItem{
		{
			ProductID: 1, // Helmet
			Quantity:  1,
		},
		{
			ProductID: 2, // Jacket
			Quantity:  1,
		},
		{
			ProductID: 3, // Gloves
			Quantity:  2, // Two pairs
		},
	}

	order := domain.Order{
		UserID: testUserID,
		Items:  orderItems,
	}

	// Create the order
	createdOrder, err := s.orderService.Create(s.ctx, order)
	assert.NoError(s.T(), err)
	assert.NotEqual(s.T(), uint64(0), createdOrder.ID)

	// 2. Verify the order was created correctly
	filter := domain.OrderFilter{
		ID: &createdOrder.ID,
	}

	retrievedOrder, err := s.orderService.Get(s.ctx, filter)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), createdOrder.ID, retrievedOrder.ID)
	assert.Equal(s.T(), testUserID, retrievedOrder.UserID)
	assert.Equal(s.T(), domain.StatusPending, retrievedOrder.Status)
	assert.Len(s.T(), retrievedOrder.Items, 3)

	// 3. Verify order items and total calculation
	// Expected total: $150 (helmet) + $250 (jacket) + $100 (2 gloves at $50 each) = $500
	assert.Equal(s.T(), 500.0, retrievedOrder.TotalAmount)

	// 4. Check that the event was published
	publishedEvents := s.eventPublisher.GetPublishedEvents()
	assert.Len(s.T(), publishedEvents, 1)
	assert.Equal(s.T(), createdOrder.ID, publishedEvents[0].ID)

	// 5. Update the order status to paid
	paidStatus := domain.StatusPaid
	updateData := domain.OrderUpdateData{
		Status: &paidStatus,
	}

	err = s.orderService.Update(s.ctx, filter, updateData)
	assert.NoError(s.T(), err)

	// 6. Verify the status was updated
	updatedOrder, err := s.orderService.Get(s.ctx, filter)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), domain.StatusPaid, updatedOrder.Status)

	// 7. Update the order status to shipped
	shippedStatus := domain.StatusShipped
	updateData = domain.OrderUpdateData{
		Status: &shippedStatus,
	}

	err = s.orderService.Update(s.ctx, filter, updateData)
	assert.NoError(s.T(), err)

	// 8. Verify the status was updated again
	updatedOrder, err = s.orderService.Get(s.ctx, filter)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), domain.StatusShipped, updatedOrder.Status)
}

// Test insufficient stock scenario
func (s *OrderServiceIntegrationTestSuite) TestInsufficientStock() {
	// Try to order more helmets than in stock
	testUserID := uint64(456)

	orderItems := []domain.OrderItem{
		{
			ProductID: 1,  // Helmet
			Quantity:  30, // Stock is only 20
		},
	}

	order := domain.Order{
		UserID: testUserID,
		Items:  orderItems,
	}

	// Attempt to create the order
	_, err := s.orderService.Create(s.ctx, order)

	// Should get an error about insufficient stock
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "insufficient stock")

	// Verify no events were published
	publishedEvents := s.eventPublisher.GetPublishedEvents()
	assert.Len(s.T(), publishedEvents, 0)
}

// Test getting orders for a user with pagination
func (s *OrderServiceIntegrationTestSuite) TestGetOrdersWithPagination() {
	// Create test user ID
	testUserID := uint64(789)

	// Create 5 orders for the same user
	for i := 0; i < 5; i++ {
		orderItems := []domain.OrderItem{
			{
				ProductID: 3,             // Gloves
				Quantity:  uint64(i + 1), // Different quantities
			},
		}

		order := domain.Order{
			UserID: testUserID,
			Items:  orderItems,
		}

		_, err := s.orderService.Create(s.ctx, order)
		assert.NoError(s.T(), err)
	}

	// Test pagination with 2 items per page
	filter := domain.OrderFilter{
		UserID: &testUserID,
	}

	// Get page 1 (first 2 orders)
	orders1, total, err := s.orderService.GetAll(s.ctx, filter, 1, 2)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(5), total) // Total count is 5
	assert.Len(s.T(), orders1, 2)        // But we only get 2 per page

	// Get page 2 (next 2 orders)
	orders2, total, err := s.orderService.GetAll(s.ctx, filter, 2, 2)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(5), total)
	assert.Len(s.T(), orders2, 2)

	// Get page 3 (last order)
	orders3, total, err := s.orderService.GetAll(s.ctx, filter, 3, 2)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(5), total)
	assert.Len(s.T(), orders3, 1) // Only 1 order on the last page

	// Verify we got different orders on each page
	orderIDs := make(map[uint64]bool)

	// Collect IDs from page 1
	for _, order := range orders1 {
		orderIDs[order.ID] = true
	}

	// Collect IDs from page 2
	for _, order := range orders2 {
		// Check that we don't have any duplicates
		assert.False(s.T(), orderIDs[order.ID], "Order ID should not be duplicated across pages")
		orderIDs[order.ID] = true
	}

	// Collect IDs from page 3
	for _, order := range orders3 {
		// Check that we don't have any duplicates
		assert.False(s.T(), orderIDs[order.ID], "Order ID should not be duplicated across pages")
		orderIDs[order.ID] = true
	}

	// Verify we got all 5 orders
	assert.Len(s.T(), orderIDs, 5)
}

// Run the test suite
func TestOrderServiceIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping order service integration tests in short mode")
	}

	suite.Run(t, new(OrderServiceIntegrationTestSuite))
}
