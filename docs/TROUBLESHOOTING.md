# 🔧 Troubleshooting Guide - NAT Management System

## Error: Connection Timeout ke MikroTik Router

### Error Message yang Umum
```
dial tcp <IP>:<PORT>: connectex: A connection attempt failed because
the connected party did not properly respond after a period of time...
```

### Penyebab & Solusi

#### 1. 🌐 Router Tidak Terkonfigurasi

**Gejala:**
- Error: "router not configured"
- List router kosong di dashboard
- File `config/routers.json` kosong

**Solusi:**
```bash
# Cek apakah ada router terdaftar
# Via UI: Login → Router Management → Add Router

# Atau via Database: Periksa tabel routers di PostgreSQL
```

**Langkah-langkah:**
1. Login sebagai Administrator
2. Buka halaman "Router Management"
3. Klik tombol "Add Router"
4. Isi form dengan informasi router:
   - Name: (contoh: JAKARTA-01)
   - Host: IP address router (contoh: 192.168.1.1)
   - Port: **8728** (default RouterOS API)
   - Username: admin (atau user dengan API access)
   - Password: password router
   - Tunnel Endpoint: (contoh: 172.22.28.5:80)
   - Public ONT URL: (contoh: http://tunnel3.ebilling.id:19701)
5. Klik "Test Connection" untuk verifikasi
6. Klik "Save Router" jika test berhasil

---

#### 2. 🔌 Port Yang Salah

**Gejala:**
- Connection timeout
- TCP connection failed
- Port non-standard (misal: 19699, 19701, dll)

**Solusi:**

**Port Standard MikroTik:**
- `8728` - RouterOS API (default, **gunakan ini!**)
- `8729` - RouterOS API-SSL
- `80` - HTTP/WebFig
- `8291` - Winbox
- `22` - SSH

**Cara Cek Port:**
1. Buka Winbox/WebFig
2. IP → Services
3. Cari service "api" atau "api-ssl"
4. Catat port yang aktif
5. Update konfigurasi router di aplikasi

**Catatan:** Port 19699/19701 biasanya untuk HTTP tunnel, **BUKAN untuk API!**

---

#### 3. 🔥 Firewall Blocking Connection

**Gejala:**
- Connection timeout
- Tidak bisa ping router
- TCP connection failed

**Solusi:**

**A. Cek Firewall Windows:**
```powershell
# Test koneksi TCP ke router
Test-NetConnection -ComputerName 192.168.1.1 -Port 8728

# Atau menggunakan telnet
telnet 192.168.1.1 8728
```

**B. Cek Firewall MikroTik:**
```routeros
# Via Winbox/Terminal
/ip firewall filter print where chain=input

# Pastikan ada rule yang allow API:
/ip firewall filter add chain=input protocol=tcp dst-port=8728 action=accept comment="Allow API Access"
```

**C. Tambah Firewall Rule (jika perlu):**
```routeros
# Allow API dari IP tertentu
/ip firewall filter add chain=input src-address=192.168.1.0/24 protocol=tcp dst-port=8728 action=accept

# Atau allow dari IP server aplikasi
/ip firewall filter add chain=input src-address=YOUR_SERVER_IP protocol=tcp dst-port=8728 action=accept
```

---

#### 4. ⚙️ API Service Tidak Aktif

**Gejala:**
- TCP connection berhasil
- RouterOS API authentication failed
- Error: "connection refused" atau "invalid user name or password"

**Solusi:**

**Aktifkan API Service:**
1. Buka Winbox
2. IP → Services
3. Cari service "api"
4. Klik 2x pada service "api"
5. Centang "Enabled"
6. Pastikan port = 8728
7. Klik OK

**Via Terminal:**
```routeros
/ip service enable api
/ip service set api port=8728
```

---

#### 5. 🔑 Username/Password Salah

**Gejala:**
- TCP connection berhasil
- Error: "invalid user name or password"
- Error: "cannot log in"

**Solusi:**

**A. Test Login Manual:**
```bash
# Test menggunakan SSH
ssh admin@192.168.1.1

# Test menggunakan Winbox
# Masukkan IP, username, password
```

**B. Cek User Permissions:**
```routeros
# Via Terminal/Winbox
/user print

# Pastikan user memiliki policy "api"
/user group print

# Add API policy jika belum ada
/user set admin group=full
```

**C. Reset Password (jika lupa):**
```routeros
# Via console/serial
/user set admin password=newpassword123
```

---

#### 6. 🌍 Network Tidak Bisa Reach Router

**Gejala:**
- Cannot resolve hostname
- Connection timeout
- Ping failed

**Solusi:**

**A. Test Koneksi Dasar:**
```powershell
# Test ping
ping 192.168.1.1

# Test DNS (jika pakai hostname)
nslookup router.example.com

# Test traceroute
tracert 192.168.1.1
```

**B. Cek Routing:**
```powershell
# Windows
route print

# Tambah static route jika perlu
route add 192.168.1.0 mask 255.255.255.0 192.168.0.1
```

**C. Cek Network Interface:**
```powershell
ipconfig /all
# Pastikan IP address dalam satu network dengan router
```

---

## 🛠️ Diagnostic Tool

### Menggunakan Router Diagnostic Tool

Kami menyediakan tool khusus untuk diagnosis koneksi:

**Compile tool:**
```bash
cd tools
go build -o router-diagnostic.exe router-diagnostic.go
```

**Jalankan diagnostic:**
```bash
router-diagnostic.exe <host> <port> <username> <password>

# Contoh:
router-diagnostic.exe 160.19.144.8 8728 admin password123
router-diagnostic.exe 192.168.1.1 8728 admin ""
```

**Output Tool:**
- ✅ DNS Resolution
- ✅ TCP Connection (multiple timeouts)
- ✅ RouterOS API Authentication
- ✅ Get Router Identity
- ✅ Get System Resources
- 📊 Diagnostic Summary
- 💡 Suggestions untuk setiap failed test

---

## 📋 Checklist Troubleshooting

Ikuti checklist ini secara berurutan:

- [ ] **Router online?**
  ```bash
  ping <router-ip>
  ```

- [ ] **Port correct? (8728 bukan 19699)**
  - Cek di Winbox: IP → Services → api

- [ ] **API service enabled?**
  - Cek di Winbox: IP → Services → api (harus enabled)

- [ ] **Credentials correct?**
  - Test login via Winbox/SSH

- [ ] **Firewall tidak blocking?**
  - Cek Windows Firewall
  - Cek MikroTik Firewall Filter

- [ ] **User memiliki API permission?**
  ```routeros
  /user print detail
  # Cari policy, harus ada "api"
  ```

- [ ] **Network bisa reach router?**
  ```bash
  tracert <router-ip>
  ```

- [ ] **Jalankan diagnostic tool**
  ```bash
  router-diagnostic.exe <host> 8728 <user> <pass>
  ```

---

## 🔍 Advanced Debugging

### Enable Debug Logging

**Edit file `.env`:**
```env
DEBUG=true
LOG_LEVEL=debug
```

**Restart aplikasi:**
```bash
# Stop aplikasi (Ctrl+C)
# Start lagi
./nat-supabase-logs.exe
```

**Check log output:**
Log akan menunjukkan detail setiap connection attempt:
```
🔄 Attempt 1/3: Connecting to ROUTER-01 at 192.168.1.1:8728 (timeout: 15s)
⚠️  Attempt 1: TCP connection to ROUTER-01 failed: dial tcp 192.168.1.1:8728: i/o timeout
⏳ Waiting 2s before retry...
🔄 Attempt 2/3: Connecting to ROUTER-01 at 192.168.1.1:8728 (timeout: 30s)
✅ TCP connection successful on attempt 2
✅ Successfully connected to ROUTER-01 on attempt 2
```

### Check Database Configuration

```sql
-- Connect ke PostgreSQL
SELECT * FROM routers WHERE enabled = true;

-- Cek router configuration
SELECT id, name, host, port, username, enabled, created_at
FROM routers
ORDER BY created_at DESC;
```

### Network Capture (Advanced)

Gunakan Wireshark untuk capture traffic:
1. Filter: `tcp.port == 8728`
2. Start capture saat aplikasi connect
3. Cek apakah ada TCP SYN/ACK
4. Cek apakah ada RouterOS API handshake

---

## ❓ FAQ

### Q: Kenapa port 19699/19701 tidak work?

**A:** Port tersebut adalah port HTTP tunnel untuk akses ONT, **BUKAN port RouterOS API!**
- Port API MikroTik yang benar adalah **8728**
- Update konfigurasi router di aplikasi dengan port yang benar

### Q: Error "router not configured" padahal sudah add router?

**A:** Aplikasi menggunakan PostgreSQL database. Cek:
1. Router tersimpan di database (bukan hanya file JSON)
2. Router dalam status `enabled = true`
3. Reload aplikasi setelah add router

### Q: Connection timeout tapi bisa ping?

**A:** Ping (ICMP) berbeda dengan TCP connection. Cek:
1. Port API (8728) tidak di-block firewall
2. Service API enabled di MikroTik
3. Test dengan: `telnet <router-ip> 8728`

### Q: Berhasil connect via Winbox tapi aplikasi gagal?

**A:** Winbox menggunakan port 8291, aplikasi menggunakan API port 8728. Cek:
1. API service enabled
2. API port tidak di-block
3. User memiliki API permission

### Q: Koneksi lambat/sering timeout?

**A:** Aplikasi sudah di-update dengan:
- Timeout 15-45 detik (dari 5 detik)
- 3x retry dengan exponential backoff
- Detailed logging untuk troubleshooting

Jika masih lambat:
1. Cek bandwidth network
2. Cek CPU load di MikroTik
3. Reduce jumlah concurrent connections

---

## 📞 Bantuan Lebih Lanjut

Jika masih mengalami masalah:

1. **Jalankan diagnostic tool dan simpan output:**
   ```bash
   router-diagnostic.exe <host> 8728 <user> <pass> > diagnostic-output.txt
   ```

2. **Collect log aplikasi:**
   - Enable DEBUG mode
   - Reproduce error
   - Copy log output

3. **Dokumentasikan:**
   - Error message lengkap
   - Router configuration (hide password)
   - Network topology
   - Diagnostic tool output

4. **Create issue di GitHub dengan informasi di atas**

---

## ✅ Resolved Issues

Setelah mengikuti guide ini, masalah yang umum resolved:

✅ Connection timeout → Update port ke 8728
✅ API auth failed → Enable API service
✅ Router not configured → Add router via UI
✅ Firewall blocking → Add firewall rules
✅ Slow connection → Retry logic sudah improved

---

**Last Updated:** 2025-10-16
**Version:** 4.1
