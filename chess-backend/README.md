# Go Chess Backend API

A go based chess API built with Gin and Postgres. It creates games, validates moves, tracks history, and exposes status endpoints that the frontend can consume.

## Requirements

- Go 1.25+ (see `go.mod`)
- Postgres 15+ (or compatible)
- `goose` for migrations (optional but recommended)

Install goose if needed:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Configuration

The server reads configuration from environment variables (defaults shown):

```bash
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=chess
DB_PASSWORD=
DB_NAME=chess_db
DB_SSLMODE=disable
```

## Database setup

Run migrations (from `chess-backend/`):

```bash
goose -dir migrations postgres "postgres://chess:password@localhost:5432/chess_db?sslmode=disable" up
```

## Run the API

```bash
go run ./cmd/api
```
or using [air](https://github.com/air-verse/air)
```bash
air
```

The API listens on `http://localhost:8080` by default.

## API endpoints

Base path: `/api/v1`

- `POST /games` - create a new game (optional body: `{ "fen": "..." }`)
- `GET /games/:id` - get game state
- `GET /games/:id/legal-moves?from=e2` - list legal UCI moves (optionally filter by from-square)
- `POST /games/:id/moves` - make a move (`{ "uci": "e2e4" }`)
- `GET /games/:id/status` - get status flags/result
- `GET /games/:id/history` - list move history (UCI)
- `POST /games/:id/resign` - resign (`{ "color": "white" | "black" }`)
- `POST /games/:id/offer-draw` - offer a draw (`{ "color": "white" | "black" }`)
- `POST /games/:id/accept-draw` - accept a draw (`{ "color": "white" | "black" }`)

Health check:

- `GET /health`

## Example requests

Create a game:

```bash
curl -s -X POST http://localhost:8080/api/v1/games
```

Make a move:

```bash
curl -s -X POST http://localhost:8080/api/v1/games/<game-id>/moves \
  -H 'Content-Type: application/json' \
  -d '{"uci":"e2e4"}'
```

## Tests

```bash
go test ./internal/...
```

Run the full suite (may download modules):

```bash
go test ./...
```


## Notes

- The API uses Postgres storage in `cmd/api/main.go`. There is an in-memory store (`internal/store/memory.go`) that is not wired into the server.
- Move notation is UCI (e.g., `e2e4`, `g1f3`, `e7e8q`).

