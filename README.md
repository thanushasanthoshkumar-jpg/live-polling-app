# Live Polling Web Application

A full-stack real-time polling app.

- **Frontend:** React (Vite) + Tailwind CSS
- **Backend:** Go (Gin)
- **Database:** MongoDB — user accounts, poll metadata, durable vote records
- **Realtime & Cache:** Redis — live vote counts (`HINCRBY` on a Redis Hash) + Pub/Sub broadcasting to WebSocket clients

## How real-time voting works

1. A client submits `POST /api/polls/:id/vote`.
2. The backend validates the poll is active and the option exists.
3. It atomically increments the count in a Redis hash (`HINCRBY poll_votes:<id> <option_id> 1`).
4. It publishes the updated counts as JSON to `poll_updates:<poll_id>` (Redis Pub/Sub).
5. Every server instance's `/ws/polls/:id` WebSocket handler subscribes to that channel and pushes the message to all connected browsers watching that poll — no refresh needed.
6. The vote is also persisted as a durable document in MongoDB (`votes` collection) for auditing/analytics, independent of the live Redis counter.

## Project structure

```
/backend    Go + Gin API, JWT auth, Mongo + Redis clients, WebSocket hub
/frontend   React (Vite) app — auth pages, poll creation, public voting page
```

## Prerequisites

- Go 1.22+
- Node.js 18+
- A running MongoDB instance (local or Atlas)
- A running Redis instance (local or hosted)

Quickest local option: use Docker for the datastores only —

```bash
docker run -d --name polling-mongo -p 27017:27017 mongo:7
docker run -d --name polling-redis -p 6379:6379 redis:7
```

## Backend setup

```bash
cd backend
cp .env.example .env
# edit .env if your Mongo/Redis URLs or JWT secret differ from the defaults
go mod tidy
go run main.go
```

The API starts on `http://localhost:8080` by default.

### `backend/.env`

```env
MONGO_URL=mongodb://localhost:27017
MONGO_DB_NAME=live_polling
REDIS_URL=redis://localhost:6379
JWT_SECRET=replace_with_a_long_random_secret
PORT=8080
```

## Frontend setup

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

The app starts on `http://localhost:5173` by default and talks to the backend at the URL in `.env`.

### `frontend/.env`

```env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

## API summary

| Method | Path                     | Auth      | Description                                  |
|--------|--------------------------|-----------|-----------------------------------------------|
| POST   | `/api/auth/register`     | Public    | Create an account, returns a JWT              |
| POST   | `/api/auth/login`        | Public    | Log in, returns a JWT                         |
| POST   | `/api/polls`             | JWT       | Create a poll                                 |
| GET    | `/api/polls`             | JWT       | List polls created by the current user        |
| GET    | `/api/polls/:id`         | Public    | Get poll metadata + live vote counts          |
| POST   | `/api/polls/:id/vote`    | Public    | Cast a vote (`{ "option_id": "..." }`)        |
| GET    | `/ws/polls/:id`          | Public    | WebSocket — live vote count broadcasts        |

## Notes on production hardening

This is a complete, runnable reference implementation. Before shipping it as-is, consider:

- Stronger duplicate-vote prevention (currently a simple per-IP fingerprint check is *not* enforced server-side against re-voting — add a unique index or Redis `SETNX` per voter+poll if that's a requirement).
- Rate limiting on `/api/polls/:id/vote` and `/api/auth/*`.
- Restricting WebSocket `CheckOrigin` to your actual frontend domain(s).
- Running multiple backend replicas behind a load balancer — the Redis Pub/Sub design already supports this correctly, since every replica subscribes independently.
- TLS termination (`wss://` / `https://`) in front of both servers.
