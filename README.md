# Bike Parts Inventory

A single-node bicycle parts inventory management system built with Go + Vanilla JS.

## Features

- Parts CRUD with soft delete
- Stock increase / decrease with movement history
- Low stock email alerts via Gmail SMTP (triggered on decrease + periodic scheduler)
- Idempotency key support for stock mutations
- Vanilla JS frontend (sidebar layout, senior-friendly)
- E2E tests with in-memory SQLite

## Tech Stack

- **Backend**: Go, Gin, SQLite (`go-sqlite3`)
- **Frontend**: Vanilla JS, HTML, CSS (no build step)
- **Email**: Gmail SMTP via `gomail.v2`

## Setup

```bash
cp .env.example .env
# fill in your Gmail credentials

make run
# open http://localhost:8080
```

## Make Commands

| Command | Description |
|---------|-------------|
| `make run` | Start the server |
| `make seed` | Seed the DB with sample data (requires `sqlite3` CLI) |
| `make build` | Build binary to `bin/bikeparts` |
| `make e2e_test` | Run e2e tests |
| `make db_clear` | Delete the SQLite DB file |
| `make docker_build` | Build Docker image |
| `make docker_run` | Run in Docker (DB persisted to `./db`) |
| `make docker_clean` | Remove Docker image |

## Environment Variables

| Key | Description |
|-----|-------------|
| `PORT` | HTTP port (default: `8080`) |
| `DB_PATH` | SQLite file path (default: `./db/data.db`) |
| `SCHEMA_PATH` | Schema SQL path (default: `./db/schema.sql`) |
| `EMAIL_USER` | Gmail address |
| `EMAIL_PASS` | Gmail app password |
| `EMAIL_TO` | Alert recipient email |
| `SMTP_PORT` | SMTP port (default: `587`) |

## API

### Parts

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/parts` | List all parts |
| GET | `/api/parts/:id` | Get part by ID |
| POST | `/api/parts` | Create part |
| PUT | `/api/parts/:id` | Update part |
| DELETE | `/api/parts/:id` | Soft delete part |

### Stock

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/parts/:id/increase` | Increase stock |
| POST | `/api/parts/:id/decrease` | Decrease stock |

### Notifications

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/notifications` | List active low stock notifications |

> Stock mutation endpoints require an `Idempotency-Key` header (UUID) to prevent duplicate operations.
