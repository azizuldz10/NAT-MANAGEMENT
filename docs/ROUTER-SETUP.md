# 📖 Router Setup Guide - NAT Management System

## Panduan Lengkap Setup Router MikroTik

### 📌 Prerequisites

Sebelum mulai, pastikan Anda memiliki:

- [ ] Akses Administrator ke NAT Management System
- [ ] Akses Winbox/WebFig ke Router MikroTik
- [ ] Username & Password router dengan full access
- [ ] IP address router yang bisa direach dari server
- [ ] Koneksi network antara server dan router

---

## 🚀 Quick Start

### Step 1: Persiapan Router MikroTik

#### 1.1 Enable API Service

**Via Winbox:**
1. Connect ke router menggunakan Winbox
2. Buka menu **IP → Services**
3. Klik 2x pada service **"api"**
4. Centang **"Enabled"**
5. Pastikan Port = **8728** (default)
6. Klik **OK**

**Via Terminal:**
```routeros
/ip service
print
set api port=8728
enable api
```

**Verifikasi:**
```routeros
/ip service print
# Output harus menunjukkan api service enabled
```

#### 1.2 Buat User untuk API Access (Opsional)

Untuk keamanan, sebaiknya buat user khusus untuk aplikasi:

```routeros
# Buat group dengan API permission
/user group add name=nat-api-users policy=api,read,write,policy,test

# Buat user baru
/user add name=nat-api password=YourStrongPassword123 group=nat-api-users

# Verifikasi
/user print detail where name=nat-api
```

#### 1.3 Setup Firewall Rule (Jika Diperlukan)

Jika ada firewall di router, allow API access:

```routeros
# Allow API dari IP server aplikasi
/ip firewall filter add chain=input \
    src-address=YOUR_SERVER_IP \
    protocol=tcp dst-port=8728 \
    action=accept \
    comment="NAT Management API Access"

# Atau allow dari subnet
/ip firewall filter add chain=input \
    src-address=192.168.1.0/24 \
    protocol=tcp dst-port=8728 \
    action=accept \
    comment="NAT Management API Access"
```

**⚠️ Security Note:**
- Jangan expose API port ke internet!
- Gunakan VPN jika akses remote
- Gunakan API-SSL (port 8729) untuk produksi

---

### Step 2: Konfigurasi di NAT Management System

#### 2.1 Login ke Aplikasi

1. Buka browser: `http://localhost:8080`
2. Login sebagai **Administrator**
   - Username: `admin`
   - Password: `admin123`

#### 2.2 Tambah Router Baru

1. Klik menu **"Router Management"** di sidebar
2. Klik tombol **"Add Router"**
3. Isi form dengan informasi berikut:

**Form Fields:**

| Field | Deskripsi | Contoh | Required |
|-------|-----------|--------|----------|
| **Router Name** | Nama unik router | `JAKARTA-01` | ✅ |
| **Host/IP Address** | IP address router | `192.168.1.1` | ✅ |
| **Port** | RouterOS API port | `8728` | ✅ |
| **Username** | Username MikroTik | `admin` | ✅ |
| **Password** | Password router | `password123` | ✅ |
| **Tunnel Endpoint** | Internal tunnel IP:port | `172.22.28.5:80` | ✅ |
| **Public ONT URL** | Public URL untuk ONT | `http://tunnel-example.domain.com:19701` | ✅ |
| **Description** | Deskripsi router (opsional) | `Router Cabang Jakarta` | ❌ |
| **Status** | Enabled/Disabled | `Enabled` | ✅ |

**Penjelasan Fields:**

1. **Router Name:**
   - Nama unik untuk identifikasi router
   - Gunakan naming convention: `CABANG-NOURUT`
   - Contoh: `JAKARTA-01`, `BANDUNG-02`, `SURABAYA-03`

2. **Host/IP Address:**
   - IP address router yang bisa direach dari server
   - Bisa menggunakan IP private (LAN) jika server satu network
   - Bisa menggunakan IP public (WAN) jika server berbeda network
   - Contoh: `192.168.1.1`, `10.10.10.1`, `203.194.112.50`

3. **Port:**
   - **WAJIB 8728** (RouterOS API port)
   - Bukan 19699/19701 (itu untuk HTTP tunnel)
   - Bukan 8291 (itu untuk Winbox)
   - Bukan 80 (itu untuk WebFig)

4. **Username & Password:**
   - User MikroTik yang memiliki full access atau minimal API permission
   - Gunakan user `admin` atau user custom yang sudah dibuat

5. **Tunnel Endpoint:**
   - IP internal dan port untuk NAT destination
   - Format: `IP:PORT`
   - Contoh: `172.22.28.5:80`
   - Ini adalah endpoint yang akan di-NAT ke public URL

6. **Public ONT URL:**
   - URL public yang bisa diakses dari internet
   - Format: `http://domain:port` atau `https://domain:port`
   - Contoh: `http://tunnel-example.domain.com:19701`
   - Ini adalah URL yang diakses user untuk remote ONT

#### 2.3 Test Connection

Sebelum save, **WAJIB test connection:**

1. Klik tombol **"Test Connection"**
2. Tunggu proses test (5-15 detik)
3. Perhatikan hasilnya:

**✅ Success:**
```
Connection successful!
Router: ROUTER-EXAMPLE
Version: 6.49.10
Board: RB750Gr3
```

**❌ Failed:**
```
Connection failed: dial tcp 192.168.1.1:8728: i/o timeout
```

**Jika Failed:**
- Cek [Troubleshooting Guide](TROUBLESHOOTING.md)
- Verifikasi IP dan port
- Pastikan API service enabled
- Cek firewall rules

#### 2.4 Save Router Configuration

Jika test connection berhasil:

1. Klik tombol **"Save Router"**
2. Tunggu konfirmasi "Router created successfully"
3. Router akan muncul di list

---

## 🔐 Security Best Practices

### 1. Gunakan User Khusus untuk API

**Jangan gunakan user `admin` untuk aplikasi!**

Buat user khusus dengan permission terbatas:

```routeros
# Group dengan permission minimal
/user group add name=nat-app \
    policy=api,read,write,test

# User untuk aplikasi
/user add name=nat-management \
    password=StrongRandomPassword123! \
    group=nat-app
```

### 2. Restrict API Access by IP

```routeros
# Allow hanya dari IP server
/ip firewall filter add chain=input \
    src-address=!YOUR_SERVER_IP \
    protocol=tcp dst-port=8728 \
    action=drop \
    comment="Block API except from NAT Management Server"
```

### 3. Gunakan API-SSL (Rekomendasi untuk Produksi)

```routeros
# Enable API-SSL
/ip service enable api-ssl
/ip service set api-ssl port=8729

# Generate certificate (jika belum ada)
/certificate add name=api-ssl common-name=router.local
/certificate sign api-ssl

# Set certificate untuk API-SSL
/ip service set api-ssl certificate=api-ssl
```

Update aplikasi configuration untuk gunakan port 8729.

### 4. Rotate Password Secara Berkala

```routeros
# Update password user
/user set nat-management password=NewStrongPassword456!
```

Jangan lupa update di aplikasi juga.

### 5. Monitor API Access

```routeros
# Lihat active API connections
/user active print where via=api

# Log API access (optional)
/system logging add topics=api action=memory
/system logging add topics=api,!debug action=disk
```

---

## 📊 Multiple Routers Setup

### Scenario: Cabang Multi-lokasi

Untuk setup multi-router (cabang berbeda):

**1. Planning:**

| Router | Lokasi | IP Address | Port | Tunnel Endpoint | Public URL |
|--------|--------|------------|------|-----------------|------------|
| JAKARTA-01 | Jakarta | 192.168.1.1 | 8728 | 172.22.28.5:80 | tunnel-branch1.yourdomain.com:19701 |
| BANDUNG-01 | Bandung | 192.168.2.1 | 8728 | 172.22.29.5:80 | tunnel-branch2.yourdomain.com:19702 |
| SURABAYA-01 | Surabaya | 10.10.10.1 | 8728 | 172.22.30.5:80 | tunnel-branch3.yourdomain.com:19703 |

**2. Add semua router satu per satu:**
- Ulangi Step 2 untuk setiap router
- Gunakan naming convention yang konsisten
- Test setiap router setelah add

**3. Role-Based Access (Optional):**

Buat user dengan akses terbatas per cabang:

```sql
-- Via Database PostgreSQL
-- User head1 hanya bisa akses JAKARTA-01
INSERT INTO user_router_access (user_id, router_name)
SELECT id, 'JAKARTA-01' FROM users WHERE username = 'head1';
```

Atau via UI: **User Management → Edit User → Select Routers**

---

## 🔍 Verification & Testing

Setelah setup, verifikasi semua router:

### 1. Dashboard Check

1. Buka **NAT Management** page
2. Cek apakah semua router muncul di dropdown
3. Test search PPPoE di setiap router

### 2. Connection Health

1. Buka **Router Management** page
2. Lihat status connection setiap router:
   - ✅ **Connected** (hijau) = OK
   - ❌ **Disconnected** (merah) = Problem
3. Klik **"Test"** untuk test ulang

### 3. Run Diagnostic Tool

Untuk setiap router:

```bash
cd tools
router-diagnostic.exe 192.168.1.1 8728 admin password123
```

Simpan output untuk reference.

### 4. Test NAT Operations

1. Pilih router di dropdown
2. Test search PPPoE username
3. Test update NAT configuration
4. Verify changes di Winbox/WebFig

---

## 🚨 Common Issues & Solutions

### Issue 1: "Router not configured"

**Penyebab:** Belum ada router di database

**Solusi:** Follow Step 2 untuk add router

---

### Issue 2: Connection Timeout

**Penyebab:**
- Port salah (19699 instead of 8728)
- API service disabled
- Firewall blocking

**Solusi:**
- Verify port = 8728
- Enable API service
- Check firewall rules

**Detailed troubleshooting:** [TROUBLESHOOTING.md](TROUBLESHOOTING.md)

---

### Issue 3: "Authentication Failed"

**Penyebab:**
- Username/password salah
- User tidak memiliki API permission

**Solusi:**
```routeros
# Cek user permission
/user print detail where name=admin

# Set group ke full jika perlu
/user set admin group=full
```

---

### Issue 4: Router Appears Offline

**Penyebab:**
- Router benar-benar offline
- Network issue
- API service crash

**Solusi:**
1. Ping router: `ping 192.168.1.1`
2. Check API service: `/ip service print`
3. Restart API: `/ip service restart api`
4. Reboot router (last resort): `/system reboot`

---

## 📝 Configuration Checklist

Before going to production:

- [ ] API service enabled di semua router
- [ ] Firewall rules configured properly
- [ ] Test connection berhasil untuk semua router
- [ ] Password sudah di-rotate (bukan default)
- [ ] User khusus untuk aplikasi sudah dibuat
- [ ] Role-based access sudah dikonfigurasi
- [ ] Backup configuration router (`.backup` & `.rsc`)
- [ ] Documentation network topology lengkap
- [ ] Emergency access procedure documented
- [ ] Monitoring & alerting setup

---

## 🎯 Next Steps

Setelah setup router selesai:

1. **Setup User Management:**
   - Buat user untuk setiap cabang
   - Assign router access per user
   - Test login dengan user non-admin

2. **Configure NAT Rules:**
   - Setup ONT NAT rule di MikroTik
   - Comment rule dengan "REMOTE ONT PELANGGAN"
   - Test NAT update via aplikasi

3. **Monitor & Maintain:**
   - Check **Activity Logs** regularly
   - Monitor connection health
   - Rotate passwords setiap 3 bulan
   - Backup configuration weekly

4. **Training:**
   - Train staff cara gunakan aplikasi
   - Documented common procedures
   - Setup support channel

---

## 📞 Support

Jika butuh bantuan:

1. Check [Troubleshooting Guide](TROUBLESHOOTING.md)
2. Run diagnostic tool dan collect output
3. Check application logs
4. Create issue di GitHub dengan detail lengkap

---

## 📚 Additional Resources

- [MikroTik API Documentation](https://wiki.mikrotik.com/wiki/Manual:API)
- [RouterOS Services](https://wiki.mikrotik.com/wiki/Manual:IP/Services)
- [Firewall Configuration](https://wiki.mikrotik.com/wiki/Manual:IP/Firewall/Filter)
- [Troubleshooting Guide](TROUBLESHOOTING.md)

---

**Last Updated:** 2025-10-16
**Version:** 4.1
**Author:** NAT Management Team
