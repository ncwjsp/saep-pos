# CLAUDE.md

QR table-ordering system for Thai restaurants (POS + scan-to-order). Customers scan a printed QR at their table to order from their phone; the kitchen sees orders live; staff close the table and take payment at the counter. Building toward a real SaaS, starting as a vertical slice.

## Stack — locked decisions, do not propose alternatives

- Backend: Go + Gin, `github.com/ncwjsp/saep-pos`
- DB: PostgreSQL via pgx + sqlc (raw SQL, no ORM). Migrations: golang-migrate
- Current phase only: in-memory storage (slices + sync.Mutex), no DB yet
- Frontend: Next.js (App Router) + TypeScript + Tailwind + shadcn/ui
- State: React Query (server), Zustand (cart)
- Real-time: SSE (server → kitchen). Customer side polls every 5s. No WebSockets
- Auth (later phase): JWT via golang-jwt, bcrypt passwords
- Money: integers in satang. Never float, never decimal strings
- i18n: Thai default, English toggle, via next-intl (later phase)
- Deploy target: Railway (Go + Postgres), Vercel (frontend), Cloudflare R2 (images)

## Architecture rules

- Monorepo: `/backend` (Go), `/frontend` (Next.js)
- Backend layout: `cmd/api/main.go` for wiring; `internal/handlers` (thin Gin handlers: bind → call service → return JSON); `internal/middleware`; business logic in plain packages (`internal/orders`, `internal/menu`, `internal/tables`, `internal/sse`) that import nothing from Gin; `internal/db` for sqlc output; `migrations/`
- Multi-tenancy: every query is scoped by `restaurant_id`. No query without it. Tenancy comes from JWT middleware via context, never from request body
- QR design: each table has a random `qr_token` (22+ chars). Customer routes are keyed on the token only (`/t/:qrToken/...`) — customers never see internal IDs
- Sessions: a table has at most one open session, enforced by a partial unique index on `table_id WHERE status = 'open'`. First order auto-opens a session inside the same transaction
- Order items snapshot the menu price at order time into `order_items.price_satang`
- Order status flow: pending → preparing → served (or cancelled). No other transitions

## Division of labor — important

Claude Code writes: scaffolding, CRUD handlers, frontend components, migrations, sqlc queries, QR/PDF generation, tests, config.

**Review only — do not write or rewrite these; the author writes them by hand:**
- `internal/sse` (the hub: subscribe/unsubscribe/publish, channel lifecycle)
- The order-creation transaction (session auto-open, concurrency handling)
- Tenancy middleware

For those files: point out bugs, race conditions, and idiomatic issues; suggest what `go test -race` would catch; explain tradeoffs. Never produce replacement implementations unless explicitly asked "write this for me."

## Conventions

- Small commits, imperative messages ("add order submission endpoint")
- Errors: wrap with context (`fmt.Errorf("creating order: %w", err)`), return JSON `{"error": "..."}` with correct status codes
- JSON fields: snake_case. Timestamps: RFC 3339 UTC
- Table/menu names support Thai text everywhere — never assume ASCII
- Frontend: mobile-first for customer pages; big touch targets; ugly is fine before polish
- No premature abstraction: no repositories-of-repositories, no microservices, no Redis, no Docker beyond local compose

## Commands

- Backend run: `cd backend && go run ./cmd/api`
- Test: `go test ./...` and `go test -race ./...`
- Lint: `go vet ./...`
- Frontend: `cd frontend && npm run dev`

## Current phase — day-one vertical slice

Hardcoded single restaurant and table ("demo"), in-memory storage. Checkpoints:
1. GET /t/demo/menu — hardcoded menu (done when curl returns JSON)
2. POST /t/demo/orders — validate, snapshot prices, store in memory
3. SSE hub + GET /kitchen/stream (HAND-WRITTEN — review only)
4. Customer page /t/demo — menu, cart, submit
5. Kitchen page /kitchen — EventSource, live order cards
6. End-to-end demo: phone orders, laptop kitchen updates live

Later phases in order: real schema + migrations + sqlc → admin auth + menu/table CRUD → sessions + billing + floor view → QR PDF + i18n + deploy.
