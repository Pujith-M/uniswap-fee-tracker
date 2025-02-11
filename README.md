# Uniswap Fee Tracker

A backend system that tracks transaction fees in USDT for Uniswap WETH-USDC transactions.

## Features

- Real-time transaction fee tracking
- Historical batch data processing
- RESTful API for querying transaction fees
- Swagger documentation
- Docker containerization

## Tech Stack

- Go (Backend)
- PostgreSQL (Database)
- Docker & Docker Compose (Containerization)
- Swagger (API Documentation)

## Project Structure

```
.
├── cmd/                  # Application entrypoints
├── internal/            # Private application code
│   ├── api/            # API handlers and routes
│   ├── models/         # Data models
│   ├── service/        # Business logic
│   └── repository/     # Data access layer
├── pkg/                # Public libraries
├── tests/              # Integration and e2e tests
└── docs/              # Documentation
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make (optional)

### Development Setup

1. Clone the repository
```bash
git clone <repository-url>
cd uniswap-fee-tracker
```

2. Run tests
```bash
go test ./...
```

3. Run with Docker Compose
```bash
docker-compose up
```

### API Documentation

API documentation will be available at `/swagger/index.html` once the server is running.

## Testing

The project follows Test-Driven Development (TDD) practices. To run tests:

```bash
go test -v ./...
```

## License

[MIT License](LICENSE)
