<p align="center">
	<img src="https://go.dev/images/go-logo-blue.svg" alt="Go official logo" width="340">
</p>

<h1 align="center">Subscription Service</h1>

<p align="center">
	A professional Go study project focused on practical concurrency with goroutines, channels, WaitGroups, PostgreSQL, Redis, and MailHog.
</p>

This repository is a study project for learning Go concurrency in a realistic web application. It combines a subscription-style app with background work handled through goroutines, channels, and sync.WaitGroup coordination.

The project uses:

- goroutines for asynchronous work such as email delivery and PDF generation
- channels for mail and error communication between concurrent workers
- sync.WaitGroup to track background jobs and support orderly shutdown
- PostgreSQL for application data
- Redis for session storage
- MailHog for local email capture and inspection

## What This Project Demonstrates

The codebase focuses on practical concurrency patterns instead of isolated examples. In particular, it shows how to:

- start background workers from HTTP request flows
- coordinate concurrent tasks with WaitGroups
- send work and errors through channels
- keep web requests responsive while longer-running work finishes asynchronously
- integrate concurrency with external services such as SMTP, Redis, and PostgreSQL

## Tech Stack And Versions

### Application

- Go 1.25.7 as declared in [go.mod](/home/jgsch/repositories/subscription_service/go.mod)
- Chi router v5.2.5
- pgx v5.8.0 for PostgreSQL access
- SCS v2.9.0 with Redis-backed sessions

### Infrastructure

- PostgreSQL 14.2 from [docker-compose.yml](/home/jgsch/repositories/subscription_service/docker-compose.yml)
- Redis using the `redis:alpine` image tag in [docker-compose.yml](/home/jgsch/repositories/subscription_service/docker-compose.yml)
- MailHog using the `mailhog/mailhog:latest` image tag in [docker-compose.yml](/home/jgsch/repositories/subscription_service/docker-compose.yml)

Note: PostgreSQL is pinned to a specific version. Redis and MailHog currently use floating image tags rather than pinned release versions.

## Requirements

To run this repository locally, install or have access to:

- Go 1.25.7 or a compatible local Go toolchain
- Docker
- Docker Compose v2 or compatible `docker compose` support
- GNU Make
- a free local TCP port set for:
	- `8080` for the Go web app
	- `5432` for PostgreSQL
	- `6379` for Redis
	- `1025` for MailHog SMTP
	- `8025` for the MailHog web UI

You also need:

- the environment file [.env](/home/jgsch/repositories/subscription_service/.env)
- the seed schema/data file [db.sql](/home/jgsch/repositories/subscription_service/db.sql)
- a writable `tmp/` directory at the repository root for generated PDF files

## Environment Configuration

This repository already includes an [.env](/home/jgsch/repositories/subscription_service/.env) file with PostgreSQL settings.

Current database configuration:

- `DB_HOST=localhost`
- `DB_PORT=5432`
- `DB_USER=postgres`
- `DB_PASSWORD=password`
- `DB_NAME=concurrency`
- `DB_SSLMODE=disable`

Optional Redis environment variables are supported by the codebase. If omitted, the application uses defaults such as `localhost:6379`.

## Local Services

Start the infrastructure stack first:

```bash
docker compose up -d
```

Services exposed locally:

- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- MailHog SMTP: `localhost:1025`
- MailHog UI: `http://localhost:8025`

## Database Setup

The application expects the `concurrency` database schema and seed data from [db.sql](/home/jgsch/repositories/subscription_service/db.sql).

After starting Docker services, load the schema:

```bash
docker compose exec -T postgres psql -U postgres -d concurrency < db.sql
```

This seeds:

- the `users` table
- the `plans` table
- the `user_plans` table
- initial plan records
- an initial admin user record

## Additional Setup

Create the temporary output directory used for generated PDF attachments:

```bash
mkdir -p tmp
```

## How To Run

### Option 1: Use Make

Build and run the application:

```bash
make run
```

The Make target:

- builds the binary from `./cmd/web`
- loads environment variables from `.env`
- starts the application locally

Useful Make targets:

- `make build` to compile the binary
- `make run` to build and start the app
- `make start` to run in the foreground
- `make start-bg` to run in the background
- `make stop` to stop the binary
- `make restart` to restart the app
- `make clean` to remove the built binary
- `make test` to run the test suite

### Option 2: Run Directly With Go

```bash
go run ./cmd/web
```

If you use this path, make sure your environment variables from `.env` are exported in your shell first.

## Application URLs

Once the application is running:

- web app: `http://localhost:8080`
- MailHog UI: `http://localhost:8025`
- test email route: `http://localhost:8080/test-email`

## Study Focus: Concurrency In This Repository

If you are using this repository to study Go concurrency, the most relevant parts are in the web application under `cmd/web`.

Key patterns implemented there include:

- background email processing
- channel-based error reporting
- WaitGroup tracking for async work
- concurrent invoice/manual generation during subscription workflows
- worker-style loops using `select`

## Typical Local Workflow

```bash
docker compose up -d
docker compose exec -T postgres psql -U postgres -d concurrency < db.sql
mkdir -p tmp
make run
```

Then open `http://localhost:8080` in your browser.

## Notes

- PostgreSQL data is persisted under `./db-data/postgres/`
- Redis data is persisted under `./db-data/redis/`
- outgoing mail is captured by MailHog rather than sent externally
- the app listens on port `8080`
- the repository includes a prebuilt `myapp` binary, but rebuilding locally is recommended

## Testing

Run the full test suite with:

```bash
make test
```

Or directly:

```bash
go test ./...
```
