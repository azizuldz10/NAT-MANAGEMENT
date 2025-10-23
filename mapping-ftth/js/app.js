class App {
    constructor() {
        this.dataManager = new DataManager();
        this.mapController = new MapController(this.dataManager);
        this.pageController = new PageController(this.dataManager);
        this.modal = null;
        this.currentEditNode = null;
        this.apiBase = 'http://localhost:8080/api';
        this.statusRefreshInterval = null;
        this.refreshIntervalSeconds = 120; // 2 menit
        this.mikrotikEnabled = true;
        this.init();
    }

    async init() {
        await this.dataManager.init();
        this.setupEventListeners();
        this.pageController.showPage('dashboard');
        await this.loadMikrotikConfig();
        this.startMikrotikStatusSync();
        window.app = this;
    }

    setupEventListeners() {
        document.getElementById('addServerBtn').addEventListener('click', () => {
            this.showNodeForm('server');
        });

        document.getElementById('addOltBtn').addEventListener('click', () => {
            this.showNodeForm('olt');
        });

        document.getElementById('addOdcBtn').addEventListener('click', () => {
            this.showNodeForm('odc');
        });

        document.getElementById('addOdpBtn').addEventListener('click', () => {
            this.showNodeForm('odp');
        });

        document.getElementById('addPelangganBtn').addEventListener('click', () => {
            this.showNodeForm('pelanggan');
        });

        document.getElementById('downloadTemplateBtn').addEventListener('click', () => {
            this.downloadExcelTemplate();
        });

        document.getElementById('importExcelBtn').addEventListener('click', () => {
            document.getElementById('excelFileInput').click();
        });

        document.getElementById('excelFileInput').addEventListener('change', (e) => {
            this.importExcelPelanggan(e.target.files[0]);
        });

        document.getElementById('searchCoordBtn').addEventListener('click', () => {
            this.searchCoordinates();
        });

        document.getElementById('searchCoordinates').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.searchCoordinates();
            }
        });

        document.getElementById('exportBtn').addEventListener('click', () => {
            this.exportData();
        });

        document.getElementById('importBtn').addEventListener('click', () => {
            document.getElementById('importFile').click();
        });

        document.getElementById('importFile').addEventListener('change', (e) => {
            this.importData(e.target.files[0]);
        });

        document.getElementById('clearAllBtn').addEventListener('click', () => {
            this.clearAllData();
        });

        document.getElementById('saveMikrotikConfigBtn').addEventListener('click', () => {
            this.saveMikrotikConfig();
        });

        document.getElementById('loadMikrotikConfigBtn').addEventListener('click', () => {
            this.loadMikrotikConfig();
        });

        document.getElementById('testMikrotikBtn').addEventListener('click', () => {
            this.testMikrotikConnection();
        });

        document.getElementById('syncNowBtn').addEventListener('click', () => {
            this.fetchMikrotikStatus();
            this.showToast('Syncing status...', 2000);
        });

        document.getElementById('toggleSyncBtn').addEventListener('click', (e) => {
            const btn = e.target.closest('button');
            this.mikrotikEnabled = !this.mikrotikEnabled;
            
            if (this.mikrotikEnabled) {
                btn.innerHTML = '<i class="fas fa-pause"></i> Pause Sync';
                btn.classList.remove('btn-success');
                btn.classList.add('btn-secondary');
                this.startMikrotikStatusSync();
            } else {
                btn.innerHTML = '<i class="fas fa-play"></i> Resume Sync';
                btn.classList.remove('btn-secondary');
                btn.classList.add('btn-success');
                this.stopMikrotikStatusSync();
            }
        });

        this.setupModal();
        this.setupFormHandlers();

        const navItems = document.querySelectorAll('.nav-item[data-page="map"]');
        navItems.forEach(item => {
            item.addEventListener('click', () => {
                setTimeout(() => {
                    if (!this.mapController.map) {
                        this.mapController.initMap('map');
                    } else {
                        this.mapController.map.invalidateSize();
                        this.mapController.loadExistingData();
                    }
                }, 100);
            });
        });
    }

    setupModal() {
        this.modal = document.getElementById('formModal');
        
        const closeButtons = this.modal.querySelectorAll('.close-modal');
        closeButtons.forEach(btn => {
            btn.addEventListener('click', () => {
                this.hideModal();
            });
        });

        window.addEventListener('click', (e) => {
            if (e.target === this.modal) {
                this.hideModal();
            }
        });
    }

    setupFormHandlers() {
        document.getElementById('saveNodeBtn').addEventListener('click', () => {
            this.saveNode();
        });

        document.getElementById('pickFromMapBtn').addEventListener('click', () => {
            this.pickCoordinateFromMap();
        });
    }

    showNodeForm(type, nodeId = null) {
        this.currentEditNode = nodeId;
        
        const isEdit = nodeId !== null;
        const node = isEdit ? this.dataManager.getNode(nodeId) : null;

        document.getElementById('modalTitle').textContent = isEdit 
            ? `Edit ${this.dataManager.getTypeLabel(type)}` 
            : `Tambah ${this.dataManager.getTypeLabel(type)}`;

        document.getElementById('formNodeId').value = nodeId || '';
        document.getElementById('formNodeType').value = type;
        document.getElementById('formName').value = node ? node.name : '';
        document.getElementById('formCoordinates').value = node ? `${node.lat},${node.lng}` : '';

        const parentGroup = document.getElementById('formParentGroup');
        const parentSelect = document.getElementById('formParent');
        
        if (type !== 'server') {
            const parents = this.dataManager.getAvailableParents(type);
            
            if (parents.length === 0) {
                const parentType = {
                    'olt': 'Server',
                    'odc': 'OLT',
                    'odp': 'ODC',
                    'pelanggan': 'ODP'
                }[type];
                
                this.showToast(`Belum ada ${parentType}. Tambahkan ${parentType} terlebih dahulu!`, 5000);
                return;
            }

            parentGroup.style.display = 'block';
            
            const parentTypeLabel = {
                'olt': 'Server',
                'odc': 'OLT',
                'odp': 'ODC',
                'pelanggan': 'ODP'
            }[type];
            
            document.getElementById('formParentLabel').innerHTML = `${parentTypeLabel} <span class="required">*</span>`;
            
            parentSelect.innerHTML = '<option value="">-- Pilih --</option>' + 
                parents.map(p => `<option value="${p.id}" ${node && node.parentId === p.id ? 'selected' : ''}>${p.name}</option>`).join('');
        } else {
            parentGroup.style.display = 'none';
        }

        const pelangganFields = document.getElementById('pelangganExtraFields');
        if (type === 'pelanggan') {
            pelangganFields.style.display = 'block';
            document.getElementById('formPppoe').value = node ? node.pppoe || '' : '';
            document.getElementById('formProfile').value = node ? node.profile || '' : '';
            document.getElementById('formWhatsapp').value = node ? node.whatsapp || '' : '';
        } else {
            pelangganFields.style.display = 'none';
        }

        this.showModal();
    }

    async saveNode() {
        const form = document.getElementById('nodeForm');
        if (!form.checkValidity()) {
            form.reportValidity();
            return;
        }

        const nodeId = document.getElementById('formNodeId').value;
        const type = document.getElementById('formNodeType').value;
        const name = document.getElementById('formName').value.trim();
        const coordinates = document.getElementById('formCoordinates').value.trim();
        const parentId = document.getElementById('formParent').value;

        if (!name || !coordinates) {
            this.showToast('Mohon lengkapi semua field yang diperlukan', 3000);
            return;
        }

        if (type !== 'server' && !parentId) {
            this.showToast('Mohon pilih parent node', 3000);
            return;
        }

        const coords = this.parseCoordinates(coordinates);
        if (!coords) {
            this.showToast('Format koordinat salah! Gunakan format: -latitude,longitude', 4000);
            return;
        }

        const { lat, lng } = coords;

        if (!this.isValidCoordinate(lat, lng)) {
            this.showToast('Koordinat tidak valid! Latitude: -90 s/d 90, Longitude: -180 s/d 180', 4000);
            return;
        }

        const data = {
            name: name,
            lat: parseFloat(lat),
            lng: parseFloat(lng),
            parentId: parentId || null
        };

        if (type === 'pelanggan') {
            data.pppoe = document.getElementById('formPppoe').value.trim();
            data.profile = document.getElementById('formProfile').value.trim();
            data.whatsapp = document.getElementById('formWhatsapp').value.trim();
        }

        try {
            if (nodeId) {
                const updated = await this.dataManager.updateNode(nodeId, data);
                if (updated) {
                    this.mapController.updateMarker(updated);
                    this.showToast('Data berhasil diupdate', 3000);
                }
            } else {
                const node = await this.dataManager.addNode(type, data);
                this.mapController.addMarker(node);
                this.showToast('Data berhasil ditambahkan', 3000);
            }

            this.hideModal();
            this.pageController.refreshCurrentPage();
            this.mapController.redrawConnections();
        } catch (error) {
            this.showToast('Error: ' + error.message, 5000);
        }
    }

    editNode(nodeId) {
        const node = this.dataManager.getNode(nodeId);
        if (node) {
            this.showNodeForm(node.type, nodeId);
        }
    }

    async deleteNodeConfirm(nodeId) {
        const node = this.dataManager.getNode(nodeId);
        if (!node) return;

        const children = this.dataManager.getChildren(nodeId);
        
        let message = `Yakin ingin menghapus ${node.name}?`;
        if (children.length > 0) {
            message = `Node ini memiliki ${children.length} child node. Hapus child node terlebih dahulu!`;
            this.showToast(message, 5000);
            return;
        }

        if (confirm(message)) {
            const result = await this.dataManager.deleteNode(nodeId);
            if (result.success) {
                this.mapController.removeMarker(nodeId);
                this.showToast('Data berhasil dihapus', 3000);
                this.pageController.refreshCurrentPage();
            } else {
                this.showToast(result.message, 5000);
            }
        }
    }

    viewOnMap(nodeId) {
        this.pageController.showPage('map');
        
        setTimeout(() => {
            if (!this.mapController.map) {
                this.mapController.initMap('map');
            } else {
                this.mapController.map.invalidateSize();
            }
            this.mapController.panToNode(nodeId);
        }, 200);
    }

    pickCoordinateFromMap() {
        this.pageController.showPage('map');
        this.hideModal();
        
        setTimeout(() => {
            if (!this.mapController.map) {
                this.mapController.initMap('map');
            } else {
                this.mapController.map.invalidateSize();
            }

            this.showToast('Klik pada peta untuk memilih koordinat', 5000);
            
            this.mapController.enablePickCoordinate((lat, lng) => {
                document.getElementById('formCoordinates').value = `${lat.toFixed(6)},${lng.toFixed(6)}`;
                
                const type = document.getElementById('formNodeType').value;
                const nodeId = document.getElementById('formNodeId').value;
                
                this.showNodeForm(type, nodeId || null);
                this.showToast('Koordinat berhasil dipilih', 3000);
            });
        }, 200);
    }

    searchCoordinates() {
        const input = document.getElementById('searchCoordinates').value.trim();
        
        if (!input) {
            this.showToast('Masukkan koordinat terlebih dahulu', 3000);
            return;
        }

        const coords = this.parseCoordinates(input);
        
        if (!coords) {
            this.showToast('Format koordinat salah! Gunakan: -xxxxxxx,yxxxxxx', 4000);
            return;
        }

        const { lat, lng } = coords;

        if (!this.isValidCoordinate(lat, lng)) {
            this.showToast('Koordinat tidak valid! Latitude: -90 s/d 90, Longitude: -180 s/d 180', 4000);
            return;
        }

        if (!this.mapController.map) {
            this.mapController.initMap('map');
        }

        this.mapController.panToCoordinates(lat, lng);
        this.showToast(`Menampilkan lokasi: ${lat.toFixed(6)}, ${lng.toFixed(6)}`);
    }

    parseCoordinates(input) {
        const cleaned = input.replace(/\s+/g, '');
        const parts = cleaned.split(',');

        if (parts.length !== 2) {
            return null;
        }

        const lat = parseFloat(parts[0]);
        const lng = parseFloat(parts[1]);

        if (isNaN(lat) || isNaN(lng)) {
            return null;
        }

        return { lat, lng };
    }

    isValidCoordinate(lat, lng) {
        return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180;
    }

    async exportData() {
        try {
            const data = await this.dataManager.exportData();
            const json = JSON.stringify(data, null, 2);
            const blob = new Blob([json], { type: 'application/json' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `ftth-data-${new Date().getTime()}.json`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
            this.showToast('Data berhasil di-export', 3000);
        } catch (error) {
            this.showToast('Error: ' + error.message, 5000);
        }
    }

    async importData(file) {
        if (!file) return;

        const reader = new FileReader();
        reader.onload = async (e) => {
            try {
                const data = JSON.parse(e.target.result);
                await this.dataManager.importData(data);
                this.mapController.loadExistingData();
                this.pageController.refreshCurrentPage();
                this.showToast('Data berhasil di-import', 3000);
            } catch (error) {
                this.showToast('Error: ' + error.message, 4000);
            }
        };
        reader.readAsText(file);
    }

    clearAllData() {
        if (confirm('Yakin ingin menghapus SEMUA data? Tindakan ini tidak dapat dibatalkan!')) {
            if (confirm('Konfirmasi sekali lagi: Hapus semua data?')) {
                this.dataManager.clearAll();
                this.mapController.clearMap();
                this.pageController.refreshCurrentPage();
                this.showToast('Semua data berhasil dihapus', 3000);
            }
        }
    }

    downloadExcelTemplate() {
        const odps = this.dataManager.getNodesByType('odp');
        
        if (odps.length === 0) {
            this.showToast('Belum ada ODP. Tambahkan ODP terlebih dahulu!', 4000);
            return;
        }

        const odpNames = odps.map(odp => odp.name).join(', ');
        
        const templateData = [
            ['Nama', 'PPPOE', 'No WhatsApp', 'Koordinat', 'Nama ODP'],
            ['Contoh: Budi Santoso', 'user001@pppoe', '628123456789', '-6.123456,106.789012', odps[0].name],
            ['Contoh: Ani Wijaya', 'user002@pppoe', '628987654321', '-6.123789,106.789345', odps[0].name],
            [],
            ['PETUNJUK:'],
            ['1. Kolom Nama: Wajib diisi, nama pelanggan'],
            ['2. Kolom PPPOE: Opsional, username PPPOE pelanggan'],
            ['3. Kolom No WhatsApp: Opsional, nomor WA format 628xxx'],
            ['4. Kolom Koordinat: WAJIB diisi, format: -latitude,longitude (contoh: -6.123456,106.789012)'],
            ['5. Kolom Nama ODP: WAJIB diisi, pilih dari ODP yang tersedia'],
            [],
            ['ODP yang tersedia:', odpNames],
            [],
            ['CATATAN:'],
            ['- Hapus baris contoh sebelum upload'],
            ['- Pastikan format koordinat benar: -lat,lng tanpa spasi'],
            ['- Nama ODP harus sesuai dengan yang ada di sistem'],
            ['- File akan divalidasi saat upload']
        ];

        const ws = XLSX.utils.aoa_to_sheet(templateData);
        
        ws['!cols'] = [
            { wch: 25 },
            { wch: 20 },
            { wch: 20 },
            { wch: 30 },
            { wch: 25 }
        ];

        const headerStyle = {
            font: { bold: true, color: { rgb: "FFFFFF" } },
            fill: { fgColor: { rgb: "3498DB" } },
            alignment: { horizontal: "center", vertical: "center" }
        };

        ['A1', 'B1', 'C1', 'D1', 'E1'].forEach(cell => {
            if (ws[cell]) {
                ws[cell].s = headerStyle;
            }
        });

        const wb = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(wb, ws, 'Template Pelanggan');
        
        XLSX.writeFile(wb, `Template_Import_Pelanggan_${new Date().getTime()}.xlsx`);
        this.showToast('Template Excel berhasil didownload', 3000);
    }

    async importExcelPelanggan(file) {
        if (!file) return;

        if (!file.name.match(/\.(xlsx|xls)$/)) {
            this.showToast('File harus berformat Excel (.xlsx atau .xls)', 4000);
            return;
        }

        const reader = new FileReader();
        reader.onload = async (e) => {
            try {
                const data = new Uint8Array(e.target.result);
                const workbook = XLSX.read(data, { type: 'array' });
                
                const firstSheet = workbook.Sheets[workbook.SheetNames[0]];
                const jsonData = XLSX.utils.sheet_to_json(firstSheet, { header: 1 });
                
                if (jsonData.length < 2) {
                    this.showToast('File Excel kosong atau tidak valid', 4000);
                    return;
                }

                const headers = jsonData[0];
                const expectedHeaders = ['Nama', 'PPPOE', 'No WhatsApp', 'Koordinat', 'Nama ODP'];
                
                const headersMatch = expectedHeaders.every((header, index) => 
                    headers[index] && headers[index].toString().trim() === header
                );

                if (!headersMatch) {
                    this.showToast('Format file tidak sesuai template! Download template terlebih dahulu.', 5000);
                    return;
                }

                const rows = jsonData.slice(1).filter(row => 
                    row.length > 0 && row[0] && row[0].toString().trim() !== '' && !row[0].toString().startsWith('Contoh')
                );

                if (rows.length === 0) {
                    this.showToast('Tidak ada data pelanggan yang valid di file Excel', 4000);
                    return;
                }

                let successCount = 0;
                let errorCount = 0;
                const errors = [];

                for (let index = 0; index < rows.length; index++) {
                    const row = rows[index];
                    const rowNum = index + 2;
                    const nama = row[0] ? row[0].toString().trim() : '';
                    const pppoe = row[1] ? row[1].toString().trim() : '';
                    const whatsapp = row[2] ? row[2].toString().trim() : '';
                    const koordinat = row[3] ? row[3].toString().trim() : '';
                    const namaOdp = row[4] ? row[4].toString().trim() : '';

                    if (!nama || !koordinat || !namaOdp) {
                        errors.push(`Baris ${rowNum}: Data tidak lengkap (Nama, Koordinat, dan Nama ODP wajib diisi)`);
                        errorCount++;
                        return;
                    }

                    const odp = this.dataManager.getNodesByType('odp').find(o => o.name === namaOdp);
                    if (!odp) {
                        errors.push(`Baris ${rowNum}: ODP "${namaOdp}" tidak ditemukan`);
                        errorCount++;
                        return;
                    }

                    const coords = this.parseCoordinates(koordinat);
                    if (!coords) {
                        errors.push(`Baris ${rowNum}: Format koordinat salah "${koordinat}"`);
                        errorCount++;
                        return;
                    }

                    const { lat, lng } = coords;
                    if (!this.isValidCoordinate(lat, lng)) {
                        errors.push(`Baris ${rowNum}: Koordinat tidak valid "${koordinat}"`);
                        errorCount++;
                        return;
                    }

                    const pelangganData = {
                        name: nama,
                        lat: lat,
                        lng: lng,
                        parentId: odp.id,
                        pppoe: pppoe,
                        whatsapp: whatsapp
                    };

                    try {
                        const newNode = await this.dataManager.addNode('pelanggan', pelangganData);
                        
                        if (this.mapController.map) {
                            this.mapController.addMarker(newNode);
                        }
                        
                        successCount++;
                    } catch (err) {
                        errors.push(`Baris ${rowNum}: Gagal menambahkan data - ${err.message}`);
                        errorCount++;
                    }
                }

                document.getElementById('excelFileInput').value = '';

                if (errors.length > 0 && errors.length <= 5) {
                    console.error('Import errors:', errors);
                }

                if (this.mapController.map) {
                    this.mapController.redrawConnections();
                }
                
                this.pageController.refreshCurrentPage();

                if (successCount > 0 && errorCount === 0) {
                    this.showToast(`✓ Berhasil import ${successCount} pelanggan`, 4000);
                } else if (successCount > 0 && errorCount > 0) {
                    this.showToast(`⚠ Import selesai: ${successCount} berhasil, ${errorCount} gagal. Cek console untuk detail.`, 6000);
                } else {
                    this.showToast(`✗ Import gagal: ${errorCount} error. Periksa format data.`, 5000);
                }

            } catch (error) {
                console.error('Import error:', error);
                this.showToast('Error membaca file Excel: ' + error.message, 5000);
            }
        };
        
        reader.readAsArrayBuffer(file);
    }

    showModal() {
        this.modal.classList.add('show');
    }

    hideModal() {
        this.modal.classList.remove('show');
        document.getElementById('nodeForm').reset();
        this.currentEditNode = null;
        this.mapController.disablePickCoordinate();
    }

    showToast(message, duration = 3000) {
        const existingToast = document.querySelector('.toast');
        if (existingToast) {
            existingToast.remove();
        }

        const toast = document.createElement('div');
        toast.className = 'toast';
        toast.textContent = message;

        document.body.appendChild(toast);

        setTimeout(() => {
            toast.style.animation = 'slideOut 0.3s ease';
            setTimeout(() => {
                if (toast.parentNode) {
                    toast.remove();
                }
            }, 300);
        }, duration);
    }

    async startMikrotikStatusSync() {
        if (!this.mikrotikEnabled) {
            console.log('Mikrotik sync disabled');
            return;
        }

        await this.fetchMikrotikStatus();
        
        this.statusRefreshInterval = setInterval(() => {
            this.fetchMikrotikStatus();
        }, this.refreshIntervalSeconds * 1000);
        
        console.log(`Mikrotik status sync started (every ${this.refreshIntervalSeconds}s)`);
    }

    stopMikrotikStatusSync() {
        if (this.statusRefreshInterval) {
            clearInterval(this.statusRefreshInterval);
            this.statusRefreshInterval = null;
            console.log('Mikrotik status sync stopped');
        }
    }

    async fetchMikrotikStatus() {
        try {
            const response = await fetch(`${this.apiBase}/mikrotik/status`);
            const result = await response.json();

            if (result.success) {
                // Update last sync time
                const now = new Date();
                const timeStr = now.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
                const lastSyncEl = document.getElementById('lastSyncTime');
                if (lastSyncEl) {
                    lastSyncEl.textContent = timeStr;
                    lastSyncEl.style.color = '#27ae60';
                }

                await this.dataManager.loadAllData();
                
                if (this.mapController.map) {
                    this.dataManager.getNodesByType('pelanggan').forEach(pelanggan => {
                        if (this.mapController.markers[pelanggan.id]) {
                            this.mapController.updateMarker(pelanggan);
                        }
                    });
                }
                
                if (this.pageController.currentPage === 'pelanggan') {
                    this.pageController.renderPelangganTable();
                }
                
                if (this.pageController.currentPage === 'dashboard') {
                    this.pageController.renderDashboard();
                }
                
                console.log(`Status updated: ${result.data.active_connections} active connections, ${result.data.online_customers} customers online`);
            } else {
                console.error('Mikrotik API error:', result.error || 'Unknown error');
                const lastSyncEl = document.getElementById('lastSyncTime');
                if (lastSyncEl) {
                    lastSyncEl.textContent = 'Error';
                    lastSyncEl.style.color = '#e74c3c';
                }
            }
        } catch (error) {
            console.error('Failed to fetch Mikrotik status:', error);
            const lastSyncEl = document.getElementById('lastSyncTime');
            if (lastSyncEl) {
                lastSyncEl.textContent = 'Failed';
                lastSyncEl.style.color = '#e74c3c';
            }
        }
    }

    async loadMikrotikConfig() {
        try {
            const response = await fetch(`${this.apiBase}/mikrotik/config`);
            const result = await response.json();
            
            if (result.success && result.data) {
                const config = result.data;
                document.getElementById('mikrotikHost').value = config.host || '';
                document.getElementById('mikrotikPort').value = config.port || 8728;
                document.getElementById('mikrotikUser').value = config.user || '';
                document.getElementById('mikrotikPassword').value = '';
                
                this.showToast('Konfigurasi berhasil dimuat', 2000);
            } else {
                this.showToast('Gagal memuat konfigurasi', 3000);
            }
        } catch (error) {
            this.showToast('Error: ' + error.message, 4000);
        }
    }

    async saveMikrotikConfig() {
        const host = document.getElementById('mikrotikHost').value.trim();
        const port = parseInt(document.getElementById('mikrotikPort').value);
        const user = document.getElementById('mikrotikUser').value.trim();
        const password = document.getElementById('mikrotikPassword').value;

        if (!host || !port || !user) {
            this.showToast('Host, Port, dan Username wajib diisi!', 3000);
            return;
        }

        try {
            const response = await fetch(`${this.apiBase}/mikrotik/config`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    host: host,
                    port: port,
                    user: user,
                    password: password
                })
            });

            const result = await response.json();
            
            if (result.success) {
                this.showToast('✓ Konfigurasi berhasil disimpan!', 3000);
                document.getElementById('mikrotikPassword').value = '';
            } else {
                this.showToast('✗ Gagal menyimpan: ' + (result.error || 'Unknown error'), 4000);
            }
        } catch (error) {
            this.showToast('✗ Error: ' + error.message, 4000);
        }
    }

    async testMikrotikConnection() {
        const statusEl = document.getElementById('mikrotikStatus');
        statusEl.textContent = 'Testing...';
        statusEl.style.color = '#f39c12';
        
        try {
            const response = await fetch(`${this.apiBase}/mikrotik/test`);
            const result = await response.json();
            
            if (result.success) {
                statusEl.textContent = '✓ Connected';
                statusEl.style.color = '#27ae60';
                this.showToast('✓ Koneksi Mikrotik berhasil!', 3000);
                return true;
            } else {
                statusEl.textContent = '✗ Failed: ' + (result.error || 'Unknown error');
                statusEl.style.color = '#e74c3c';
                this.showToast(`✗ Koneksi gagal: ${result.error}`, 5000);
                return false;
            }
        } catch (error) {
            statusEl.textContent = '✗ Error: ' + error.message;
            statusEl.style.color = '#e74c3c';
            this.showToast('✗ Error: ' + error.message, 5000);
            return false;
        }
    }

    toggleMikrotikSync(enabled) {
        this.mikrotikEnabled = enabled;
        
        if (enabled) {
            this.startMikrotikStatusSync();
            this.showToast('Mikrotik sync diaktifkan', 3000);
        } else {
            this.stopMikrotikStatusSync();
            this.showToast('Mikrotik sync dinonaktifkan', 3000);
        }
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new App();
});
