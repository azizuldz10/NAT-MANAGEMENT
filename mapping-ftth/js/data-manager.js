class DataManager {
    constructor() {
        this.nodes = this.loadFromStorage('ftth_nodes') || [];
        this.connections = this.loadFromStorage('ftth_connections') || [];
    }

    loadFromStorage(key) {
        try {
            const data = localStorage.getItem(key);
            return data ? JSON.parse(data) : null;
        } catch (e) {
            console.error('Error loading from storage:', e);
            return null;
        }
    }

    saveToStorage(key, data) {
        try {
            localStorage.setItem(key, JSON.stringify(data));
        } catch (e) {
            console.error('Error saving to storage:', e);
        }
    }

    generateId() {
        return 'node_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    addNode(type, data) {
        const node = {
            id: this.generateId(),
            type: type,
            name: data.name,
            lat: parseFloat(data.lat),
            lng: parseFloat(data.lng),
            parentId: data.parentId || null,
            createdAt: new Date().toISOString()
        };

        if (type === 'pelanggan') {
            node.pppoe = data.pppoe || '';
            node.whatsapp = data.whatsapp || '';
            node.status = 'offline';
            node.lastSeen = null;
        }

        this.nodes.push(node);
        this.saveToStorage('ftth_nodes', this.nodes);

        if (node.parentId) {
            this.addConnection(node.parentId, node.id);
        }

        return node;
    }

    updateNode(id, updates) {
        const index = this.nodes.findIndex(n => n.id === id);
        if (index !== -1) {
            const oldParentId = this.nodes[index].parentId;
            
            this.nodes[index] = { 
                ...this.nodes[index], 
                ...updates,
                lat: updates.lat ? parseFloat(updates.lat) : this.nodes[index].lat,
                lng: updates.lng ? parseFloat(updates.lng) : this.nodes[index].lng
            };
            
            if (updates.parentId !== undefined && oldParentId !== updates.parentId) {
                if (oldParentId) {
                    this.deleteConnection(oldParentId, id);
                }
                if (updates.parentId) {
                    this.addConnection(updates.parentId, id);
                }
            }
            
            this.saveToStorage('ftth_nodes', this.nodes);
            return this.nodes[index];
        }
        return null;
    }

    deleteNode(id) {
        const children = this.getChildren(id);
        
        if (children.length > 0) {
            return {
                success: false,
                message: `Tidak bisa menghapus node ini karena masih memiliki ${children.length} child node. Hapus child node terlebih dahulu.`
            };
        }

        this.nodes = this.nodes.filter(n => n.id !== id);
        this.connections = this.connections.filter(c => c.from !== id && c.to !== id);
        this.saveToStorage('ftth_nodes', this.nodes);
        this.saveToStorage('ftth_connections', this.connections);
        
        return { success: true };
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

    addConnection(fromId, toId) {
        const exists = this.connections.find(c => 
            (c.from === fromId && c.to === toId) || 
            (c.from === toId && c.to === fromId)
        );

        if (!exists) {
            const connection = {
                id: 'conn_' + Date.now(),
                from: fromId,
                to: toId,
                createdAt: new Date().toISOString()
            };
            this.connections.push(connection);
            this.saveToStorage('ftth_connections', this.connections);
            return connection;
        }
        return null;
    }

    deleteConnection(fromId, toId) {
        this.connections = this.connections.filter(c => 
            !((c.from === fromId && c.to === toId) || 
              (c.from === toId && c.to === fromId))
        );
        this.saveToStorage('ftth_connections', this.connections);
    }

    getAllConnections() {
        return this.connections;
    }

    getNodeConnections(nodeId) {
        return this.connections.filter(c => c.from === nodeId || c.to === nodeId);
    }

    clearAll() {
        this.nodes = [];
        this.connections = [];
        localStorage.removeItem('ftth_nodes');
        localStorage.removeItem('ftth_connections');
    }

    exportData() {
        return {
            nodes: this.nodes,
            connections: this.connections,
            exportDate: new Date().toISOString(),
            version: '2.0'
        };
    }

    importData(data) {
        if (data.nodes) this.nodes = data.nodes;
        if (data.connections) this.connections = data.connections;
        this.saveToStorage('ftth_nodes', this.nodes);
        this.saveToStorage('ftth_connections', this.connections);
        return true;
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

    getStats() {
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

    updatePelangganStatus(activeConnections) {
        const pelanggans = this.getNodesByType('pelanggan');
        const activePPPOE = activeConnections.map(conn => conn.name.toLowerCase());
        const timestamp = new Date().toISOString();
        
        let updatedCount = 0;
        
        pelanggans.forEach(pelanggan => {
            if (!pelanggan.pppoe) {
                return;
            }
            
            const isOnline = activePPPOE.includes(pelanggan.pppoe.toLowerCase());
            
            if (isOnline && pelanggan.status !== 'online') {
                pelanggan.status = 'online';
                pelanggan.lastSeen = timestamp;
                updatedCount++;
            } else if (!isOnline && pelanggan.status !== 'offline') {
                pelanggan.status = 'offline';
                updatedCount++;
            } else if (isOnline) {
                pelanggan.lastSeen = timestamp;
            }
        });
        
        if (updatedCount > 0) {
            this.saveToStorage('ftth_nodes', this.nodes);
        }
        
        return updatedCount;
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
