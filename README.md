# Go Book Backend

A step-by-step Go backend learning project that builds a library management API with Go, PostgreSQL, `chi`, `pgx`, `sqlc`, and Docker. Each lesson is designed to be self-contained so you can study the progression from a simple HTTP server to a production-style API with health checks, Swagger docs, and container orchestration.

## Project structure

The repository contains one lesson per folder:

- `lesson01` — minimal Go HTTP server with `chi`
- `lesson02` — PostgreSQL connection and first `sqlc` integration
- `lesson03` — CRUD endpoints for books
- `lesson04` — service/store refactor for cleaner separation
- `lesson05` — users and borrowing/return flows
- `lesson06` — continued API growth and transaction-oriented patterns
- `lesson07` — automated API tests
- `lesson08` — Docker setup for the app and database
- `lesson09` — Swagger/OpenAPI documentation and API docs
- `lesson10` — health checks, graceful shutdown, and Kubernetes deployment assets

Each lesson has its own `go.mod` and usually its own `db/` and `internal/` layout.

## Prerequisites

- Go 1.21+ (use the version defined in each lesson's `go.mod`)
- PostgreSQL
- Docker + Docker Compose (for lessons 08-10)
- `sqlc` for generating typed database code
- `goose` for applying migrations
- `swag` if you regenerate Swagger documentation

## Quick start

### 1. Start with lesson 01

```bash
cd lesson01
go mod tidy
go run ./cmd/api
```

Open:

```text
http://localhost:3000/
```

### 2. Run a database-backed lesson

Inside the lesson you want to run, create a `.env` file:

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/gobooksnew?sslmode=disable
```

Then apply migrations:

```bash
goose -dir db/migrations postgres "$env:DATABASE_URL" up
```

If you are using a different shell, adjust the environment-variable syntax accordingly.

Then start the API:

```bash
cd lesson10
go mod tidy
go run ./cmd/api
```

## Common API routes

The library API follows this progression:

### Books

- `GET /books`
- `GET /books/{id}`
- `POST /books`
- `PUT /books/{id}`
- `DELETE /books/{id}`

### Users

- `GET /users`
- `GET /users/{id}`
- `POST /users`
- `PUT /users/{id}`
- `DELETE /users/{id}`

### Borrowing workflow

- `POST /books/borrow`
- `POST /books/return`
- `GET /books/user/{id}/borrowed`
- `GET /books/overdue`

### Operational endpoints

- `GET /swagger/*` — Swagger UI
- `GET /health/live` — liveness probe
- `GET /health/ready` — readiness probe

## Example requests

Create a user:

```bash
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

Create a book:

```bash
curl -X POST http://localhost:3000/books \
  -H "Content-Type: application/json" \
  -d '{"title":"The Go Programming Language","author":"Alan Donovan","description":"A comprehensive guide to Go","year_of_publication":2015}'
```

Borrow a book:

```bash
curl -X POST http://localhost:3000/books/borrow \
  -H "Content-Type: application/json" \
  -d '{"book_id":1,"user_id":1,"days":7}'
```

Return a book:

```bash
curl -X POST http://localhost:3000/books/return \
  -H "Content-Type: application/json" \
  -d '{"book_id":1,"user_id":1}'
```

Check the Swagger UI:

```text
http://localhost:3000/swagger/index.html
```

## Docker workflow

Lessons 08-10 include Docker support.

### Run the stack with Docker Compose

```bash
cd lesson10
docker compose up --build
```

This starts:

- PostgreSQL database
- migration container
- API container on port `3000`

## Regenerating generated code

For any lesson with `sqlc.yaml`:

```bash
cd lesson05
sqlc generate
```

For Swagger docs in newer lessons:

```bash
cd lesson10
swag init -g cmd/api/main.go -o docs
```

## Migration notes

- Run migrations from the same lesson directory as the app you are starting.
- The schema changes between lessons, so mixing lesson folders is a common source of errors.
- Always ensure the database URL matches the lesson configuration.

## Troubleshooting

### `DATABASE_URL environment variable is not set`

Create a `.env` file in the lesson folder and make sure it includes a valid `DATABASE_URL`.

### `failed to connect to database`

Confirm PostgreSQL is running and the host, port, user, password, and database name match the configured URL.

### `relation does not exist`

Run migrations from the same lesson directory before starting the API.

### `404` for expected routes

Check which lesson you started. Older lessons do not include later routes like users or borrowing endpoints.

### Port `3000` already in use

Stop the process currently using that port or change the `Addr` in the lesson's `main.go`.

## Learning path

The best way to study this repository is to move through the lessons in order:

1. `lesson01` — minimal API foundation
2. `lesson02` — data access with PostgreSQL and `sqlc`
3. `lesson03` — basic CRUD
4. `lesson04` — cleaner architecture
5. `lesson05` — richer business workflows
6. `lesson06+` — resilience, testing, deployment readiness
7. `lesson08-10` — Docker, docs, and Kubernetes-ready runtime

## License

This project is intended for learning and experimentation. Use it as a reference for building Go backend APIs.
