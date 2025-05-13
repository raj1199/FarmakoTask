# Coupon System API

A production-ready coupon system implementation in Go for a medicine ordering platform. The system provides functionality for coupon creation, validation, and management with support for different types of coupons and concurrent usage.

## Features

- Admin coupon management (CRUD operations)
- Multiple coupon types (one-time, multi-use, time-based)
- Product and category-specific coupons
- Concurrent validation handling
- Redis-based caching
- PostgreSQL for persistent storage
- Swagger/OpenAPI documentation
- Docker containerization

## Tech Stack

- Go 1.21+
- PostgreSQL (Database)
- Redis (Caching)
- Gin (Web Framework)
- GORM (ORM)
- Docker & Docker Compose

## Setup Instructions

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL (if running locally)
- Redis (if running locally)

### Running with Docker

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd coupon-system
   ```

2. Start the services:
   ```bash
   docker compose up --build
   ```

   This will start:
   - The API server on port 8080
   - PostgreSQL on port 5432
   - Redis on port 6379

### Running Locally

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Set up environment variables:
   ```bash
   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/coupon_system?sslmode=disable"
   export REDIS_URL="localhost:6379"
   ```

3. Run the application:
   ```bash
   go run cmd/server/main.go
   ```

## API Documentation

### Endpoints

#### Admin Endpoints
- `POST /admin/coupons` - Create a new coupon
  ```json
  {
    "code": "SAVE20",
    "expiry_date": "2024-12-31T23:59:59Z",
    "usage_type": "multi_use",
    "discount_type": "percentage",
    "discount_value": 20,
    "min_order_value": 100,
    "max_usage_per_user": 5
  }
  ```

#### Public Endpoints
- `GET /coupons/applicable` - Get applicable coupons for cart
  ```json
  {
    "cart_items": [
      {
        "id": "med_123",
        "category": "painkiller",
        "price": 100
      }
    ],
    "order_total": 700
  }
  ```

- `POST /coupons/validate` - Validate a coupon
  ```json
  {
    "coupon_code": "SAVE20",
    "cart_items": [...],
    "order_total": 700
  }
  ```

## Architectural Design

### Component Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   API Layer     │────▶│ Service Layer   │────▶│    Repository   │
│  (Gin Handlers) │     │ (Business Logic)│     │     Layer       │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                                                        │
                                                        ▼
                                                ┌─────────────────┐
                                                │   Database &    │
                                                │     Cache       │
                                                └─────────────────┘
```

### Directory Structure

```
.
├── cmd/
│   └── server/          # Application entrypoint
├── internal/
│   ├── api/            # API handlers
│   ├── models/         # Data models
│   ├── repository/     # Database operations
│   ├── service/        # Business logic
│   └── cache/          # Caching layer
└── docker/            # Docker related files
```

## Concurrency and Caching Notes

### Concurrency Handling

1. **Database-Level Concurrency**
   - Uses PostgreSQL transactions for atomic operations
   - Implements optimistic locking for coupon usage counts
   - Handles race conditions in coupon redemption

2. **Application-Level Concurrency**
   - Goroutines for concurrent request handling
   - Context-based request cancellation
   - Thread-safe coupon validation

### Caching Strategy

1. **Redis Caching**
   - Caches frequently accessed coupons
   - TTL-based cache invalidation
   - Distributed caching for scalability

2. **Cache Invalidation**
   - Automatic invalidation on coupon updates
   - TTL-based expiry for time-sensitive data
   - Cache warming for popular coupons

### Locking Mechanisms

1. **Distributed Locking**
   - Redis-based distributed locks for coupon usage
   - Prevents duplicate coupon redemption
   - Handles concurrent validation requests

2. **Transaction Isolation**
   - SERIALIZABLE isolation level for critical operations
   - Prevents phantom reads and write skew
   - Ensures data consistency

## Security Considerations

- Input validation using validator package
- Rate limiting on API endpoints
- JWT-based authentication for admin endpoints
- SQL injection prevention using GORM
- XSS protection through proper response encoding

## Monitoring and Metrics

- Structured logging using zerolog
- Prometheus metrics for monitoring
- Tracing support using OpenTelemetry
- Health check endpoints

## API Documentation

- Swagger UI: `http://localhost:8080/swagger/index.html`
- OpenAPI Spec: `http://localhost:8080/swagger/doc.json`

## Deployment

The application can be deployed to any cloud provider that supports Docker containers. Recommended platforms:
- AWS ECS
- Google Cloud Run
- Azure Container Apps

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 
