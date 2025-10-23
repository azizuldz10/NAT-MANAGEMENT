# Sistem Management FTTH

Aplikasi web management dan pemetaan jaringan FTTH (Fiber To The Home) dengan hierarki parent-child: **Server → OLT → ODC → ODP → Pelanggan**

## 🚀 Versi Baru: Backend Golang + SQLite

Project ini sekarang menggunakan:
- **Backend**: Golang dengan RESTful API
- **Database**: SQLite 
- **Frontend**: Vanilla JavaScript + Leaflet.js
- **Integrasi**: Mikrotik RouterOS API untuk monitoring status pelanggan

**Instalasi Backend**: Lihat [INSTALL_BACKEND.md](INSTALL_BACKEND.md)

## Fitur Utama

### 1. Dashboard
- Overview statistik semua node
- Card display dengan jumlah Server, OLT, ODC, ODP, dan Pelanggan
- UI modern dan responsive

### 2. Management CRUD
Setiap tipe node memiliki halaman management sendiri dengan fitur:
- **Table view** dengan informasi lengkap
- **Tambah node baru** dengan form modal
- **Edit node** dengan update data
- **Hapus node** dengan validasi child protection
- **Parent-child relationship** otomatis

#### Hierarki & Dependency:
- **Server**: Node utama (tidak punya parent)
- **OLT**: Harus memilih Server sebagai parent
- **ODC**: Harus memilih OLT sebagai parent
- **ODP**: Harus memilih ODC sebagai parent
- **Pelanggan**: Harus memilih ODP sebagai parent (dengan field tambahan: PPPOE, WhatsApp)

### 3. Peta Jaringan
- **Visualisasi interaktif** menggunakan Leaflet.js
- **Marker berbeda** untuk setiap tipe node dengan color coding:
  - 🔴 Server (Merah)
  - 🟠 OLT (Orange)
  - 🔵 ODC (Biru)
  - 🟣 ODP (Ungu)
  - 🟢 Pelanggan (Hijau)
- **Auto-connect**: Garis penghubung otomatis antara parent-child
- **Popup info** dengan detail node dan relationship
- **Pencarian koordinat** dengan format `-xxxxxxx,yxxxxxx`

### 4. Form Management
- **Dropdown parent selection** otomatis berdasarkan tipe node
- **Koordinat manual** atau pick dari peta
- **Validasi** untuk memastikan parent tersedia
- **Field khusus pelanggan** (PPPOE, WhatsApp)

### 5. Utilitas
- **Export JSON**: Backup semua data
- **Import JSON**: Restore data dari file
- **Clear All**: Hapus semua data dengan double confirmation
- **LocalStorage**: Auto-save setiap perubahan

## Cara Menggunakan

### 1. Membuka Aplikasi
Buka file `index.html` di browser modern (Chrome/Firefox/Edge recommended)

### 2. Menambah Node

#### a) Tambah Server
1. Navigasi ke menu **Server**
2. Klik tombol **"Tambah Server"**
3. Isi nama dan koordinat (atau klik "Pilih dari Peta")
4. Klik **"Simpan"**

#### b) Tambah OLT
1. Navigasi ke menu **OLT**
2. Klik tombol **"Tambah OLT"**
3. **Pilih Server parent** dari dropdown
4. Isi nama dan koordinat
5. Klik **"Simpan"**
6. System akan otomatis menghubungkan OLT ke Server

#### c) Tambah ODC
1. Navigasi ke menu **ODC**
2. Klik **"Tambah ODC"**
3. **Pilih OLT parent** dari dropdown
4. Isi data lengkap
5. Simpan - otomatis terkoneksi ke OLT

#### d) Tambah ODP
1. Navigasi ke menu **ODP**
2. Klik **"Tambah ODP"**
3. **Pilih ODC parent**
4. Isi data dan simpan

#### e) Tambah Pelanggan
1. Navigasi ke menu **Pelanggan**
2. Klik **"Tambah Pelanggan"**
3. **Pilih ODP parent**
4. Isi nama, PPPOE, WhatsApp, dan koordinat
5. Simpan

### 3. Edit Node
- Klik tombol **"Edit"** di tabel
- Update data yang diperlukan
- Klik **"Simpan"**

### 4. Hapus Node
- Klik tombol **"Hapus"**
- **Catatan**: Node yang memiliki child tidak bisa dihapus. Hapus child terlebih dahulu.

### 5. Lihat di Peta
- Klik tombol **"Lihat"** untuk zoom ke lokasi node di peta
- Atau navigasi ke menu **"Peta Jaringan"**
- Semua node dan koneksi akan tampil otomatis

### 6. Pencarian Koordinat
1. Navigasi ke **"Peta Jaringan"**
2. Masukkan koordinat dengan format: `-6.123456,106.789012`
3. Klik **"Cari"** atau tekan **Enter**
4. Peta akan zoom ke lokasi dengan marker temporary

### 7. Pick Koordinat dari Peta
1. Saat mengisi form tambah/edit node
2. Klik tombol **"Pilih dari Peta"**
3. Klik lokasi di peta
4. Koordinat otomatis terisi di form

## Struktur File

```
mapping-ftth/
├── index.html              # Main HTML dengan multi-page layout
├── css/
│   └── style.css          # Modern UI styling
├── js/
│   ├── app.js             # Main application & event handlers
│   ├── data-manager.js    # CRUD operations & parent-child logic
│   ├── map-controller.js  # Leaflet map management
│   └── page-controller.js # Multi-page navigation & table rendering
└── README.md              # Dokumentasi
```

## Teknologi

- **HTML5 & CSS3**: Modern semantic layout
- **Vanilla JavaScript**: No framework, pure ES6+
- **Leaflet.js 1.9.4**: Interactive mapping library
- **Font Awesome 6.4**: Icon library
- **LocalStorage API**: Client-side data persistence

## Validasi & Business Logic

### Parent-Child Rules:
1. **Server** tidak memerlukan parent
2. **OLT** WAJIB memiliki Server parent
3. **ODC** WAJIB memiliki OLT parent
4. **ODP** WAJIB memiliki ODC parent
5. **Pelanggan** WAJIB memiliki ODP parent

### Delete Protection:
- Node yang memiliki child **TIDAK BISA** dihapus
- Hapus child terlebih dahulu (bottom-up deletion)
- Sistem akan memberi warning jika ada child

### Auto-Connection:
- Saat node ditambahkan/diupdate dengan parent, koneksi otomatis dibuat
- Garis biru putus-putus menunjukkan relationship
- Update parent akan menghapus koneksi lama dan buat yang baru

## Browser Support

- ✅ Chrome 90+
- ✅ Firefox 88+
- ✅ Edge 90+
- ✅ Safari 14+

## Tips Penggunaan

1. **Workflow**: Selalu mulai dari Server → OLT → ODC → ODP → Pelanggan
2. **Backup**: Export data secara berkala untuk backup
3. **Koordinat**: Gunakan fitur "Pilih dari Peta" untuk akurasi koordinat
4. **Visualisasi**: Gunakan menu "Peta Jaringan" untuk melihat topologi lengkap
5. **Navigation**: Klik "Lihat" untuk quick jump ke lokasi node di peta

## Troubleshooting

**Q: Tidak bisa tambah OLT?**
A: Pastikan sudah ada Server. Tambahkan Server terlebih dahulu.

**Q: Tidak bisa hapus node?**
A: Node masih memiliki child. Hapus child terlebih dahulu.

**Q: Peta tidak muncul?**
A: Klik menu "Peta Jaringan" dan tunggu beberapa detik untuk loading.

**Q: Data hilang setelah refresh?**
A: Data disimpan di LocalStorage browser. Jangan clear browser data atau gunakan incognito mode.

## License

Open Source - Free to use and modify
