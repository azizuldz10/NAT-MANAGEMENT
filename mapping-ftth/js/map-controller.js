class MapController {
    constructor(dataManager) {
        this.dataManager = dataManager;
        this.map = null;
        this.markers = {};
        this.lines = {};
        this.lineAnimations = {};
        this.pickingCoordinate = false;
        this.pickCoordinateCallback = null;
        this.tempMarker = null;
    }

    initMap(elementId, center = [-6.2088, 106.8456], zoom = 13) {
        if (this.map) {
            return;
        }

        this.map = L.map(elementId).setView(center, zoom);

        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '© OpenStreetMap contributors',
            maxZoom: 19
        }).addTo(this.map);

        this.map.on('click', (e) => {
            if (this.pickingCoordinate && this.pickCoordinateCallback) {
                this.pickCoordinateCallback(e.latlng.lat, e.latlng.lng);
                this.disablePickCoordinate();
            }
        });

        this.loadExistingData();
    }

    loadExistingData() {
        this.clearMap();

        this.dataManager.getAllNodes().forEach(node => {
            this.addMarker(node);
        });

        this.dataManager.getAllConnections().forEach(conn => {
            this.drawLine(conn);
        });
    }

    addMarker(node) {
        if (this.markers[node.id]) {
            this.removeMarker(node.id);
        }

        let color = this.dataManager.getTypeColor(node.type);
        let pulseAnimation = '';
        
        if (node.type === 'pelanggan') {
            if (node.status === 'online') {
                color = '#27ae60';
                pulseAnimation = 'animation: pulse-online 2s infinite;';
            } else {
                color = '#95a5a6';
            }
        }
        
        const icon = L.divIcon({
            className: 'custom-marker',
            html: `<div style="
                background-color: ${color};
                width: 30px;
                height: 30px;
                border-radius: 50%;
                border: 3px solid white;
                box-shadow: 0 2px 5px rgba(0,0,0,0.3);
                display: flex;
                align-items: center;
                justify-content: center;
                color: white;
                font-weight: bold;
                font-size: 12px;
                ${pulseAnimation}
            ">${this.getTypeIcon(node.type)}</div>
            <style>
                @keyframes pulse-online {
                    0%, 100% { box-shadow: 0 2px 5px rgba(0,0,0,0.3); }
                    50% { box-shadow: 0 0 15px rgba(39, 174, 96, 0.8); }
                }
            </style>`,
            iconSize: [30, 30],
            iconAnchor: [15, 15]
        });

        const marker = L.marker([node.lat, node.lng], { icon: icon })
            .addTo(this.map)
            .bindPopup(this.createPopupContent(node));

        marker.nodeId = node.id;
        this.markers[node.id] = marker;
        return marker;
    }

    updateMarker(node) {
        this.addMarker(node);
        
        const connections = this.dataManager.getNodeConnections(node.id);
        connections.forEach(conn => {
            this.drawLine(conn);
        });
    }

    removeMarker(nodeId) {
        const marker = this.markers[nodeId];
        if (marker) {
            this.map.removeLayer(marker);
            delete this.markers[nodeId];
        }

        Object.keys(this.lines).forEach(key => {
            if (key.includes(nodeId)) {
                this.map.removeLayer(this.lines[key]);
                delete this.lines[key];
            }
        });
    }

    getTypeIcon(type) {
        const icons = {
            'server': 'S',
            'olt': 'O',
            'odc': 'C',
            'odp': 'P',
            'pelanggan': 'U'
        };
        return icons[type] || '?';
    }

    createPopupContent(node) {
        const parent = this.dataManager.getParent(node.id);
        const children = this.dataManager.getChildren(node.id);
        
        let content = `<div class="popup-header">${node.name}</div>`;
        content += `<div class="popup-info"><span class="popup-label">Tipe:</span> <span class="popup-value">${this.dataManager.getTypeLabel(node.type)}</span></div>`;
        
        if (parent) {
            content += `<div class="popup-info"><span class="popup-label">Parent:</span> <span class="popup-value">${parent.name}</span></div>`;
        }
        
        if (children.length > 0) {
            content += `<div class="popup-info"><span class="popup-label">Children:</span> <span class="popup-value">${children.length} node</span></div>`;
        }
        
        if (node.type === 'pelanggan') {
            const statusBadge = node.status === 'online' 
                ? '<span style="background: #27ae60; color: white; padding: 2px 8px; border-radius: 10px; font-size: 11px; font-weight: 600;">🟢 ONLINE</span>'
                : '<span style="background: #95a5a6; color: white; padding: 2px 8px; border-radius: 10px; font-size: 11px; font-weight: 600;">⚫ OFFLINE</span>';
            
            content += `<div class="popup-info"><span class="popup-label">Status:</span> ${statusBadge}</div>`;
            
            if (node.pppoe) {
                content += `<div class="popup-info"><span class="popup-label">PPPOE:</span> <span class="popup-value">${node.pppoe}</span></div>`;
            }
            if (node.profile) {
                const isIsolir = node.profile.toLowerCase().includes('isolir') || node.profile.toLowerCase().includes('isolasi');
                const profileStyle = isIsolir 
                    ? 'background: #e74c3c; color: white; padding: 2px 8px; border-radius: 10px; font-size: 11px; font-weight: 600;'
                    : '';
                const profileBadge = isIsolir 
                    ? `<span style="${profileStyle}">🔴 ${node.profile}</span>`
                    : `<span style="font-weight: 600;">${node.profile}</span>`;
                content += `<div class="popup-info"><span class="popup-label">Profile:</span> ${profileBadge}</div>`;
            }
            if (node.whatsapp) {
                content += `<div class="popup-info"><span class="popup-label">WhatsApp:</span> <span class="popup-value">${node.whatsapp}</span></div>`;
            }
            if (node.lastSeen && node.status === 'online') {
                const lastSeenDate = new Date(node.lastSeen);
                content += `<div class="popup-info"><span class="popup-label">Last Seen:</span> <span class="popup-value">${lastSeenDate.toLocaleString('id-ID')}</span></div>`;
            }
        }

        content += `<div class="popup-info"><span class="popup-label">Koordinat:</span> <span class="popup-value">${node.lat.toFixed(6)}, ${node.lng.toFixed(6)}</span></div>`;

        return content;
    }

    drawLine(connection) {
        const fromNode = this.dataManager.getNode(connection.from);
        const toNode = this.dataManager.getNode(connection.to);

        if (fromNode && toNode) {
            const lineKey1 = `${connection.from}_${connection.to}`;
            const lineKey2 = `${connection.to}_${connection.from}`;

            if (this.lines[lineKey1]) {
                this.map.removeLayer(this.lines[lineKey1]);
                delete this.lines[lineKey1];
            }
            if (this.lines[lineKey2]) {
                this.map.removeLayer(this.lines[lineKey2]);
                delete this.lines[lineKey2];
            }

            const line = L.polyline(
                [[fromNode.lat, fromNode.lng], [toNode.lat, toNode.lng]],
                {
                    color: '#3498db',
                    weight: 4,
                    opacity: 0.7,
                    dashArray: '10, 15',
                    className: 'animated-line'
                }
            ).addTo(this.map);

            const lineKey = `${connection.from}_${connection.to}`;
            this.lines[lineKey] = line;
            
            this.animateLine(line, lineKey);
        }
    }

    animateLine(line, lineKey) {
        let offset = 0;
        
        if (this.lineAnimations[lineKey]) {
            clearInterval(this.lineAnimations[lineKey]);
        }
        
        this.lineAnimations[lineKey] = setInterval(() => {
            offset += 1;
            if (offset > 25) offset = 0;
            
            const element = line.getElement();
            if (element) {
                element.style.strokeDashoffset = offset;
            }
        }, 50);
    }

    removeLine(fromId, toId) {
        const lineKey1 = `${fromId}_${toId}`;
        const lineKey2 = `${toId}_${fromId}`;

        if (this.lines[lineKey1]) {
            if (this.lineAnimations[lineKey1]) {
                clearInterval(this.lineAnimations[lineKey1]);
                delete this.lineAnimations[lineKey1];
            }
            this.map.removeLayer(this.lines[lineKey1]);
            delete this.lines[lineKey1];
        }
        if (this.lines[lineKey2]) {
            if (this.lineAnimations[lineKey2]) {
                clearInterval(this.lineAnimations[lineKey2]);
                delete this.lineAnimations[lineKey2];
            }
            this.map.removeLayer(this.lines[lineKey2]);
            delete this.lines[lineKey2];
        }
    }

    clearMap() {
        Object.values(this.markers).forEach(marker => {
            this.map.removeLayer(marker);
        });
        this.markers = {};

        Object.keys(this.lineAnimations).forEach(key => {
            clearInterval(this.lineAnimations[key]);
        });
        this.lineAnimations = {};

        Object.values(this.lines).forEach(line => {
            this.map.removeLayer(line);
        });
        this.lines = {};
    }

    panToNode(nodeId) {
        const node = this.dataManager.getNode(nodeId);
        if (node) {
            this.map.setView([node.lat, node.lng], 16, { animate: true });
            const marker = this.markers[nodeId];
            if (marker) {
                marker.openPopup();
            }
        }
    }

    panToCoordinates(lat, lng) {
        this.map.setView([lat, lng], 16, { animate: true });

        if (this.tempMarker) {
            this.map.removeLayer(this.tempMarker);
        }

        const icon = L.divIcon({
            className: 'temp-marker',
            html: `<div style="
                background-color: #e74c3c;
                width: 40px;
                height: 40px;
                border-radius: 50%;
                border: 4px solid white;
                box-shadow: 0 3px 10px rgba(0,0,0,0.5);
                display: flex;
                align-items: center;
                justify-content: center;
                color: white;
                font-weight: bold;
                font-size: 20px;
                animation: pulse 1.5s infinite;
            ">📍</div>
            <style>
                @keyframes pulse {
                    0%, 100% { transform: scale(1); }
                    50% { transform: scale(1.1); }
                }
            </style>`,
            iconSize: [40, 40],
            iconAnchor: [20, 20]
        });

        this.tempMarker = L.marker([lat, lng], { icon: icon })
            .addTo(this.map)
            .bindPopup(`
                <div class="popup-header">Koordinat</div>
                <div class="popup-info">
                    <span class="popup-label">Latitude:</span> 
                    <span class="popup-value">${lat.toFixed(6)}</span>
                </div>
                <div class="popup-info">
                    <span class="popup-label">Longitude:</span> 
                    <span class="popup-value">${lng.toFixed(6)}</span>
                </div>
            `)
            .openPopup();

        setTimeout(() => {
            if (this.tempMarker) {
                this.map.removeLayer(this.tempMarker);
                this.tempMarker = null;
            }
        }, 10000);
    }

    enablePickCoordinate(callback) {
        this.pickingCoordinate = true;
        this.pickCoordinateCallback = callback;
        this.map.getContainer().style.cursor = 'crosshair';
    }

    disablePickCoordinate() {
        this.pickingCoordinate = false;
        this.pickCoordinateCallback = null;
        this.map.getContainer().style.cursor = '';
    }

    redrawConnections() {
        Object.values(this.lines).forEach(line => {
            this.map.removeLayer(line);
        });
        this.lines = {};

        this.dataManager.getAllConnections().forEach(conn => {
            this.drawLine(conn);
        });
    }
}
