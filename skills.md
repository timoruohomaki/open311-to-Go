# Skills & Conventions — Open311-to-Go

A living engineering playbook for working in this repo: how to build/run/test,
the conventions to follow, the known gotchas, and step-by-step recipes. Pair this
with [developer-reference.md](developer-reference.md) (the API contract) and
[README.md](README.md) (project intent).

---

## Quick map

| Path | What |
|---|---|
| [`src/main.go`](src/main.go) | Entry point: config → loggers → Mongo → Sentry → HTTP server + graceful shutdown |
| [`src/config/`](src/config/) | `config.go` — env-var loader + `.env` support ([.env.example](src/.env.example)) |
| [`src/domain/models/`](src/domain/models/) | `User`, `Service`, `ServiceRequest` + XML wrappers |
| [`src/internal/api/`](src/internal/api/) | Route registration |
| [`src/internal/handlers/`](src/internal/handlers/) | HTTP handlers (`*_handler.go`) |
| [`src/internal/repository/`](src/internal/repository/) | Mongo connection + per-entity repos |
| [`src/pkg/`](src/pkg/) | Reusable: `router`, `middleware`, `logger`, `httputil`, `app` (unused) |
| [`scripts/`](scripts/) | Operational tooling — [`feed-boston.ps1`](scripts/feed-boston.ps1) (Boston 311 CSV → API importer) |

Module: `github.com/timoruohomaki/open311-to-Go` · Go **1.26.x** · net/http
routing (1.22+).

---

## Build / run / test

All commands run from `src/`.

```sh
cd src

# Configure: copy the example env file and fill it in (at minimum MONGODB_URI)
cp .env.example .env   # .env is gitignored; real env vars override it

# Run the dev server (loads ./.env if present)
go run main.go
# or: make run
# Point at a different env file: go run main.go -env=/path/to/.env

# Test
go test ./...                      # everything
go test ./pkg/middleware -v        # one package
go test ./internal/handlers -v

# Vet / format (do this before committing)
go vet ./...
gofmt -l .
```

`make build` outputs the binary to `bin/open311api` (gitignored). Configuration
is **environment variables only** — there is no config file. See
[.env.example](src/.env.example) for the full list; `config.Load()` applies
defaults and requires only `MONGODB_URI`.

---

## Architecture conventions

- **Layering:** `handlers` (HTTP) → `repository` (interface) → Mongo. Handlers
  never touch the driver; repositories never touch `http`.
- **Repositories are interfaces** (`UserRepository`, `ServiceRepository`,
  `ServiceRequestRepository`) with `Mongo*` implementations. Handlers depend on
  the interface — this is what makes the testify mocks in `*_test.go` work.
- **Naming:** handlers `*_handler.go` / `{Entity}Handler`; repos `*_repository.go`
  / `Mongo{Entity}Repository`; middleware `{Action}Middleware`.
- **Errors:** repositories return sentinel errors (`ErrNotFound`, `ErrInvalidID`,
  `ErrDatabase`); handlers map them to status codes with `errors.Is`. Follow this
  pattern for new endpoints (see [`user_handler.go`](src/internal/handlers/user_handler.go)).
- **Responses:** use `BaseHandler.SendResponse` / `SendError`; they honor the
  `Accept` header (JSON/XML) via `httputil`. For collections, wrap in the XML
  wrapper type (`models.Users{Items: ...}`) when `Accept` is XML.
- **Context:** pass the request context (`r.Context()`) down. See the timeout
  gotcha below.
- **Auth:** writes (`POST`/`PUT`/`PATCH`/`DELETE`) require a valid `X-API-Key`
  via `middleware.APIKeyMiddleware` (allowlist from `API_KEYS`); reads and
  `GET /health` are public. Empty `API_KEYS` disables write auth (dev) + warns.
- **Rate limiting:** `middleware.RateLimitMiddleware` (`RATE_LIMIT_RPM`, fixed
  window per client via X-Forwarded-For; `/health` exempt; `429` + `Retry-After`).
  0 disables (default).
- **Responses:** bare Open311 docs via `httputil.Send` (no envelope); errors via
  `httputil.SendError` (`{errors:[{code,description}]}`). JSON-first —
  `httputil.WantsXML` returns XML only for explicit non-browser XML clients.
  **Return structs/slices, never Go maps** (encoding/xml can't marshal maps).
  Middleware order: log → rate-limit → API-key → content-type.
- **Indexes:** `repository.EnsureIndexes` runs at startup (idempotent). New
  query fields should get an index there; `Create` derives a GeoJSON `location`.
- **Health:** `GET /health` pings MongoDB (`200` healthy / `503` unhealthy).
- **Extras / `properties`:** jurisdiction-specific fields with no Open311
  equivalent go in the `service_request.properties` string map (JSON object; XML
  `<property key>`; BSON subdoc). Per-jurisdiction column→field mappings live in
  [dictionaries/](dictionaries/) (reference only; the API doesn't load them).

---

## Known gotchas (read before editing)

1. **Persistence-DTO pattern (do not put `bson`/`ObjectID` in domain models).**
   Domain models (`models.*`) are pure JSON/XML DTOs with a **string** `ID`.
   Each repository defines a private `*Doc` struct that holds the `bson` tags +
   `primitive.ObjectID` `_id` and converts via `toModel()` / inline build on
   insert. New persisted fields go on the `*Doc` (mirror the json tag name), and
   `$set` keys / query filters must use those bson names. Nested types embedded
   in a `*Doc` (`UserOrganizationLink`, `ServiceAttribute`) also need `bson`
   tags. Rationale: without a `bson` tag the driver lowercases the whole field
   name, so `ID`↛`_id` and camelCase `$set` keys silently miss. See
   [developer-reference §8](developer-reference.md#8-data-model--mongodb-mapping).
2. **Operation timeout is dropped.** Repos do
   `opCtx, cancel := r.db.GetContext(); if ctx != nil { opCtx = ctx }`, so the
   configured `operationTimeoutSeconds` is never applied when a request context is
   passed. Prefer `context.WithTimeout(ctx, ...)`.
3. **Collection naming is inconsistent:** `Users` (PascalCase) vs `services` /
   `service_requests`. Normalize to lowercase during the overhaul.
4. **Two Mongo drivers:** `go.mod` has v1 (`mongo-driver` v1.17.3, used directly)
   and v2 (`mongo-driver/v2`, indirect). Pick one before adding features.
5. **Secrets:** config is env-var only (no config file); `MONGODB_URI` carries no
   password (X.509 cert auth), and the cert PEM is referenced by path
   (`MONGODB_TLS_CERT_KEY_FILE`), never committed. `.env` is gitignored.
   `SENTRY_SEND_DEFAULT_PII` defaults to `false` (GDPR). The old
   `src/config/config.json` is no longer read.
6. **`pkg/app/app.go` is dead code** — a second, unused API setup. Don't extend
   it; `internal/api` is the live path.
8. **Boston `service_code`s contain spaces and colons** — always URL-encode when
   building `/services/{service_code}` paths.

---

## Recipe: add a new endpoint

1. **Model** — add/extend a struct in `domain/models/` with `json`, `xml`, **and
   `bson`** tags. Add an XML wrapper if it's a collection.
2. **Repository** — add the method to the interface in
   `repository/repository.go`, implement it on the `Mongo*` repo, and add a
   testify mock + table tests (mirror
   [`service_request_repository_test.go`](src/internal/repository/service_request_repository_test.go)).
3. **Handler** — add a method on the `{Entity}Handler`; decode with
   `DecodeRequest`, map sentinel errors to status codes, respond via
   `SendResponse`. Add a handler test.
4. **Route** — register it in
   [`internal/api/api.go`](src/internal/api/api.go) `registerRoutes`.
5. **Contract** — document the endpoint in
   [developer-reference.md](developer-reference.md) and flip the relevant box in
   the README status list.
6. Run `go test ./...` and `go vet ./...`.

---

## Recipe: bulk-import Boston 311 data

[`scripts/feed-boston.ps1`](scripts/feed-boston.ps1) maps a data.boston.gov 311
CSV export (`case_enquiry_id, open_dt, closed_dt, case_status, type, latitude,
longitude, geom_4326, …`) to Open311 JSON and feeds it via the API. Standard
fields go top-level, jurisdiction extras under `properties`; coordinates are
preserved when valid (Boston bbox) so the server derives the `2dsphere`
`location`; Boston-local timestamps are converted to UTC (DST-aware); `&amp;`
and similar are HTML-decoded; rows with no location or no `service_code` are
skipped.

```powershell
$env:OPEN311_API_KEY = '<key>'
# Preview the mapping without sending:
./scripts/feed-boston.ps1 -CsvPath C:\path\boston.csv -First 3 -DryRun
# Bulk backfill (fast path — chunks arrays to POST /requests/bulk):
./scripts/feed-boston.ps1 -CsvPath C:\path\boston.csv -First 200000 -Bulk -BatchSize 500
# Paced per-record PUT (use when rate limiting is ON; honors Retry-After):
./scripts/feed-boston.ps1 -CsvPath C:\path\boston.csv -First 500 -DelaySeconds 6.5
```

Notes:
- **Idempotent** — keyed on `service_request_id` (Boston Case ID), so re-running
  updates in place. The full ~134k export loads in ~10 min via `-Bulk` (~266
  rows/s) vs ~7.8 h sequentially.
- **Rate limit:** the live API enforces ~10 req/min per client when
  `RATE_LIMIT_RPM` is set. For a big backfill, set `RATE_LIMIT_RPM=0` on the
  deployment, run `-Bulk`, then **restore it afterward**. The script also backs
  off on `429` automatically.
- The script holds **no secrets** — the key comes from `-ApiKey` / `$env:OPEN311_API_KEY`.

## MongoDB cert (X.509) auth — how it's wired

Hosted Mongo (Atlas) uses certificate auth. This is **implemented**:
- The connection string carries `authSource=$external` and
  `authMechanism=MONGODB-X509` (URL-encode `$` as `%24` in JSON), e.g.
  `mongodb+srv://<host>/?authSource=%24external&authMechanism=MONGODB-X509`.
  Note there is **no username/password** — the client certificate is the
  credential.
- [`config.MongoDBConfig`](src/config/config.go) carries
  `tlsCertificateKeyFile` (the **client** cert+key PEM — the X.509 user
  downloaded from Atlas, not the CA) and an optional `tlsCAFile` (leave empty to
  use system roots; Atlas's server cert chains to a public CA).
- [`repository.buildTLSConfig`](src/internal/repository/mongodb.go) loads them
  via `tls.LoadX509KeyPair(path, path)` (same path twice ⇒ cert+key in one file)
  and `connect()` applies them with `options.Client().SetTLSConfig(...)`.
- Config comes from env vars: `MONGODB_TLS_CERT_KEY_FILE` and `MONGODB_TLS_CA_FILE`
  (see [.env.example](src/.env.example)). The cert PEM lives outside the repo.

---

## Definition of done (for changes here)

- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean, `gofmt` applied
- [ ] New persisted fields have `bson` tags
- [ ] Contract change reflected in [developer-reference.md](developer-reference.md)
- [ ] No secrets added to tracked files
