# Motorcycle Store Microservices

This project is a microservices-based application for managing a motorcycle store. Each service is responsible for its own business logic, and communication between services is handled via HTTP APIs and Kafka message broker.

---

## Table of Contents
- [Project Structure](#project-structure)
- [Technology Stack](#technology-stack)
- [Microservices Overview](#microservices-overview)
- [API Endpoints](#api-endpoints)
- [Environment Variables](#environment-variables)
- [Quick Start](#quick-start)
- [Testing](#testing)
- [Monitoring](#monitoring)
- [CI/CD](#cicd)
- [FAQ](#faq)
- [Contributing](#contributing)
- [Roadmap](#roadmap)
- [Changelog](#changelog)
- [Known Issues](#known-issues)
- [License](#license)
- [Contacts & Credits](#contacts--credits)

---

## Project Structure

- **api-gateway/** — Entry point for clients, routes requests to other services.
- **inventory-service/** — Manages motorcycle inventory and stock.
- **order-service/** — Handles order processing and management.
- **user-service/** — Manages user accounts and authentication.
- **prometheus/** — Monitoring configuration for Prometheus.

## Technology Stack

- Go (Golang)
- MongoDB (database for services)
- Redis (caching, used by inventory-service)
- Kafka & Zookeeper (message broker for event-driven communication)
- Docker & Docker Compose (containerization and orchestration)
- Prometheus & Grafana (monitoring and visualization)

## Microservices Overview

### 1. API Gateway
- Exposes a unified HTTP API for clients.
- Forwards requests to the appropriate microservice.
- Handles authentication and basic validation.

### 2. Inventory Service
- Manages motorcycles, stock levels, and inventory operations.
- Stores data in MongoDB.
- Uses Redis for caching.
- Publishes/consumes events via Kafka.

### 3. Order Service
- Handles order creation, updates, and status tracking.
- Coordinates with Inventory Service to reserve or release stock.
- Stores data in MongoDB.
- Communicates with other services via Kafka.

### 4. User Service
- Manages user registration, authentication, and profiles.
- Stores user data in MongoDB.

---

## API Endpoints

### API Gateway

#### `POST /api/v1/auth/login`
Authenticate a user.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "string"
}
```
**Response:**
```json
{
  "token": "jwt-token",
  "user": { "id": "...", "email": "..." }
}
```

#### `POST /api/v1/auth/register`
Register a new user.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "string",
  "name": "John Doe"
}
```
**Response:**
```json
{
  "id": "user-id",
  "email": "user@example.com"
}
```

#### `GET /api/v1/inventory`
Get all motorcycles in stock.

**Response:**
```json
[
  { "id": "1", "model": "Yamaha MT-07", "stock": 5 },
  { "id": "2", "model": "Honda CB500F", "stock": 2 }
]
```

#### `POST /api/v1/orders`
Create a new order.

**Request:**
```json
{
  "user_id": "user-id",
  "items": [
    { "motorcycle_id": "1", "quantity": 1 }
  ]
}
```
**Response:**
```json
{
  "order_id": "order-id",
  "status": "pending"
}
```

#### `GET /api/v1/orders/{id}`
Get order details by ID.

**Response:**
```json
{
  "order_id": "order-id",
  "user_id": "user-id",
  "items": [
    { "motorcycle_id": "1", "quantity": 1 }
  ],
  "status": "pending"
}
```

---

### Inventory Service

#### `GET /inventory`
List all motorcycles.

#### `POST /inventory`
Add a new motorcycle to inventory.

#### `PUT /inventory/{id}`
Update motorcycle details.

#### `DELETE /inventory/{id}`
Remove a motorcycle from inventory.

---

### Order Service

#### `POST /orders`
Create a new order.

#### `GET /orders/{id}`
Get order by ID.

#### `GET /orders/user/{user_id}`
Get all orders for a user.

---

### User Service

#### `POST /users/register`
Register a new user.

#### `POST /users/login`
Authenticate user.

#### `GET /users/{id}`
Get user profile.

---

## Environment Variables

| Variable                  | Description                        | Example                        |
|---------------------------|------------------------------------|--------------------------------|
| MONGO_DB_URI              | MongoDB connection string           | mongodb://mongodb:27017        |
| MONGO_DB_REPLICA_SET      | Replica set name                   | rs0                            |
| MONGO_DB                  | Database name                      | inventory-service              |
| REDIS_HOSTS               | Redis host                         | redis:6379                     |
| BROKERS                   | Kafka broker address                | kafka:9092                     |
| ORDER_SERVICE_HOST        | Order service hostname              | order-service                  |
| USER_SERVICE_HOST         | User service hostname               | user-service                   |
| INVENTORY_SERVICE_HOST    | Inventory service hostname          | inventory-service              |
| REDIS_PASSWORD            | Redis password                      | yourpassword                   |

---

## Quick Start

### 1. Clone the Repository
```bash
git clone <repo-url>
cd Motorcycle-Store-Microservices
```

### 2. Start with Docker Compose
```bash
docker-compose up --build
```

### 3. Run Services Locally
```bash
task gateway
task inventory
task order
task user
```

---

## Testing

### Unit Tests
Each service contains unit tests in the `/tests` directory. Run tests with:
```bash
go test ./...
```

### Integration Tests
Integration tests require running dependencies (MongoDB, Redis, Kafka). Use Docker Compose for setup.

### Example Test (Go)
```go
func TestCreateOrder(t *testing.T) {
    // Arrange
    // ...
    // Act
    // ...
    // Assert
    // ...
}
```

---

## Monitoring
- **Prometheus** collects metrics from services (see `prometheus/prometheus.yml`).
- **Grafana** visualizes metrics (port 3001, default login: admin/admin).

---

## CI/CD

- Use GitHub Actions or GitLab CI for automated testing and deployment.
- Example workflow:
```yaml
name: CI
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.20
      - name: Build
        run: go build ./...
      - name: Test
        run: go test ./...
```

---

## FAQ

**Q: How do I add a new microservice?**
A: Create a new folder, add a Go module, update `go.work` and `docker-compose.yml`.

**Q: How do I connect to MongoDB?**
A: Use the connection string from the environment variables.

**Q: How do I reset the database?**
A: Stop Docker Compose, remove the `mongo-data` volume, and restart.

---

## Contributing

1. Fork the repository
2. Create a new branch (`git checkout -b feature/your-feature`)
3. Commit your changes
4. Push to your fork
5. Open a pull request

---

## Roadmap
- [x] Basic microservices
- [x] Docker Compose setup
- [x] Monitoring with Prometheus & Grafana
- [ ] Add OpenAPI/Swagger docs
- [ ] Add payment service
- [ ] Add notification service
- [ ] Kubernetes deployment

---

## Changelog
- v0.1.0 — Initial release
- v0.2.0 — Added monitoring
- v0.3.0 — Improved API Gateway

---

## Known Issues
- No rate limiting on API Gateway
- No email verification for users
- No payment integration

---

## License
MIT (or specify your own)

---

## Contacts & Credits
- Author: MAESTROonly, BeksultanSE and  beginore
- Contributors: See [[https://github.com/BeksultanSE/Motorcycle-Store-Microservices/commits/master/](https://github.com/BeksultanSE/Motorcycle-Store-Microservices/commits/master/)]
---

```
ASCII Architecture Diagram Example:

         +-------------+
         |  API-Gateway|
         +------+------+ 
                |         
   +------------+------------+
   |            |            |
+--v--+      +--v--+      +--v--+
|User |      |Order|      |Inventory|
|Svc  |      |Svc  |      |Svc      |
+-----+      +-----+      +---------+
                |
            +---v---+
            | Mongo  |
            +-------+
```

---

> For questions or suggestions, please open an issue or submit a pull request!