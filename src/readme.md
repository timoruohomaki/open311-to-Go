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
- `OrganizationID` (string, optional): Reference to the organization related to the request

### Service
Represents a service in the Open311 system. Key fields:
- `ID`, `ServiceCode`, `ServiceName`, `Description`, `Metadata`, `Type`, `Keywords`, `Group`
- `Attributes`: List of custom attributes (see ServiceAttribute)
- `CreatedAt`, `UpdatedAt` (time.Time)

### User
Represents a user in the system. Key fields:
- `ID`, `Email`, `FirstName`, `LastName`
- `Phone` (string, optional): User's phone number
- `Organization` (string, optional): Organization name or identifier
- `OrgType` (enum, optional): Type of organization/user (see below)
- `CreatedAt`, `UpdatedAt` (time.Time)

#### OrgType enumeration
- `unknown`: Default/unspecified
- `subcontractor`: Subcontractor identifying issues
- `supervisor`: Internal supervisor monitoring progress/service levels
- `internal`: Internal user
- `external`: External user

This allows distinguishing user roles and organizations for access control, reporting, and workflow logic.

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
- `GET /api/v1/service_requests/by_organization?organizationId=...` - Search service requests by organization
- Other endpoints for users and services are also available (see `api.go`)

---

## Observability: Sentry Integration

- Sentry is integrated for error and performance monitoring using the [Sentry Go net/http guide](https://docs.sentry.io/platforms/go/guides/http/).
- Configuration is provided in `config/config.json` under the `sentry` section (DSN, tracing, etc.).
- Sentry is initialized in `main.go` before the server starts. The API handler is wrapped with Sentry's HTTP middleware for automatic error and performance capture.
- Example config:
  ```json
  "sentry": {
    "dsn": "https://examplePublicKey@o0.ingest.sentry.io/0",
    "enableTracing": true,
    "tracesSampleRate": 1.0,
    "sendDefaultPII": true
  }
  ```
- **Note:** For production, move secrets like the Sentry DSN to a secure location (e.g., environment variables or a secrets manager).

---

## Testing

- Repository and handler logic are covered by tests:
  - `internal/repository/service_request_repository_test.go`
  - `internal/handlers/service_request_handler_test.go`
- Tests cover searching by featureId, featureGuid, both, and neither.

---

## MongoDB Setup (using MongoDB Atlas)

1. **Create a MongoDB Atlas Account:**
   - Go to https://www.mongodb.com/cloud/atlas and sign up or log in.

2. **Create a Cluster:**
   - Click "Build a Database" and choose a free or paid cluster.
   - Select your cloud provider and region, then create the cluster.

3. **Create a Database User:**
   - In the Atlas dashboard, go to "Database Access".
   - Add a new database user with a username and password.
   - Assign "Read and write to any database" or restrict as needed.

4. **Configure Network Access:**
   - In "Network Access", add your IP address or 0.0.0.0/0 (for development only) to the IP whitelist.

5. **Get the Connection String:**
   - In "Clusters", click "Connect" > "Connect your application".
   - Copy the provided connection string (e.g., `mongodb+srv://<user>:<password>@cluster0.mongodb.net/<dbname>?retryWrites=true&w=majority`).
   - Update `src/config/config.json` with your URI and database name.

6. **Create Collections:**
   - Collections are created automatically when you insert the first document.
   - For this project, you will need at least:
     - `service_requests`
     - `users`
     - `services`
     - (optionally) `organizations`, `features_of_interest`, etc.

7. **Test the Connection:**
   - Run the application. If the connection is successful, you should see a log message indicating MongoDB is connected.

**Note:**
- Never commit your real MongoDB credentials to version control.
- For production, restrict network and user access as much as possible.

