package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go-redis-cache-api/internal/cache"
	"go-redis-cache-api/internal/model"
	"go-redis-cache-api/internal/repository"
)

type ProductService struct {
	repo       *repository.ProductRepository
	redisCache *cache.RedisCache
	memCache   *cache.MemoryCache
	writeBack  chan *model.Product
}

func NewProductService(repo *repository.ProductRepository, redisCache *cache.RedisCache, memCache *cache.MemoryCache) *ProductService {
	svc := &ProductService{
		repo:       repo,
		redisCache: redisCache,
		memCache:   memCache,
		writeBack:  make(chan *model.Product, 100),
	}
	go svc.writeBackProcessor()
	return svc
}

func (s *ProductService) productKey(id string) string {
	return fmt.Sprintf("product:%s", id)
}

func (s *ProductService) allProductsKey() string {
	return "products:all"
}

func duration(format string) string {
	start := time.Now()
	defer func() {
		log.Printf("[TIMING] %s took %s", format, time.Since(start))
	}()
	return time.Since(start).String()
}

func measure(start time.Time) string {
	return time.Since(start).String()
}

func (s *ProductService) GetProductByID(ctx context.Context, id string) (*model.ProductResponse, error) {
	start := time.Now()

	var product model.Product

	err := s.redisCache.Get(ctx, s.productKey(id), &product)
	if err == nil {
		return &model.ProductResponse{
			Data:    product,
			Source:  "Redis cache",
			Latency: time.Since(start).String(),
		}, nil
	}

	err = s.memCache.Get(ctx, s.productKey(id), &product)
	if err == nil {
		return &model.ProductResponse{
			Data:    product,
			Source:  "In-memory cache",
			Latency: time.Since(start).String(),
		}, nil
	}

	productPtr, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.redisCache.Set(ctx, s.productKey(id), productPtr)
	_ = s.memCache.Set(ctx, s.productKey(id), productPtr)

	return &model.ProductResponse{
		Data:    productPtr,
		Source:  "Database",
		Latency: time.Since(start).String(),
	}, nil
}

func (s *ProductService) GetAllProducts(ctx context.Context) (*model.ProductResponse, error) {
	start := time.Now()

	var products []model.Product
	err := s.redisCache.Get(ctx, s.allProductsKey(), &products)
	if err == nil {
		return &model.ProductResponse{
			Data:    products,
			Source:  "Redis cache",
			Latency: time.Since(start).String(),
		}, nil
	}

	err = s.memCache.Get(ctx, s.allProductsKey(), &products)
	if err == nil {
		return &model.ProductResponse{
			Data:    products,
			Source:  "In-memory cache",
			Latency: time.Since(start).String(),
		}, nil
	}

	products, err = s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(products) > 0 {
		_ = s.redisCache.Set(ctx, s.allProductsKey(), products)
		_ = s.memCache.Set(ctx, s.allProductsKey(), products)
	}

	return &model.ProductResponse{
		Data:    products,
		Source:  "Database",
		Latency: time.Since(start).String(),
	}, nil
}

func (s *ProductService) CreateProductWriteThrough(ctx context.Context, req model.CreateProductRequest) (*model.ProductResponse, error) {
	start := time.Now()

	product, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	_ = s.redisCache.Set(ctx, s.productKey(product.ID), product)
	_ = s.memCache.Set(ctx, s.productKey(product.ID), product)
	_ = s.redisCache.Delete(ctx, s.allProductsKey())
	_ = s.memCache.Delete(ctx, s.allProductsKey())

	return &model.ProductResponse{
		Data:    product,
		Source:  "Database + Cache (Write-Through)",
		Latency: time.Since(start).String(),
	}, nil
}

func (s *ProductService) CreateProductWriteBack(ctx context.Context, req model.CreateProductRequest) (*model.ProductResponse, error) {
	start := time.Now()

	product := &model.Product{
		ID:          fmt.Sprintf("pending-%d", time.Now().UnixNano()),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_ = s.redisCache.SetWithTTL(ctx, s.productKey(product.ID), product, 30*time.Minute)
	_ = s.memCache.SetWithTTL(ctx, s.productKey(product.ID), product, 30*time.Minute)

	s.writeBack <- product

	return &model.ProductResponse{
		Data:    product,
		Source:  "Write-Back Cache (async DB write pending)",
		Latency: time.Since(start).String(),
	}, nil
}

func (s *ProductService) writeBackProcessor() {
	for product := range s.writeBack {
		time.Sleep(2 * time.Second)
		ctx := context.Background()

		req := model.CreateProductRequest{
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
		}
		saved, err := s.repo.Create(ctx, req)
		if err != nil {
			log.Printf("[Write-Back] Failed to persist product %s: %v", product.ID, err)
			continue
		}

		_ = s.redisCache.Delete(ctx, s.productKey(product.ID))
		_ = s.memCache.Delete(ctx, s.productKey(product.ID))
		_ = s.redisCache.Set(ctx, s.productKey(saved.ID), saved)
		_ = s.memCache.Set(ctx, s.productKey(saved.ID), saved)
		_ = s.redisCache.Delete(ctx, s.allProductsKey())
		_ = s.memCache.Delete(ctx, s.allProductsKey())

		log.Printf("[Write-Back] Product %s -> %s persisted to DB", product.ID, saved.ID)
	}
}

func (s *ProductService) UpdateProductWriteAround(ctx context.Context, id string, req model.UpdateProductRequest) (*model.ProductResponse, error) {
	start := time.Now()

	product, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	_ = s.redisCache.Delete(ctx, s.productKey(id))
	_ = s.memCache.Delete(ctx, s.productKey(id))
	_ = s.redisCache.Delete(ctx, s.allProductsKey())
	_ = s.memCache.Delete(ctx, s.allProductsKey())

	return &model.ProductResponse{
		Data:    product,
		Source:  "Database (Write-Around, cache invalidated)",
		Latency: time.Since(start).String(),
	}, nil
}

func (s *ProductService) GetProductNoCache(ctx context.Context, id string) (*model.ProductResponse, error) {
	start := time.Now()

	time.Sleep(50 * time.Millisecond)

	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.ProductResponse{
		Data:    product,
		Source:  "Database (no cache - simulated slow)",
		Latency: time.Since(start).String(),
	}, nil
}

func (s *ProductService) Benchmark(ctx context.Context, id string) (map[string]string, error) {
	results := make(map[string]string)

	start := time.Now()
	_ = s.redisCache.Get(ctx, s.productKey(id), &model.Product{})
	results["redis_cache"] = time.Since(start).String()

	start = time.Now()
	_ = s.memCache.Get(ctx, s.productKey(id), &model.Product{})
	results["memory_cache"] = time.Since(start).String()

	start = time.Now()
	_, _ = s.repo.GetByID(ctx, id)
	results["database"] = time.Since(start).String()

	return results, nil
}

func (s *ProductService) GetSession(ctx context.Context, sessionID string) (*model.ProductResponse, error) {
	start := time.Now()

	var sessionData map[string]interface{}
	err := s.redisCache.Get(ctx, fmt.Sprintf("session:%s", sessionID), &sessionData)
	if err != nil {
		sessionData = map[string]interface{}{
			"session_id": sessionID,
			"user":       "guest",
			"created_at": time.Now().Format(time.RFC3339),
		}
		_ = s.redisCache.SetWithTTL(ctx, fmt.Sprintf("session:%s", sessionID), sessionData, 5*time.Minute)
		return &model.ProductResponse{
			Data:    sessionData,
			Source:  "New session (stored in Redis cache)",
			Latency: time.Since(start).String(),
		}, nil
	}

	return &model.ProductResponse{
		Data:    sessionData,
		Source:  "Redis cache (session)",
		Latency: time.Since(start).String(),
	}, nil
}

func (s *ProductService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"caching_strategies": map[string]string{
			"read_through": "GET products - reads from cache first, falls back to DB",
			"write_through": "POST /products - writes to DB + cache simultaneously",
			"write_back":   "POST /products/write-back - writes to cache first, async DB flush",
			"write_around": "PUT /products/:id - writes to DB only, invalidates cache",
		},
		"why_caching_faster": "Redis (RAM ~100ns) vs Database (disk ~10ms) - 100x faster",
		"redis_commands": map[string]string{
			"SET":     "redis-cli SET product:123 '{\"name\":\"Laptop\"}'",
			"GET":     "redis-cli GET product:123",
			"DEL":     "redis-cli DEL product:123",
			"EXPIRE":  "redis-cli EXPIRE product:123 60",
			"FLUSHALL": "redis-cli FLUSHALL",
		},
	}

	data, _ := json.MarshalIndent(stats, "", "  ")
	stats["raw"] = string(data)

	return stats, nil
}
