# Go backend (minimal scaffold)

This folder contains a minimal Go HTTP server that runs without external
dependencies so you can iterate quickly. It provides two endpoints:

- `GET /api/health` — returns a small JSON health response.
- `POST /api/auth/verify` — accepts JSON `{ "idToken": "..." }` and
  returns a mocked user object. This is a placeholder until Firebase Admin
  verification and MongoDB upsert are added.

Run locally:

```powershell
Set-Location 'C:\Users\Jappan\ReactProjects\betting app indrajeet\backend\go'
go run .
```

The server listens on port `4001` by default.

Next steps:
- Add `firebase` initialization and `mongo` helper.
- Replace the mock verify handler with real token verification and upsert.
# Go backend (scaffold)

This folder contains a minimal Go backend scaffold so you can run the server without fetching external modules yet.

Run locally (PowerShell):

```powershell
Set-Location 'C:\Users\Jappan\ReactProjects\betting app indrajeet\backend\go'
go run .\main.go
```

Later steps:
- Add `firebase.go` to initialize Firebase Admin SDK.
- Add `mongo.go` to connect to MongoDB using `go.mongodb.org/mongo-driver` and pin a working version.
- Implement token verification handler to verify Firebase ID tokens and upsert users.
# Backend (Go) for Betting App

This Go server verifies Firebase ID tokens and upserts users into MongoDB.

Set up
1. Copy `.env.example` to `.env` and fill values. Provide either `FIREBASE_SERVICE_ACCOUNT_PATH` or `FIREBASE_SERVICE_ACCOUNT`.

2. Install dependencies and build/run

```powershell
cd backend/go
go mod tidy
go run main.go
```

If you see an error similar to "missing go.sum entry for module", run:

```powershell
cd backend/go
go mod download
go run main.go
```

This will populate `go.sum` with trusted module checksums. If you are behind a corporate proxy, ensure `GOPROXY` is set and reachable; you can also run `go env GOPROXY` to confirm.

If you are having trouble fetching a specific mongodb driver version ("unknown revision"), try:

```powershell
cd backend/go
go get go.mongodb.org/mongo-driver@latest
go mod tidy
go run main.go
```

If behind a proxy, make sure GOPROXY is reachable:

```powershell
go env -w GOPROXY=https://proxy.golang.org,direct

If `go mod download` still fails, try forcing the public proxy and retry:

```powershell
cd backend/go
go env -w GOPROXY=https://proxy.golang.org,direct
go env -w GOSUMDB=sum.golang.org
go mod download
```

If that still fails, run `.\listMongoVersions.ps1` and paste its output here — I will pick a safe version and update the `go.mod` for you.

Extra diagnostics: If `go get` still fails, list all available versions that the proxy knows about and choose one to pin:

```powershell
cd backend/go
.\listMongoVersions.ps1
# This will show versions like: vX.Y.Z vX.Y.(Z-1) ...
# Choose a stable version and run:
go get go.mongodb.org/mongo-driver@vX.Y.Z
go mod tidy
```
```

3. Endpoints
- POST /api/auth/verify — body: { "idToken": "<firebase-id-token>" }

Notes:
- Use `firebase-admin` service account for token verification.
- This is intentionally minimal. Add CORS, logging, and validation for production.
