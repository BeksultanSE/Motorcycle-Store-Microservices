package usecase

import (
	"context"
	"github.com/BeksultanSE/Assignment1-inventory/internal/adapter/mongo"
	"github.com/BeksultanSE/Assignment1-inventory/internal/domain"
	"log"
	"time"
)

type Product struct {
	aiRepo auto_inc_Repo
	repo   product_Repo
	cache  ProductCache
}

func NewProduct(aiRepo auto_inc_Repo, repo product_Repo, cache ProductCache) *Product {
	return &Product{
		aiRepo: aiRepo,
		repo:   repo,
		cache:  cache,
	}
}

func (p *Product) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	id, err := p.aiRepo.Next(ctx, mongo.CollectionProducts)
	if err != nil {
		return domain.Product{}, err
	}
	product.ID = id
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	err = p.repo.Create(ctx, product)
	if err != nil {
		return domain.Product{}, err
	}
	return domain.Product{
		ID:   id,
		Name: product.Name,
	}, nil
}

func (p *Product) Get(ctx context.Context, pf domain.ProductFilter) (domain.Product, error) {
	cachedProduct, err := p.cache.Get(ctx, *pf.ID)
	if err == nil && cachedProduct != (domain.Product{}) {
		log.Println("Cache hit:", cachedProduct)
		return cachedProduct, nil
	}

	product, err := p.repo.GetWithFilter(ctx, pf)
	if err != nil {
		return domain.Product{}, err
	}

	//caching
	if err = p.cache.Set(ctx, product); err != nil {
		log.Printf("Failed to cache product %d: %v", product.ID, err)
	}
	return product, nil
}

func (p *Product) GetAll(ctx context.Context, pf domain.ProductFilter, page, limit int64) ([]domain.Product, int, error) {
	products, totalCount, err := p.repo.GetListWithFilter(ctx, pf, page, limit)
	if err != nil {
		return nil, 0, err
	}
	return products, totalCount, nil
}

func (p *Product) Update(ctx context.Context, filter domain.ProductFilter, updated domain.ProductUpdateData) error {
	if *updated.Stock < uint64(0) {
		return domain.ErrInsufficientStock
	}
	updated.UpdatedAt = func() *time.Time { t := time.Now(); return &t }()
	err := p.repo.Update(ctx, filter, updated)
	if err != nil {
		return err
	}
	return nil
}

func (p *Product) Delete(ctx context.Context, filter domain.ProductFilter) error {
	err := p.repo.Delete(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}
