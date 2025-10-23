# FTTH Management Backend - Golang + SQLite

Backend API server untuk sistem management FTTH menggunakan Golang dan SQLite database.

## Fitur

- RESTful API untuk CRUD Router (Server, OLT, ODC, ODP)
- RESTful API untuk CRUD Pelanggan
- Integrasi Mikrotik RouterOS API
- Export/Import data JSON
- Statistics Dashboard
- SQLite database untuk storage
- CORS enabled untuk frontend

## Instalasi

### 1. Install Golang (Jika ingin build ulang)
Download dan install Golang dari https://golang.org/dl/

**Note:** File `ftth-backend.exe` sudah tersedia dan siap pakai!

### 2. Run
```bash
# Windows
ftth-backend.exe

# Linux/Mac
./ftth-backend
```

Server akan berjalan di http://localhost:8080

## API Endpoints

### Routers (Server/OLT/ODC/ODP)

- `GET /api/routers` - Get all routers
- `GET /api/routers/{id}` - Get router by ID
- `POST /api/routers` - Create new router
- `PUT /api/routers/{id}` - Update router
- `DELETE /api/routers/{id}` - Delete router

#### Request Body (POST/PUT):
```json
{
  "name": "Server 1",
  "type": "server",
  "parent_id": null,
  "coordinates": "-6.123456,106.789012"
}
```

### Pelanggan

- `GET /api/pelanggan` - Get all pelanggan
- `GET /api/pelanggan/{id}` - Get pelanggan by ID
- `POST /api/pelanggan` - Create new pelanggan
- `PUT /api/pelanggan/{id}` - Update pelanggan
- `DELETE /api/pelanggan/{id}` - Delete pelanggan

#### Request Body (POST/PUT):
```json
{
  "name": "Customer Name",
  "odp_id": "odp-uuid",
  "pppoe": "user123",
  "whatsapp": "628xxx",
  "coordinates": "-6.123456,106.789012",
  "status": "offline"
}
```

### Mikrotik Integration

- `GET /api/mikrotik/status` - Get active PPPoE connections and update pelanggan status
- `GET /api/mikrotik/test` - Test Mikrotik connection
- `GET /api/mikrotik/config` - Get Mikrotik configuration
- `PUT /api/mikrotik/config` - Update Mikrotik configuration

#### Mikrotik Config (PUT):
```json
{
  "host": "192.168.88.1",
  "user": "admin",
  "password": "password",
  "port": 8728
}
```

### Utilities

- `GET /api/stats` - Get dashboard statistics
- `GET /api/export` - Export all data as JSON
- `POST /api/import` - Import data from JSON

## Database

Database SQLite akan otomatis dibuat di `ftth.db` saat pertama kali running.

### Schema:

#### Table: routers
- id (TEXT, PRIMARY KEY)
- name (TEXT)
- type (TEXT) - 'server', 'olt', 'odc', 'odp'
- parent_id (TEXT, FOREIGN KEY)
- coordinates (TEXT)
- created_at (DATETIME)

#### Table: pelanggan
- id (TEXT, PRIMARY KEY)
- name (TEXT)
- odp_id (TEXT, FOREIGN KEY)
- pppoe (TEXT)
- whatsapp (TEXT)
- coordinates (TEXT)
- status (TEXT) - 'online', 'offline'
- created_at (DATETIME)

#### Table: mikrotik_config
- id (INTEGER, PRIMARY KEY)
- host (TEXT)
- user (TEXT)
- password (TEXT)
- port (INTEGER)

## Konfigurasi Mikrotik

1. Enable Mikrotik API service:
```
/ip service enable api
```

2. Pastikan firewall tidak memblokir port 8728

3. Update konfigurasi melalui API atau database

## Development

### Structure
```
backend/
├── main.go           # Main server & routes
├── database.go       # Database setup & models
├── handlers.go       # CRUD handlers
├── mikrotik.go       # Mikrotik API integration
├── utilities.go      # Export/Import/Stats
├── go.mod            # Dependencies
└── ftth.db           # SQLite database (auto-generated)
```

### Dependencies
- github.com/gorilla/mux - HTTP router
- github.com/rs/cors - CORS middleware
- modernc.org/sqlite - Pure Go SQLite driver (no CGO required)
- github.com/google/uuid - UUID generator

## Troubleshooting

### Port sudah digunakan
Ubah port di `main.go`:
```go
log.Fatal(http.ListenAndServe(":8080", handler))
```

### Database locked
Stop semua instance backend yang berjalan

### Mikrotik connection failed
- Periksa IP address dan credentials
- Pastikan API service enabled
- Cek firewall rules

## License

Open Source - Free to use and modify
