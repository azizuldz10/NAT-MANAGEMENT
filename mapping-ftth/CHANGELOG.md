# Changelog

## [v4.3] - 2025-10-23

### Added
- ✨ **Field Profile** untuk pelanggan
  - Form input di modal tambah/edit pelanggan
  - Kolom baru di table pelanggan
  - Tampil di popup marker map
  - Support di export/import JSON

### Changed
- 🔄 **Auto-sync interval** dari 30 detik ke 2 menit
- 💾 **Database**: Tambah kolom `profile` di table pelanggan
- 🎨 **Status badge**: Animasi pulse untuk status online

### Fixed
- 🐛 Fix Mikrotik API GetActivePPPoE - case-sensitive bug
- 🐛 Fix status sync - semua pelanggan reset offline dulu
- 🐛 COALESCE profile field untuk handle NULL values

### Technical
- Backend: Pure Go SQLite driver (no CGO required)
- Migration script: `migrate.go` untuk update database existing
- Debug logging untuk troubleshooting Mikrotik sync

---

## [v4.2] - 2025-10-23

### Added
- 🔧 **UI Konfigurasi Mikrotik** di menu Pengaturan
  - Form input IP, Port, Username, Password
  - Save/Load config dari database
  - Test koneksi button
- ⏱️ **Last Sync Time** indicator
- 📊 **Enhanced logging** untuk debugging

### Changed
- 🔄 Monitoring system dengan auto-refresh
- 💾 Mikrotik config disimpan di database (tidak di file)
- 🎨 Badge styling dengan border dan better contrast

---

## [v4.0] - 2025-10-23

### Added
- 🚀 **Backend Golang** dengan RESTful API
- 💾 **SQLite Database** untuk persistent storage
- 🔌 **Mikrotik Integration** via RouterOS API
  - Auto-sync status pelanggan online/offline
  - Test koneksi
  - Monitor active PPPoE connections
- 📡 **Real-time Status Monitoring**

### Changed
- ❌ Removed LocalStorage (browser) storage
- ✅ Replaced with SQLite database
- 🔄 All CRUD operations via API

### Endpoints
```
GET    /api/routers
POST   /api/routers
PUT    /api/routers/:id
DELETE /api/routers/:id

GET    /api/pelanggan
POST   /api/pelanggan
PUT    /api/pelanggan/:id
DELETE /api/pelanggan/:id

GET    /api/mikrotik/status
GET    /api/mikrotik/test
GET    /api/mikrotik/config
PUT    /api/mikrotik/config

GET    /api/stats
GET    /api/export
POST   /api/import
```

---

## [v3.1] - Previous Version

### Features
- Frontend only (Vanilla JS)
- LocalStorage for data persistence
- Leaflet.js for mapping
- PHP Mikrotik API integration
- Parent-child hierarchy management
- Excel import for pelanggan
- JSON export/import

---

## Field Structure

### Pelanggan Object
```json
{
  "id": "uuid",
  "name": "Nama Pelanggan",
  "odp_id": "parent-odp-uuid",
  "pppoe": "username",
  "profile": "10Mbps",        // ← NEW in v4.3
  "whatsapp": "628xxx",
  "coordinates": "-6.xxx,106.xxx",
  "status": "online",
  "created_at": "2025-10-23T..."
}
```

### Router Object
```json
{
  "id": "uuid",
  "name": "Router Name",
  "type": "server|olt|odc|odp",
  "parent_id": "parent-uuid",
  "coordinates": "-6.xxx,106.xxx",
  "created_at": "2025-10-23T..."
}
```

## Migration Guide

### From v4.2 to v4.3
```bash
cd backend
go run migrate.go    # Tambah kolom profile
```

Atau manual SQL:
```sql
ALTER TABLE pelanggan ADD COLUMN profile TEXT;
```

### From v3.x to v4.x
1. Export data dari v3.x (JSON)
2. Install Golang
3. Setup backend v4.x
4. Import data ke v4.x

## Breaking Changes

### v4.0
- ⚠️ PHP backend diganti Golang
- ⚠️ LocalStorage tidak digunakan lagi
- ⚠️ Format export JSON berubah struktur

### v4.3
- ✅ Backward compatible dengan v4.0-4.2
- ✅ Existing data tetap bisa diimport
- ✅ Profile field optional (boleh kosong)
