package integration

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/BeksultanSE/Assignment1-inventory/internal/adapter/mongo"
	"github.com/BeksultanSE/Assignment1-inventory/internal/domain"
	"github.com/BeksultanSE/Assignment1-inventory/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Define a cache miss error for the mock
var ErrCacheMiss = errors.New("cache miss")

// Integration test suite for product service
type ProductIntegrationTestSuite struct {
	suite.Suite
	ctx             context.Context
	mongoClient     *mongodriver.Client
	productRepo     *mongo.ProductRepo
	autoIncRepo     *mongo.AutoInc
	mockCache       *mockProductCache
	productService  *usecase.Product
	testCollections []string
}

// Mock product cache for testing
type mockProductCache struct {
	cache map[uint64]domain.Product
}

func newMockProductCache() *mockProductCache {
	return &mockProductCache{
		cache: make(map[uint64]domain.Product),
	}
}

func (m *mockProductCache) Get(ctx context.Context, productID uint64) (domain.Product, error) {
	if product, ok := m.cache[productID]; ok {
		return product, nil
	}
	return domain.Product{}, ErrCacheMiss
}

func (m *mockProductCache) Set(ctx context.Context, product domain.Product) error {
	m.cache[product.ID] = product
	return nil
}

func (m *mockProductCache) SetMany(ctx context.Context, products []domain.Product) error {
	for _, product := range products {
		m.cache[product.ID] = product
	}
	return nil
}

func (m *mockProductCache) Delete(ctx context.Context, productID uint64) error {
	delete(m.cache, productID)
	return nil
}

// Setup test suite
func (s *ProductIntegrationTestSuite) SetupSuite() {
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
	dbName := "inventory_test_db"

	// Create repositories
	s.autoIncRepo = mongo.NewAutoInc(s.mongoClient.Database(dbName))
	s.productRepo = mongo.NewProductRepo(s.mongoClient.Database(dbName))

	// Store collection names for cleanup
	s.testCollections = []string{mongo.CollectionProducts, mongo.CollectionAutoInc}

	// Store context
	s.ctx = context.Background()
}

// SetupTest runs before each test
func (s *ProductIntegrationTestSuite) SetupTest() {
	// Create a new mock cache for each test to avoid cross-test cache pollution
	s.mockCache = newMockProductCache()

	// Create product service with real repositories and fresh mock cache
	s.productService = usecase.NewProduct(s.autoIncRepo, s.productRepo, s.mockCache)
}

// Cleanup after each test
func (s *ProductIntegrationTestSuite) TearDownTest() {
	// Clean up test collections after each test
	for _, coll := range s.testCollections {
		_, err := s.mongoClient.Database("inventory_test_db").Collection(coll).DeleteMany(s.ctx, bson.M{})
		if err != nil {
			s.T().Logf("Failed to clean up collection %s: %v", coll, err)
		}
	}
}

// Cleanup after all tests
func (s *ProductIntegrationTestSuite) TearDownSuite() {
	// Close MongoDB connection
	if s.mongoClient != nil {
		err := s.mongoClient.Disconnect(s.ctx)
		if err != nil {
			s.T().Logf("Failed to disconnect from MongoDB: %v", err)
		}
	}
}

// Test creating a product and then retrieving it
func (s *ProductIntegrationTestSuite) TestCreateAndGetProduct() {
	// Create a test product
	product := domain.Product{
		Name:     "Test Motorcycle Helmet",
		Category: "Safety",
		Price:    150.0,
		Stock:    20,
	}

	// Create the product
	createdProduct, err := s.productService.Create(s.ctx, product)
	assert.NoError(s.T(), err)
	assert.NotEqual(s.T(), uint64(0), createdProduct.ID)

	// Debug log to see what's in the created product
	log.Printf("Created product: %+v", createdProduct)

	// Note: createdProduct only contains ID and possibly Name,
	// but not all fields are populated by the Create method

	// Get the product
	filter := domain.ProductFilter{
		ID: &createdProduct.ID,
	}

	retrievedProduct, err := s.productService.Get(s.ctx, filter)
	assert.NoError(s.T(), err)

	// Debug log to see what's in the retrieved product
	log.Printf("Retrieved product: %+v", retrievedProduct)

	assert.Equal(s.T(), createdProduct.ID, retrievedProduct.ID)

	// Compare with the original product values
	// since createdProduct doesn't have all fields populated
	assert.Equal(s.T(), product.Name, retrievedProduct.Name, "Name mismatch")
	assert.Equal(s.T(), product.Category, retrievedProduct.Category, "Category mismatch")
	assert.Equal(s.T(), product.Price, retrievedProduct.Price, "Price mismatch")
	assert.Equal(s.T(), product.Stock, retrievedProduct.Stock, "Stock mismatch")
}

// Test product update functionality
func (s *ProductIntegrationTestSuite) TestUpdateProduct() {
	// Create a test product
	product := domain.Product{
		Name:     "Test Motorcycle Gloves",
		Category: "Accessories",
		Price:    50.0,
		Stock:    30,
	}

	// Create the product
	createdProduct, err := s.productService.Create(s.ctx, product)
	assert.NoError(s.T(), err)

	// Update the product
	filter := domain.ProductFilter{
		ID: &createdProduct.ID,
	}

	newName := "Premium Motorcycle Gloves"
	newPrice := 75.0
	newStock := uint64(25)

	updateData := domain.ProductUpdateData{
		Name:  &newName,
		Price: &newPrice,
		Stock: &newStock,
	}

	err = s.productService.Update(s.ctx, filter, updateData)
	assert.NoError(s.T(), err)

	// Get the product to verify update
	retrievedProduct, err := s.productService.Get(s.ctx, filter)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), newName, retrievedProduct.Name)
	assert.Equal(s.T(), newPrice, retrievedProduct.Price)
	assert.Equal(s.T(), newStock, retrievedProduct.Stock)
	assert.Equal(s.T(), product.Category, retrievedProduct.Category) // Category shouldn't change
}

// Test getting products with pagination
func (s *ProductIntegrationTestSuite) TestGetProductsWithPagination() {
	// Create multiple products
	testProducts := []domain.Product{
		{Name: "Helmet A", Category: "Safety", Price: 100.0, Stock: 10},
		{Name: "Helmet B", Category: "Safety", Price: 120.0, Stock: 15},
		{Name: "Jacket A", Category: "Apparel", Price: 200.0, Stock: 5},
		{Name: "Jacket B", Category: "Apparel", Price: 250.0, Stock: 8},
		{Name: "Gloves A", Category: "Accessories", Price: 40.0, Stock: 20},
	}

	for _, product := range testProducts {
		_, err := s.productService.Create(s.ctx, product)
		assert.NoError(s.T(), err)
	}

	// Test pagination - page 1, limit 2
	products, total, err := s.productService.GetAll(s.ctx, domain.ProductFilter{}, 1, 2)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 5, int(total))
	assert.Len(s.T(), products, 2)

	// Test pagination - page 2, limit 2
	products, total, err = s.productService.GetAll(s.ctx, domain.ProductFilter{}, 2, 2)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 5, int(total))
	assert.Len(s.T(), products, 2)

	// Test pagination - page 3, limit 2
	products, total, err = s.productService.GetAll(s.ctx, domain.ProductFilter{}, 3, 2)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 5, int(total))
	assert.Len(s.T(), products, 1) // Only 1 product on the last page

	// Test filtering by category
	safetyCategory := "Safety"
	filter := domain.ProductFilter{
		Category: &safetyCategory,
	}

	products, total, err = s.productService.GetAll(s.ctx, filter, 1, 10)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, int(total))
	assert.Len(s.T(), products, 2)

	for _, product := range products {
		assert.Equal(s.T(), "Safety", product.Category)
	}
}

// Test deleting a product
func (s *ProductIntegrationTestSuite) TestDeleteProduct() {
	// Create a test product
	product := domain.Product{
		Name:     "Test Product to Delete",
		Category: "Test",
		Price:    75.0,
		Stock:    5,
	}

	// Create the product
	createdProduct, err := s.productService.Create(s.ctx, product)
	assert.NoError(s.T(), err)

	// Delete the product
	filter := domain.ProductFilter{
		ID: &createdProduct.ID,
	}

	err = s.productService.Delete(s.ctx, filter)
	assert.NoError(s.T(), err)

	// Try to get the deleted product
	_, err = s.productService.Get(s.ctx, filter)
	assert.Error(s.T(), err) // Should return an error
}

// Run the test suite
func TestProductIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ProductIntegrationTestSuite))
}
