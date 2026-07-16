# Go Redis Cache API

Demonstrasi implementasi **4 caching strategies** menggunakan **Go + Redis** — berdasarkan materi Modul 3 (Caching) Topik Khusus.

## Fitur

| Strategy | Endpoint | Cara Kerja |
|---|---|---|
| **Read-Through** | `GET /api/products/{id}` | Cek Redis → cek In-Memory → fallback ke DB, lalu simpan ke cache |
| **Write-Through** | `POST /api/products` | Simpan ke DB + Redis secara simultan |
| **Write-Back** | `POST /api/products/write-back` | Simpan ke Redis dulu, return 202, goroutine async flush ke DB |
| **Write-Around** | `PUT /api/products/{id}` | Update DB langsung, invalidate cache |

## Tech Stack

- **Go 1.21** — chi router, pgx (PostgreSQL driver), go-redis
- **Redis 7** — In-memory cache
- **PostgreSQL 16** — Database
- **Docker Compose** — Orchestrasi container

## Cara Menjalankan

```bash
docker compose up --build
```

Server akan berjalan di `http://localhost:8080`.

## Endpoints

| Method | Path | Deskripsi |
|---|---|---|
| `GET` | `/` | Informasi API |
| `GET` | `/api/products` | Ambil semua produk (Read-Through) |
| `GET` | `/api/products/{id}` | Ambil produk by ID (Read-Through) |
| `GET` | `/api/products/{id}/slow` | Ambil produk tanpa cache (benchmark) |
| `POST` | `/api/products` | Buat produk (Write-Through) |
| `POST` | `/api/products/write-back` | Buat produk (Write-Back) |
| `PUT` | `/api/products/{id}` | Update produk (Write-Around) |
| `GET` | `/api/products/benchmark/{id}` | Benchmark latency Redis vs Memory vs DB |
| `GET` | `/api/session/{sessionId}` | Session caching demo |
| `GET` | `/api/cache/stats` | Informasi strategi caching |

## Perbandingan Latency

| Layer | Latency |
|---|---|
| Redis (in-memory) | ~100 ns |
| In-Memory Cache (Go) | ~50 ns |
| PostgreSQL (disk) | ~10 ms |

Redis **100x lebih cepat** dari database untuk data yang sering dibaca.

## Struktur Proyek

```
├── cmd/server/main.go        # Entry point
├── internal/
│   ├── cache/redis.go        # Redis cache client
│   ├── cache/memory.go       # In-memory cache
│   ├── config/config.go      # Environment config
│   ├── handler/product.go    # HTTP handlers
│   ├── model/product.go      # Data model
│   ├── repository/product.go # PostgreSQL access
│   └── service/product.go    # Business logic + caching
├── init.sql                  # Seed data
├── docker-compose.yml
└── Dockerfile
```
