# Contributing to CodeStream

## Local Development

### Prerequisites

- Go 1.27+
- Docker + Docker Compose
- PostgreSQL 16 and Redis 7 (via Docker)

### Setup

1. Clone the repository.
2. Copy environment file:

   ```bash
   cp .env.example .env
   ```

3. Start infrastructure:

   ```bash
   docker compose up -d postgres redis
   ```

4. Run migrations:

   ```bash
   export DATABASE_URL=postgres://codestream:codestream@localhost:5432/codestream?sslmode=disable
   make migrate-up
   ```

5. Run the server:

   ```bash
   go run ./cmd/server
   ```

## Running Tests

```bash
make test
```

## Running Linter

```bash
make lint
```

## Formatting

```bash
make fmt
```

## CI Pipeline

```bash
make ci
```

## Adding a Migration

1. Create a new pair of files in `migrations/`:

   ```bash
   touch migrations/XXX_description.up.sql
   touch migrations/XXX_description.down.sql
   ```

2. Write SQL in the `.up.sql` file to apply the change.
3. Write SQL in the `.down.sql` file to revert it.
4. Run `make migrate-up` to apply.

## Project Structure

```
cmd/server/          # Entry point
internal/            # Application code
  authz/             # Authorization
  config/            # Environment configuration
  handler/           # HTTP and WebSocket handlers
  logger/            # Logging setup
  middleware/        # Gin middleware
  models/            # GORM models
  repository/        # Data access
  service/           # Business logic
  ws/                # WebSocket hub and connection
migrations/          # SQL migrations
scripts/             # Helper scripts
```
