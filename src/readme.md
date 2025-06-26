# Source Code

### Project structure:

- `config/` - Configuration files and logic
- `domain/models/` - Core data models (ServiceRequest, Service, User)
- `internal/api/` - API initialization and route registration
- `internal/handlers/` - HTTP handlers for API endpoints
- `internal/repository/` - Data access layer (MongoDB repositories)
- `pkg/` - Utilities, middleware, logger, router, and HTTP helpers

---

## Data Models

### ServiceRequest
Represents a request in the system. Key fields:
- `ID` (string): Internal unique identifier
- `ServiceRequestID` (string): Public request identifier
- `Status`, `StatusNotes`, `ServiceName`, `ServiceCode`, `Description`, `AgencyResponsible`, `ServiceNotice` (string): Metadata fields
- `RequestedDatetime`, `UpdatedDatetime`, `ExpectedDatetime` (time.Time): Timestamps
- `Address`, `AddressID`, `Zipcode` (string): Location/address info
- `Latitude`, `Longitude` (float64): Geospatial coordinates
- `MediaURL` (string): Optional media attachment
- `FeatureID` (string, optional): OGC API Features canonical URI for a geospatial feature
- `FeatureGuid` (string, optional): Unique identifier for a geospatial feature within its collection

### Service
Represents a service in the Open311 system. Key fields:
- `ID`, `ServiceCode`, `ServiceName`, `Description`, `Metadata`, `Type`, `Keywords`, `Group`
- `Attributes`: List of custom attributes (see ServiceAttribute)
- `CreatedAt`, `UpdatedAt` (time.Time)

### User
Represents a user in the system. Key fields:
- `ID`, `Email`, `FirstName`, `LastName`
- `CreatedAt`, `UpdatedAt` (time.Time)

---

## Repository Layer

- Each model has a corresponding repository in `internal/repository/`.
- `ServiceRequestRepository` provides `FindByFeature(ctx, featureID, featureGuid)` to search for service requests by OGC feature URI or GUID.
- Example usage:
  ```go
  repo.FindByFeature(ctx, "https://example.com/ogcapi/collections/parks/items/park-42", "park-42")
  ```

---

## Handler Layer

- Handlers are in `internal/handlers/`.
- `ServiceRequestHandler` exposes `SearchServiceRequestsByFeature`, which handles:
  - `GET /api/v1/service_requests/search?featureId=...&featureGuid=...`
- Returns all service requests matching the given feature URI and/or GUID.

---

## API Endpoints

- `GET /api/v1/service_requests/search?featureId=...&featureGuid=...` - Search service requests by geospatial feature
- Other endpoints for users and services are also available (see `api.go`)

---

## Testing

- Repository and handler logic are covered by tests:
  - `internal/repository/service_request_repository_test.go`
  - `internal/handlers/service_request_handler_test.go`
- Tests cover searching by featureId, featureGuid, both, and neither.

