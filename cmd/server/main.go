package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-redis-cache-api/internal/cache"
	"go-redis-cache-api/internal/config"
	"go-redis-cache-api/internal/handler"
	"go-redis-cache-api/internal/repository"
	"go-redis-cache-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgPool.Close()

	if err := pgPool.Ping(ctx); err != nil {
		log.Printf("WARNING: PostgreSQL not reachable: %v", err)
		log.Println("The server will start, but DB features will be unavailable.")
	} else {
		log.Println("Connected to PostgreSQL")
	}

	redisCache := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPass, cfg.CacheTTL)
	if err := redisCache.Ping(ctx); err != nil {
		log.Printf("WARNING: Redis not reachable: %v", err)
	} else {
		log.Printf("Connected to Redis at %s (TTL: %ds)", cfg.RedisAddr, cfg.CacheTTL)
	}
	defer redisCache.Close()

	memCache := cache.NewMemoryCache(cfg.CacheTTL)

	productRepo := repository.NewProductRepository(pgPool)
	productSvc := service.NewProductService(productRepo, redisCache, memCache)
	productHandler := handler.NewProductHandler(productSvc)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Go Redis Cache API - Topik Khusus Modul 3 (Caching)","endpoints":["GET /api/products","GET /api/products/{id}","GET /api/products/{id}/slow","POST /api/products","POST /api/products/write-back","PUT /api/products/{id}","GET /api/products/benchmark/{id}","GET /api/session/{sessionId}","GET /api/cache/stats"]}`))
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/products", productHandler.GetAllProducts)
		r.Post("/products", productHandler.CreateProductWriteThrough)
		r.Post("/products/write-back", productHandler.CreateProductWriteBack)
		r.Get("/products/{id}", productHandler.GetProductByID)
		r.Get("/products/{id}/slow", productHandler.GetProductNoCache)
		r.Put("/products/{id}", productHandler.UpdateProductWriteAround)
		r.Get("/products/benchmark/{id}", productHandler.Benchmark)

		r.Get("/session/{sessionId}", productHandler.GetSession)
		r.Get("/cache/stats", productHandler.GetCacheStats)
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		log.Println("================================================")
		log.Println("  GO REDIS CACHE API")
		log.Println("  Topik Khusus - Modul 3 (Caching)")
		log.Println("================================================")
		log.Println()
		log.Println("  Endpoints:")
		log.Println("  [Read-Through]  GET    /api/products")
		log.Println("  [Read-Through]  GET    /api/products/{id}")
		log.Println("  [No Cache]      GET    /api/products/{id}/slow")
		log.Println("  [Write-Through] POST   /api/products")
		log.Println("  [Write-Back]    POST   /api/products/write-back")
		log.Println("  [Write-Around]  PUT    /api/products/{id}")
		log.Println("  [Benchmark]     GET    /api/products/benchmark/{id}")
		log.Println("  [Session]       GET    /api/session/{sessionId}")
		log.Println("  [Stats]         GET    /api/cache/stats")
		log.Println()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped gracefully")
}
