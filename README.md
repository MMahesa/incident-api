# incident-api

API kecil untuk nyatet incident operasional. Saya buat ini buat latihan bikin service Go yang simpel tapi tetap kepakai untuk kasus monitoring atau tiket internal.

## Fitur

- REST API menggunakan `net/http`
- CRUD incident
- Ambil detail incident per ID
- Endpoint statistik sederhana
- Penyimpanan data ke file JSON
- Health check
- Filter dan pagination
- Header `X-Request-ID`
- Test dasar

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

Secara default server berjalan di `:8080`.
Data akan disimpan di `data/incidents.json`.

Untuk ganti port:

```bash
PORT=8081 go run ./cmd/api
```

## Cara Coba Cepat

Lihat daftar incident:

```bash
curl http://localhost:8080/v1/incidents
```

Tambah incident baru:

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

Lihat detail incident:

```bash
curl http://localhost:8080/v1/incidents/1
```

Lihat statistik:

```bash
curl http://localhost:8080/v1/incidents/stats
```

Contoh filter:

```bash
curl "http://localhost:8080/v1/incidents?status=investigating&service=auth-service&limit=5&offset=0"
```

## Terminal Snapshot

![Terminal output](./assets/terminal-output.png)
