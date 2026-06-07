# Developer Reference — Open311-to-Go

This document is the **API contract reference** for this project. It describes the
Open311 GeoReport v2 endpoints as we implement them, the City of Boston (BOS:311)
flavor we follow, the Helsinki extensions we adopt, and our own spatial /
asset-management extensions for the **spatial-data-lake** project.

> Status: **living document.** Sections marked _(planned)_ are not yet
> implemented in code. Sections marked _(drift)_ describe where the current
> code diverges from this contract.

Reference implementations:
- **Boston BOS:311** (our primary reference, because our dataset is from Boston):
  - Docs: <https://311.boston.gov/open311/docs>
  - Live base URL: `https://311.boston.gov/open311/v2/`
- **Open311 GeoReport v2 spec**: <http://wiki.open311.org/GeoReport_v2/>
- **Helsinki extensions**: <https://dev.hel.fi/apis/open311>

---

## 1. Conventions

### Base URL & versioning
```
https://<host>/open311/v2/...           # this project (decided target)
https://<host>/api/v1/...               # current code (to be migrated)
https://311.boston.gov/open311/v2/...   # Boston reference
```
**Decided:** we adopt the GeoReport convention `/open311/v2/` as the route
prefix. The current code uses `/api/v1/`; migrating it is part of the overhaul
(see [§9 Drift](#9-current-state-vs-contract-drift)).

### Formats & content negotiation
- Every resource is available as **JSON** and **XML**.
- GeoReport convention puts the format in the path extension:
  `/services.json`, `/services.xml`, `/requests/{id}.json`.
- This project _additionally_ negotiates via the `Accept` header. It is
  **JSON-first**: XML is returned only when the client explicitly prefers it
  (`Accept: application/xml` / `text/xml`) and is not a browser — a browser's
  `Accept` contains `text/html`, so browsers and default clients get JSON
  (`httputil.WantsXML`). `Content-Type` is required and validated on `POST`/`PUT`
  by `ContentTypeMiddleware`.
- Response bodies must be structs/slices, **not Go maps** — `encoding/xml`
  cannot marshal maps, which would break the XML path.
- **Response shape:** bare Open311 documents — success responses write the data
  directly (a JSON array/object, or the XML collection element) with no envelope;
  errors use the `errors` document (see Error format). Helpers:
  `httputil.Send` / `httputil.SendError`.

### Jurisdiction
`jurisdiction_id` is required only when an endpoint serves multiple
jurisdictions. Boston serves a single jurisdiction and does not require it; this
project is single-jurisdiction by default. Accept it as an optional parameter for
spec compatibility.

### Authentication
- **This project (decided):** `X-API-Key: <key>` header, validated against an
  `API_KEYS` allowlist — consistent with the sibling [nps-api](https://github.com/timoruohomaki/nps-api).
- **Boston reference, for comparison:** `Authorization: Bearer <key>` (preferred)
  or `api_key` query/form param. We may *additionally* accept `api_key` for
  Open311 client compatibility, but `X-API-Key` is our primary scheme.
- Read endpoints (service list/definition, request lookup) and `/health` are
  public. Writes (`POST`/`PUT`/`PATCH`/`DELETE`) require a valid key.
- **Implemented** as `middleware.APIKeyMiddleware` (allowlist from the `API_KEYS`
  env var). If `API_KEYS` is empty, write auth is disabled and the server logs a
  warning at startup.

### Health check
`GET /health` **and** `GET /open311/v2/health` (public) — the prefixed path is
needed because the fronting proxy routes only `/open311/v2/*` to the service (the
bare `/health` is intercepted by the gateway). Pings MongoDB and returns `200`
`{"status":"healthy","database":"ok",...}` when reachable, or `503`
`{"status":"unhealthy","database":"unreachable"}` otherwise — suitable for load
balancer / container liveness probes and for confirming DB connectivity.

### Rate limiting
- **Implemented** as `middleware.RateLimitMiddleware`: a fixed-window per-client
  cap from `RATE_LIMIT_RPM` (0 = disabled, the default). `/health` is exempt.
  Client identity prefers the first `X-Forwarded-For` hop (set by the fronting
  proxy), falling back to `RemoteAddr`.
- On exceed: `429 Too Many Requests` with a `Retry-After` header.
- Boston's public reference uses **10 requests/minute**.

### Dates
ISO 8601 with timezone, e.g. `2026-06-07T08:15:30-05:00` or
`2026-06-07T13:15:30Z`. All text is UTF-8.

### Date-range limits (Boston)
Date-based search parameters are capped at a **90-day** span. List responses are
bounded by the smaller of the 90-day window or **1,000** requests.

### Error format
```json
{ "errors": [ { "code": 400, "description": "service_code is required" } ] }
```
XML equivalent:
```xml
<errors><error><code>400</code><description>service_code is required</description></error></errors>
```
Common codes: `400` invalid request, `403` missing/invalid API key, `404` not
found, `415` unsupported media type, `429` rate limited.

---

## 2. GET Service List

`GET /services.{format}`

| Param | Req? | Notes |
|---|---|---|
| `jurisdiction_id` | conditional | only for multi-jurisdiction servers |

**Service object fields:**

| Field | Type | Notes |
|---|---|---|
| `service_code` | string | unique id. **Boston uses long colon-delimited codes**, not numbers (see below) |
| `service_name` | string | human name |
| `description` | string | optional; Boston often omits |
| `metadata` | boolean | `true` ⇒ a service definition with attributes exists |
| `type` | string | `realtime` \| `batch` \| `blackbox` |
| `keywords` | string | comma-separated; optional |
| `group` | string | UI grouping |

**Boston example** (`service_code` is a hierarchical string):
```json
[
  {
    "service_code": "Public Works Department:Highway Maintenance:Request for Pothole Repair",
    "service_name": "Pothole",
    "metadata": false,
    "type": "batch",
    "group": "Roads & Sidewalks"
  },
  {
    "service_code": "Public Works Department:Street Cleaning:Pick up Dead Animal",
    "service_name": "Dead Animal Pickup",
    "metadata": true,
    "type": "batch",
    "group": "Pets, Wildlife, & Dead Animals"
  }
]
```
> ⚠️ **Design implication:** because Boston `service_code`s contain spaces and
> colons, any code used in a URL path (`/services/{service_code}`) must be
> URL-encoded. Plan storage/validation around opaque string codes, not integers.

---

## 3. GET Service Definition

`GET /services/{service_code}.{format}`

Returned only when the service's `metadata` is `true`.

**`service_definition` fields:**

| Field | Type | Notes |
|---|---|---|
| `service_code` | string | echoes the requested code |
| `attributes[]` | array | custom fields for this service |

**`attribute` fields:**

| Field | Type | Notes |
|---|---|---|
| `variable` | boolean | `false` ⇒ display-only (e.g. instructions), not submitted |
| `code` | string | attribute id, submitted as `attribute[code]=value` |
| `datatype` | string | `string` \| `number` \| `datetime` \| `text` \| `singlevaluelist` \| `multivaluelist` |
| `required` | boolean | |
| `datatype_description` | string | hint shown to the user |
| `order` | int | display order |
| `description` | string | label |
| `values[]` | array | for list types: `{ "key": ..., "name": ... }` |

```json
{
  "service_code": "Public Works Department:Highway Maintenance:Request for Pothole Repair",
  "attributes": []
}
```
> Note: many Boston services return an empty `attributes` array even when
> `metadata` is true.

---

## 4. POST Service Request

`POST /requests.{format}` — requires API key.

> **Implementation deviations from strict GeoReport** (intentional, consistent
> with this JSON/XML-first API):
> - **Body format:** we accept `application/json` or `application/xml`, *not*
>   GeoReport's `application/x-www-form-urlencoded`.
> - **`service_request_id`:** assigned synchronously at creation (the new
>   ObjectID hex), so responses always return `service_request_id` and never a
>   `token`. Defaults applied on create: `status=open`, `requested_datetime`/
>   `updated_datetime=now`.
> - **Required on create:** `service_code` and a location (`lat`+`long`, or
>   `address`, or `address_id`).

| Param | Req? | Notes |
|---|---|---|
| `api_key` | yes | or `Authorization: Bearer` header |
| `jurisdiction_id` | conditional | |
| `service_code` | yes | from the service list |
| `lat` + `long` | one-of | WGS84 |
| `address_string` | one-of | |
| `address_id` | one-of | |
| `attribute[<code>]` | conditional | required when the service definition says so; repeatable for `multivaluelist` |
| `description` | no | ≤ 4,000 chars |
| `email`, `first_name`, `last_name`, `phone` | no | reporter |
| `device_id`, `account_id` | no | |
| `media_url` | no | image URL (see Helsinki external-media extension §7) |

**Response** (`201`-ish; GeoReport returns the request stub):
```json
[ { "service_request_id": "638344", "token": null, "service_notice": "...", "account_id": null } ]
```
If the backend assigns ids asynchronously, it returns a `token` instead of a
`service_request_id`; resolve it via [§6 tokens](#6-get-service_request_id-from-token).

---

## 4a. PUT Service Request (upsert) — project extension

`PUT /requests/{service_request_id}.{format}` — requires API key.

Not part of GeoReport v2. Added for **idempotent, re-runnable bulk feeds** (e.g.
importing a periodic Boston 311 dump): inserts the request if absent, fully
**replaces** it if present, keyed on `service_request_id`. The existing MongoDB
`_id` is preserved on replace.

> **Semantics:**
> - **Body format:** `application/json` or `application/xml` (same as POST).
> - **Key:** the `{service_request_id}` in the URL is authoritative and overrides
>   any `service_request_id` in the body.
> - **Required:** same as POST — `service_code` and a location (`lat`+`long`, or
>   `address`, or `address_id`).
> - **`updated_datetime`:** **preserved when supplied** (defaults to now only when
>   absent) — this is the key difference from POST, which always stamps now. Lets
>   a feed carry the source's own update/close timestamps. `status` defaults to
>   `open` and `requested_datetime` to now when absent.
> - **Replace, not patch:** the body is the full resource; omitted fields are not
>   merged from the existing document.
> - **Status codes:** `201 Created` when newly created, `200 OK` when an existing
>   request was updated. Body is the stored request (as a single-element list,
>   like the other request endpoints).

---

## 4b. DELETE Service Request — project extension

`DELETE /requests/{service_request_id}.{format}` — requires API key.

Not part of GeoReport v2. Provided for **administrative cleanup** — removing test
records or mis-imported rows from a bulk feed, keyed on `service_request_id`.

> **Semantics:**
> - **Status codes:** `200 OK` on success (body `{"message":"Service request
>   deleted successfully"}`), `404 Not Found` when the id does not exist.
> - Hard delete (the document is removed, not flagged). There is no soft-delete or
>   status-based archival.

---

## 4c. POST Service Requests (bulk upsert) — project extension

`POST /requests/bulk` — requires API key.

Not part of GeoReport v2. For **high-throughput re-runnable feeds** (backfilling a
full Boston 311 dump): upserts many records in a single MongoDB `BulkWrite`
(unordered), collapsing per-record round-trips. Same upsert semantics as
`PUT /requests/{id}` (keyed on `service_request_id`, preserves a supplied
`updated_datetime`).

> **Semantics:**
> - **Body:** a JSON **array** of service requests, or an XML `<requests>`
>   document. Max **1000** records per call — chunk larger inputs.
> - **Per-record validation:** each needs `service_request_id`, `service_code`,
>   and a location (`lat`+`long`, `address`, or `address_id`). Invalid records are
>   rejected individually and reported; they do **not** abort the batch.
> - **In-batch de-duplication:** records sharing a `service_request_id` within one
>   call are collapsed (last wins) so they don't collide on the unique index.
> - **Status code:** `200 OK` with a summary; `400` only when the whole payload is
>   malformed, empty, or over the cap.

**Response:**
```json
{ "requested": 500, "created": 480, "updated": 18, "failed": 2,
  "errors": [ { "index": 7, "service_request_id": "", "message": "service_request_id is required" } ] }
```

> Throughput: against the live cluster, batched bulk upserts ingest the full
> ~134k-row export in minutes, versus ~7.8 h for sequential single PUTs.

---

## 5. GET Service Requests

**Single:** `GET /requests/{service_request_id}.{format}`
**List:** `GET /requests.{format}`

**List query parameters:**

| Param | Notes |
|---|---|
| `service_request_id` | comma-separated; overrides other filters |
| `service_code` | comma-separated |
| `status` | `open` \| `closed`, comma-separated |
| `start_date` / `end_date` | ISO 8601, ≤ 90-day span (defaults to last 90 days) |
| **Boston:** `q` | free-text search |
| **Boston:** `updated_after` / `updated_before` | ISO 8601, ≤ 90 days |
| **Boston:** `page` / `per_page` | `per_page` max **100** |

**`service_request` fields:**

| Field | Type |
|---|---|
| `service_request_id` | string |
| `status` | `open` \| `closed` |
| `status_notes` | string |
| `service_name` | string |
| `service_code` | string |
| `description` | string |
| `agency_responsible` | string |
| `service_notice` | string |
| `requested_datetime` | ISO 8601 |
| `updated_datetime` | ISO 8601 |
| `expected_datetime` | ISO 8601 |
| `address` | string |
| `address_id` | string |
| `zipcode` | string |
| `lat` | float (WGS84) |
| `long` | float (WGS84) |
| `media_url` | string |
| `token` | string (Boston includes it on every request) |

---

## 6. GET service_request_id from token

`GET /tokens/{token}.{format}` → `[{ "service_request_id": "...", "token": "..." }]`

Used to resolve an asynchronously-assigned id after a `POST`.

---

## 7. Extensions we adopt

### 7.1 Boston `extensions=true`
Passing `extensions=true` enriches responses:
- **Service definitions** add: `active` (bool), `notice` (string),
  `updated_at` (ISO 8601).
- **Service requests** add: `details`, `attributes`, `extended_attributes`
  (includes `x`/`y` coordinates in **ESRI:102686** projection and a `photos`
  array), and a `notes` array.

> For the spatial-data-lake we care about `extended_attributes` (projected
> coordinates + photos). We store canonical WGS84 `lat`/`long` and may derive or
> carry projected coordinates separately.

### 7.2 Helsinki extensions
- **Localization:** **deferred — English only for now.** We will not implement
  localized/multilingual text fields in this phase. When revisited, decide
  between a locale-keyed structure and a `*_fi` / `*_sv` / `*_en` suffix
  convention.
- **External media server:** images are not uploaded inline; `media_url` (and a
  list for multiple images) points to an external media server. Validate/allow-list
  hosts on ingest. _(Adopted.)_

### 7.3 Inline `properties` extension (Boston extras + PSK 5970)
**Implemented.** `service_request` carries an inline `properties` object — an open
map of **string** key/value pairs for fields with no Open311 equivalent. It serves
two purposes:
- **Per-jurisdiction extras** — e.g. Boston's `assigned_team`, `closure_reason`,
  `on_time`, `pwd_district`, `ward`, `precinct`, `neighborhood` (mapping in
  [dictionaries/boston-311.yaml](dictionaries/boston-311.yaml)).
- **PSK 5970 / ISO 55000** case-and-event annotations linking feedback to assets.

```json
{
  "service_request_id": "BCS-00059336",
  "status": "closed",
  "properties": {
    "closure_reason": "Resolved",
    "on_time": "ONTIME",
    "pwd_district": "1B",
    "asset_id": "BRIDGE-0042",
    "psk5970:event_class": "inspection"
  }
}
```
- **JSON:** a plain object. **XML:** `<properties><property key="k">v</property>…`
  (stable, key-sorted; round-trips via a custom (un)marshaler on the `Properties`
  type). **BSON:** a `properties` subdocument. Empty ⇒ omitted from all formats.
- Unknown keys pass through unchanged; the API does not enforce a vocabulary
  (the dictionary is reference-only). Values are strings.

### 7.4 NPS API
**NPS = Net Promoter Score** (satisfaction feedback), *not* National Park
Service. It is a separate sibling service:
<https://github.com/timoruohomaki/nps-api>, deployed at
`https://api.ruohomaki.fi/nps`. In the spatial-data-lake it contributes the
**citizen/user satisfaction signal** that can be correlated with service
requests and assets.

Same stack as this project (Go 1.26+, MongoDB Atlas, Sentry, Docker, GitHub
Actions), so it doubles as an **architectural reference** for several overhaul
decisions (see notes below).

**Endpoints:**

| Method | Path | Auth | Notes |
|---|---|---|---|
| `GET` | `/nps/health` | none | `{"status":"healthy","timestamp":"..."}` |
| `POST` | `/nps/api/v1/feedback` | `X-API-Key` (when `API_KEYS` set) | submit feedback |

**Feedback payload** (canonical schema in the repo's `docs/feedback-v1.json`):

| Field | Type | Notes |
|---|---|---|
| `schema_version` | string | e.g. `"1.0"` |
| `app` | string | non-empty, e.g. `"idefinity"` |
| `app_version` | string | e.g. `"0.1.0"` |
| `platform` | string | validated against `ALLOWED_PLATFORMS` allowlist |
| `timestamp` | ISO 8601 | |
| `nps_rating` | number | 0–10 |
| `nps_category` | string | `promoter` \| `passive` \| `detractor` |
| `timezone` | string | e.g. `"Europe/Helsinki"` |
| `comment` | string | optional |

Responses: `201` created, `400` invalid JSON, `422` validation failure.

**Patterns worth adopting from nps-api during this overhaul:**
- **Env-var config** (`MONGODB_URI`, `MONGODB_DATABASE`, `PORT`, `SENTRY_DSN`,
  `SENTRY_ENVIRONMENT`, …) instead of a plaintext `config.json` with embedded
  secrets. **Adopted** — see [.env.example](src/.env.example). (`API_KEYS` /
  `ALLOWED_PLATFORMS` arrive with the auth step.)
- **`X-API-Key` + `API_KEYS` allowlist** auth. **Decided:** this project
  standardizes on `X-API-Key` (consistent with nps-api); `api_key` may be
  accepted additionally for Open311 client compatibility.
- A public **`/health`** endpoint (this project currently has none).
- **Schema versioning** of payloads (`schema_version`) — apply the same to our
  `properties` / PSK 5970 extension.
- **Deployment / containerization is external:** handled by the **backend01**
  devops project (same model as nps-api). This repo does not carry a Dockerfile
  or compose; it just needs to build and read its env vars. Sentry provides
  error + uptime monitoring (uptime checks can poll `GET /health`).

> **Decided:** NPS is **purely a data source** for now — open311-to-Go does not
> call or aggregate it directly. Revisit only if new needs arise.

---

## 8. Data model & MongoDB mapping

Go models live in [`src/domain/models/`](src/domain/models/). MongoDB collections:

| Collection | Model | Notes |
|---|---|---|
| `service_requests` | `ServiceRequest` | snake_case ✅ |
| `services` | `Service` | lowercase |
| `Users` | `User` | **PascalCase — inconsistent**, normalize during overhaul |

### BSON mapping — persistence-DTO pattern (implemented)
The Mongo driver, **absent a `bson` tag, lowercases the entire Go field name**
(`FirstName` → `firstname`, `ID` → `id`). The original models relied on `json`
tags that do **not** apply to BSON, which caused: `ID` never mapping to Mongo's
`_id` (empty ids in responses) and `Update` `$set` writing camelCase keys that
reads never saw.

**How it's fixed:** persistence is separated from transport. Each repository
defines a private `*Doc` struct that carries the `bson` tags and an
`primitive.ObjectID` `_id`, and converts to/from the domain model. The domain
models (`models.User`, `models.Service`, `models.ServiceRequest`) stay pure
JSON/XML DTOs with a **string** `ID` — handlers, tests, and the API contract are
untouched, and BSON/ObjectID concerns never leak out of the repository layer.

```go
// in the repository package
type userDoc struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"` // omitempty ⇒ Mongo generates on insert
    Email     string             `bson:"email"`
    FirstName string             `bson:"firstName"`
    // ...
}
func (d userDoc) toModel() models.User { /* d.ID.Hex() → model.ID */ }
```
Rules: `bson` tags mirror the model's `json` tags; `$set` keys and query filters
use those same `bson` names; nested types embedded in a `*Doc` (e.g.
`UserOrganizationLink`, `ServiceAttribute`) carry `bson` tags too. The
`service_requests` filters use `featureId` / `featureGuid` / `organizationId`.

### Spatial storage
Store geometry as **GeoJSON** in MongoDB and add a `2dsphere` index to support
spatial queries (`$near`, `$geoWithin`) for the data-lake. Keep canonical
coordinates in WGS84; the Open311 `lat`/`long` fields map to GeoJSON
`[long, lat]` order (note the swap).

### Collection topology — regular vs. time-series

**Decision: `service_requests`, `services`, and `users` are regular
collections, NOT MongoDB time-series collections.**

It's tempting to model service requests as time-series data because each one
arrives with a timestamp. But a service request is a **mutable entity with a
lifecycle** (its `status` moves open → in_progress → closed, and
`status_notes` / `updated_datetime` / `expected_datetime` change over time), not
an immutable measurement. MongoDB time-series collections are purpose-built for
append-only data points, and their limitations are disqualifying here (verified
against current MongoDB docs):

| Time-series limitation | Open311 / data-lake need it breaks |
|---|---|
| Updates may modify **only the `metaField`** — no other field is updatable | `status`, `status_notes`, `updated_datetime`, `expected_datetime` all change |
| **No unique indexes** | unique `service_request_id` |
| **No `$near` / `$nearSphere`**; geospatial only via the `$geoNear` aggregation, and `2d` indexes only on the `metaField` | spatial `$near` / `$geoWithin` queries are core to the data-lake |
| **No direct deletes** (TTL only) | admin delete / GDPR erasure |
| No change streams, no schema validation, no CSFLE, no writes in transactions | XML schema validation, future change-driven pipelines |
| Collection type is fixed at creation; can't convert either direction | a wrong choice is expensive to undo |

So the **entity of record** lives in a regular collection. These indexes are
created idempotently at startup by `repository.EnsureIndexes`:
- a **unique** index on `service_request_id`
- a **`2dsphere`** index on a GeoJSON `location` field (`[long, lat]`)
- secondary indexes on `status`, `organizationId`, `featureId`, and
  `requested_datetime` / `updated_datetime` (for Boston's date-range queries)
- plus unique `service_code` (`services`) and sparse-unique `email` (`Users`)

`Create` derives the GeoJSON `location` from the request's `lat`/`long`. **Data
imported by an external pipeline must populate a `location` GeoJSON field** to be
covered by the `2dsphere` index (documents missing it are simply not geo-indexed).

**Where time-series *is* the right tool (future):** an **append-only event
stream** — exactly the "cases and **events**" half of PSK 5970 / ISO 55000.
When we add a status-change / activity log or asset condition/inspection
measurements, model *those* as a separate time-series collection
(`timeField = timestamp`, `metaField = service_request_id` or `asset_id`). That
data never mutates, so it gets the columnar compression and fast time-bucketed
analytics time-series is designed for. Rule of thumb: **the record of state →
regular collection; the immutable history of changes → time-series.** (NPS
feedback is the same append-only shape, but lives in its own
[nps-api](https://github.com/timoruohomaki/nps-api) service.)

---

## 9. Current state vs contract (drift)

| Area | Contract / target | Current code |
|---|---|---|
| Version prefix | `/open311/v2/` (decided) | ✅ migrated from `/api/v1/` |
| Service list | `GET /services` | ✅ implemented |
| Service definition | `GET /services/{code}` | ⚠️ uses `{id}` (Mongo `_id`), not `service_code` |
| Service CRUD | not in Open311 (admin only) | `POST/PUT/DELETE /services` exist |
| Service requests | `GET /requests`, `GET /requests/{id}`, `POST /requests` | ✅ implemented (+ `PUT /requests/{id}` idempotent upsert, `POST /requests/bulk` bulk upsert, `DELETE /requests/{id}` admin cleanup, `/requests/search` & `/requests/by_organization` extensions) |
| Tokens | `GET /tokens/{id}` | ⏭️ not implemented — ids assigned synchronously |
| Users | not part of Open311 | `GET /users`, `GET /users/{id}`; CRUD commented out |
| Auth | `X-API-Key` + allowlist | ✅ `X-API-Key` on writes |
| Rate limiting | 10/min, `429` + `Retry-After` | ✅ configurable (`RATE_LIMIT_RPM`, default off) |
| Health check | `GET /health` (DB ping) | ✅ implemented |
| Response shape | bare Open311 docs / `errors` | ✅ normalized (no `{status,data}` envelope) |
| Jurisdiction extras | `properties` extension | ✅ inline `properties` (JSON/XML/BSON); Boston mapped in [dictionaries/](dictionaries/boston-311.yaml) |
| XML schema validation | required | not started |
| BSON mapping | `_id` mapped, names consistent | ✅ fixed (persistence-DTO pattern) |
| Storage / collections | regular collections + GeoJSON `2dsphere`, unique `service_request_id` (decided; not time-series) | ✅ provisioned via `EnsureIndexes`; `open311-boston` backfilled with full Boston 311 export (~134k docs) |

---

## 10. Overhaul checklist (high level)

- [x] BSON `_id` mapping fixed via persistence-DTO pattern in repositories
- [ ] Normalize collection naming (`users` lowercase)
- [x] Canonical request endpoints `GET /requests`, `GET /requests/{id}`, `POST /requests` (tokens skipped — synchronous ids)
- [x] Idempotent `PUT /requests/{id}` upsert (re-runnable bulk feeds; preserves supplied `updated_datetime`)
- [x] `DELETE /requests/{id}` (admin cleanup of test / mis-imported records)
- [x] `POST /requests/bulk` (BulkWrite upsert for high-throughput backfills; feeder in [scripts/feed-boston.ps1](scripts/feed-boston.ps1))
- [x] Migrate route prefix `/api/v1` → `/open311/v2`
- [ ] Service definition lookup by `service_code` (currently by Mongo `_id`)
- [x] `X-API-Key` auth on writes (`API_KEYS` allowlist) + public `GET /health` (DB ping)
- [ ] Rate limiting (Boston: 10 req/min, `429` + `Retry-After`)
- [x] Provision indexes via `repository.EnsureIndexes` (unique `service_request_id`, `2dsphere` on GeoJSON `location`, secondaries; unique `service_code`/`email`)
- [x] Rate limiting (`RATE_LIMIT_RPM`, fixed-window, `429` + `Retry-After`)
- [x] Response-envelope normalization (bare Open311 docs + `errors` format)
- [ ] XML schema validation
- [ ] External-media (Helsinki) support — _localization deferred; English only_
- [x] Inline `properties` extension (Boston extras + PSK 5970), JSON/XML/BSON; example dictionary in [dictionaries/boston-311.yaml](dictionaries/boston-311.yaml)
- [ ] Integrate the NPS (Net Promoter Score) API as a satisfaction data source ([nps-api](https://github.com/timoruohomaki/nps-api))
- [x] MongoDB X.509 (`MONGODB-X509` / `$external`) cert auth wired in `connect()` (see [config.example.json](src/config/config.example.json))
- [x] Config migrated to env vars (12-factor; `.env` for local dev, see [.env.example](src/.env.example))
- [ ] Resolve mongo-driver **v1 vs v2** (both currently pulled in)
- [ ] Fix `Makefile` `build` target (`-o main.go` overwrites source)
