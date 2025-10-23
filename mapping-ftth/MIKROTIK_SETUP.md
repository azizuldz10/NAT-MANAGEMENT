# Setup Mikrotik Integration

Panduan lengkap untuk mengaktifkan integrasi Mikrotik dengan sistem FTTH Management.

## Prerequisites

1. **Mikrotik RouterOS** dengan PPPoE Server aktif
2. **PHP 7.0+** installed di server web
3. **Port 8728** terbuka di firewall Mikrotik
4. User Mikrotik dengan permission API

## Step 1: Enable Mikrotik API

### Via Winbox/WebFig:
1. Masuk ke Mikrotik Router
2. Buka menu **IP → Services**
3. Enable service **"api"**
4. Set port ke **8728** (default)
5. Klik **Apply**

### Via Terminal:
```
/ip service enable api
/ip service set api port=8728
```

## Step 2: Configure Firewall (Jika Perlu)

Jika Mikrotik memblokir koneksi API, tambahkan firewall rule:

```
/ip firewall filter add chain=input protocol=tcp dst-port=8728 action=accept comment="Allow API Access"
```

## Step 3: Create API User (Recommended)

Untuk keamanan, buat user khusus untuk API:

```
/user add name=ftth-api group=full password=YourSecurePassword
```

Atau via Winbox:
1. Buka **System → Users**
2. Klik **Add New** (+)
3. Name: `ftth-api`
4. Group: `full` (atau buat group custom dengan permission terbatas)
5. Password: Isi password yang kuat
6. Klik **OK**

## Step 4: Configure api/config.php

Edit file `api/config.php` di folder project:

```php
<?php
$MIKROTIK_HOST = '192.168.88.1';  // IP Mikrotik Anda
$MIKROTIK_USER = 'ftth-api';      // Username yang dibuat
$MIKROTIK_PASS = 'YourSecurePassword';  // Password user
$MIKROTIK_PORT = 8728;             // Port API (default 8728)
?>
```

**PENTING:**
- Ganti `192.168.88.1` dengan IP Mikrotik Anda
- Gunakan user dengan permission yang tepat
- Password HARUS diisi (tidak boleh kosong untuk production)

## Step 5: Test Connection

1. Buka aplikasi FTTH Management
2. Navigasi ke menu **Pengaturan**
3. Scroll ke section **Mikrotik Integration**
4. Klik button **"Test Koneksi"**
5. Jika berhasil, akan muncul notifikasi: "✓ Koneksi Mikrotik berhasil!"

## Step 6: Verify PPPoE Data Match

Pastikan username **PPPOE** di data pelanggan **SAMA PERSIS** dengan username di Mikrotik:

**Contoh:**
- Di Mikrotik PPPoE Server: `user001@pppoe`
- Di data pelanggan: `user001@pppoe` ✅

**Case-Insensitive:** Sistem sudah handle uppercase/lowercase, jadi `User001@PPPOE` akan match dengan `user001@pppoe`

## Troubleshooting

### Error: "Connection failed"

**Kemungkinan Penyebab:**
1. IP Mikrotik salah
2. Port 8728 tidak terbuka
3. Firewall memblokir koneksi
4. API service belum di-enable

**Solusi:**
```bash
# Test ping ke Mikrotik
ping 192.168.88.1

# Test port 8728
telnet 192.168.88.1 8728

# Via Mikrotik terminal
/ip service print
```

### Error: "Login failed"

**Kemungkinan Penyebab:**
1. Username/password salah
2. User tidak punya permission API

**Solusi:**
- Pastikan username dan password benar di `config.php`
- Pastikan user punya group `full` atau permission API

### Status Tidak Update

**Kemungkinan Penyebab:**
1. PPPOE username tidak match
2. Auto-refresh di-pause
3. Pelanggan tidak punya data PPPOE

**Solusi:**
- Cek kolom PPPOE di data pelanggan sudah diisi
- Pastikan username sama dengan yang di Mikrotik
- Klik "Resume Sync" jika di-pause
- Klik "Sync Sekarang" untuk manual sync

## How It Works

### Auto-Sync Process:
1. Setiap **30 detik**, sistem fetch data dari Mikrotik API
2. Endpoint: `/ppp/active/print`
3. Data active connections dibandingkan dengan database pelanggan
4. Status pelanggan di-update (online/offline)
5. Map markers dan table di-refresh otomatis

### Visual Indicators:
- 🟢 **Hijau + Pulse Animation** = Online
- ⚫ **Abu-abu** = Offline
- **Badge di Table** = Status realtime
- **Last Seen** = Timestamp terakhir online

## Security Best Practices

1. **Jangan gunakan user admin** untuk API
2. **Gunakan password yang kuat** (min 12 karakter, campuran huruf/angka/simbol)
3. **Batasi IP access** jika memungkinkan:
   ```
   /ip firewall filter add chain=input src-address=192.168.1.100 protocol=tcp dst-port=8728 action=accept
   /ip firewall filter add chain=input protocol=tcp dst-port=8728 action=drop
   ```
4. **Gunakan HTTPS** untuk web interface
5. **Backup config.php** dan jangan commit ke public repository

## Performance Notes

- **Refresh Interval:** 30 detik (bisa diubah di `app.js`)
- **Network Impact:** Minimal (~1KB per request)
- **Server Load:** Ringan (hanya parsing JSON)
- **Mikrotik CPU:** < 1% per request

## API Endpoints

### Get Active Connections
```
GET api/mikrotik-api.php?action=status
```

Response:
```json
{
  "success": true,
  "data": [
    {
      "name": "user001@pppoe",
      "address": "10.10.10.2",
      "uptime": "1h2m3s",
      "caller-id": "00:11:22:33:44:55",
      "service": "pppoe"
    }
  ]
}
```

### Test Connection
```
GET api/mikrotik-api.php?action=test
```

Response:
```json
{
  "success": true,
  "message": "Connection successful"
}
```

## Advanced Configuration

### Change Refresh Interval

Edit `app.js`:
```javascript
this.refreshIntervalSeconds = 30;  // Ubah ke nilai yang diinginkan (dalam detik)
```

### Custom User Permission

Buat group dengan permission minimal:
```
/user group add name=ftth-readonly policy=api,read
/user add name=ftth-api group=ftth-readonly password=SecurePass
```

## Support

Jika masih ada masalah:
1. Cek console browser (F12) untuk error JavaScript
2. Cek file PHP error log
3. Cek Mikrotik log: `/log print where topics~"api"`
4. Pastikan waktu server dan Mikrotik sudah sinkron (NTP)

---

**Selamat!** Sistem monitoring Mikrotik sekarang aktif! 🎉
