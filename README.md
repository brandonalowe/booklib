# BookLib

Simple Go backend with a small embedded frontend.

Run locally:

```bash
# build
go build ./cmd/server
# run
./cmd/server
# or
# go run ./cmd/server
```

Open http://localhost:8080 to view the frontend. The frontend calls the API endpoints:
- GET /books
- GET /books/{id}
- POST /books
- PUT /books/{id}
- DELETE /books/{id}

Notes:
- Static frontend files are embedded in the binary from `cmd/server/web` using Go's `embed`.
- For deployment to Fly or Render, serve the built binary as normal and allow port configuration via environment (currently hard-coded to :8080).
