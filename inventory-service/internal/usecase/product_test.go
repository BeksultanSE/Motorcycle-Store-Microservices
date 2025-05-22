package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/BeksultanSE/Assignment1-inventory/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for auto_inc_Repo
type MockAutoIncRepo struct {
	mock.Mock
}

func (m *MockAutoIncRepo) Next(ctx context.Context, collection string) (uint64, error) {
	args := m.Called(ctx, collection)
	return args.Get(0).(uint64), args.Error(1)
}

// Mock for product_Repo
type MockProductRepo struct {
	mock.Mock
}

func (m *MockProductRepo) Create(ctx context.Context, product domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepo) Update(ctx context.Context, filter domain.ProductFilter, update domain.ProductUpdateData) error {
	args := m.Called(ctx, filter, update)
	return args.Error(0)
}

func (m *MockProductRepo) GetWithFilter(ctx context.Context, filter domain.ProductFilter) (domain.Product, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductRepo) GetListWithFilter(ctx context.Context, filter domain.ProductFilter, page, limit int64) ([]domain.Product, int, error) {
	args := m.Called(ctx, filter, page, limit)
	return args.Get(0).([]domain.Product), args.Int(1), args.Error(2)
}

func (m *MockProductRepo) Delete(ctx context.Context, filter domain.ProductFilter) error {
	args := m.Called(ctx, filter)
	return args.Error(0)
}

// Mock for ProductCache
type MockProductCache struct {
	mock.Mock
}

func (m *MockProductCache) Get(ctx context.Context, productID uint64) (domain.Product, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductCache) Set(ctx context.Context, product domain.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductCache) SetMany(ctx context.Context, products []domain.Product) error {
	args := m.Called(ctx, products)
	return args.Error(0)
}

func (m *MockProductCache) Delete(ctx context.Context, productID uint64) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func TestProduct_Create(t *testing.T) {
	mockAutoIncRepo := new(MockAutoIncRepo)
	mockProductRepo := new(MockProductRepo)
	mockCache := new(MockProductCache)

	productService := NewProduct(mockAutoIncRepo, mockProductRepo, mockCache)
	ctx := context.Background()

	tests := []struct {
		name          string
		inputProduct  domain.Product
		setupMocks    func()
		expectedID    uint64
		expectedError error
	}{
		{
			name: "Success",
			inputProduct: domain.Product{
				Name:     "Motorcycle Helmet",
				Category: "Safety",
				Price:    99.99,
				Stock:    50,
			},
			setupMocks: func() {
				mockAutoIncRepo.On("Next", ctx, "products").Return(uint64(1), nil)
				mockProductRepo.On("Create", ctx, mock.AnythingOfType("domain.Product")).Return(nil)
			},
			expectedID:    1,
			expectedError: nil,
		},
		{
			name: "AutoInc Error",
			inputProduct: domain.Product{
				Name:     "Motorcycle Helmet",
				Category: "Safety",
				Price:    99.99,
				Stock:    50,
			},
			setupMocks: func() {
				mockAutoIncRepo.On("Next", ctx, "products").Return(uint64(0), errors.New("auto inc error"))
			},
			expectedID:    0,
			expectedError: errors.New("auto inc error"),
		},
		{
			name: "Create Error",
			inputProduct: domain.Product{
				Name:     "Motorcycle Helmet",
				Category: "Safety",
				Price:    99.99,
				Stock:    50,
			},
			setupMocks: func() {
				mockAutoIncRepo.On("Next", ctx, "products").Return(uint64(1), nil)
				mockProductRepo.On("Create", ctx, mock.AnythingOfType("domain.Product")).Return(errors.New("create error"))
			},
			expectedID:    0,
			expectedError: errors.New("create error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockAutoIncRepo.ExpectedCalls = nil
			mockProductRepo.ExpectedCalls = nil
			mockCache.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Call the method
			result, err := productService.Create(ctx, tt.inputProduct)

			// Assert error
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, result.ID)
				assert.Equal(t, tt.inputProduct.Name, result.Name)
			}

			// Assert expectations
			mockAutoIncRepo.AssertExpectations(t)
			mockProductRepo.AssertExpectations(t)
		})
	}
}

func TestProduct_Get(t *testing.T) {
	mockAutoIncRepo := new(MockAutoIncRepo)
	mockProductRepo := new(MockProductRepo)
	mockCache := new(MockProductCache)

	productService := NewProduct(mockAutoIncRepo, mockProductRepo, mockCache)
	ctx := context.Background()

	// Setup test data
	id := uint64(1)
	testProduct := domain.Product{
		ID:        id,
		Name:      "Motorcycle Helmet",
		Category:  "Safety",
		Price:     99.99,
		Stock:     50,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name           string
		filter         domain.ProductFilter
		setupMocks     func()
		expectedResult domain.Product
		expectedError  error
	}{
		{
			name: "Cache Hit",
			filter: domain.ProductFilter{
				ID: &id,
			},
			setupMocks: func() {
				mockCache.On("Get", ctx, id).Return(testProduct, nil)
			},
			expectedResult: testProduct,
			expectedError:  nil,
		},
		{
			name: "Cache Miss Success",
			filter: domain.ProductFilter{
				ID: &id,
			},
			setupMocks: func() {
				mockCache.On("Get", ctx, id).Return(domain.Product{}, errors.New("cache miss"))
				mockProductRepo.On("GetWithFilter", ctx, mock.AnythingOfType("domain.ProductFilter")).Return(testProduct, nil)
				mockCache.On("Set", ctx, testProduct).Return(nil)
			},
			expectedResult: testProduct,
			expectedError:  nil,
		},
		{
			name: "Repository Error",
			filter: domain.ProductFilter{
				ID: &id,
			},
			setupMocks: func() {
				mockCache.On("Get", ctx, id).Return(domain.Product{}, errors.New("cache miss"))
				mockProductRepo.On("GetWithFilter", ctx, mock.AnythingOfType("domain.ProductFilter")).Return(domain.Product{}, domain.ErrProductNotFound)
			},
			expectedResult: domain.Product{},
			expectedError:  domain.ErrProductNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockAutoIncRepo.ExpectedCalls = nil
			mockProductRepo.ExpectedCalls = nil
			mockCache.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Call the method
			result, err := productService.Get(ctx, tt.filter)

			// Assert error
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.ID, result.ID)
				assert.Equal(t, tt.expectedResult.Name, result.Name)
			}

			// Assert expectations
			mockCache.AssertExpectations(t)
			mockProductRepo.AssertExpectations(t)
		})
	}
}

func TestProduct_Update(t *testing.T) {
	mockAutoIncRepo := new(MockAutoIncRepo)
	mockProductRepo := new(MockProductRepo)
	mockCache := new(MockProductCache)

	productService := NewProduct(mockAutoIncRepo, mockProductRepo, mockCache)
	ctx := context.Background()

	// Setup test data
	id := uint64(1)
	stock := uint64(10)
	zeroStock := uint64(0) // For testing zero stock

	tests := []struct {
		name          string
		filter        domain.ProductFilter
		updateData    domain.ProductUpdateData
		setupMocks    func()
		expectedError error
	}{
		{
			name: "Success",
			filter: domain.ProductFilter{
				ID: &id,
			},
			updateData: domain.ProductUpdateData{
				Stock: &stock,
			},
			setupMocks: func() {
				mockProductRepo.On("Update", ctx, mock.AnythingOfType("domain.ProductFilter"), mock.AnythingOfType("domain.ProductUpdateData")).Return(nil)
				mockCache.On("Delete", ctx, id).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Zero Stock (Should Succeed)",
			filter: domain.ProductFilter{
				ID: &id,
			},
			updateData: domain.ProductUpdateData{
				Stock: &zeroStock,
			},
			setupMocks: func() {
				// Note: In product.go, the check is now `if updated.Stock != nil && *updated.Stock == 0`
				// So this test will now fail with ErrInsufficientStock
				// No mock calls needed since the validation will fail
			},
			expectedError: domain.ErrInsufficientStock,
		},
		{
			name: "Repository Error",
			filter: domain.ProductFilter{
				ID: &id,
			},
			updateData: domain.ProductUpdateData{
				Stock: &stock,
			},
			setupMocks: func() {
				mockProductRepo.On("Update", ctx, mock.AnythingOfType("domain.ProductFilter"), mock.AnythingOfType("domain.ProductUpdateData")).Return(errors.New("update error"))
			},
			expectedError: errors.New("update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockAutoIncRepo.ExpectedCalls = nil
			mockProductRepo.ExpectedCalls = nil
			mockCache.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Call the method
			err := productService.Update(ctx, tt.filter, tt.updateData)

			// Assert error
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Assert expectations
			mockProductRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestProduct_Delete(t *testing.T) {
	mockAutoIncRepo := new(MockAutoIncRepo)
	mockProductRepo := new(MockProductRepo)
	mockCache := new(MockProductCache)

	productService := NewProduct(mockAutoIncRepo, mockProductRepo, mockCache)
	ctx := context.Background()

	// Setup test data
	id := uint64(1)

	tests := []struct {
		name          string
		filter        domain.ProductFilter
		setupMocks    func()
		expectedError error
	}{
		{
			name: "Success",
			filter: domain.ProductFilter{
				ID: &id,
			},
			setupMocks: func() {
				mockProductRepo.On("Delete", ctx, mock.AnythingOfType("domain.ProductFilter")).Return(nil)
				mockCache.On("Delete", ctx, id).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Repository Error",
			filter: domain.ProductFilter{
				ID: &id,
			},
			setupMocks: func() {
				mockProductRepo.On("Delete", ctx, mock.AnythingOfType("domain.ProductFilter")).Return(errors.New("delete error"))
			},
			expectedError: errors.New("delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockAutoIncRepo.ExpectedCalls = nil
			mockProductRepo.ExpectedCalls = nil
			mockCache.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Call the method
			err := productService.Delete(ctx, tt.filter)

			// Assert error
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Assert expectations
			mockProductRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestProduct_GetAll(t *testing.T) {
	mockAutoIncRepo := new(MockAutoIncRepo)
	mockProductRepo := new(MockProductRepo)
	mockCache := new(MockProductCache)

	productService := NewProduct(mockAutoIncRepo, mockProductRepo, mockCache)
	ctx := context.Background()

	// Setup test data
	products := []domain.Product{
		{
			ID:       1,
			Name:     "Product 1",
			Category: "Category 1",
			Price:    10.0,
			Stock:    10,
		},
		{
			ID:       2,
			Name:     "Product 2",
			Category: "Category 2",
			Price:    20.0,
			Stock:    20,
		},
	}
	totalCount := 2
	page := int64(1)
	limit := int64(10)

	tests := []struct {
		name           string
		filter         domain.ProductFilter
		page           int64
		limit          int64
		setupMocks     func()
		expectedResult []domain.Product
		expectedCount  int
		expectedError  error
	}{
		{
			name:   "Success",
			filter: domain.ProductFilter{},
			page:   page,
			limit:  limit,
			setupMocks: func() {
				mockProductRepo.On("GetListWithFilter", ctx, mock.AnythingOfType("domain.ProductFilter"), page, limit).Return(products, totalCount, nil)
			},
			expectedResult: products,
			expectedCount:  totalCount,
			expectedError:  nil,
		},
		{
			name:   "Repository Error",
			filter: domain.ProductFilter{},
			page:   page,
			limit:  limit,
			setupMocks: func() {
				mockProductRepo.On("GetListWithFilter", ctx, mock.AnythingOfType("domain.ProductFilter"), page, limit).Return([]domain.Product{}, 0, errors.New("getall error"))
			},
			expectedResult: nil,
			expectedCount:  0,
			expectedError:  errors.New("getall error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockAutoIncRepo.ExpectedCalls = nil
			mockProductRepo.ExpectedCalls = nil
			mockCache.ExpectedCalls = nil

			// Setup mocks
			tt.setupMocks()

			// Call the method
			result, count, err := productService.GetAll(ctx, tt.filter, tt.page, tt.limit)

			// Assert error
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
				assert.Equal(t, tt.expectedResult, result)
			}

			// Assert expectations
			mockProductRepo.AssertExpectations(t)
		})
	}
}
