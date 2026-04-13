# incident-api

Project backend portfolio berbasis Go untuk mengelola incident dan tiket operasional.

## Gambaran

- REST API menggunakan `net/http`
- Fitur CRUD incident untuk alur kerja backend atau NOC
- Penyimpanan berbasis file dengan validasi input
- Endpoint health check untuk pengecekan service
- Filtering dan pagination pada daftar incident
- Header `X-Request-ID` untuk traceability request
- Seed data awal untuk demo cepat
- Cakupan test dasar pada layer handler

## Endpoint

- `GET /healthz`
- `GET /v1/incidents`
- `POST /v1/incidents`
- `PUT /v1/incidents/{id}`
- `DELETE /v1/incidents/{id}`

## Kegunaan Project

Project ini cocok untuk menunjukkan dasar kemampuan backend, terutama untuk kasus operasional seperti:

- pencatatan incident layanan
- pelacakan status penanganan gangguan
- pengelolaan tiket internal sederhana
- dasar API untuk dashboard monitoring atau sistem NOC

## Menjalankan Project

```bash
go run ./cmd/api
```

Secara default server berjalan di `:8080`.
Data akan disimpan di `data/incidents.json`.

Jika ingin mengganti port:

```bash
PORT=8081 go run ./cmd/api
```

## Contoh Request

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

## Contoh Filter

```bash
curl "http://localhost:8080/v1/incidents?status=investigating&service=auth-service&limit=5&offset=0"
```
