class DataManager {
    constructor() {
        this.apiBase = 'http://localhost:8080/api';
        this.nodes = [];
        this.connections = [];
        this.initialized = false;
    }

    async init() {
        if (!this.initialized) {
            await this.loadAllData();
            this.initialized = true;
        }
    }

    async loadAllData() {
        try {
            // Load routers
            const routersRes = await fetch(`${this.apiBase}/routers`);
            const routersData = await routersRes.json();
            
            // Load pelanggan
            const pelangganRes = await fetch(`${this.apiBase}/pelanggan`);
            const pelangganData = await pelangganRes.json();
            
            if (routersData.success && pelangganData.success) {
                // Convert API data to internal format
                this.nodes = [];
                
                // Add routers
                if (routersData.data) {
                    routersData.data.forEach(router => {
                        this.nodes.push(this.convertRouterToNode(router));
                    });
                }
                
                // Add pelanggan
                if (pelangganData.data) {
                    pelangganData.data.forEach(pelanggan => {
                        this.nodes.push(this.convertPelangganToNode(pelanggan));
                    });
                }
                
                // Build connections
                this.rebuildConnections();
            }
        } catch (error) {
            console.error('Error loading data:', error);
        }
    }

    convertRouterToNode(router) {
        const [lat, lng] = router.coordinates.split(',').map(c => parseFloat(c.trim()));
        return {
            id: router.id,
            type: router.type,
            name: router.name,
            lat: lat,
            lng: lng,
            parentId: router.parent_id,
            createdAt: router.created_at
        };
    }

    convertPelangganToNode(pelanggan) {
        const [lat, lng] = pelanggan.coordinates.split(',').map(c => parseFloat(c.trim()));
        return {
            id: pelanggan.id,
            type: 'pelanggan',
            name: pelanggan.name,
            lat: lat,
            lng: lng,
            parentId: pelanggan.odp_id,
            pppoe: pelanggan.pppoe || '',
            profile: pelanggan.profile || '',
            whatsapp: pelanggan.whatsapp || '',
            status: pelanggan.status || 'offline',
            createdAt: pelanggan.created_at
        };
    }

    rebuildConnections() {
        this.connections = [];
        this.nodes.forEach(node => {
            if (node.parentId) {
                this.connections.push({
                    id: `conn_${node.parentId}_${node.id}`,
                    from: node.parentId,
                    to: node.id
                });
            }
        });
    }

    async addNode(type, data) {
        try {
            const coordinates = `${data.lat},${data.lng}`;
            
            if (type === 'pelanggan') {
                const response = await fetch(`${this.apiBase}/pelanggan`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        name: data.name,
                        odp_id: data.parentId,
                        pppoe: data.pppoe || '',
                        profile: data.profile || '',
                        whatsapp: data.whatsapp || '',
                        coordinates: coordinates,
                        status: 'offline'
                    })
                });
                
                const result = await response.json();
                if (result.success) {
                    await this.loadAllData();
                    return this.getNode(result.data.id);
                }
                throw new Error(result.error || 'Failed to create pelanggan');
            } else {
                const response = await fetch(`${this.apiBase}/routers`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        name: data.name,
                        type: type,
                        parent_id: data.parentId || null,
                        coordinates: coordinates
                    })
                });
                
                const result = await response.json();
                if (result.success) {
                    await this.loadAllData();
                    return this.getNode(result.data.id);
                }
                throw new Error(result.error || 'Failed to create router');
            }
        } catch (error) {
            console.error('Error adding node:', error);
            throw error;
        }
    }

    async updateNode(id, updates) {
        try {
            const node = this.getNode(id);
            if (!node) return null;
            
            const coordinates = updates.lat && updates.lng ? 
                `${updates.lat},${updates.lng}` : 
                `${node.lat},${node.lng}`;
            
            if (node.type === 'pelanggan') {
                const response = await fetch(`${this.apiBase}/pelanggan/${id}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        name: updates.name || node.name,
                        odp_id: updates.parentId !== undefined ? updates.parentId : node.parentId,
                        pppoe: updates.pppoe !== undefined ? updates.pppoe : node.pppoe,
                        profile: updates.profile !== undefined ? updates.profile : node.profile,
                        whatsapp: updates.whatsapp !== undefined ? updates.whatsapp : node.whatsapp,
                        coordinates: coordinates,
                        status: updates.status || node.status
                    })
                });
                
                const result = await response.json();
                if (result.success) {
                    await this.loadAllData();
                    return this.getNode(id);
                }
                throw new Error(result.error || 'Failed to update pelanggan');
            } else {
                const response = await fetch(`${this.apiBase}/routers/${id}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        name: updates.name || node.name,
                        type: node.type,
                        parent_id: updates.parentId !== undefined ? updates.parentId : node.parentId,
                        coordinates: coordinates
                    })
                });
                
                const result = await response.json();
                if (result.success) {
                    await this.loadAllData();
                    return this.getNode(id);
                }
                throw new Error(result.error || 'Failed to update router');
            }
        } catch (error) {
            console.error('Error updating node:', error);
            throw error;
        }
    }

    async deleteNode(id) {
        try {
            const node = this.getNode(id);
            if (!node) {
                return {
                    success: false,
                    message: 'Node tidak ditemukan'
                };
            }
            
            const children = this.getChildren(id);
            if (children.length > 0) {
                return {
                    success: false,
                    message: `Tidak bisa menghapus node ini karena masih memiliki ${children.length} child node. Hapus child node terlebih dahulu.`
                };
            }
            
            const endpoint = node.type === 'pelanggan' ? 
                `${this.apiBase}/pelanggan/${id}` : 
                `${this.apiBase}/routers/${id}`;
            
            const response = await fetch(endpoint, { method: 'DELETE' });
            const result = await response.json();
            
            if (result.success) {
                await this.loadAllData();
                return { success: true };
            }
            
            return {
                success: false,
                message: result.error || 'Failed to delete node'
            };
        } catch (error) {
            console.error('Error deleting node:', error);
            return {
                success: false,
                message: error.message
            };
        }
    }

    getNode(id) {
        return this.nodes.find(n => n.id === id);
    }

    getAllNodes() {
        return this.nodes;
    }

    getNodesByType(type) {
        return this.nodes.filter(n => n.type === type);
    }

    getChildren(nodeId) {
        return this.nodes.filter(n => n.parentId === nodeId);
    }

    getParent(nodeId) {
        const node = this.getNode(nodeId);
        if (node && node.parentId) {
            return this.getNode(node.parentId);
        }
        return null;
    }

    getAvailableParents(childType) {
        const parentTypes = {
            'olt': 'server',
            'odc': 'olt',
            'odp': 'odc',
            'pelanggan': 'odp'
        };

        const parentType = parentTypes[childType];
        if (!parentType) return [];

        return this.getNodesByType(parentType);
    }

    getAllConnections() {
        return this.connections;
    }

    getNodeConnections(nodeId) {
        return this.connections.filter(c => c.from === nodeId || c.to === nodeId);
    }

    async clearAll() {
        // This would need to delete all data via API
        // For now, just clear local cache
        this.nodes = [];
        this.connections = [];
    }

    async exportData() {
        try {
            const response = await fetch(`${this.apiBase}/export`);
            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Error exporting data:', error);
            throw error;
        }
    }

    async importData(data) {
        try {
            const response = await fetch(`${this.apiBase}/import`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            
            const result = await response.json();
            if (result.success) {
                await this.loadAllData();
                return true;
            }
            throw new Error(result.error || 'Import failed');
        } catch (error) {
            console.error('Error importing data:', error);
            throw error;
        }
    }

    getTypeLabel(type) {
        const labels = {
            'server': 'Server',
            'olt': 'OLT',
            'odc': 'ODC',
            'odp': 'ODP',
            'pelanggan': 'Pelanggan'
        };
        return labels[type] || type;
    }

    getTypeColor(type) {
        const colors = {
            'server': '#e74c3c',
            'olt': '#f39c12',
            'odc': '#3498db',
            'odp': '#9b59b6',
            'pelanggan': '#27ae60'
        };
        return colors[type] || '#95a5a6';
    }

    async getStats() {
        try {
            const response = await fetch(`${this.apiBase}/stats`);
            const result = await response.json();
            
            if (result.success) {
                const pelanggans = this.getNodesByType('pelanggan');
                const onlineCount = pelanggans.filter(p => p.status === 'online').length;
                
                return {
                    server: result.data.server_count,
                    olt: result.data.olt_count,
                    odc: result.data.odc_count,
                    odp: result.data.odp_count,
                    pelanggan: result.data.pelanggan_count,
                    pelangganOnline: onlineCount,
                    pelangganOffline: result.data.pelanggan_count - onlineCount,
                    total: result.data.server_count + result.data.olt_count + result.data.odc_count + result.data.odp_count + result.data.pelanggan_count
                };
            }
        } catch (error) {
            console.error('Error getting stats:', error);
        }
        
        // Fallback to local calculation
        const pelanggans = this.getNodesByType('pelanggan');
        const onlineCount = pelanggans.filter(p => p.status === 'online').length;
        
        return {
            server: this.getNodesByType('server').length,
            olt: this.getNodesByType('olt').length,
            odc: this.getNodesByType('odc').length,
            odp: this.getNodesByType('odp').length,
            pelanggan: pelanggans.length,
            pelangganOnline: onlineCount,
            pelangganOffline: pelanggans.length - onlineCount,
            total: this.nodes.length
        };
    }

    async updatePelangganStatus() {
        try {
            const response = await fetch(`${this.apiBase}/mikrotik/status`);
            const result = await response.json();
            
            if (result.success) {
                await this.loadAllData();
                return result.data.active_connections || 0;
            }
        } catch (error) {
            console.error('Error updating pelanggan status:', error);
        }
        return 0;
    }

    getPelangganStatusSummary() {
        const pelanggans = this.getNodesByType('pelanggan');
        const online = pelanggans.filter(p => p.status === 'online');
        const offline = pelanggans.filter(p => p.status === 'offline');
        
        return {
            total: pelanggans.length,
            online: online.length,
            offline: offline.length,
            withPPPOE: pelanggans.filter(p => p.pppoe && p.pppoe.trim() !== '').length
        };
    }
}
