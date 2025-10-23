class PageController {
    constructor(dataManager) {
        this.dataManager = dataManager;
        this.currentPage = 'dashboard';
        this.setupNavigation();
    }

    setupNavigation() {
        document.querySelectorAll('.nav-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const page = item.dataset.page;
                if (page) {
                    this.showPage(page);
                }
            });
        });
    }

    showPage(pageName) {
        document.querySelectorAll('.page').forEach(page => {
            page.classList.remove('active');
        });

        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });

        const pageElement = document.getElementById(`page-${pageName}`);
        if (pageElement) {
            pageElement.classList.add('active');
        }

        const navItem = document.querySelector(`.nav-item[data-page="${pageName}"]`);
        if (navItem) {
            navItem.classList.add('active');
        }

        this.currentPage = pageName;

        switch (pageName) {
            case 'dashboard':
                this.renderDashboard();
                break;
            case 'server':
                this.renderServerTable();
                break;
            case 'olt':
                this.renderOltTable();
                break;
            case 'odc':
                this.renderOdcTable();
                break;
            case 'odp':
                this.renderOdpTable();
                break;
            case 'pelanggan':
                this.renderPelangganTable();
                break;
        }
    }

    renderDashboard() {
        const stats = this.dataManager.getStats();
        
        document.getElementById('stat-server').textContent = stats.server;
        document.getElementById('stat-olt').textContent = stats.olt;
        document.getElementById('stat-odc').textContent = stats.odc;
        document.getElementById('stat-odp').textContent = stats.odp;
        document.getElementById('stat-pelanggan').textContent = stats.pelanggan;
    }

    renderServerTable() {
        const tbody = document.getElementById('serverTableBody');
        const servers = this.dataManager.getNodesByType('server');

        if (servers.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="5" class="empty-state">
                        <i class="fas fa-server"></i>
                        <p>Belum ada data server</p>
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = servers.map(server => {
            const oltCount = this.dataManager.getChildren(server.id).length;
            const date = new Date(server.createdAt).toLocaleString('id-ID');
            
            return `
                <tr>
                    <td><strong>${server.name}</strong></td>
                    <td>${server.lat.toFixed(6)}, ${server.lng.toFixed(6)}</td>
                    <td><span class="badge badge-info">${oltCount} OLT</span></td>
                    <td>${date}</td>
                    <td class="table-actions">
                        <button class="btn btn-sm btn-primary" onclick="window.app.editNode('${server.id}')">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="window.app.deleteNodeConfirm('${server.id}')">
                            <i class="fas fa-trash"></i> Hapus
                        </button>
                        <button class="btn btn-sm btn-secondary" onclick="window.app.viewOnMap('${server.id}')">
                            <i class="fas fa-map-marker-alt"></i> Lihat
                        </button>
                    </td>
                </tr>
            `;
        }).join('');
    }

    renderOltTable() {
        const tbody = document.getElementById('oltTableBody');
        const olts = this.dataManager.getNodesByType('olt');

        if (olts.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="5" class="empty-state">
                        <i class="fas fa-hdd"></i>
                        <p>Belum ada data OLT</p>
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = olts.map(olt => {
            const parent = this.dataManager.getParent(olt.id);
            const odcCount = this.dataManager.getChildren(olt.id).length;
            
            return `
                <tr>
                    <td><strong>${olt.name}</strong></td>
                    <td>${parent ? parent.name : '-'}</td>
                    <td>${olt.lat.toFixed(6)}, ${olt.lng.toFixed(6)}</td>
                    <td><span class="badge badge-info">${odcCount} ODC</span></td>
                    <td class="table-actions">
                        <button class="btn btn-sm btn-primary" onclick="window.app.editNode('${olt.id}')">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="window.app.deleteNodeConfirm('${olt.id}')">
                            <i class="fas fa-trash"></i> Hapus
                        </button>
                        <button class="btn btn-sm btn-secondary" onclick="window.app.viewOnMap('${olt.id}')">
                            <i class="fas fa-map-marker-alt"></i> Lihat
                        </button>
                    </td>
                </tr>
            `;
        }).join('');
    }

    renderOdcTable() {
        const tbody = document.getElementById('odcTableBody');
        const odcs = this.dataManager.getNodesByType('odc');

        if (odcs.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="5" class="empty-state">
                        <i class="fas fa-box"></i>
                        <p>Belum ada data ODC</p>
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = odcs.map(odc => {
            const parent = this.dataManager.getParent(odc.id);
            const odpCount = this.dataManager.getChildren(odc.id).length;
            
            return `
                <tr>
                    <td><strong>${odc.name}</strong></td>
                    <td>${parent ? parent.name : '-'}</td>
                    <td>${odc.lat.toFixed(6)}, ${odc.lng.toFixed(6)}</td>
                    <td><span class="badge badge-info">${odpCount} ODP</span></td>
                    <td class="table-actions">
                        <button class="btn btn-sm btn-primary" onclick="window.app.editNode('${odc.id}')">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="window.app.deleteNodeConfirm('${odc.id}')">
                            <i class="fas fa-trash"></i> Hapus
                        </button>
                        <button class="btn btn-sm btn-secondary" onclick="window.app.viewOnMap('${odc.id}')">
                            <i class="fas fa-map-marker-alt"></i> Lihat
                        </button>
                    </td>
                </tr>
            `;
        }).join('');
    }

    renderOdpTable() {
        const tbody = document.getElementById('odpTableBody');
        const odps = this.dataManager.getNodesByType('odp');

        if (odps.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="5" class="empty-state">
                        <i class="fas fa-inbox"></i>
                        <p>Belum ada data ODP</p>
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = odps.map(odp => {
            const parent = this.dataManager.getParent(odp.id);
            const pelangganCount = this.dataManager.getChildren(odp.id).length;
            
            return `
                <tr>
                    <td><strong>${odp.name}</strong></td>
                    <td>${parent ? parent.name : '-'}</td>
                    <td>${odp.lat.toFixed(6)}, ${odp.lng.toFixed(6)}</td>
                    <td><span class="badge badge-success">${pelangganCount} Pelanggan</span></td>
                    <td class="table-actions">
                        <button class="btn btn-sm btn-primary" onclick="window.app.editNode('${odp.id}')">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="window.app.deleteNodeConfirm('${odp.id}')">
                            <i class="fas fa-trash"></i> Hapus
                        </button>
                        <button class="btn btn-sm btn-secondary" onclick="window.app.viewOnMap('${odp.id}')">
                            <i class="fas fa-map-marker-alt"></i> Lihat
                        </button>
                    </td>
                </tr>
            `;
        }).join('');
    }

    renderPelangganTable() {
        const tbody = document.getElementById('pelangganTableBody');
        const pelanggans = this.dataManager.getNodesByType('pelanggan');

        if (pelanggans.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="7" class="empty-state">
                        <i class="fas fa-users"></i>
                        <p>Belum ada data pelanggan</p>
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = pelanggans.map(pelanggan => {
            const parent = this.dataManager.getParent(pelanggan.id);
            const statusBadge = pelanggan.status === 'online' 
                ? '<span class="badge badge-success"><i class="fas fa-circle"></i> Online</span>'
                : '<span class="badge badge-danger"><i class="fas fa-circle"></i> Offline</span>';
            
            return `
                <tr>
                    <td><strong>${pelanggan.name}</strong></td>
                    <td>${statusBadge}</td>
                    <td>${parent ? parent.name : '-'}</td>
                    <td>${pelanggan.pppoe || '-'}</td>
                    <td>${pelanggan.profile || '-'}</td>
                    <td>${pelanggan.whatsapp || '-'}</td>
                    <td>${pelanggan.lat.toFixed(6)}, ${pelanggan.lng.toFixed(6)}</td>
                    <td class="table-actions">
                        <button class="btn btn-sm btn-primary" onclick="window.app.editNode('${pelanggan.id}')">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="window.app.deleteNodeConfirm('${pelanggan.id}')">
                            <i class="fas fa-trash"></i> Hapus
                        </button>
                        <button class="btn btn-sm btn-secondary" onclick="window.app.viewOnMap('${pelanggan.id}')">
                            <i class="fas fa-map-marker-alt"></i> Lihat
                        </button>
                    </td>
                </tr>
            `;
        }).join('');
    }

    refreshCurrentPage() {
        this.showPage(this.currentPage);
    }
}
