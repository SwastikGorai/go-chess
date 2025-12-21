# Go Chess Frontend

A React + Vite UI for the Go chess backend. Play locally (client-side rules) or host/join an online game backed by the API.

## Features

- Local pass-and-play mode with client-side move validation
- Online mode that syncs with the Go API (create/join by Game ID)
- Theme controls for board and pieces
- Resign, draw offer, and game status overlays

## Requirements

- Node.js 18+ (recommended)

## Quick start

```bash
npm install
npm run dev
```

Vite will print the local dev URL (usually `http://localhost:5173`).

## Configuration

The app uses the backend API base URL from, in order:

1. `VITE_API_BASE_URL` (build-time env)
2. `window.CHESS_API_BASE_URL` (runtime, if injected)
3. `http://localhost:8080` (default)

Example `.env.local`:

```bash
VITE_API_BASE_URL=http://localhost:8080
```

## Scripts

- `npm run dev` - start the Vite dev server
- `npm run build` - build for production
- `npm run preview` - preview the production build
- `npm run lint` - run ESLint

## Notes and limitations

- Local mode uses simplified rules (no castling/en passant yet).
- Online mode uses server-side rules/validation from the Go backend.

