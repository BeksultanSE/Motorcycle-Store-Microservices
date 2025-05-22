package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/BeksultanSE/Assignment1-order/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockAutoIncRepo struct {
	mock.Mock
}

func (m *MockAutoIncRepo) Next(ctx context.Context, coll string) (uint64, error) {
	args := m.Called(ctx, coll)
	return args.Get(0).(uint64), args.Error(1)
}

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order domain.Order, id uint64) error {
	args := m.Called(ctx, order, id)
	return args.Error(0)
}

func (m *MockOrderRepository) Update(ctx context.Context, filter domain.OrderFilter, update domain.OrderUpdateData) error {
	args := m.Called(ctx, filter, update)
	return args.Error(0)
}

func (m *MockOrderRepository) GetWithFilter(ctx context.Context, filter domain.OrderFilter) (domain.Order, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(domain.Order), args.Error(1)
}

func (m *MockOrderRepository) GetAllWithFilter(ctx context.Context, filter domain.OrderFilter, page, limit int64) ([]domain.Order, int64, error) {
	args := m.Called(ctx, filter, page, limit)
	return args.Get(0).([]domain.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockOrderRepository) Delete(ctx context.Context, filter domain.OrderFilter) error {
	args := m.Called(ctx, filter)
	return args.Error(0)
}

type MockInventoryClient struct {
	mock.Mock
}

func (m *MockInventoryClient) GetProduct(ctx context.Context, productID uint64) (domain.Product, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).(domain.Product), args.Error(1)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishOrderCreated(ctx context.Context, event domain.Order) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestOrder_Create(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAutoInc := new(MockAutoIncRepo)
	mockRepo := new(MockOrderRepository)
	mockInventory := new(MockInventoryClient)
	mockPublisher := new(MockEventPublisher)

	orderSvc := NewOrder(mockAutoInc, mockRepo, mockInventory, mockPublisher)

	// Test data
	testOrderID := uint64(1)
	testUserID := uint64(123)
	testProductID := uint64(456)

	orderItems := []domain.OrderItem{
		{
			ProductID: testProductID,
			Quantity:  2,
		},
	}

	input := domain.Order{
		UserID: testUserID,
		Items:  orderItems,
	}

	product := domain.Product{
		ID:    testProductID,
		Name:  "Test Product",
		Price: 100.0,
		Stock: 5,
	}

	// Mocks setup
	mockInventory.On("GetProduct", ctx, testProductID).Return(product, nil)
	mockAutoInc.On("Next", ctx, "orders").Return(testOrderID, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("domain.Order"), testOrderID).Return(nil)
	mockPublisher.On("PublishOrderCreated", ctx, mock.AnythingOfType("domain.Order")).Return(nil)

	// Test execution
	result, err := orderSvc.Create(ctx, input)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, testOrderID, result.ID)
	assert.Equal(t, testUserID, result.UserID)
	assert.Equal(t, domain.StatusPending, result.Status)

	// Verify mocks were called correctly
	mockInventory.AssertExpectations(t)
	mockAutoInc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestOrder_Create_InsufficientStock(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAutoInc := new(MockAutoIncRepo)
	mockRepo := new(MockOrderRepository)
	mockInventory := new(MockInventoryClient)
	mockPublisher := new(MockEventPublisher)

	orderSvc := NewOrder(mockAutoInc, mockRepo, mockInventory, mockPublisher)

	// Test data
	testProductID := uint64(456)

	orderItems := []domain.OrderItem{
		{
			ProductID: testProductID,
			Quantity:  10, // More than stock
		},
	}

	input := domain.Order{
		UserID: uint64(123),
		Items:  orderItems,
	}

	product := domain.Product{
		ID:    testProductID,
		Name:  "Test Product",
		Price: 100.0,
		Stock: 5, // Less than requested
	}

	// Mocks setup
	mockInventory.On("GetProduct", ctx, testProductID).Return(product, nil)

	// Test execution
	_, err := orderSvc.Create(ctx, input)

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stock")

	// Verify only necessary mocks were called
	mockInventory.AssertExpectations(t)
}

func TestOrder_Get(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAutoInc := new(MockAutoIncRepo)
	mockRepo := new(MockOrderRepository)
	mockInventory := new(MockInventoryClient)
	mockPublisher := new(MockEventPublisher)

	orderSvc := NewOrder(mockAutoInc, mockRepo, mockInventory, mockPublisher)

	// Test data
	testOrderID := uint64(1)
	testUserID := uint64(123)
	testProductID := uint64(456)

	orderItems := []domain.OrderItem{
		{
			ProductID: testProductID,
			Quantity:  2,
		},
	}

	filter := domain.OrderFilter{
		ID: &testOrderID,
	}

	expectedOrder := domain.Order{
		ID:     testOrderID,
		UserID: testUserID,
		Items:  orderItems,
		Status: domain.StatusPending,
	}

	product := domain.Product{
		ID:    testProductID,
		Name:  "Test Product",
		Price: 100.0,
		Stock: 5,
	}

	// Mocks setup
	mockRepo.On("GetWithFilter", ctx, filter).Return(expectedOrder, nil)
	mockInventory.On("GetProduct", ctx, testProductID).Return(product, nil)

	// Test execution
	result, err := orderSvc.Get(ctx, filter)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, testOrderID, result.ID)
	assert.Equal(t, testUserID, result.UserID)
	assert.Equal(t, domain.StatusPending, result.Status)
	assert.Equal(t, 200.0, result.TotalAmount) // 2 * 100.0

	// Verify mocks were called correctly
	mockRepo.AssertExpectations(t)
	mockInventory.AssertExpectations(t)
}

func TestOrder_GetAll(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAutoInc := new(MockAutoIncRepo)
	mockRepo := new(MockOrderRepository)
	mockInventory := new(MockInventoryClient)
	mockPublisher := new(MockEventPublisher)

	orderSvc := NewOrder(mockAutoInc, mockRepo, mockInventory, mockPublisher)

	// Test data
	testProductID := uint64(456)

	orders := []domain.Order{
		{
			ID:     1,
			UserID: 123,
			Items: []domain.OrderItem{
				{
					ProductID: testProductID,
					Quantity:  2,
				},
			},
			Status: domain.StatusPending,
		},
		{
			ID:     2,
			UserID: 123,
			Items: []domain.OrderItem{
				{
					ProductID: testProductID,
					Quantity:  1,
				},
			},
			Status: domain.StatusShipped,
		},
	}

	product := domain.Product{
		ID:    testProductID,
		Name:  "Test Product",
		Price: 100.0,
		Stock: 5,
	}

	filter := domain.OrderFilter{
		UserID: &orders[0].UserID,
	}

	// Mocks setup
	mockRepo.On("GetAllWithFilter", ctx, filter, int64(1), int64(10)).Return(orders, int64(2), nil)
	mockInventory.On("GetProduct", ctx, testProductID).Return(product, nil).Times(2)

	// Test execution
	results, total, err := orderSvc.GetAll(ctx, filter, 1, 10)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
	assert.Equal(t, uint64(1), results[0].ID)
	assert.Equal(t, uint64(2), results[1].ID)
	assert.Equal(t, 200.0, results[0].TotalAmount) // 2 * 100.0
	assert.Equal(t, 100.0, results[1].TotalAmount) // 1 * 100.0

	// Verify mocks were called correctly
	mockRepo.AssertExpectations(t)
	mockInventory.AssertExpectations(t)
}

func TestOrder_Update(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAutoInc := new(MockAutoIncRepo)
	mockRepo := new(MockOrderRepository)
	mockInventory := new(MockInventoryClient)
	mockPublisher := new(MockEventPublisher)

	orderSvc := NewOrder(mockAutoInc, mockRepo, mockInventory, mockPublisher)

	// Test data
	testOrderID := uint64(1)

	filter := domain.OrderFilter{
		ID: &testOrderID,
	}

	status := domain.StatusPaid
	update := domain.OrderUpdateData{
		Status: &status,
	}

	// Mocks setup
	mockRepo.On("Update", ctx, filter, mock.AnythingOfType("domain.OrderUpdateData")).Return(nil)

	// Test execution
	err := orderSvc.Update(ctx, filter, update)

	// Assertions
	assert.NoError(t, err)

	// Verify mock UpdatedAt was set
	mockRepo.AssertCalled(t, "Update", ctx, filter, mock.MatchedBy(func(u domain.OrderUpdateData) bool {
		return u.Status == &status && u.UpdatedAt != nil
	}))
}

func TestOrder_Delete(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAutoInc := new(MockAutoIncRepo)
	mockRepo := new(MockOrderRepository)
	mockInventory := new(MockInventoryClient)
	mockPublisher := new(MockEventPublisher)

	orderSvc := NewOrder(mockAutoInc, mockRepo, mockInventory, mockPublisher)

	// Test data
	testOrderID := uint64(1)

	filter := domain.OrderFilter{
		ID: &testOrderID,
	}

	// Mocks setup
	mockRepo.On("Delete", ctx, filter).Return(nil)

	// Test execution
	err := orderSvc.Delete(ctx, filter)

	// Assertions
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestOrder_Create_RepositoryError(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAutoInc := new(MockAutoIncRepo)
	mockRepo := new(MockOrderRepository)
	mockInventory := new(MockInventoryClient)
	mockPublisher := new(MockEventPublisher)

	orderSvc := NewOrder(mockAutoInc, mockRepo, mockInventory, mockPublisher)

	// Test data
	testOrderID := uint64(1)
	testUserID := uint64(123)
	testProductID := uint64(456)

	orderItems := []domain.OrderItem{
		{
			ProductID: testProductID,
			Quantity:  2,
		},
	}

	input := domain.Order{
		UserID: testUserID,
		Items:  orderItems,
	}

	product := domain.Product{
		ID:    testProductID,
		Name:  "Test Product",
		Price: 100.0,
		Stock: 5,
	}

	// Mocks setup
	mockInventory.On("GetProduct", ctx, testProductID).Return(product, nil)
	mockAutoInc.On("Next", ctx, "orders").Return(testOrderID, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("domain.Order"), testOrderID).Return(errors.New("database error"))

	// Test execution
	_, err := orderSvc.Create(ctx, input)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())

	// Verify mocks were called correctly
	mockInventory.AssertExpectations(t)
	mockAutoInc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
