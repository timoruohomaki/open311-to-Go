# Open311-to-Go


## What is Open311?

Open311 is a form of technology that provides open channels of communication for issues that concern public space and public services. Primarily, Open311 refers to a standardized protocol for location-based collaborative issue-tracking. By offering free web API access to an existing 311 service, Open311 is an evolution of the phone-based 311 systems that many cities in North America offer.

Unlike the synchronous one-to-one communication of a 311 call center, Open311 technologies use the internet to enable these interactions to be asynchronous and many-to-many. This means that several different people can openly exchange information centered around a single public issue. This open model allows people to provide more actionable information for those who need it most and it encourages the public to be engaged with civic issues because they know their voices are being heard. Yet Open311 isn't just about this more open internet-enabled model for 311 services, it's also about making sure the technology itself is open so that 311 services and applications are interoperable and can be used everywhere.

The key features of Opn311 can be summarized followingly:
*  **Open Data Format:** Open311 provides a standardized data format for issues to be reported and tracked. This makes it easy for different systems to communicate with one another.
*  **APIs:** Open311 offers Application Programming Interfaces (APIs) for developers, allowing third-party applications to integrate with government services. For example, a smartphone app can be created to enable users to report problems directly from their phones.
*  **Transparency:** Open311 promotes transparency by allowing local governments to share the status and progress of reported issues with the public, making it easier for citizens to see how their concerns are being addressed.

Source: https://www.open311.org/learn/

## Why Go?

Most of the Open311 implementations so far are based on Python on Django framework. While easy to implement, that might not be the most optimal starting point for a highly performant and scalable API backend.

## Anything new here?

This implementation inludes the additions the City of Helsinki added on Open311, mainly support for other languages and locales and support for external media server in cases where images are included in the ticket.

Source: https://dev.hel.fi/apis/open311 

This implementation also uses MongoDB as a backend, also utilizing its spatial functions. XML formats are supported with schemas.

Due to the experimental nature of this implementation, the schema for service request is extended with inline properties object, containing user-annotated properties as key-value pairs. This approach makes it possible to support use cases where there are additional properties in the service request, e.g. because of supporting a specific standard such as the Finnish PSK 5970 that defines the schema for data record of cases and events. With this approach, the goal is to link citizen feedback with ISO 55000 asset management practises.

## Project context: spatial-data-lake

This API is a building block of the **spatial-data-lake** project — a municipal
asset-management data structure that combines Open311 with an NPS (Net Promoter
Score) satisfaction-feedback API ([nps-api](https://github.com/timoruohomaki/nps-api),
a sibling Go/MongoDB service). Our
dataset comes from the **City of Boston (BOS:311)**, so Boston's implementation is
our primary reference for the API contract, extended where useful (as is the
convention in most cities). We adopt the City of Helsinki extensions for external
media servers and localization, run hosted MongoDB with **certificate (X.509)
authentication**, and use Sentry for monitoring.

> This project had been dormant for some time; an overhaul is in progress. See
> the [overhaul checklist](developer-reference.md#10-overhaul-checklist-high-level).

## API contract & documentation

- **[developer-reference.md](developer-reference.md)** — the API contract:
  GeoReport v2 endpoints as we implement them, the Boston flavor, Helsinki and
  PSK 5970 extensions, the data model / MongoDB mapping, and where the current
  code still diverges from the contract.
- **[skills.md](skills.md)** — engineering playbook: build/run/test commands,
  conventions, known gotchas, and recipes for adding endpoints.

Reference docs: [Boston BOS:311](https://311.boston.gov/open311/docs) ·
[GeoReport v2 spec](http://wiki.open311.org/GeoReport_v2/) ·
[Helsinki Open311](https://dev.hel.fi/apis/open311)

## Development framework and versions

* This work uses golang version 1.26.x [^1] (`go.mod` pins 1.26.4). The work depends on the new Go net/http routing capabilities so a version of 1.22 or newer is required.
* Deployment and containerization are handled externally by the **backend01** devops project (same model as the sibling nps-api); this repo carries no Dockerfile. The dev server is Ubuntu 22.04 hosted at api.spatialworks.fi
* The development is done using Visual Studio Code - however it shouldn't make any difference what editor to use
* Sentry is used for telemetry (free version will be enough).
* API examples and tests collection is managed with [Bruno](https://www.usebruno.com/)

## Implementation Status

Infrastructure (done):

* [x]  Github action for Ubuntu ci/cd pipeline
* [x]  Logging (Syslog)
* [x]  Observability (Sentry)
* [x]  MongoDB database backend
* [x]  Content negotiation (JSON / XML) middleware

Open311 GeoReport v2 endpoints (served under `/open311/v2/` — see [developer-reference.md](developer-reference.md)):

* [x]  GET Service List — `GET /open311/v2/services`
* [ ]  GET Service Definition — `GET /open311/v2/services/{id}` _(by Mongo `_id`; `service_code` lookup pending)_
* [x]  POST Service Request — `POST /open311/v2/requests`
* [x]  GET Service Request by id — `GET /open311/v2/requests/{id}`
* [x]  GET Service Requests (list) — `GET /open311/v2/requests`
* [ ]  GET service_request_id from token — _skipped; ids assigned synchronously_

> Routes are served under `/open311/v2/` (migrated from `/api/v1/`). The project
> also exposes spatial-lookup extensions (`/requests/search`,
> `/requests/by_organization`) and `GET /users`. POST accepts JSON/XML (not
> GeoReport form-urlencoded); see deviations in developer-reference.

Cross-cutting (not started):

* [x]  API auth — `X-API-Key` on writes (`API_KEYS` allowlist); reads public
* [x]  `GET /health` — liveness + MongoDB connectivity (503 when DB unreachable)
* [x]  Rate limiting (`RATE_LIMIT_RPM`, fixed window, `429` + `Retry-After`; default off)
* [x]  Bare Open311 response shape (no `{status,data}` envelope; `errors` format)
* [x]  MongoDB X.509 certificate authentication (wired; see [.env.example](src/.env.example))
* [ ]  Schema validation on XML messages
* [x]  GeoJSON storage + `2dsphere` spatial index (via `EnsureIndexes`; `Create` derives `location`)
* [ ]  TLS termination (handled at the proxy / backend01)
* [x]  BSON tag / `_id` mapping fix (persistence-DTO pattern; see [developer-reference §8](developer-reference.md#8-data-model--mongodb-mapping))
* [ ]  External media server (Helsinki) — _localization deferred; English only_
* [ ]  Inline `properties` (PSK 5970) passthrough
* [ ]  NPS (Net Promoter Score) API integration as satisfaction data source

## What is the motivation for this?

This work is to support my master's thesis work on large scale asset management on urban digital twins. This should be up and running by the end of 2025-ish.

## Credits

* This work heavily relies on the concept of distributed services Travis Jeffery provided in his book [Distributed Services with Go](https://a.co/d/g5mhjd8).
* Credits also to Ishan Shrestha on RestAPI and MongoDB best practises, [blog here](https://medium.com/@ishan.shrestha356/scalable-json-restapi-using-go-lang-and-mongodb-cf9699c5f6e8)

## Project Structure

The project follows a modular Go structure:

```
open311-to-Go/
  src/
    config/         # Env-var config loader + .env.example
    domain/         # Domain models (service, user, serviceRequest)
    internal/
      api/          # API setup and route registration
      handlers/     # HTTP handlers for business logic
      repository/   # MongoDB and repository interfaces
    pkg/
      app/          # Legacy/alternate App setup — currently unused (live path is internal/api)
      httputil/     # HTTP utilities (params, response helpers)
      logger/       # Logging framework (syslog, file, stdout)
      middleware/   # HTTP middleware (logging, content-type)
      router/       # Custom router implementation
    main.go         # Main entry point
    Makefile        # Build and test commands
```

## Configuration & running

Configuration is **environment variables only** (12-factor) — there is no config
file. [`src/.env.example`](src/.env.example) lists every variable with defaults;
only `MONGODB_URI` is required.

```sh
cd src
cp .env.example .env     # .env is gitignored; fill in MONGODB_URI etc.
go run main.go           # loads ./.env if present; real env vars take precedence
# make run               # same thing
# make build             # outputs ./bin/open311api
```

MongoDB uses **X.509 certificate auth** (`MONGODB-X509` / `authSource=$external`),
so the URI carries no password; `MONGODB_TLS_CERT_KEY_FILE` points to the client
certificate + key PEM (kept outside the repo).

## Tests

The project includes unit tests for core logging functionality:

- **Logging Middleware:**
  - Ensures HTTP requests are logged in strict Apache Combined Log Format (suitable for analytics tools like Matomo).
  - Verifies correct logging of status code, response size, referer, and user-agent.
- **Syslog Logger:**
  - Verifies that the logger can be created with syslog configuration enabled (UDP, remote syslog server).
  - Ensures no errors occur when enabling syslog logging (even if no syslog server is running).

To run the tests:

```sh
cd src
# Run all tests
go test ./...
# Run only logging middleware tests
go test ./pkg/middleware -v
# Run only logger tests
go test ./pkg/logger -v
```



[^1]: It should be always possible to continue development using the latest stable release

