# URL Shortener API

A production-style URL shortener API built with Go and PostgreSQL, deployed and live on Render. This project was built deliberately as a learning exercise to practice production-grade backend patterns — proper project structure, raw SQL with `pgx`, database migrations, containerization with a non-root user, concurrency-safe rate limiting, structured logging, a real test suite, and a live deployment — rather than to ship a novel product.

**Live demo:** https://url-shortener-api-xiyp.onrender.com

## Why This Project

I'm pivoting from frontend-leaning full-stack work toward backend engineering, with Go as the primary focus. This is the first project in that learning path. The goal wasn't just "make a URL shortener" — it was to practice the engineering habits that separate a toy script from a service that could survive in production: clean package boundaries, explicit error handling, observability, safe concurrency, test coverage that actually proves correctness rather than just exercising code, and an actual deployment with a real database, not just `localhost`.

## Tech Stack

- **Language:** Go
- **Database:** PostgreSQL
- **DB Driver:** `pgx/v5` (raw SQL, no ORM — see Architecture Decisions)
- **Migrations:** `golang-migrate`
- **Validation:** `go-playground/validator`
- **Containerization:** Docker, Docker Compose (local), Dockerfile-based deploy (Render)
- **Logging:** `log/slog` (structured JSON logging)
- **Hosting:** Render (Web Service + managed PostgreSQL)

## API Endpoints

| Method | Path | Description | Success | Failure |
|---|---|---|---|---|
| `GET` | `/health` | Health check | `200` | — |
| `POST` | `/urls` | Create a short URL | `201` | `400`, `500` |
| `GET` | `/urls/{shortCode}` | Redirect to original URL | `302` | `400`, `404`, `500` |

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
  "short_url": "https://url-shortener-api-xiyp.onrender.com/urls/aB3xKp9q1L"
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

**Concurrency-safe in-memory rate limiting (token bucket)** — Implemented per-client (per-IP) token bucket rate limiting using a custom generic `SafeMap[T]`. The core safety property: refill calculation, the allow/reject decision, and the write-back all happen inside a single mutex-protected `Update` call, eliminating the read-then-write race window that a naive `Get`/compute/`Set` sequence would have. This was load-tested with concurrent requests (via `hey`) to confirm token counts stay correct under real concurrency, not just in theory — see Testing below for how this is also covered by an automated, race-detector-verified test.

**Rate limit headers follow the IETF draft spec** — Returns `RateLimit-Limit`, `RateLimit-Remaining`, and `RateLimit-Reset` (seconds remaining until full reset, per the non-prefixed `RateLimit-*` draft spec) rather than the older `X-RateLimit-*` / absolute-timestamp convention some APIs (e.g. GitHub) use. `Retry-After` is included on `429` responses specifically, representing seconds until at least one token is available again (not seconds until the bucket is full).

**Structured request logging middleware** — Every request is logged via `slog` (method, path, status code, duration, client IP) using a response writer wrapper that intercepts `WriteHeader` to capture the actual status code, since `http.ResponseWriter` doesn't expose what was written after the fact. Sits outside the rate limiter in the middleware chain, so rejected (`429`) requests are logged too, not just successful ones.

**Non-root Docker user** — The production image creates a dedicated unprivileged system user (`adduser -S`) and runs the final binary as that user, not root. Multi-stage build ensures the final image contains only the compiled binary — no source code, no build tools.

**No port exposure on PostgreSQL (local dev)** — In local Docker Compose, the `db` service does not publish its port to the host. The database is only reachable from the `app` service over Docker's internal network. To inspect data manually, connect via `docker exec` (wrapped in `make db`) rather than exposing the port — a habit carried over deliberately from how this would need to run in production.

**Health-checked service startup ordering (local dev)** — `docker-compose.yml` uses `depends_on` with `condition: service_healthy` (backed by a `pg_isready` healthcheck) rather than a bare `depends_on`, which only waits for the container to start, not for Postgres to actually be ready to accept connections.

**Deployed on Render via its managed PostgreSQL, not the Compose stack directly** — Render doesn't run `docker-compose.yml` as-is; each service is deployed independently. The Go app is deployed as a Render Web Service built from the existing `Dockerfile`, and the database is a separate Render-managed PostgreSQL instance rather than a self-managed container. `InitDB` connects using whatever connection details Render provides for its managed Postgres, same `pgx/v5` pool underneath — local Docker Compose and Render are just two different ways of supplying the same connection info to the same code.

**Migrations run manually against the deployed database** — Render doesn't run `golang-migrate` for you. Schema changes are applied by pointing the same `migrate/migrate` Docker command used locally at Render's external database connection string instead of the local one, run from a developer machine. There's no migration step wired into the deploy pipeline yet — see Known Limitations.

## Testing

The test suite is split deliberately into unit and integration tests, each targeting a different kind of correctness.

**Unit tests** (`go test ./internal/utils/...`)
- `SafeMap[T]`'s `Update` method — spins up 1000 concurrent goroutines incrementing the same key and asserts the final value is exactly 1000. Run with `-race` to confirm there's no read-modify-write gap, not just that the test happens to pass once.
- `GenerateShortCode` — correct default/custom length, output only uses the defined charset, distinct values across calls.
- `GetClientIP` — covers `X-Forwarded-For`, `X-Real-IP`, no-header fallback to `RemoteAddr`, and malformed header fallthrough, each with an isolated request per subtest.

**Integration tests** (`go test ./internal/handlers/...`)
- Run against a real, separate PostgreSQL instance (`postgres-test` service), not mocked — schema is loaded directly from the actual migration file via `//go:embed`, so the test schema can't drift from the real one.
- Each test truncates the `urls` table first for isolation.
- `POST /urls` — asserts status code, response shape, field correctness, and a database-level side effect check (queries the row back out and compares against the response).
- `GET /urls/{shortCode}` — covers invalid short code (`400`), valid-but-nonexistent short code (`404`), and a successful redirect (`302`) with `Location` header verification.

Run everything:
```bash
go test -race ./...
```

## Known Limitations

- **No eviction on the in-memory rate limiter.** Every unique client IP that hits the API stays in memory for the life of the process. Fine for a learning project and a low-traffic demo; would need either a periodic cleanup sweep or a move to Redis (with TTLs) for long-running production use.
- **IP-based rate limiting only.** There's no authentication layer, so per-API-key or per-user limiting isn't possible yet. IP is a reasonable default for an unauthenticated public API but is coarser than per-user limiting (shared IPs, NAT, VPNs all share a bucket).
- **Single-instance rate limiting.** The in-memory map only works correctly with one running instance of the app. Multi-instance deployment would require a shared store (Redis) for the rate limiter to remain correct.
- **Migrations are applied manually.** There's no CI/CD step that runs `migrate up` automatically against the production database on deploy — it's a manual command run locally against Render's external connection string. Fine at this scale, but a real gap for an actual production workflow.
- **Render free tier behavior.** The deployed instance may spin down on inactivity and take a few seconds to respond to the first request after idling — expected behavior of the hosting tier, not the application.

## Project Structure

```
url-shortener-api/
├── cmd/
│   └── api/
│       └── main.go          # entry point — wires deps, starts server
├── internal/
│   ├── database/             # DB connection pool, queries
│   ├── handlers/             # HTTP handlers + integration tests
│   ├── middleware/           # rate limiting, request logging
│   ├── models/                # request/response/db row structs
│   ├── testutils/             # test DB setup, schema loading, truncation helper
│   └── utils/                 # short code generation, SafeMap, IP extraction (+ unit tests)
├── migrations/                # versioned schema changes (golang-migrate, embedded for tests)
├── Dockerfile                  # multi-stage build, non-root user (also used by Render)
├── docker-compose.yml          # local dev only
├── Makefile                    # migrate-up/down/create, up, down, clean, clean-hard, db
└── .env.example                 # local config template
```

## Setup (Local Development)

### Prerequisites
- Docker & Docker Compose
- `make` (if you wanna use `make` commands)
- Go (for running tests locally, outside Docker)

### Environment Variables

Copy the `.env.example` file to `.env` (and `.env.test.example` to `.env.test` for running tests):

```bash
cp .env.example .env
cp .env.test.example .env.test
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
| `make migrate-up` | Apply pending migrations (local) |
| `make migrate-up-render` | Apply pending migrations in render db |
| `make migrate-down` | Roll back the last migration (local) |
| `make up` | Starts the containers |
| `make down` | Stops the containers |
| `make clean` | Stop containers, remove volumes/networks (keeps images) |
| `make clean-hard` | Same as `clean`, also removes all images |
| `make db` | Open a `psql` shell into the running database |

### Inspecting the database

Local: the database port is intentionally not exposed to the host. To inspect data:

```bash
make db
```

Render: connect with `psql` (or via Docker, no local install needed) using Render's **External Database URL** from the dashboard:

```bash
psql "<RENDER_EXTERNAL_DATABASE_URL>"
```

## Deployment

Deployed on Render as a Web Service (built from the repo's `Dockerfile`) paired with a separate Render-managed PostgreSQL instance.

- `BASE_URL` is set to the live Render domain (`https://...onrender.com`) so generated short URLs point at the public, internet-reachable address rather than `localhost`.
- Database connection details point at Render's managed PostgreSQL rather than a local container — same `InitDB` code path, different connection info.
- Schema migrations are applied by running the same `golang-migrate` Docker command used locally, pointed at Render's external database connection string instead.

## What's Next

- Automate migrations as part of the deploy step instead of running them manually
- Optional: migrate rate limiter state to Redis with TTL-based eviction
- Optional: periodic cleanup sweep for stale rate limit entries if Redis migration is deferred
- Optional: custom domain instead of the default Render subdomain