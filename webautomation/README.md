# 🚀 ONT WiFi Extractor - Universal Auto-Detection

Ekstrak informasi WiFi (SSID & Password) dari berbagai model ONT secara otomatis menggunakan web automation.

## ✨ Features

- 🔍 **Auto-Detection** - Otomatis mendeteksi model ONT
- 🌐 **Multi-Model Support** - Support 4+ model ONT berbeda
- 🤖 **Web Automation** - Menggunakan Playwright untuk browser automation
- 📊 **JSON Output** - Export hasil ke file JSON
- 🐛 **Debug Mode** - Mode debug dengan screenshot untuk troubleshooting
- 💻 **CLI Interface** - Command-line interface yang mudah digunakan

## 📋 Supported Models

| Model | Brand | Interface Type | Status |
|-------|-------|---------------|--------|
| GM220-S | Fiberhome | Frame-based | ✅ Working |
| AccesGo / OLD_MODEL | Various | Menu-based (NETWORK → WLAN) | ✅ Working |
| ZXHN F450 | ZTE | Standard login + iframe dashboard | ✅ Working |
| ZXHN F477V2 | ZTE | WELCOME page + icon menu | ✅ Working |

## 🛠️ Installation

### Prerequisites

- Node.js v14+
- npm atau yarn

### Install Dependencies

```bash
cd webautomation
npm install
```

### Install Browser (Chromium)

```bash
npm run install-browsers
```

## 🎯 Usage

### Quick Start (Auto-Detection)

```bash
# Menggunakan launcher (auto-detect model)
node ont-extractor-launcher.js <ONT_URL> [username] [password] [--debug]
```

### Examples

#### 1. Auto-detect dengan default credentials (admin/admin)

```bash
node ont-extractor-launcher.js http://192.168.1.1/
```

#### 2. Dengan custom credentials

```bash
node ont-extractor-launcher.js http://192.168.1.1/ admin mypassword
```

#### 3. Dengan debug mode (show browser + screenshots)

```bash
node ont-extractor-launcher.js http://192.168.1.1/ admin admin --debug
```

#### 4. ZTE F477V2 Example

```bash
node ont-extractor-launcher.js http://tunnel3.ebilling.id:15634/ admin suportadmin --debug
```

### NPM Scripts

```bash
# Auto-detect dan extract
npm start <url> <username> <password>

# Dengan debug mode
npm run extract:debug <url> <username> <password>

# Show help
npm run extract:help

# Extract model tertentu (manual)
npm run extract:gm220s <url> <username> <password> --debug
npm run extract:zte-f450 <url> <username> <password> --debug
npm run extract:zte-f477v2 <url> <username> <password> --debug
```

## 📤 Output

### Console Output

```
============================================================
         ZTE ZXHN F477V2 - WiFi Information
============================================================
ONT URL        : http://tunnel3.ebilling.id:15634/
Model          : ZXHN F477V2
Login Username : admin
────────────────────────────────────────────────────────────
SSID           : MyWiFiNetwork
Password       : MySecurePassword123
Security       : WPAand11i
Encryption     : TKIPandAESEncryption
Authentication : WPA/WPA2-PSK
────────────────────────────────────────────────────────────
Extracted At   : 2025-10-16T16:15:21.569Z
============================================================
```

### JSON Output

Output disimpan ke file JSON sesuai model:

- `wifi_info.json` - Untuk GM220-S & OLD_MODEL
- `zte_wifi_info.json` - Untuk ZTE F450
- `zte_f477v2_wifi_info.json` - Untuk ZTE F477V2

Example `zte_f477v2_wifi_info.json`:

```json
{
  "ssid": "MyWiFiNetwork",
  "password": "MySecurePassword123",
  "security": "WPAand11i",
  "encryption": "TKIPandAESEncryption",
  "authentication": "WPA/WPA2-PSK",
  "extracted_at": "2025-10-16T16:15:21.569Z",
  "ont_url": "http://tunnel3.ebilling.id:15634/",
  "ont_model": "ZXHN F477V2",
  "credentials": {
    "username": "admin",
    "password_used": "suportadmin"
  }
}
```

## 🔧 Manual Extraction (Specific Model)

Jika auto-detection gagal atau ingin menggunakan extractor tertentu:

### GM220-S / AccesGo

```bash
node ont-wifi-extractor.js http://192.168.1.1/ admin admin --debug
```

### ZTE F450

```bash
node zte-f450-extractor.js http://192.168.1.1/ admin admin --debug
```

### ZTE F477V2

```bash
node zte-f477v2-extractor.js http://192.168.1.1/ admin suportadmin --debug
```

## 🐛 Troubleshooting

### 1. Browser tidak ter-install

```bash
npm run install-browsers
```

### 2. Timeout / Connection Error

- Pastikan ONT device accessible dari network
- Coba increase timeout dengan edit file extractor
- Gunakan `--debug` mode untuk melihat screenshot

### 3. Auto-detection salah

Gunakan manual extraction dengan file extractor spesifik:

```bash
# Paksa gunakan extractor tertentu
node zte-f477v2-extractor.js <url> <user> <pass> --debug
```

### 4. Login gagal

- Periksa username dan password
- Beberapa model menggunakan credentials berbeda:
  - F450: biasanya `admin/admin`
  - F477V2: bisa `admin/suportadmin` atau `admin/admin`
- Gunakan `--debug` mode untuk melihat login page screenshot

### 5. Password tidak ter-extract

- Gunakan `--debug` mode
- Periksa screenshot `*_security_settings.png`
- Pastikan page sudah fully loaded
- Beberapa model perlu scroll atau klik "Click here to display"

## 📁 Project Structure

```
webautomation/
├── ont-extractor-launcher.js    # 🚀 Main launcher (auto-detect)
├── ont-wifi-extractor.js        # GM220-S & OLD_MODEL extractor
├── zte-f450-extractor.js        # ZTE F450 extractor
├── zte-f477v2-extractor.js      # ZTE F477V2 extractor
├── package.json                 # Dependencies & scripts
├── README.md                    # Documentation (this file)
├── wifi_info.json               # Output (GM220-S/OLD_MODEL)
├── zte_wifi_info.json           # Output (F450)
└── zte_f477v2_wifi_info.json    # Output (F477V2)
```

## 🔒 Security Notes

- Tool ini menggunakan credentials untuk login ke ONT device
- Credentials tidak disimpan atau dikirim ke external server
- Gunakan dengan tanggung jawab dan hanya untuk device yang Anda miliki
- Screenshot debug mode mungkin berisi informasi sensitif

## 🤝 Contributing

Untuk menambahkan support model ONT baru:

1. Research interface web ONT menggunakan browser
2. Identifikasi:
   - Login flow
   - Menu navigation structure
   - Location of SSID dan Password
3. Buat extractor baru berdasarkan template yang ada
4. Update auto-detection logic di `ont-extractor-launcher.js`
5. Test dengan device asli
6. Update README dengan model baru

## 📝 License

ISC

## 🙏 Credits

Built with:
- [Playwright](https://playwright.dev/) - Browser automation
- Node.js - Runtime environment

---

**⚠️ Disclaimer:** Tool ini dibuat untuk tujuan legitimate administration dari ONT device yang Anda miliki. Gunakan dengan bijak dan sesuai hukum yang berlaku.
