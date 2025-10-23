# Cara Menjalankan Aplikasi FTTH Management

Panduan lengkap instalasi dan menjalankan aplikasi dengan backend Mikrotik.

## Opsi 1: Menggunakan XAMPP (Paling Mudah) ⭐ RECOMMENDED

### Download & Install XAMPP

1. **Download XAMPP:**
   - Link: https://www.apachefriends.org/download.html
   - Pilih versi untuk Windows
   - Download minimal PHP 7.0+

2. **Install XAMPP:**
   - Jalankan installer
   - Install ke `C:\xampp` (default)
   - Centang: Apache, PHP, phpMyAdmin (MySQL opsional)
   - Klik Install

### Setup Project

1. **Copy Project ke XAMPP:**
   ```
   Copy folder: mapping-ftth
   Ke folder: C:\xampp\htdocs\
   
   Hasil: C:\xampp\htdocs\mapping-ftth\
   ```

2. **Start XAMPP:**
   - Buka **XAMPP Control Panel**
   - Klik tombol **Start** pada **Apache**
   - Apache akan running di port 80
   - Status akan berubah hijau

3. **Test Akses:**
   - Buka browser
   - Akses: `http://localhost/mapping-ftth/`
   - Aplikasi akan terbuka! ✅

### Configure Mikrotik

1. **Edit Config:**
   - Buka file: `C:\xampp\htdocs\mapping-ftth\api\config.php`
   - Edit dengan Notepad atau VS Code:
   ```php
   <?php
   $MIKROTIK_HOST = '192.168.88.1';  // Ganti dengan IP Mikrotik Anda
   $MIKROTIK_USER = 'admin';          // Username Mikrotik
   $MIKROTIK_PASS = 'password';       // Password Mikrotik
   $MIKROTIK_PORT = 8728;
   ?>
   ```

2. **Save File** (Ctrl+S)

3. **Test Koneksi:**
   - Refresh browser: `http://localhost/mapping-ftth/`
   - Klik menu **"Pengaturan"**
   - Klik **"Test Koneksi"**
   - Jika berhasil: "✓ Koneksi Mikrotik berhasil!"

---

## Opsi 2: Menggunakan WAMP Server

### Install WAMP

1. **Download WAMP:**
   - Link: https://www.wampserver.com/en/download-wampserver-64bits/
   - Install ke `C:\wamp64`

2. **Start WAMP:**
   - Klik icon WAMP di system tray
   - Pastikan icon berwarna **HIJAU** (online)

### Setup Project

1. **Copy Project:**
   ```
   Copy folder: mapping-ftth
   Ke: C:\wamp64\www\
   
   Hasil: C:\wamp64\www\mapping-ftth\
   ```

2. **Akses:**
   - Browser: `http://localhost/mapping-ftth/`

3. **Configure Mikrotik** (sama seperti XAMPP)

---

## Opsi 3: Menggunakan PHP Built-in Server (Development Only)

**Cara Cepat untuk Testing:**

### Prerequisite:
- PHP sudah terinstall di Windows
- Check dengan: `php --version`

### Jalankan:

1. **Buka Command Prompt:**
   ```cmd
   cd C:\OPREKV1\NAT\nat-management-appV3.1\NAT4.2\mapping-ftth
   ```

2. **Start PHP Server:**
   ```cmd
   php -S localhost:8000
   ```

3. **Akses:**
   - Browser: `http://localhost:8000/`

**CATATAN:** 
- ⚠️ PHP built-in server **TIDAK mendukung .htaccess**
- ⚠️ **TIDAK untuk production**, hanya development
- ⚠️ Mikrotik API mungkin **TIDAK BERFUNGSI** karena single-threaded

---

## Opsi 4: Menggunakan Laragon (Modern & Ringan)

### Install Laragon

1. **Download:**
   - Link: https://laragon.org/download/
   - Pilih versi Lite atau Full

2. **Install:**
   - Install ke folder default
   - Centang: Apache, PHP, MySQL (opsional)

3. **Setup Project:**
   ```
   Copy folder: mapping-ftth
   Ke: C:\laragon\www\
   ```

4. **Start Laragon:**
   - Klik **Start All**
   - Akses: `http://localhost/mapping-ftth/`

---

## Setup Lengkap Step-by-Step (XAMPP)

### 1. Install XAMPP
```
✓ Download XAMPP dari apachefriends.org
✓ Install ke C:\xampp
✓ Selesai instalasi
```

### 2. Copy Project
```
✓ Copy folder "mapping-ftth"
✓ Paste ke C:\xampp\htdocs\
✓ Path final: C:\xampp\htdocs\mapping-ftth\
```

### 3. Start Apache
```
✓ Buka XAMPP Control Panel
✓ Klik Start pada Apache
✓ Tunggu status jadi hijau
✓ Port 80 akan aktif
```

### 4. Configure Mikrotik
```
✓ Edit file: htdocs\mapping-ftth\api\config.php
✓ Isi IP Mikrotik
✓ Isi username dan password
✓ Save file
```

### 5. Test Application
```
✓ Browser: http://localhost/mapping-ftth/
✓ Aplikasi terbuka
✓ Klik menu "Pengaturan"
✓ Klik "Test Koneksi" pada Mikrotik Integration
✓ Tunggu notifikasi sukses
```

### 6. Setup Mikrotik (jika belum)
```bash
# Di Mikrotik Terminal atau Winbox:
/ip service enable api
/ip service set api port=8728
/user add name=ftth-api group=full password=SecurePass123
```

### 7. Test Full System
```
✓ Tambah data pelanggan dengan PPPOE
✓ Pastikan pelanggan connect ke Mikrotik PPPoE
✓ Tunggu 30 detik (auto-sync)
✓ Status pelanggan berubah jadi ONLINE (hijau)
✓ Marker di peta berubah warna hijau + pulse
✓ SUCCESS! 🎉
```

---

## Troubleshooting

### Apache Tidak Mau Start

**Problem:** Port 80 already in use

**Solusi:**
1. Stop aplikasi yang pakai port 80:
   - Skype (ubah setting)
   - IIS (disable di Windows Features)
   - World Wide Web Publishing Service

2. Atau ubah port Apache:
   - Edit: `C:\xampp\apache\conf\httpd.conf`
   - Cari: `Listen 80`
   - Ubah jadi: `Listen 8080`
   - Restart Apache
   - Akses: `http://localhost:8080/mapping-ftth/`

### PHP Error: "Call to undefined function fsockopen"

**Solusi:**
1. Edit `php.ini`: `C:\xampp\php\php.ini`
2. Cari: `;extension=sockets`
3. Hapus `;` jadi: `extension=sockets`
4. Save dan restart Apache

### CORS Error (jika API di server lain)

**Solusi:** Sudah ditangani di `mikrotik-api.php`:
```php
header('Access-Control-Allow-Origin: *');
```

### Mikrotik Connection Timeout

**Cek:**
```bash
# Test ping
ping 192.168.88.1

# Test port (via telnet atau tool lain)
telnet 192.168.88.1 8728
```

**Solusi:**
- Pastikan port 8728 tidak diblokir firewall
- Cek Mikrotik firewall rules
- Pastikan PC dan Mikrotik satu network

---

## Production Deployment

### Untuk Server Production:

**1. Setup Proper Web Server:**
- Gunakan Apache atau Nginx
- Enable HTTPS (SSL Certificate)
- Setup virtual host

**2. Security:**
```apache
# .htaccess untuk protect api/config.php
<Files "config.php">
    Order Allow,Deny
    Deny from all
</Files>
```

**3. Performance:**
- Enable gzip compression
- Enable browser caching
- Optimize PHP (OPcache)

**4. Monitoring:**
- Setup error logging
- Monitor API response time
- Setup uptime monitoring

---

## Quick Commands Reference

### XAMPP Commands:
```cmd
# Start Apache
C:\xampp\xampp-control.exe

# View Apache Error Log
C:\xampp\apache\logs\error.log

# View PHP Errors
C:\xampp\php\php_error.log
```

### Access URLs:
```
Application: http://localhost/mapping-ftth/
API Test:    http://localhost/mapping-ftth/api/mikrotik-api.php?action=test
API Status:  http://localhost/mapping-ftth/api/mikrotik-api.php?action=status
```

### Mikrotik Commands:
```bash
# Check API status
/ip service print

# Check active PPPoE
/ppp active print

# Check API user
/user print

# Enable API
/ip service enable api
```

---

## Struktur Folder Final

```
C:\xampp\htdocs\mapping-ftth\
├── api\
│   ├── config.php          ← Edit ini untuk Mikrotik config
│   └── mikrotik-api.php    ← Backend API
├── css\
│   └── style.css
├── js\
│   ├── app.js
│   ├── data-manager.js
│   ├── map-controller.js
│   └── page-controller.js
├── index.html              ← Entry point
├── MIKROTIK_SETUP.md
├── INSTALLATION.md         ← File ini
└── README.md
```

---

## Video Tutorial (Outline)

**Part 1: Install XAMPP** (5 menit)
1. Download XAMPP
2. Install
3. Start Apache
4. Test localhost

**Part 2: Setup Project** (3 menit)
1. Copy project ke htdocs
2. Open in browser
3. Navigate menu

**Part 3: Configure Mikrotik** (5 menit)
1. Enable Mikrotik API
2. Create user
3. Edit config.php
4. Test connection

**Part 4: Test Full System** (5 menit)
1. Add pelanggan
2. Connect via PPPoE
3. Watch status update
4. Check map markers

---

**Selamat!** Aplikasi siap digunakan! 🎉

Jika ada masalah, cek:
1. Apache running (hijau di XAMPP)
2. File config.php sudah diedit
3. Mikrotik API enabled
4. Network antara PC dan Mikrotik lancar
