# Testing Monitoring Status Pelanggan

## Cara Kerja System

### 1. Konsep
- **Online**: Username PPPOE pelanggan **ada** di list Active PPPoE Connections Mikrotik
- **Offline**: Username PPPOE pelanggan **tidak ada** di list Active PPPoE Connections

### 2. Proses Sync
1. Setiap **2 menit** system akan:
   - Connect ke Mikrotik Router via API
   - Ambil list `/ppp/active/print` 
   - Reset semua pelanggan jadi **OFFLINE**
   - Loop active connections, match dengan field PPPOE pelanggan
   - Update status jadi **ONLINE** jika match (case-insensitive)

### 3. Visual Indicator
- Status di table pelanggan:
  - 🟢 **ONLINE** - Badge hijau dengan icon berkedip
  - 🔴 **OFFLINE** - Badge merah
- Status di peta:
  - Marker pelanggan berubah warna sesuai status

## Cara Testing

### Setup Awal

1. **Enable Mikrotik API**
   ```
   /ip service enable api
   ```

2. **Konfigurasi di UI**
   - Menu "Pengaturan"
   - Isi form Mikrotik config
   - Klik "Simpan Konfigurasi"
   - Klik "Test Koneksi" → harus success

### Test Scenario

#### Test 1: Pelanggan Tanpa PPPOE
```
Status: OFFLINE (permanent)
Reason: Field PPPOE kosong, tidak ada yang di-match
```

#### Test 2: Pelanggan dengan PPPOE (User Belum Login)
```
Contoh:
- Nama: Egi Setiawan Sempu
- PPPOE: egisetiawansempu
- Status: OFFLINE

Mikrotik Active Connections: (kosong/tidak ada egisetiawansempu)
Result: Status tetap OFFLINE
```

#### Test 3: Pelanggan Login ke Mikrotik
```
Langkah:
1. User "egisetiawansempu" login ke PPPoE
2. Tunggu max 2 menit (atau klik "Sync Sekarang")
3. System ambil active connections dari Mikrotik
4. Match "egisetiawansempu" dengan database
5. Status berubah jadi ONLINE

Di table pelanggan:
✓ Badge hijau "ONLINE" dengan icon berkedip
✓ Marker di peta berubah hijau
```

#### Test 4: Pelanggan Logout/Disconnect
```
Langkah:
1. User "egisetiawansempu" disconnect dari PPPoE
2. Tunggu max 2 menit (atau klik "Sync Sekarang")
3. System ambil active connections (tidak ada egisetiawansempu)
4. Status berubah jadi OFFLINE

Di table pelanggan:
✗ Badge merah "OFFLINE"
✗ Marker di peta berubah warna default
```

## Monitoring

### Interval Sync
- **Auto**: Setiap 2 menit
- **Manual**: Tombol "Sync Sekarang"

### Status Display (Menu Pengaturan)
```
Status Koneksi: ✓ Connected
Auto-sync: Setiap 2 menit
Last Sync: 14:30:45
```

### Console Log
Buka Browser Console (F12):
```
Status updated: 5 active connections, 3 customers online
```

- `active_connections`: Total PPPoE yang login di Mikrotik
- `customers_online`: Pelanggan di database yang match dengan active connections

## Troubleshooting

### Status tidak update
1. Check Mikrotik config di menu Pengaturan
2. Klik "Test Koneksi" - harus success
3. Check field PPPOE pelanggan sudah terisi
4. Check username PPPOE di Mikrotik sama persis (case-insensitive)
5. Cek console log browser untuk error

### Status selalu OFFLINE
1. Pastikan pelanggan punya field PPPOE yang terisi
2. Pastikan username PPPOE login di Mikrotik
3. Check typo di username PPPOE (system case-insensitive)
4. Manual sync dengan tombol "Sync Sekarang"

### Mikrotik Connection Failed
1. Check IP address dan port (default 8728)
2. Check username/password
3. Pastikan Mikrotik API enabled
4. Check firewall tidak block port 8728
5. Test dengan Winbox/Telnet dulu

## Example Data

### Pelanggan di Database
```json
{
  "id": "uuid-123",
  "name": "Egi Setiawan Sempu",
  "pppoe": "egisetiawansempu",
  "status": "offline"  // akan update otomatis
}
```

### Active Connections dari Mikrotik
```json
[
  {
    "name": "egisetiawansempu",
    "address": "10.10.10.2",
    "uptime": "1h30m",
    "caller_id": "00:11:22:33:44:55"
  },
  {
    "name": "usertest",
    "address": "10.10.10.3",
    "uptime": "45m",
    "caller_id": "AA:BB:CC:DD:EE:FF"
  }
]
```

### Matching Logic
```
1. Loop active connections
2. Ambil field "name" (username PPPOE)
3. Query database: WHERE LOWER(pppoe) = LOWER('egisetiawansempu')
4. Jika match: UPDATE status = 'online'
5. Jika tidak match: tetap offline
```

## Performance

- **Interval**: 2 menit (120 detik)
- **Connection Time**: ~1-2 detik per sync
- **Database Update**: Batch update (reset all → update online)
- **UI Refresh**: Auto refresh table dan map setelah sync

## Security Note

⚠️ **PENTING**: Simpan credentials Mikrotik dengan aman
- Password disimpan di database backend
- Tidak ditampilkan di UI (masked)
- Gunakan user Mikrotik dengan permission minimal (read-only untuk PPP)
