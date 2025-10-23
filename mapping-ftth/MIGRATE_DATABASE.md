# Database Migration - Tambah Field Profile

## Untuk Database yang Sudah Ada

Jika Anda sudah punya database `ftth.db` yang sudah berisi data pelanggan, Anda perlu menambahkan kolom `profile` secara manual.

### Option 1: Menggunakan DB Browser for SQLite

1. Download **DB Browser for SQLite**: https://sqlitebrowser.org/dl/
2. Install dan buka `ftth.db` dengan DB Browser
3. Menu **Execute SQL** (Tab SQL)
4. Jalankan query berikut:
   ```sql
   ALTER TABLE pelanggan ADD COLUMN profile TEXT;
   ```
5. Klik **Execute**
6. Save dan close

### Option 2: Menggunakan Go Script

Buat file `migrate.go` di folder backend:

```go
package main

import (
    "database/sql"
    "fmt"
    _ "modernc.org/sqlite"
)

func main() {
    db, err := sql.Open("sqlite", "./ftth.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    _, err = db.Exec(`ALTER TABLE pelanggan ADD COLUMN profile TEXT`)
    if err != nil {
        fmt.Println("Error (might already exist):", err)
    } else {
        fmt.Println("Column 'profile' added successfully!")
    }
}
```

Run:
```bash
go run migrate.go
```

### Option 3: Fresh Start (Hapus Database)

**WARNING: Ini akan menghapus semua data!**

1. Stop backend
2. Hapus file `ftth.db`
3. Start backend (database baru akan otomatis dibuat dengan schema lengkap)

### Option 4: Export → Fresh → Import

Cara paling aman:

1. **Export data existing**:
   - Buka aplikasi
   - Menu "Pengaturan" → "Export JSON"
   - Save file backup

2. **Hapus database**:
   - Stop backend
   - Delete `ftth.db`

3. **Start backend baru**:
   - Backend akan create database dengan schema baru

4. **Import data**:
   - Menu "Pengaturan" → "Import JSON"
   - Pilih file backup tadi

## Verifikasi

Setelah migration, cek di browser console:
```javascript
// Ambil 1 pelanggan
fetch('http://localhost:8080/api/pelanggan')
  .then(r => r.json())
  .then(d => console.log(d.data[0]))
```

Harus ada field `profile`:
```json
{
  "id": "...",
  "name": "...",
  "pppoe": "...",
  "profile": "",  // ← Field baru
  "whatsapp": "...",
  ...
}
```

## Fitur Profile

Setelah migration berhasil:

### 1. Form Pelanggan
Akan ada field baru "Profile" antara PPPOE dan WhatsApp

### 2. Table Pelanggan
Kolom baru "Profile" untuk menampilkan paket/speed

### 3. Map Popup
Info profile akan muncul di popup marker pelanggan:
```
Status: Online
PPPOE: username
Profile: 10Mbps    ← Field baru
WhatsApp: 628xxx
```

### Contoh Pengisian Profile:
- `10Mbps`
- `20Mbps - Premium`
- `Paket Basic`
- `Dedicated 50M`
- dll (free text)

## Troubleshooting

### Error: duplicate column name
Berarti kolom sudah ada, skip migration.

### Field profile tidak muncul di UI
1. Hard refresh browser (Ctrl+Shift+R)
2. Clear browser cache
3. Check backend sudah restart dengan build terbaru

### Data profile tidak tersimpan
1. Check backend console untuk error
2. Pastikan migration SQL sudah jalan
3. Test dengan curl:
   ```bash
   curl http://localhost:8080/api/pelanggan
   ```
