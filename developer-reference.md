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
- This project _additionally_ negotiates via the `Accept` header
  (`application/json` default, `application/xml`). `Content-Type` is required and
  validated on `POST`/`PUT` by `ContentTypeMiddleware`.

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
  public. `POST /requests` (and any future writes) require a key.
- _(drift)_ This project currently implements **no** API authentication.

### Rate limiting (Boston)
- Default **10 requests/minute**; request an application key for more.
- `429 Too Many Requests` when exceeded; honor the `Retry-After` header.

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

`POST /requests.{format}` — requires API key. Body is
`application/x-www-form-urlencoded; charset=utf-8`.

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

### 7.3 Project extension — inline `properties` (PSK 5970)
To link citizen feedback to **ISO 55000** asset-management practice, we extend
`service_request` with an inline `properties` object of user-annotated
key/value pairs. This supports records conforming to the Finnish **PSK 5970**
standard (schema for cases and events).
```json
{
  "service_request_id": "...",
  "properties": {
    "asset_id": "BRIDGE-0042",
    "psk5970:event_class": "inspection",
    "psk5970:condition_grade": "3"
  }
}
```
> _(planned)_ Exact key vocabulary TBD. Keep it an open map so unknown keys pass
> through; validate known namespaces (`psk5970:*`) when a schema is finalized.

### 7.4 NPS API
**NPS = Net Promoter Score** (satisfaction feedback), *not* National Park
Service. It is a separate sibling service:
<https://github.com/timoruohomaki/nps-api>, deployed at
`https://api.ruohomaki.fi/nps`. In the spatial-data-lake it contributes the
**citizen/user satisfaction signal** that can be correlated with service
requests and assets.

Same stack as this project (Go 1.24+, MongoDB Atlas, Sentry, Docker, GitHub
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
- Containerized **Docker + Nginx** deployment (vs. this project's stated Azure
  Functions target — reconcile).

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

So the **entity of record** lives in a regular collection. Provision
`service_requests` with:
- a **unique** index on `service_request_id`
- a **`2dsphere`** index on a GeoJSON `location` field (`[long, lat]`)
- secondary indexes on `status`, `organizationId`, `featureId`, and
  `requested_datetime` / `updated_datetime` (for Boston's date-range queries)

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
| Version prefix | `/open311/v2/` (decided) | `/api/v1/` |
| Service list | `GET /services` | ✅ implemented |
| Service definition | `GET /services/{code}` | ✅ (`{id}`, not `{service_code}`) |
| Service CRUD | not in Open311 (admin only) | `POST/PUT/DELETE /services` exist |
| Service requests | `GET /requests`, `GET /requests/{id}`, `POST /requests`, `GET /tokens/{id}` | **none of these**; instead `GET /service_requests/search` and `/by_organization` (project-specific spatial lookups) |
| Users | not part of Open311 | `GET /users`, `GET /users/{id}`; CRUD commented out |
| Auth | `X-API-Key` + allowlist + rate limit | none |
| XML schema validation | required | not started |
| BSON mapping | `_id` mapped, names consistent | ✅ fixed (persistence-DTO pattern) |
| Storage / collections | regular collections + GeoJSON `2dsphere`, unique `service_request_id` (decided; not time-series) | no indexes defined yet |

---

## 10. Overhaul checklist (high level)

- [x] BSON `_id` mapping fixed via persistence-DTO pattern in repositories
- [ ] Normalize collection naming (`users` lowercase)
- [ ] Implement the canonical request endpoints (`/requests`, `/requests/{id}`, `POST /requests`, `/tokens/{id}`)
- [ ] Migrate route prefix `/api/v1` → `/open311/v2`
- [ ] API authentication (`X-API-Key` + `API_KEYS` allowlist) and rate limiting
- [ ] Provision indexes: unique `service_request_id`, `2dsphere` on GeoJSON `location`, secondary on `status`/`organizationId`/`featureId`/`*_datetime` (regular collections — see §8 Collection topology)
- [ ] XML schema validation
- [ ] External-media (Helsinki) support — _localization deferred; English only_
- [ ] `properties` (PSK 5970) passthrough + validation
- [ ] Integrate the NPS (Net Promoter Score) API as a satisfaction data source ([nps-api](https://github.com/timoruohomaki/nps-api))
- [x] MongoDB X.509 (`MONGODB-X509` / `$external`) cert auth wired in `connect()` (see [config.example.json](src/config/config.example.json))
- [x] Config migrated to env vars (12-factor; `.env` for local dev, see [.env.example](src/.env.example))
- [ ] Resolve mongo-driver **v1 vs v2** (both currently pulled in)
- [ ] Fix `Makefile` `build` target (`-o main.go` overwrites source)
