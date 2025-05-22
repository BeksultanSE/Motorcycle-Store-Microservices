package integration

import (
	"context"
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

// ProductServiceIntegrationTestSuite for testing the product service with real dependencies
type ProductServiceIntegrationTestSuite struct {
	suite.Suite
	ctx             context.Context
	mongoClient     *mongodriver.Client
	productRepo     *mongo.ProductRepo
	autoIncRepo     *mongo.AutoInc
	mockCache       *mockProductCache
	productService  *usecase.Product
	testCollections []string
}

// Setup test suite
func (s *ProductServiceIntegrationTestSuite) SetupSuite() {
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
	dbName := "inventory_service_test_db"

	// Create repositories
	s.autoIncRepo = mongo.NewAutoInc(s.mongoClient.Database(dbName))
	s.productRepo = mongo.NewProductRepo(s.mongoClient.Database(dbName))

	// Store collection names for cleanup
	s.testCollections = []string{mongo.CollectionProducts, mongo.CollectionAutoInc}

	// Store context
	s.ctx = context.Background()
}

// SetupTest runs before each test
func (s *ProductServiceIntegrationTestSuite) SetupTest() {
	// Create a new mock cache for each test to avoid cross-test cache pollution
	s.mockCache = newMockProductCache()

	// Create product service with real repositories and fresh mock cache
	s.productService = usecase.NewProduct(s.autoIncRepo, s.productRepo, s.mockCache)
}

// Cleanup after each test
func (s *ProductServiceIntegrationTestSuite) TearDownTest() {
	// Clean up test collections after each test
	for _, coll := range s.testCollections {
		_, err := s.mongoClient.Database("inventory_service_test_db").Collection(coll).DeleteMany(s.ctx, bson.M{})
		if err != nil {
			s.T().Logf("Failed to clean up collection %s: %v", coll, err)
		}
	}
}

// Cleanup after all tests
func (s *ProductServiceIntegrationTestSuite) TearDownSuite() {
	// Close MongoDB connection
	if s.mongoClient != nil {
		err := s.mongoClient.Disconnect(s.ctx)
		if err != nil {
			s.T().Logf("Failed to disconnect from MongoDB: %v", err)
		}
	}
}

// Test creating a product and verifying cache update
func (s *ProductServiceIntegrationTestSuite) TestProductCreateAndCache() {
	// Create a test product
	product := domain.Product{
		Name:     "Premium Helmet",
		Category: "Safety",
		Price:    200.0,
		Stock:    15,
	}

	// Create the product
	createdProduct, err := s.productService.Create(s.ctx, product)
	assert.NoError(s.T(), err)
	assert.NotEqual(s.T(), uint64(0), createdProduct.ID)

	// Get the product (should check cache first)
	filter := domain.ProductFilter{
		ID: &createdProduct.ID,
	}

	retrievedProduct, err := s.productService.Get(s.ctx, filter)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), createdProduct.ID, retrievedProduct.ID)
	assert.Equal(s.T(), product.Name, retrievedProduct.Name)
	assert.Equal(s.T(), product.Category, retrievedProduct.Category)
	assert.Equal(s.T(), product.Price, retrievedProduct.Price)
	assert.Equal(s.T(), product.Stock, retrievedProduct.Stock)

	// Verify cache has been populated
	cachedProduct, err := s.mockCache.Get(s.ctx, createdProduct.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), createdProduct.ID, cachedProduct.ID)
}

// Test updating a product and verifying cache invalidation
func (s *ProductServiceIntegrationTestSuite) TestProductUpdateAndCacheInvalidation() {
	// Create a test product
	product := domain.Product{
		Name:     "Standard Jacket",
		Category: "Apparel",
		Price:    150.0,
		Stock:    10,
	}

	// Create the product
	createdProduct, err := s.productService.Create(s.ctx, product)
	assert.NoError(s.T(), err)

	// Update the product
	filter := domain.ProductFilter{
		ID: &createdProduct.ID,
	}

	newName := "Premium Jacket"
	newPrice := 200.0
	newStock := uint64(5)

	updateData := domain.ProductUpdateData{
		Name:  &newName,
		Price: &newPrice,
		Stock: &newStock,
	}

	err = s.productService.Update(s.ctx, filter, updateData)
	assert.NoError(s.T(), err)

	// Get the updated product (should update cache)
	retrievedProduct, err := s.productService.Get(s.ctx, filter)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), newName, retrievedProduct.Name)
	assert.Equal(s.T(), newPrice, retrievedProduct.Price)
	assert.Equal(s.T(), newStock, retrievedProduct.Stock)

	// Verify cache has been updated
	cachedProduct, err := s.mockCache.Get(s.ctx, createdProduct.ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), newName, cachedProduct.Name)
	assert.Equal(s.T(), newPrice, cachedProduct.Price)
	assert.Equal(s.T(), newStock, cachedProduct.Stock)
}

// Test product deletion and cache invalidation
func (s *ProductServiceIntegrationTestSuite) TestProductDeleteAndCacheInvalidation() {
	// Create a test product
	product := domain.Product{
		Name:     "Gloves for Deletion",
		Category: "Accessories",
		Price:    45.0,
		Stock:    8,
	}

	// Create the product
	createdProduct, err := s.productService.Create(s.ctx, product)
	assert.NoError(s.T(), err)

	// Ensure product is in cache by getting it first
	filter := domain.ProductFilter{
		ID: &createdProduct.ID,
	}
	_, err = s.productService.Get(s.ctx, filter)
	assert.NoError(s.T(), err)

	// Delete the product
	err = s.productService.Delete(s.ctx, filter)
	assert.NoError(s.T(), err)

	// Try to get the deleted product
	_, err = s.productService.Get(s.ctx, filter)
	assert.Error(s.T(), err) // Should return an error

	// Verify it's deleted from cache
	_, err = s.mockCache.Get(s.ctx, createdProduct.ID)
	assert.Error(s.T(), err) // Should return a cache miss error
}

// Test batch operations and cache management
func (s *ProductServiceIntegrationTestSuite) TestBatchOperationsAndCache() {
	// Create multiple products
	testProducts := []domain.Product{
		{Name: "Item 1", Category: "Category A", Price: 100.0, Stock: 10},
		{Name: "Item 2", Category: "Category A", Price: 150.0, Stock: 15},
		{Name: "Item 3", Category: "Category B", Price: 200.0, Stock: 20},
	}

	var createdProducts []domain.Product
	for _, product := range testProducts {
		created, err := s.productService.Create(s.ctx, product)
		assert.NoError(s.T(), err)
		createdProducts = append(createdProducts, created)
	}

	// Get all products for Category A
	categoryA := "Category A"
	filter := domain.ProductFilter{
		Category: &categoryA,
	}

	products, total, err := s.productService.GetAll(s.ctx, filter, 1, 10)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 2, int(total)) // Convert int64 to int for comparison
	assert.Len(s.T(), products, 2)

	// Verify cache has been populated with these products
	for _, product := range createdProducts {
		if product.Category == "Category A" {
			cachedProduct, err := s.mockCache.Get(s.ctx, product.ID)
			assert.NoError(s.T(), err)
			assert.Equal(s.T(), product.ID, cachedProduct.ID)
		}
	}

	// Update stock for one product
	productIDToUpdate := createdProducts[0].ID
	filterUpdate := domain.ProductFilter{
		ID: &productIDToUpdate,
	}

	newStock := uint64(5)
	updateData := domain.ProductUpdateData{
		Stock: &newStock,
	}

	err = s.productService.Update(s.ctx, filterUpdate, updateData)
	assert.NoError(s.T(), err)

	// Verify the stock update
	updatedProduct, err := s.productService.Get(s.ctx, filterUpdate)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), newStock, updatedProduct.Stock)
}

// Run the test suite
func TestProductServiceIntegrationSuite(t *testing.T) {
	suite.Run(t, new(ProductServiceIntegrationTestSuite))
}
