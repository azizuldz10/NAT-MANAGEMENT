# Panduan Instalasi Backend Golang

## Prerequisites

1. **Golang** (versi 1.21 atau lebih baru)
   - Download: https://golang.org/dl/
   - Install dan pastikan `go` command tersedia di terminal

**Note:** GCC/CGO **TIDAK** diperlukan karena menggunakan pure Go SQLite driver (modernc.org/sqlite)

## Langkah Instalasi

### 1. Install Dependencies

Buka terminal/command prompt di folder `backend`:

```bash
cd C:\OPREKV1\NAT\nat-management-appV3.1\NAT4.2\mapping-ftth\backend
go mod download
```

### 2. Build Backend (Opsional - sudah ada ftth-backend.exe)

```bash
go build -o ftth-backend.exe
```

**Note:** File `ftth-backend.exe` sudah tersedia, tidak perlu build ulang kecuali ada perubahan code.

### 3. Jalankan Backend

```bash
# Windows
ftth-backend.exe

# Atau langsung run tanpa build
go run .
```

Backend akan berjalan di: **http://localhost:8080**

### 4. Buka Frontend

Buka file `index.html` di browser, atau gunakan local server:

```bash
# Jika punya Python
python -m http.server 8000

# Atau menggunakan VS Code Live Server extension
```

Akses di: http://localhost:8000

## Konfigurasi Mikrotik

1. **Enable Mikrotik API**
   
   Login ke Mikrotik via Winbox/SSH, jalankan:
   ```
   /ip service enable api
   ```

2. **Update Config**
   
   Setelah backend running, bisa update config via API:
   ```bash
   curl -X PUT http://localhost:8080/api/mikrotik/config \
     -H "Content-Type: application/json" \
     -d '{"host":"192.168.88.1","user":"admin","password":"yourpass","port":8728}'
   ```

   Atau bisa langsung edit database `ftth.db` table `mikrotik_config`.

## Testing API

### Test Connection Mikrotik
```bash
curl http://localhost:8080/api/mikrotik/test
```

### Get Statistics
```bash
curl http://localhost:8080/api/stats
```

### Get All Routers
```bash
curl http://localhost:8080/api/routers
```

### Create Router (Server)
```bash
curl -X POST http://localhost:8080/api/routers \
  -H "Content-Type: application/json" \
  -d '{
    "name":"Server Jakarta",
    "type":"server",
    "parent_id":null,
    "coordinates":"-6.200000,106.816666"
  }'
```

## Troubleshooting

### Error: port 8080 sudah digunakan
- Ubah port di `main.go`:
  ```go
  log.Fatal(http.ListenAndServe(":8081", handler))
  ```

### Error: database locked
- Stop semua instance backend yang running
- Delete file `ftth.db` dan restart (data akan hilang)

### Frontend tidak bisa connect ke backend
- Pastikan backend running di port 8080
- Check browser console untuk error CORS
- Pastikan frontend mengakses dari http://localhost bukan file:///

## Struktur File

```
backend/
├── main.go           # Entry point & routing
├── database.go       # Database & models
├── handlers.go       # CRUD handlers
├── mikrotik.go       # Mikrotik integration
├── utilities.go      # Export/import/stats
├── go.mod            # Dependencies
├── go.sum            # Checksum dependencies
├── ftth.db           # SQLite database (auto-created)
└── README.md         # API documentation
```

## Production Deployment

### Build untuk Windows
```bash
GOOS=windows GOARCH=amd64 go build -o ftth-backend-windows.exe
```

### Build untuk Linux
```bash
GOOS=linux GOARCH=amd64 go build -o ftth-backend-linux
```

### Run as Service (Windows)

Gunakan NSSM (Non-Sucking Service Manager):
1. Download NSSM: https://nssm.cc/download
2. Install service:
   ```cmd
   nssm install FTTHBackend "C:\path\to\ftth-backend.exe"
   nssm start FTTHBackend
   ```

### Run as Service (Linux)

Create systemd service `/etc/systemd/system/ftth-backend.service`:
```ini
[Unit]
Description=FTTH Backend Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/ftth-backend
ExecStart=/opt/ftth-backend/ftth-backend
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable dan start:
```bash
sudo systemctl enable ftth-backend
sudo systemctl start ftth-backend
```

## Backup & Restore

### Backup
```bash
# Copy database file
cp ftth.db ftth-backup-$(date +%Y%m%d).db

# Atau via API export
curl http://localhost:8080/api/export > backup.json
```

### Restore
```bash
# Restore dari database
cp ftth-backup-20251023.db ftth.db

# Atau via API import
curl -X POST http://localhost:8080/api/import \
  -H "Content-Type: application/json" \
  -d @backup.json
```

## Support

Jika ada masalah, check:
1. Logs di console backend
2. Browser console untuk frontend errors
3. Network tab untuk API request/response
