package repository

import (
	"context"
	"fmt"
	"time"

	"go-redis-cache-api/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) Create(ctx context.Context, req model.CreateProductRequest) (*model.Product, error) {
	product := &model.Product{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO products (id, name, description, price, stock, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		product.ID, product.Name, product.Description, product.Price, product.Stock,
		product.CreatedAt, product.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert product: %w", err)
	}

	return product, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*model.Product, error) {
	product := &model.Product{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, price, stock, created_at, updated_at
		 FROM products WHERE id = $1`, id,
	).Scan(&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Stock, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("query product: %w", err)
	}
	return product, nil
}

func (r *ProductRepository) GetAll(ctx context.Context) ([]model.Product, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, description, price, stock, created_at, updated_at
		 FROM products ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price,
			&p.Stock, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, id string, req model.UpdateProductRequest) (*model.Product, error) {
	product := &model.Product{}
	err := r.pool.QueryRow(ctx,
		`UPDATE products
		 SET name = $1, description = $2, price = $3, stock = $4, updated_at = $5
		 WHERE id = $6
		 RETURNING id, name, description, price, stock, created_at, updated_at`,
		req.Name, req.Description, req.Price, req.Stock, time.Now(), id,
	).Scan(&product.ID, &product.Name, &product.Description, &product.Price,
		&product.Stock, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}
	return product, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}
