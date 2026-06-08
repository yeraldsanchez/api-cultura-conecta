# api-cultura-conecta

REST API for the Cultura Conecta platform, built with Go, Gin, and PostgreSQL.

## Requirements

- Go 1.22+
- PostgreSQL 16
- Docker & Docker Compose (optional, but easier)

## Setup

Copy the example env file and fill in your values:

```bash
cp .env.example .env
```

The required variables are:

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string (`pgx5://user:pass@host:port/db`) |
| `JWT_SECRET` | Secret used to sign JWT tokens |
| `CORS_ALLOWED_ORIGINS` | Comma-separated list of allowed origins |
| `SENTRY_DSN` | Sentry DSN (optional, leave blank to skip) |
| `APP_ENV` | Environment name (`development`, `staging`, `production`) |

## Running with Docker

The quickest way to get everything up is with Docker Compose. It spins up PostgreSQL and the API together:

```bash
docker compose up --build
```

The API will be available at `http://localhost:8080`. Migrations run automatically on startup.

## Running locally

Start a PostgreSQL instance (or point `DATABASE_URL` to an existing one), then:

```bash
go run ./cmd/api
```

## Running tests

```bash
go test ./...
```

## API routes

All routes are prefixed with `/api/v1`.

| Method | Path              | Description            |
|--------|-------------------|------------------------|
| POST   | `/auth/register`  | Register a new user    |
| POST   | `/auth/login`     | Login and get a JWT    |
| POST   | `/users`          | Create a user profile  |
| GET    | `/interests`      | List interests         |
| POST   | `/interests`      | Create an interest     |
| GET    | `/focus-types`    | List focus types       |
| POST   | `/focus-types`    | Create a focus type    |
| GET    | `/cultural-works` | List cultural works    |
| POST   | `/cultural-works` | Create a cultural work |
| GET    | `/groups`         | List groups            |
| POST   | `/groups`         | Create a group         |

## Deployment

Pushing to `master` triggers a GitHub Actions workflow that builds a Docker image, pushes it to GHCR, and deploys to Render via deploy hook.
