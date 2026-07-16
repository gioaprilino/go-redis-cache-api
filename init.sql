CREATE TABLE IF NOT EXISTS products (
    id          VARCHAR(36) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price       DECIMAL(12,2) NOT NULL DEFAULT 0,
    stock       INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at DESC);

INSERT INTO products (id, name, description, price, stock, created_at, updated_at) VALUES
    ('a1b2c3d4-0001-4000-8000-000000000001', 'Laptop ASUS ROG', 'Gaming laptop with RTX 4060', 15000000, 10, NOW(), NOW()),
    ('a1b2c3d4-0002-4000-8000-000000000002', 'Mouse Logitech G102', 'RGB gaming mouse', 250000, 50, NOW(), NOW()),
    ('a1b2c3d4-0003-4000-8000-000000000003', 'Keyboard Mechanical', 'Cherry MX Blue switch', 750000, 30, NOW(), NOW()),
    ('a1b2c3d4-0004-4000-8000-000000000004', 'Monitor Samsung 27"', '4K IPS monitor', 4500000, 15, NOW(), NOW()),
    ('a1b2c3d4-0005-4000-8000-000000000005', 'Headset SteelSeries', '7.1 surround sound', 1200000, 20, NOW(), NOW());
