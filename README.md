# Bike Parts Inventory

A simple inventory management system for bike parts, built with Go.

## Features

- Parts inventory with stock tracking
- Stock movement history (increase / decrease)
- Low stock email alerts via Gmail SMTP
- Soft delete support

## Tech Stack

- **Go** + Gin
- **SQLite**

## Setup

```bash
cp .env.example .env
# fill in your values

make run
```

## Environment Variables

| Key | Description |
|-----|-------------|
| `PORT` | HTTP port (default: 8080) |
| `DB_PATH` | SQLite file path |
| `EMAIL_USER` | Gmail address |
| `EMAIL_PASS` | Gmail app password |
| `EMAIL_TO` | Alert recipient |
| `SMTP_PORT` | SMTP port (default: 587) |

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/mail_test` | Test email alert |
