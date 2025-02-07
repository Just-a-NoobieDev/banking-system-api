# Project banking-system

A modern banking system application that handles basic banking operations.

## Getting Started

These instructions will help you set up and run the project on your local machine for development and testing purposes.

### Prerequisites

Before you begin, ensure you have the following installed:

- Go (version 1.23 or higher)
- Docker and Docker Compose
- Make

### Installation

1. Clone the repository

```bash
git clone https://github.com/yourusername/banking-system.git
cd banking-system
```

2. Set up the environment variables (if any)

```bash
cp .env.example .env
# Edit .env file with your configurations
```

## Available Make Commands

The project includes several make commands to help you with development:

| Command               | Description                                    |
| --------------------- | ---------------------------------------------- |
| `make all`            | Run build with tests                           |
| `make build`          | Build the application                          |
| `make run`            | Run the application                            |
| `make docker-run`     | Create and start the database container        |
| `make docker-down`    | Shutdown the database container                |
| `make itest`          | Run integration tests                          |
| `make watch`          | Live reload the application during development |
| `make test`           | Run the test suite                             |
| `make clean`          | Clean up binary from the last build            |
| `make migrate-create` | Create a new database migration file           |
| `make migrate-up`     | Run all pending database migrations            |
| `make migrate-down`   | Rollback the last database migration           |

## Project Structure

```
banking-system/
├── cmd/                    # Application entry points
│   └── api/               # API server
│       └── main.go        # Main application entry point
├── docs/                  # Documentation files
│   ├── docs.go           # Generated API documentation
│   ├── swagger.json      # Swagger API specification in JSON
│   └── swagger.yaml      # Swagger API specification in YAML
├── internal/             # Private application and library code
│   ├── database/         # Database management
│   │   ├── migrations/   # Database migration files
│   │   ├── models/       # Data models
│   │   │   ├── account.go
│   │   │   ├── soa.go
│   │   │   ├── transaction.go
│   │   │   ├── types.go
│   │   │   └── user.go
│   │   ├── repositories/ # Data access layer
│   │   │   ├── account.go
│   │   │   ├── soa.go
│   │   │   ├── transaction.go
│   │   │   └── user.go
│   │   └── database.go
│   ├── pdf/             # PDF generation
│   │   └── statement.go
│   ├── server/          # HTTP server implementation
│   │   ├── account.go
│   │   ├── auth.go
│   │   ├── routes.go
│   │   ├── routes_test.go
│   │   ├── server.go
│   │   ├── soa.go
│   │   ├── transactions.go
│   │   └── user.go
│   └── utils/           # Internal utilities
│       └── http.go
├── statements/          # Generated statement PDFs
├── tmp/                # Temporary files
├── utils/              # Global utilities
│   ├── http.go
│   └── jwt.go
├── .air.toml           # Air live reload configuration
├── .dockerignore       # Docker ignore file
├── .env                # Environment variables
├── .env.example        # Example environment file
├── .gitignore         # Git ignore file
├── docker-compose.yml  # Docker compose configuration
├── Dockerfile         # Docker build file
├── go.mod             # Go module file
├── go.sum             # Go module checksum
├── main               # Binary executable
├── Makefile          # Build automation
└── README.md         # Project documentation
```

## API Documentation

The API provides the following endpoints:

### Authentication

| Method | Endpoint             | Description         |
| ------ | -------------------- | ------------------- |
| POST   | `/api/auth/register` | Register a new user |
| POST   | `/api/auth/login`    | Login a user        |

### Account Management

| Method | Endpoint                    | Description                                    |
| ------ | --------------------------- | ---------------------------------------------- |
| POST   | `/api/account/create`       | Create a new account                           |
| GET    | `/api/account/get`          | Get account details by ID                      |
| GET    | `/api/account/get-accounts` | Get all accounts with filtering and pagination |
| DELETE | `/api/account/delete`       | Delete an account                              |

### Transactions

| Method | Endpoint                    | Description           |
| ------ | --------------------------- | --------------------- |
| POST   | `/api/transaction/deposit`  | Make a deposit        |
| POST   | `/api/transaction/withdraw` | Make a withdrawal     |
| GET    | `/api/transaction/`         | Get all transactions  |
| GET    | `/api/transaction/get`      | Get transaction by ID |

### User Management

| Method | Endpoint                    | Description              |
| ------ | --------------------------- | ------------------------ |
| GET    | `/api/user/view-balance`    | View user's balance      |
| PUT    | `/api/user/update-profile`  | Update user profile      |
| PUT    | `/api/user/update-password` | Update user password     |
| GET    | `/api/user/me`              | Get current user details |

### Statement of Account (SOA)

| Method | Endpoint             | Description                      |
| ------ | -------------------- | -------------------------------- |
| POST   | `/api/soa/generate`  | Generate a statement of account  |
| GET    | `/api/soa/generated` | Get list of generated statements |
| GET    | `/api/soa/download`  | Download a specific statement    |

### System

| Method | Endpoint  | Description         |
| ------ | --------- | ------------------- |
| GET    | `/health` | Check system health |

All endpoints except `/api/auth/register` and `/api/auth/login` require authentication using a Bearer token in the Authorization header.

## Database Schema

### Users Table

- `id`: SERIAL PRIMARY KEY
- `first_name`: VARCHAR(255) NOT NULL
- `last_name`: VARCHAR(255) NOT NULL
- `email`: VARCHAR(255) NOT NULL UNIQUE
- `password`: VARCHAR(255) NOT NULL
- `created_at`: TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
- `updated_at`: TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP

### Accounts Table

- `id`: SERIAL PRIMARY KEY
- `user_id`: INT NOT NULL (Foreign key to users.id)
- `balance`: DECIMAL(10, 2) NOT NULL DEFAULT 0.00
- `currency`: currency_type NOT NULL DEFAULT 'USD'
- `account_name`: VARCHAR(255)
- `account_description`: TEXT
- `created_at`: TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
- `updated_at`: TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP

### Transactions Table

- `id`: SERIAL PRIMARY KEY
- `account_id`: INT NOT NULL (Foreign key to accounts.id)
- `user_id`: INT (Foreign key to users.id)
- `amount`: DECIMAL(10, 2) NOT NULL
- `transaction_type`: transaction_type NOT NULL
- `reference_id`: VARCHAR(255)
- `status`: transaction_status
- `created_at`: TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
- `updated_at`: TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP

### Statements Table

- `id`: SERIAL PRIMARY KEY
- `user_id`: INT (Foreign key to users.id)
- `pdf_url`: TEXT NOT NULL
- `statement_date`: DATE NOT NULL
- `created_at`: TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
- `updated_at`: TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP

### Enums

- `currency_type`: ['USD', 'EUR', 'GBP']
- `transaction_type`: ['DEPOSIT', 'WITHDRAWAL', 'TRANSFER']
- `transaction_status`: ['pending', 'completed', 'failed']

### Indexes

- `idx_transactions_account_id` on transactions(account_id)
- `idx_transactions_created_at` on transactions(created_at)
- `idx_accounts_user_id` on accounts(user_id)
- `idx_statements_account_id` on statements(account_id)

## Running Tests

The project includes both unit tests and integration tests:

```bash
# Run unit tests
make test

# Run integration tests (requires database)
make itest
```

## Built With

- Go
- PostgreSQL
- Docker
- Standard Library
- Swagger
- Go Migrate
- Go PDF

## Monitoring

- Grafana
- Prometheus
- Nginx
- API