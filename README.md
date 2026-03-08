# Go with Gin and Zerolog

A RESTful API server built with [Gin](https://github.com/gin-gonic/gin) and structured logging via [Zerolog](https://github.com/rs/zerolog). It connects to a Microsoft SQL Server database and includes Swagger documentation out of the box.

## Tech Stack

| Technology | Purpose |
|---|---|
| [Go 1.25](https://go.dev/) | Language runtime |
| [Gin](https://github.com/gin-gonic/gin) | HTTP web framework |
| [Zerolog](https://github.com/rs/zerolog) | Structured JSON logging |
| [go-mssqldb](https://github.com/microsoft/go-mssqldb) | Microsoft SQL Server driver |
| [Swaggo](https://github.com/swaggo/swag) | Auto-generated Swagger/OpenAPI docs |

## Project Structure

```
.
├── cmd/
│   └── api/
│       ├── main.go              # Application entrypoint
│       └── Dockerfile           # Multi-stage Docker build (scratch)
├── docs/                        # Auto-generated Swagger documentation
├── internal/
│   ├── errs/                    # Custom HTTP error types
│   ├── middleware/               # Gin middleware (error handler)
│   └── modules/
│       ├── application/         # HTTP server, routing, graceful shutdown
│       ├── config/              # Environment-based configuration
│       ├── database/            # SQL Server connection & pool management
│       ├── health/              # Liveness & readiness health checks
│       └── swagger/             # Swagger UI controller
├── .env.example                 # Environment variable reference
├── Makefile                     # Common development commands
├── go.mod
└── go.sum
```

## Prerequisites

- **Go** >= 1.25
- **Microsoft SQL Server** (running and accessible)
- **Docker** (optional, for containerized builds)

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/chatnarongt/go-with-gin-and-zerolog.git
cd go-with-gin-and-zerolog
```

### 2. Configure environment variables

Copy the example file and fill in your database credentials:

```bash
cp .env.example .env.local
```

Edit `.env.local` with your settings. See [Environment Variables](#environment-variables) for details.

### 3. Run the server

```bash
make run
```

The server will start on the port specified by `APP_PORT` (default: `8080`).

## Development Commands

| Command | Description |
|---|---|
| `make run` | Run the server with `.env` + `.env.local` |
| `make run-test` | Run the server with `.env` + `.env.test` |
| `make run-staging` | Run the server with `.env` + `.env.staging` |
| `make build` | Build the binary to `bin/api` |
| `make test` | Run all tests with verbose output |
| `make swag` | Regenerate Swagger documentation |

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `APP_ENVIRONMENT` | `development`, `test`, `staging`, `production` | `development` |
| `APP_PORT` | Server listen port (1–65535) | `8080` |
| `APP_LOG_LEVEL` | Zerolog level: -1=trace, 0=debug … 5=panic | _(empty)_ |
| `APP_ENABLE_SWAGGER` | Enable Swagger UI (`true` / `false`) | `true` |
| `DB_HOST` | Database host address | `localhost` |
| `DB_PORT` | Database port | `1433` |
| `DB_USER` | Database username | `sa` |
| `DB_PASSWORD` | Database password | _(empty)_ |
| `DB_NAME` | Database name | `master` |
| `DB_MAX_IDLE_CONNECTION_SIZE` | Idle connections in pool | `0` |
| `DB_MAX_CONNECTION_SIZE` | Max connections in pool | `1` |
| `DB_MAX_CONNECTION_IDLE_TIME` | Max idle time per connection (seconds) | `300` |
| `DB_MAX_CONNECTION_LIFE_TIME` | Max lifetime per connection (seconds) | `3600` |

## Docker

Build and run the container:

```bash
# Build the image
docker build -f cmd/api/Dockerfile -t go-with-gin-and-zerolog .

# Run the container
docker run -p 8080:8080 \
  -e APP_ENVIRONMENT=production \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=1433 \
  -e DB_USER=sa \
  -e DB_PASSWORD=yourpassword \
  -e DB_NAME=master \
  go-with-gin-and-zerolog
```

The production image uses a multi-stage build with a `scratch` base, resulting in a minimal and secure container that contains only the statically-compiled Go binary, CA certificates, and timezone data.

## API Documentation

When `APP_ENABLE_SWAGGER=true`, Swagger UI is available at:

```
http://localhost:8080/swagger/index.html
```

To regenerate the documentation after modifying API annotations:

```bash
make swag
```

## License

This project is open source. See the repository for license details.
