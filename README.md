# Go-Based Auction System API

This project is a Go-based API for an auction system, providing functionality for user management, product listings, and bidding processes.

The auction system API is designed to facilitate online auctions, allowing users to create accounts, list products for sale, and participate in bidding. It leverages a PostgreSQL database for data persistence and follows a clean architecture pattern, separating concerns into distinct layers such as API handlers, services, and data store.

Key features of the system include:
- User authentication and session management
- Product creation and management
- Auction creation and bidding functionality
- RESTful API endpoints for all core operations
- Database migrations for easy schema management
- Docker support for simplified deployment and development

## Repository Structure

The repository is organized as follows:

- `cmd/`: Contains the main entry points for the application
  - `api/`: Main API server entry point
  - `terndotenv/`: Utility for environment variable management
- `internal/`: Core application code
  - `api/`: API handlers and routing
  - `jsonutils/`: JSON utility functions
  - `services/`: Business logic services
  - `store/`: Data persistence layer
    - `pgstore/`: PostgreSQL specific implementation
      - `migrations/`: Database migration files
      - `queries/`: SQL query files
  - `usecase/`: Use case implementations
  - `validator/`: Input validation utilities
- `docker-compose.yml`: Docker Compose configuration for local development

Key files:
- `cmd/api/main.go`: Entry point for the API server
- `internal/api/routes.go`: API route definitions
- `internal/store/pgstore/db.go`: Database connection and management
- `internal/store/pgstore/sqlc.yml`: SQL code generation configuration

## Usage Instructions

### Installation

Prerequisites:
- Go 1.16 or later
- Docker and Docker Compose
- PostgreSQL 13 or later (if not using Docker)

Steps:
1. Clone the repository
2. Navigate to the project root
3. Copy the `.env.example` file to `.env` and adjust the values as needed
4. Run `docker-compose up -d` to start the PostgreSQL database
5. Run `go mod download` to install dependencies
6. Execute `go run cmd/api/main.go` to start the API server

### Configuration

The application uses environment variables for configuration. Key variables include:

- `GOBID_DATABASE_USER`: PostgreSQL username
- `GOBID_DATABASE_PASSWORD`: PostgreSQL password
- `GOBID_DATABASE_NAME`: PostgreSQL database name
- `GOBID_DATABASE_PORT`: PostgreSQL port (default: 5432)

### API Endpoints

The API provides the following main endpoints:

- `/api/users`: User management
- `/api/products`: Product listings
- `/api/auctions`: Auction management
- `/api/bids`: Bidding operations

For detailed API documentation, refer to the API specification document (not included in this repository).

### Database Migrations

To run database migrations:

1. Install the `migrate` tool: `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`
2. Run migrations: `migrate -database "${DATABASE_URL}" -path internal/store/pgstore/migrations up`

### Testing

To run tests:

```sh
go test ./...
```

### Troubleshooting

Common issues:

1. Database connection errors
   - Ensure PostgreSQL is running and accessible
   - Verify database credentials in the `.env` file
   - Check network connectivity if using a remote database

2. API server fails to start
   - Confirm all required environment variables are set
   - Ensure the database has been migrated to the latest version
   - Check for port conflicts and adjust the API server port if necessary

Debugging:
- Set the `DEBUG=true` environment variable for verbose logging
- Logs are output to stdout/stderr; use your system's logging mechanism to capture them
- Database query logs can be enabled in PostgreSQL for deeper investigation

Performance optimization:
- Monitor database query performance using PostgreSQL's query planner
- Use connection pooling to manage database connections efficiently
- Implement caching for frequently accessed data to reduce database load

## Data Flow

The request data flow through the application follows these steps:

1. Client sends HTTP request to API server
2. API router in `internal/api/routes.go` directs request to appropriate handler
3. Handler in `internal/api` processes request, performs authentication if required
4. Handler calls appropriate service in `internal/services`
5. Service implements business logic, interacting with the store layer as needed
6. Store layer in `internal/store/pgstore` executes database operations
7. Results propagate back through the layers
8. API handler formats and returns HTTP response to client

```
[Client] <-> [API Router] <-> [API Handler] <-> [Service] <-> [Store] <-> [Database]
```

Note: The use case layer (`internal/usecase`) may be involved for complex operations, sitting between the API handlers and services.

## Infrastructure

The project uses Docker Compose for local development infrastructure. Key resources include:

- PostgreSQL Database:
  - Type: Docker container
  - Image: postgres:latest
  - Port: Configurable, defaults to 5432
  - Environment variables:
    - POSTGRES_USER: Set from GOBID_DATABASE_USER
    - POSTGRES_PASSWORD: Set from GOBID_DATABASE_PASSWORD
    - POSTGRES_DB: Set from GOBID_DATABASE_NAME
  - Persistent volume: Local driver for data storage

The `docker-compose.yml` file defines this infrastructure, allowing for easy setup and teardown of the development environment.