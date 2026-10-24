# Smart Mess

Today's mess menus, crowd levels and meal ratings for IIIT Pune students.
Implements the flow in `ARCHITECTURE.md`: OTP login, published menus with a
Regular/Jain split, live crowd levels from the sub managers, one review per
student per meal, and comment moderation for the main managers.

## Layout

```
smart-mess/
├── db/migrations/       # goose SQL migrations (the single schema source)
├── server/              # Go (chi + pgx + sqlc) REST API on :8080
├── client/              # React 19 + Vite + Tailwind 4 SPA (PWA)
├── docker-compose.yml   # local Postgres
└── docs/                # accessibility-report.md
```

## Prerequisites

- Docker Desktop (for Postgres)
- Go 1.24+
- Node 22+ with npm

## Run the backend

```bash
cd smart-mess
docker compose up -d db          # local Postgres on :5432

# migrations (goose v3); schema lives in server/db/migrations
go install github.com/pressly/goose/v3/cmd/goose@latest
cd server
DATABASE_URL=postgres://smartmess:smartmess@localhost:5432/smartmess?sslmode=disable \
  goose -dir db/migrations postgres "$DATABASE_URL" up
go run ./seed --config seed/managers.json    # users + default items
```

`seed/managers.json` is the demo seed. To seed your own, copy
`seed/managers.example.json`, fill in the real mess-manager emails and the
per-manager mess assignment, then run `go run ./seed --config <file>`.

Then start the API (replace the values as needed):

```bash
cd smart-mess/server
DATABASE_URL="postgres://smartmess:smartmess@localhost:5432/smartmess?sslmode=disable" \
JWT_SECRET="change-me" OTP_SECRET="change-me-too" \
ALLOWED_EMAIL_DOMAIN="iiitp.ac.in" PORT=8080 \
go run ./cmd/api
```

Without SMTP credentials the server logs OTP codes to the console
(development mode). Two more optional dev helpers:

- `DEV_CLOCK="2026-02-03T13:30:00+05:30"` — run against a fixed wall clock,
  useful to reproduce a lunch-window flow (`cmd/api/main.go` ignores it unless
  `APP_ENV` is development).
- Server tests: `DATABASE_URL_TEST="postgres://smartmess:smartmess@localhost:5432/smartmess_test?sslmode=disable" go test ./... -count=1`.

## Run the frontend

```bash
cd smart-mess/client
npm install
npm run dev        # http://localhost:5173, expects the API on :8080
```

Set `VITE_API_URL` to point the client at a different API base. `npm run build`
produces the PWA (manifest + workbox precache service worker) in `dist/`.

## Accessibility

CVD-safe palettes (`data-cvd` on `<html>`, 6 modes), 44px targets, text-scale,
reduced-motion and focus-visible support are built in; see
`docs/accessibility-report.md` for the automated results (axe on real
rendered components: zero serious/critical findings).
