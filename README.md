# incident-api

Layanan API sederhana berbasis Go untuk pencatatan dan pengelolaan incident operasional.

## Fitur

- REST API menggunakan `net/http`
- Operasi CRUD untuk data incident
- Endpoint detail incident berdasarkan ID
- Endpoint statistik ringkas untuk rekap incident
- Penyimpanan data berbasis file JSON
- Health check untuk verifikasi service
- Filter dan pagination pada daftar incident
- Header `X-Request-ID` untuk identifikasi request
- Pengujian dasar pada layer API

## Endpoint

- `GET /healthz`
- `GET /v1/incidents`
- `GET /v1/incidents/{id}`
- `GET /v1/incidents/stats`
- `POST /v1/incidents`
- `PUT /v1/incidents/{id}`
- `DELETE /v1/incidents/{id}`

## Menjalankan Project

```bash
go run ./cmd/api
```

Secara default server berjalan pada `:8080`.
Data disimpan pada `data/incidents.json`.

Untuk menggunakan port lain:

```bash
PORT=8081 go run ./cmd/api
```

## Contoh Penggunaan

Melihat daftar incident:

```bash
curl http://localhost:8080/v1/incidents
```

Menambahkan incident baru:

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

Melihat detail incident:

```bash
curl http://localhost:8080/v1/incidents/1
```

Melihat statistik incident:

```bash
curl http://localhost:8080/v1/incidents/stats
```

Menggunakan filter:

```bash
curl "http://localhost:8080/v1/incidents?status=investigating&service=auth-service&limit=5&offset=0"
```

## Terminal Snapshot

![Terminal output](./assets/terminal-output.png)
