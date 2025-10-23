<?php
// Mikrotik Configuration
// Edit sesuai dengan router Mikrotik Anda

$MIKROTIK_HOST = '192.168.88.1';  // IP Address Mikrotik
$MIKROTIK_USER = 'admin';          // Username Mikrotik
$MIKROTIK_PASS = '';               // Password Mikrotik
$MIKROTIK_PORT = 8728;             // API Port (default 8728)

// CATATAN:
// 1. Pastikan Mikrotik API service sudah enabled
// 2. Command di Mikrotik: /ip service enable api
// 3. Pastikan firewall tidak memblokir port 8728
// 4. User harus punya permission untuk akses API
?>
