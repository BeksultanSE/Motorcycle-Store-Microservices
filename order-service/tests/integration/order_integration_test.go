package integration

import (
	"context"
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

// Integration test suite for order service
type OrderIntegrationTestSuite struct {
	suite.Suite
	ctx             context.Context
	mongoClient     *mongodriver.Client
	orderRepo       *mongo.OrderRepo
	autoIncRepo     *mongo.AutoInc
	mockInventory   *mockInventoryClient
	mockPublisher   *mockEventPublisher
	orderService    *usecase.Order
	testCollections []string
}

// Mock inventory client for testing
type mockInventoryClient struct{}

func (m *mockInventoryClient) GetProduct(ctx context.Context, productID uint64) (domain.Product, error) {
	// Return test product data
	return domain.Product{
		ID:        productID,
		Name:      "Test Product",
		Category:  "Test Category",
		Price:     100.0,
		Stock:     10,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// Mock event publisher for testing
type mockEventPublisher struct{}

func (m *mockEventPublisher) PublishOrderCreated(ctx context.Context, event domain.Order) error {
	// In integration test, we just simulate success
	return nil
}

// Setup test suite
func (s *OrderIntegrationTestSuite) SetupSuite() {
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
	dbName := "order_test_db"

	// Create repositories
	s.autoIncRepo = mongo.NewAutoInc(s.mongoClient.Database(dbName))
	s.orderRepo = mongo.NewOrderRepo(s.mongoClient.Database(dbName))

	// Create mocks
	s.mockInventory = &mockInventoryClient{}
	s.mockPublisher = &mockEventPublisher{}

	// Create order service with real repositories and mock clients
	s.orderService = usecase.NewOrder(s.autoIncRepo, s.orderRepo, s.mockInventory, s.mockPublisher)

	// Store collection names for cleanup
	s.testCollections = []string{mongo.CollectionOrders, mongo.CollectionAutoInc}

	// Store context
	s.ctx = context.Background()
}

// Cleanup after each test
func (s *OrderIntegrationTestSuite) TearDownTest() {
	// Clean up test collections after each test
	for _, coll := range s.testCollections {
		_, err := s.mongoClient.Database("order_test_db").Collection(coll).DeleteMany(s.ctx, bson.M{})
		if err != nil {
			s.T().Logf("Failed to clean up collection %s: %v", coll, err)
		}
	}
}

// Cleanup after all tests
func (s *OrderIntegrationTestSuite) TearDownSuite() {
	// Close MongoDB connection
	if s.mongoClient != nil {
		err := s.mongoClient.Disconnect(s.ctx)
		if err != nil {
			s.T().Logf("Failed to disconnect from MongoDB: %v", err)
		}
	}
}

// Test creating an order and then retrieving it
func (s *OrderIntegrationTestSuite) TestCreateAndGetOrder() {
	// Create a test order
	testUserID := uint64(123)
	testProductID := uint64(456)

	orderItems := []domain.OrderItem{
		{
			ProductID: testProductID,
			Quantity:  2,
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
	assert.Equal(s.T(), testUserID, createdOrder.UserID)
	assert.Equal(s.T(), domain.StatusPending, createdOrder.Status)

	// Get the order
	filter := domain.OrderFilter{
		ID: &createdOrder.ID,
	}

	retrievedOrder, err := s.orderService.Get(s.ctx, filter)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), createdOrder.ID, retrievedOrder.ID)
	assert.Equal(s.T(), testUserID, retrievedOrder.UserID)
	assert.Equal(s.T(), domain.StatusPending, retrievedOrder.Status)
	assert.Len(s.T(), retrievedOrder.Items, 1)
	assert.Equal(s.T(), "Test Product", retrievedOrder.Items[0].Name)
	assert.Equal(s.T(), 100.0, retrievedOrder.Items[0].Price)
	assert.Equal(s.T(), 200.0, retrievedOrder.TotalAmount) // 2 items * $100 each
}

// Test order update functionality
func (s *OrderIntegrationTestSuite) TestUpdateOrder() {
	// Create a test order
	testUserID := uint64(123)
	testProductID := uint64(456)

	orderItems := []domain.OrderItem{
		{
			ProductID: testProductID,
			Quantity:  2,
		},
	}

	order := domain.Order{
		UserID: testUserID,
		Items:  orderItems,
	}

	// Create the order
	createdOrder, err := s.orderService.Create(s.ctx, order)
	assert.NoError(s.T(), err)

	// Update the order status
	filter := domain.OrderFilter{
		ID: &createdOrder.ID,
	}

	newStatus := domain.StatusPaid
	updateData := domain.OrderUpdateData{
		Status: &newStatus,
	}

	err = s.orderService.Update(s.ctx, filter, updateData)
	assert.NoError(s.T(), err)

	// Get the order to verify update
	retrievedOrder, err := s.orderService.Get(s.ctx, filter)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), domain.StatusPaid, retrievedOrder.Status)
}

// Test getting all orders for a user
func (s *OrderIntegrationTestSuite) TestGetAllOrders() {
	// Create test user ID
	testUserID := uint64(123)
	testProductID := uint64(456)

	// Create multiple orders for the same user
	for i := 0; i < 3; i++ {
		orderItems := []domain.OrderItem{
			{
				ProductID: testProductID,
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

	// Get all orders for the user
	filter := domain.OrderFilter{
		UserID: &testUserID,
	}

	orders, total, err := s.orderService.GetAll(s.ctx, filter, 1, 10)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), int64(3), total)
	assert.Len(s.T(), orders, 3)

	// Verify the orders have different quantities and correct total amounts
	var quantities []uint64
	var totalAmounts []float64

	for _, order := range orders {
		assert.Equal(s.T(), testUserID, order.UserID)
		assert.Len(s.T(), order.Items, 1)
		quantities = append(quantities, order.Items[0].Quantity)
		totalAmounts = append(totalAmounts, order.TotalAmount)
	}

	// Check that we have orders with quantities 1, 2, and 3
	assert.Contains(s.T(), quantities, uint64(1))
	assert.Contains(s.T(), quantities, uint64(2))
	assert.Contains(s.T(), quantities, uint64(3))

	// Check corresponding total amounts: $100, $200, $300
	assert.Contains(s.T(), totalAmounts, 100.0)
	assert.Contains(s.T(), totalAmounts, 200.0)
	assert.Contains(s.T(), totalAmounts, 300.0)
}

// Test deleting an order
func (s *OrderIntegrationTestSuite) TestDeleteOrder() {
	// Create a test order
	testUserID := uint64(123)
	testProductID := uint64(456)

	orderItems := []domain.OrderItem{
		{
			ProductID: testProductID,
			Quantity:  2,
		},
	}

	order := domain.Order{
		UserID: testUserID,
		Items:  orderItems,
	}

	// Create the order
	createdOrder, err := s.orderService.Create(s.ctx, order)
	assert.NoError(s.T(), err)

	// Delete the order
	filter := domain.OrderFilter{
		ID: &createdOrder.ID,
	}

	err = s.orderService.Delete(s.ctx, filter)
	assert.NoError(s.T(), err)

	// Try to get the deleted order
	_, err = s.orderService.Get(s.ctx, filter)
	assert.Error(s.T(), err) // Should return an error since the order is deleted
}

// Run the test suite
func TestOrderIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(OrderIntegrationTestSuite))
}
