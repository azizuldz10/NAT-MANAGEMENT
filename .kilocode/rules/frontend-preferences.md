## Brief overview
- Panduan ringkas untuk pengembangan UI/UX aplikasi NAT Management berbasis HTML/CSS/JS (tanpa framework besar), menyesuaikan preferensi yang telah diterapkan di proyek.
- Fokus: konsistensi tema, responsivitas mobile, keamanan dasar di sisi frontend, struktur aset statis, dan gaya komunikasi kerja.
- Rujukan file utama: [web.templates.nat_management.html](web/templates/nat_management.html), [web.templates.login.html](web/templates/login.html), [web.static.css.theme.css](web/static/css/theme.css), [cmd.main.go](cmd/main.go), [internal.middleware.secure_auth.go](internal/middleware/secure_auth.go).

## Gaya komunikasi
- Gunakan Bahasa Indonesia, ringkas, teknis, dan langsung ke pokok masalah. Hindari basa-basi.
- Ketika mengusulkan perubahan, sertakan lokasi file dan baris yang relevan agar mudah ditinjau, contoh:
  - Logo login di [web.templates.login.html](web/templates/login.html:353).
  - Variabel CSS sidebar di [web.static.css.theme.css](web/static/css/theme.css:50).
- Jika ada trade-off, jelaskan dampaknya secara singkat (aksesibilitas, performa, keamanan).

## Alur pengembangan (workflow)
- Lakukan perubahan UI konsisten dengan tema global di [web.static.css.theme.css](web/static/css/theme.css).
- Uji di desktop dan mobile. Pastikan toggling kelas sidebar berfungsi:
  - Toggle ‘collapsed’/‘open’ dikelola di [initializeSidebar()](web/templates/nat_management.html:442).
- Verifikasi aset statis terlayani dengan benar melalui route static di [cmd.main.go](cmd/main.go:102).
- Setelah perubahan server-side (route static, CSP), restart server lokal untuk efek penuh.

## Konvensi kode
- HTML
  - Hindari framework CSS besar; gunakan komponen custom yang sudah ada.
  - Struktur semantik: header, nav, main, footer; sidebar memakai [web.templates.nat_management.html](web/templates/nat_management.html:62).
- CSS
  - Manfaatkan custom properties (CSS variables) untuk konsistensi:
    - Contoh penggunaan variabel: [var(--sidebar-width)](web/static/css/theme.css:50), [var(--transition-smooth)](web/static/css/theme.css:29).
  - Hindari inline style pada HTML kecuali untuk override minimal.
- JavaScript
  - Gunakan event handler terpisah dan class toggling:
    - Contoh toggle di [initializeSidebar()](web/templates/nat_management.html:442).
  - Hindari ketergantungan CDN yang tidak perlu; prefer self-host assets.

## UI/UX preferensi
- Sidebar kiri harus presisi, tidak overlap dan responsif:
  - Penempatan tombol hamburger disesuaikan: lihat aturan di [web.static.css.theme.css](web/static/css/theme.css:129,155,1157).
- Halaman login:
  - Tampilkan brand logo di header (gunakan path /image). Contoh:
    - [img.alt="NAT Management Logo"](web/templates/login.html:353) dengan sumber /image/logo.png.
  - Hindari elemen “keyboard shortcuts layout” pada dashboard; referensi penghapusan di [web.templates.nat_management.html](web/templates/nat_management.html:15,399,542,419,1436).
- Mobile
  - Body scroll locking saat sidebar ‘open’; atur via JS di [web.templates.nat_management.html](web/templates/nat_management.html:463).

## Aset & static route
- Semua brand asset ditempatkan di ‘image/’ dan disajikan via /image/*:
  - Konfigurasi route static di [cmd.main.go](cmd/main.go:102) untuk /image dan /static.
- Gunakan path absolut berbasis origin (contoh: /image/logo.png), bukan relatif.

## Keamanan & header
- CSP dasar diset di [internal.middleware.SecureAuthMiddleware.setSecurityHeaders()](internal/middleware/secure_auth.go:95).
- Rekomendasi (ketika siap):
  - Hilangkan ‘unsafe-inline’ untuk scripts/styles; gunakan nonce/hash.
  - Tambahkan HSTS di lingkungan produksi (via reverse proxy atau middleware terpisah).

## Kinerja
- Pertahankan animasi halus namun sediakan fallback untuk pengguna dengan prefers-reduced-motion (tambahkan di [web.static.css.theme.css](web/static/css/theme.css:29)).
- Optimalkan aset (cache-control, ETag) di layer reverse proxy (lihat panduan di [docs.DEPLOYMENT.md](docs/DEPLOYMENT.md)).

## Testing & verifikasi
- Verifikasi rendering logo:
  - HEAD /image/logo.png harus 200 OK (cek via curl).
- Uji toggling sidebar di desktop/mobile, pastikan body scroll lock berfungsi; rujuk [initializeSidebar()](web/templates/nat_management.html:442).
- Hard refresh (Ctrl+F5) setelah perubahan route static atau CSP untuk menghindari cache usang.

## Dokumentasi perubahan
- Setiap perubahan signifikan, sertakan:
  - File & baris, alasan, dampak singkat, cara roll-back.
  - Contoh: “Geser hamburger agar tidak overlap” di [web.static.css.theme.css](web/static/css/theme.css:129,155,1157).

## Hal yang dihindari
- Mengaktifkan kembali “keyboard shortcuts layout” di dashboard kecuali ada kebutuhan khusus.
- Menggunakan path aset yang tidak terlayani oleh server (hindari /SS untuk brand; gunakan /image).
