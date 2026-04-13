# incident-api

Backend portfolio project in Go for managing incidents and operational tickets.

## Highlights

- REST API with `net/http`
- Incident CRUD basics for NOC/backend workflows
- File-based persistence with validation
- Health endpoint for service checks
- Filtering and pagination on incident listing
- Request ID response header for traceability
- Seed data for quick demo
- Basic handler test coverage

## Endpoints

- `GET /healthz`
- `GET /v1/incidents`
- `POST /v1/incidents`
- `PUT /v1/incidents/{id}`
- `DELETE /v1/incidents/{id}`

## Run

```bash
go run ./cmd/api
```

Server listens on `:8080` by default.
Data is persisted in `data/incidents.json`.

## Example Request

```bash
curl -X POST http://localhost:8080/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{
    "title":"Packet loss on uplink",
    "service":"edge-router",
    "severity":"high",
    "status":"investigating",
    "description":"Packet loss above threshold",
    "owner":"noc-team"
  }'
```

## Example Filters

```bash
curl "http://localhost:8080/v1/incidents?status=investigating&service=auth-service&limit=5&offset=0"
```
