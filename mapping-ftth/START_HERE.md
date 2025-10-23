# Quick Start Guide

## Cara Menjalankan Aplikasi

### 1. Start Backend Server

**Option A - Double click:**
```
backend/start.bat
```

**Option B - Command line:**
```bash
cd backend
ftth-backend.exe
```

Backend akan running di: **http://localhost:8080**

### 2. Buka Frontend

**Option A - Double click:**
Langsung buka file `index.html` di browser

**Option B - Local Server (Recommended):**

Menggunakan Python:
```bash
python -m http.server 8000
```
Akses: http://localhost:8000

Atau menggunakan VS Code "Live Server" extension

### 3. Selesai!

- Frontend akan otomatis connect ke backend di port 8080
- Mulai tambahkan data Server → OLT → ODC → ODP → Pelanggan
- Semua data tersimpan di database `backend/ftth.db`

## Konfigurasi Mikrotik (Opsional)

Untuk enable monitoring status pelanggan:

1. **Enable Mikrotik API:**
   ```
   /ip service enable api
   ```

2. **Update Config di Settings:**
   - Buka menu "Pengaturan" di aplikasi
   - Klik "Test Koneksi" untuk verifikasi
   - Status pelanggan akan auto-update setiap 30 detik

## Troubleshooting

### Backend tidak jalan
- Pastikan port 8080 tidak digunakan aplikasi lain
- Check apakah `ftth-backend.exe` ada di folder backend
- Jika tidak ada, run: `go build -o ftth-backend.exe`

### Frontend tidak connect
- Pastikan backend sudah running
- Buka browser console (F12) untuk cek error
- Pastikan akses dari http://localhost bukan file:///

### Data tidak muncul
- Refresh halaman (F5)
- Check backend console untuk error
- Database ada di `backend/ftth.db`

## Fitur Utama

✅ **Management Hierarchical**
- Server → OLT → ODC → ODP → Pelanggan
- CRUD lengkap untuk semua entity
- Parent-child relationship otomatis

✅ **Peta Interaktif**
- Visualisasi semua node di peta
- Auto-connect parent-child dengan garis
- Click marker untuk detail
- Pick koordinat dari peta

✅ **Import Excel**
- Template Excel untuk bulk import pelanggan
- Download template dari menu Pelanggan
- Validasi otomatis

✅ **Mikrotik Integration**
- Monitor status online/offline pelanggan
- Auto-sync setiap 30 detik
- Test koneksi dari menu Settings

✅ **Export/Import**
- Backup data ke JSON
- Restore dari backup
- Transfer data antar sistem

## API Documentation

Lihat file `backend/README.md` untuk dokumentasi lengkap API endpoints.

## Need Help?

1. Check dokumentasi:
   - [README.md](README.md) - Overview sistem
   - [INSTALL_BACKEND.md](INSTALL_BACKEND.md) - Setup backend
   - [backend/README.md](backend/README.md) - API docs

2. Check logs:
   - Backend console untuk server errors
   - Browser console (F12) untuk frontend errors
