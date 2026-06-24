# URL Shortener API

A production-style URL shortener API built with Go and PostgreSQL. This project was built deliberately as a learning exercise to practice production-grade backend patterns — proper project structure, raw SQL with `pgx`, database migrations, containerization with a non-root user, concurrency-safe rate limiting, and structured logging — rather than to ship a novel product.

## Why This Project

I'm pivoting from frontend-leaning full-stack work toward backend engineering, with Go as the primary focus. This is the first project in that learning path. The goal wasn't just "make a URL shortener" — it was to practice the engineering habits that separate a toy script from a service that could survive in production: clean package boundaries, explicit error handling, observability, and safe concurrency.

## Tech Stack

- **Language:** Go
- **Database:** PostgreSQL
- **DB Driver:** `pgx/v5` (raw SQL, no ORM — see Architecture Decisions)
- **Migrations:** `golang-migrate`
- **Validation:** `go-playground/validator`
- **Containerization:** Docker, Docker Compose
- **Logging:** `log/slog` (structured JSON logging)

## API Endpoints

| Method | Path | Description | Success | Failure |
|---|---|---|---|---|
| `GET` | `/health` | Health check | `200` | — |
| `POST` | `/urls` | Create a short URL | `201` | `400`, `500` |
| `GET` | `/urls/{shortCode}` | Redirect to original URL | `302` | `404`, `500` |

### `POST /urls`

**Request:**
```json
{ "url": "https://example.com/some/long/path" }
```

**Response (201):**
```json
{
  "short_code": "aB3xKp9q1L",
  "original_url": "https://example.com/some/long/path",
  "short_url": "http://localhost:8080/urls/aB3xKp9q1L"
}
```

## Architecture Decisions

**`/cmd` + `/internal` layout** — Follows the [Go standard project layout](https://github.com/golang-standards/project-layout). `cmd/api/main.go` is a thin entry point that wires dependencies together; all real logic lives in `internal/`, which Go's compiler enforces as private to this module. Deliberately skipped `/pkg`, `/api`, `/web`, `/third_party` — they're either controversial conventions or solve problems this project doesn't have yet.

**Raw SQL with `pgx/v5`, no ORM** — Chose to hand-write SQL instead of using an ORM (or even `sqlc`) specifically to learn how database interactions actually work under the hood — connection pooling, named args, row scanning, error codes. `sqlc` is a natural next step once these fundamentals are solid.

**PostgreSQL over MongoDB** — An earlier version of this project (a CLI tool, kept as project history) used MongoDB. This rebuild intentionally switches to PostgreSQL to practice relational schema design, constraints, and migrations.

**Short codes generated in-app, not derived from DB IDs** — Short codes are randomly generated in Go rather than encoding the database's auto-increment ID. This avoids leaking sequential/guessable IDs and decouples the public-facing code from internal row identity.

**Collision handling via retry, not pre-check** — Since short codes are randomly generated, collisions are possible (if rare). Rather than checking for existence before every insert (an extra round-trip on the common path), the insert is attempted directly and retried up to 5 times only on a Postgres unique-violation (`23505`). This is the optimistic/retry-on-conflict pattern — cheaper than a check-then-insert on the happy path. A `UNIQUE` constraint on `short_code` is the real source of truth; the retry logic is just a graceful way to handle the rare conflict it catches.

**Collisions are a 500, not a 409** — A 409 Conflict implies the *client* claimed something already taken. Here, short codes are server-generated, so an unresolved collision after 5 retries is a server-side failure, not a client error.

**Fail-fast startup** — The app pings the database once on boot inside `InitDB`. If the connection fails, it logs and exits immediately rather than starting in a broken state. `/health` is a static liveness check — if the app is running at all, by definition `InitDB` already succeeded.

**Concurrency-safe in-memory rate limiting (token bucket)** — Implemented per-client (per-IP) token bucket rate limiting using a custom generic `SafeMap[T]`. The core safety property: refill calculation, the allow/reject decision, and the write-back all happen inside a single mutex-protected `Update` call, eliminating the read-then-write race window that a naive `Get`/compute/`Set` sequence would have. This was load-tested with concurrent requests (via `hey`) to confirm token counts stay correct under real concurrency, not just in theory.

**Rate limit headers follow the IETF draft spec** — Returns `RateLimit-Limit`, `RateLimit-Remaining`, and `RateLimit-Reset` (seconds remaining until full reset, per the non-prefixed `RateLimit-*` draft spec) rather than the older `X-RateLimit-*` / absolute-timestamp convention some APIs (e.g. GitHub) use. `Retry-After` is included on `429` responses specifically, representing seconds until at least one token is available again (not seconds until the bucket is full).

**Non-root Docker user** — The production image creates a dedicated unprivileged system user (`adduser -S`) and runs the final binary as that user, not root. Multi-stage build ensures the final image contains only the compiled binary — no source code, no build tools.

**No port exposure on PostgreSQL** — The `db` service does not publish its port to the host, even in local development. The database is only reachable from the `app` service over Docker's internal network. To inspect data manually, connect via `docker exec` into the running container rather than exposing the port — a habit carried over deliberately from how this would need to run in production.

**Health-checked service startup ordering** — `docker-compose.yml` uses `depends_on` with `condition: service_healthy` (backed by a `pg_isready` healthcheck) rather than a bare `depends_on`, which only waits for the container to start, not for Postgres to actually be ready to accept connections.

## Known Limitations

- **No eviction on the in-memory rate limiter.** Every unique client IP that hits the API stays in memory for the life of the process. Fine for a learning project and short-lived demos; would need either a periodic cleanup sweep or a move to Redis (with TTLs) for long-running production use.
- **IP-based rate limiting only.** There's no authentication layer, so per-API-key or per-user limiting isn't possible yet. IP is a reasonable default for an unauthenticated public API but is coarser than per-user limiting (shared IPs, NAT, VPNs all share a bucket).
- **Single-instance rate limiting.** The in-memory map only works correctly with one running instance of the app. Multi-instance deployment would require a shared store (Redis) for the rate limiter to remain correct.

## Project Structure

```
url-shortener-api/
├── cmd/
│   └── api/
│       └── main.go          # entry point — wires deps, starts server
├── internal/
│   ├── database/             # DB connection pool, queries
│   ├── handlers/             # HTTP handlers
│   ├── middleware/           # rate limiting, etc.
│   ├── models/                # request/response/db row structs
│   └── utils/                 # short code generation, SafeMap, IP extraction
├── migrations/                # versioned schema changes (golang-migrate)
├── Dockerfile                  # multi-stage build, non-root user
├── docker-compose.yml
├── Makefile                    # migrate-up/down/create, clean, clean-hard
└── .env                         # local config (not committed)
```

## Setup

### Prerequisites
- Docker & Docker Compose
- `make` (if you wanna use `make` commands)

### Environment Variables

Copy the `.env.example` file to `.env`:

```bash
cp .env.example .env
```

### Run

```bash
# start the app + database
make up

# run database migrations
make migrate-up
```

### Useful Makefile commands

| Command | Description |
|---|---|
| `make migrate-create name=<name>` | Create a new migration file pair |
| `make migrate-up` | Apply pending migrations |
| `make migrate-down` | Roll back the last migration |
| `make up` | Starts the containers |
| `make down` | Stops the containers |
| `make clean` | Stop containers, remove volumes/networks (keeps images) |
| `make clean-hard` | Same as `clean`, also removes all images |

### Inspecting the database

The database port is intentionally not exposed to the host. To inspect data:

```bash
make db
```

## What's Next

- Structured request logging middleware (method, path, status, duration, client IP)
- Integration tests using `httptest`
- Deploy to a live environment (AWS/Render) for a working demo link
- Optional: migrate rate limiter state to Redis with TTL-based eviction