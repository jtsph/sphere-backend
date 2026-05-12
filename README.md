# Sphere Backend

A secure Go backend for The Sphere Online services, built to support:

- Blockchain dashboard and validator data
- Sentience learning and documentation endpoints
- Investment options and community staking information
- Minecraft server status for the educative world
- Authentication and session-based access control

## Architecture

The backend is implemented as a single Go service with:

- `main.go` — server bootstrap and HTTP server configuration
- `config.go` — environment-based settings for port, database, and CORS
- `store.go` — SQLite-backed persistence and demo seed data
- `handlers.go` — REST endpoints, secure headers, rate limiting, and request logging
- `models.go` — typed domain models for blocks, validators, users, learn content, investment products, and servers

## Frontend

A companion frontend app is available under `frontend/`.
It is built with Vite, React, TypeScript, and proxies `/api` requests to the local backend.

## Running locally

1. Install Go 1.22 or later.
2. From the project root:

```bash
cd c:\Users\HP\sphere-backend
go mod tidy
go run .
```

3. Visit `http://localhost:8080/api/v1/health` to verify the service.

## Environment

Optional environment variables:

- `PORT` — server port (default: `8080`)
- `DATABASE_URL` — SQLite DSN (default: `file:sphere.db?_foreign_keys=on`)
- `JWT_SECRET` — session signing secret placeholder
- `CORS_ORIGIN` — allowed origin for API requests (`*` by default)

## API Endpoints

- `GET /api/v1/health`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/dashboard`
- `GET /api/v1/blocks`
- `GET /api/v1/validators`
- `GET /api/v1/sentience/learn`
- `GET /api/v1/sentience/docs`
- `GET /api/v1/invest`
- `GET /api/v1/minecraft`

## Next steps for The Sphere Online

1. Add admin and content management APIs for blocks, validators, learn resources, and investment campaigns.
2. Add TLS support or run the service behind a reverse proxy for production.
3. Connect the frontend domains to the API and map routes from `www.thesphere.online`, `sentience.thesphere.online`, and `cortex.thesphere.online`.
4. Add analytics, deployment scripts, and a CI pipeline for secure updates.
