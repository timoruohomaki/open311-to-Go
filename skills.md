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
| [`src/config/`](src/config/) | `config.go` loader + `config.json` (gitignored) |
| [`src/domain/models/`](src/domain/models/) | `User`, `Service`, `ServiceRequest` + XML wrappers |
| [`src/internal/api/`](src/internal/api/) | Route registration |
| [`src/internal/handlers/`](src/internal/handlers/) | HTTP handlers (`*_handler.go`) |
| [`src/internal/repository/`](src/internal/repository/) | Mongo connection + per-entity repos |
| [`src/pkg/`](src/pkg/) | Reusable: `router`, `middleware`, `logger`, `httputil`, `app` (unused) |

Module: `github.com/timoruohomaki/open311-to-Go` · Go **1.24.x** · net/http
routing (1.22+).

---

## Build / run / test

All commands run from `src/`.

```sh
cd src

# Run the dev server (reads ./config/config.json)
go run main.go -config="./config/config.json"
# or: make run

# Test
go test ./...                      # everything
go test ./pkg/middleware -v        # one package
go test ./internal/handlers -v

# Vet / format (do this before committing)
go vet ./...
gofmt -l .
```

> ⚠️ **Do not use `make build`.** The target is `go build -o main.go`, which
> writes the compiled binary over the `main.go` **source file**. Use
> `go build -o bin/open311api .` instead until the Makefile is fixed.

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

---

## Known gotchas (read before editing)

1. **BSON tags are mandatory.** Without a `bson` tag the driver lowercases the
   whole field name, so `ID`↛`_id` and camelCase `$set` keys silently miss. Every
   persisted field needs an explicit `bson` tag; ids should be
   `primitive.ObjectID` with `bson:"_id,omitempty"`. See
   [developer-reference §8](developer-reference.md#8-data-model--mongodb-mapping).
2. **Operation timeout is dropped.** Repos do
   `opCtx, cancel := r.db.GetContext(); if ctx != nil { opCtx = ctx }`, so the
   configured `operationTimeoutSeconds` is never applied when a request context is
   passed. Prefer `context.WithTimeout(ctx, ...)`.
3. **Collection naming is inconsistent:** `Users` (PascalCase) vs `services` /
   `service_requests`. Normalize to lowercase during the overhaul.
4. **Two Mongo drivers:** `go.mod` has v1 (`mongo-driver` v1.17.3, used directly)
   and v2 (`mongo-driver/v2`, indirect). Pick one before adding features.
5. **Secrets:** `src/config/config.json` holds the Mongo URI with credentials.
   It is gitignored (✅ never committed) but is plaintext — move to env/secret
   store and use **X.509 cert auth** for hosted Mongo (project requirement).
   `sentry.sendDefaultPII` is `true` — reconsider for GDPR.
6. **Decode errors are swallowed** in the service-request repo cursor loops
   (`continue` on `Decode` error). Log them.
7. **`pkg/app/app.go` is dead code** — a second, unused API setup. Don't extend
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

## Recipe: add MongoDB cert (X.509) auth

When wiring hosted Mongo with certificate auth (target setup):
- Connection string uses `authMechanism=MONGODB-X509` and
  `tls=true` / `tlsCertificateKeyFile=<pem>`.
- Configure via the driver `options.Client().SetTLSConfig(...)` rather than
  embedding paths in the URI where possible.
- Keep cert paths and any remaining secrets out of `config.json` (env vars).
- Update `repository/mongodb.go` `connect()` and the config struct accordingly.

---

## Definition of done (for changes here)

- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean, `gofmt` applied
- [ ] New persisted fields have `bson` tags
- [ ] Contract change reflected in [developer-reference.md](developer-reference.md)
- [ ] No secrets added to tracked files
