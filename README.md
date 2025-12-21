# Go Chess

Monorepo for a Go chess backend API and a React + Vite frontend. The backend owns rules, game state, and persistence; the frontend provides a rich UI that can play locally or sync with the API.

## Repository layout

- `chess-backend/` - Go API server (Gin + Postgres)
- `chess-frontend/` - React + Vite UI

## Quick start

### Backend

```bash
cd chess-backend
# Configure DB env vars (see chess-backend/README.md)
# Run migrations (requires goose)
goose -dir migrations postgres "postgres://chess:password@localhost:5432/chess_db?sslmode=disable" up
# Start the API
go run ./cmd/api # or use air (https://github.com/air-verse/air)
```

### Frontend

```bash
cd chess-frontend
npm install
npm run dev
```

By default the frontend talks to `http://localhost:8080`.

## Configuration

- Backend env vars live in `chess-backend/README.md`.
- Frontend API base URL can be set via `VITE_API_BASE_URL`.

Example:

```bash
# chess-frontend/.env.local
VITE_API_BASE_URL=http://localhost:8080
```

## Docs

- `chess-backend/README.md` - API setup, endpoints, and data store details
- `chess-frontend/README.md` - UI setup and usage

