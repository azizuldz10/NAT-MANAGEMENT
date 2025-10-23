/**
 * Advanced Search & Filter System
 * Features:
 * - Multi-field search (username, IP, caller ID, MAC, etc.)
 * - Advanced filters (router, status, time range)
 * - Saved filter presets
 * - Real-time filtering
 * - Filter history
 */

class AdvancedSearchFilter {
    constructor(options = {}) {
        // Configuration
        this.config = {
            searchFields: options.searchFields || ['username', 'ip_address', 'caller_id'],
            filterFields: options.filterFields || {},
            dataSource: options.dataSource || [],
            onFilter: options.onFilter || null,
            containerId: options.containerId || 'search-filter-container',
            storageKey: options.storageKey || 'nat_saved_filters',
            debounceDelay: options.debounceDelay || 300,
        };

        // State
        this.state = {
            searchQuery: '',
            activeFilters: {},
            savedFilters: [],
            filterHistory: [],
            currentResults: [],
            isOpen: false,
        };

        // Elements
        this.container = null;
        this.searchInput = null;
        this.filterPanel = null;

        // Debounce timer
        this.debounceTimer = null;

        // Initialize
        this.loadSavedFilters();
        this.createUI();
        this.bindEvents();

        console.log('✅ Advanced Search & Filter initialized');
    }

    // ============================
    // UI CREATION
    // ============================

    createUI() {
        const targetContainer = document.getElementById(this.config.containerId);
        if (!targetContainer) {
            console.error('❌ Advanced Search: Container not found with ID:', this.config.containerId);
            return;
        }

        console.log('✅ Advanced Search: Container found, creating UI...');

        // Create enhanced search interface
        this.container = document.createElement('div');
        this.container.className = 'advanced-search-container';
        this.container.innerHTML = `
            <div class="search-wrapper">
                <div class="search-input-group">
                    <span class="search-icon"><i class="fas fa-search"></i></span>
                    <input type="text" class="search-input" placeholder="Search username...">
                    <button class="search-clear-btn" style="display: none;">
                        <i class="fas fa-times"></i>
                    </button>
                    <button class="search-filter-toggle-btn" title="Advanced Filters">
                        <i class="fas fa-sliders-h"></i>
                        <span class="filter-badge" style="display: none;">0</span>
                    </button>
                </div>

                <!-- Filter Panel -->
                <div class="filter-panel" style="display: none;">
                    <div class="filter-panel-header">
                        <h6><i class="fas fa-filter"></i> Advanced Filters</h6>
                        <div class="filter-panel-actions">
                            <button class="btn-link filter-save-btn" title="Save Filter">
                                <i class="fas fa-save"></i> Save
                            </button>
                            <button class="btn-link filter-load-btn" title="Load Saved Filter">
                                <i class="fas fa-folder-open"></i> Load
                            </button>
                            <button class="btn-link filter-clear-btn" title="Clear All Filters">
                                <i class="fas fa-eraser"></i> Clear
                            </button>
                        </div>
                    </div>

                    <div class="filter-panel-body">
                        <!-- Router Filter -->
                        <div class="filter-group">
                            <label class="filter-label">
                                <i class="fas fa-server"></i> Router
                            </label>
                            <select class="filter-select" data-filter="router">
                                <option value="">All Routers</option>
                            </select>
                        </div>

                        <!-- Connection Status Filter -->
                        <div class="filter-group">
                            <label class="filter-label">
                                <i class="fas fa-signal"></i> Connection Status
                            </label>
                            <select class="filter-select" data-filter="status">
                                <option value="">All Status</option>
                                <option value="active">Active</option>
                                <option value="idle">Idle</option>
                            </select>
                        </div>

                        <!-- Uptime Filter -->
                        <div class="filter-group">
                            <label class="filter-label">
                                <i class="fas fa-clock"></i> Minimum Uptime
                            </label>
                            <select class="filter-select" data-filter="uptime">
                                <option value="">Any</option>
                                <option value="300">5+ minutes</option>
                                <option value="1800">30+ minutes</option>
                                <option value="3600">1+ hour</option>
                                <option value="21600">6+ hours</option>
                                <option value="86400">1+ day</option>
                            </select>
                        </div>

                        <!-- IP Range Filter -->
                        <div class="filter-group">
                            <label class="filter-label">
                                <i class="fas fa-network-wired"></i> IP Range
                            </label>
                            <div class="filter-input-group">
                                <input type="text" class="filter-input" data-filter="ip_start" placeholder="From (e.g., 192.168.1.1)">
                                <span class="filter-separator">to</span>
                                <input type="text" class="filter-input" data-filter="ip_end" placeholder="To (e.g., 192.168.1.255)">
                            </div>
                        </div>

                        <!-- Active Filters Display -->
                        <div class="active-filters" style="display: none;">
                            <label class="filter-label">Active Filters:</label>
                            <div class="active-filters-tags"></div>
                        </div>
                    </div>

                    <!-- Saved Filters List -->
                    <div class="saved-filters-list" style="display: none;">
                        <div class="saved-filters-header">
                            <h6><i class="fas fa-bookmark"></i> Saved Filters</h6>
                            <button class="btn-link saved-filters-close">
                                <i class="fas fa-times"></i>
                            </button>
                        </div>
                        <div class="saved-filters-body"></div>
                    </div>
                </div>

                <!-- Results Summary -->
                <div class="search-results-summary">
                    <span class="results-count">
                        <i class="fas fa-users"></i>
                        <strong id="filtered-count">0</strong> of <span id="total-count">0</span> clients
                    </span>
                    <button class="btn-link export-results-btn" title="Export Results">
                        <i class="fas fa-download"></i> Export
                    </button>
                </div>
            </div>
        `;

        targetContainer.innerHTML = '';
        targetContainer.appendChild(this.container);

        // Store element references
        this.searchInput = this.container.querySelector('.search-input');
        this.filterPanel = this.container.querySelector('.filter-panel');
        this.clearBtn = this.container.querySelector('.search-clear-btn');
        this.filterToggleBtn = this.container.querySelector('.search-filter-toggle-btn');
        this.filterBadge = this.container.querySelector('.filter-badge');
        this.activeFiltersContainer = this.container.querySelector('.active-filters');
        this.activeFiltersTags = this.container.querySelector('.active-filters-tags');
        this.savedFiltersList = this.container.querySelector('.saved-filters-list');

        console.log('✅ Advanced Search: UI created successfully!');
        console.log('   - Search Input:', !!this.searchInput);
        console.log('   - Clear Button:', !!this.clearBtn);
        console.log('   - Filter Toggle:', !!this.filterToggleBtn);
    }

    // ============================
    // EVENT BINDING
    // ============================

    bindEvents() {
        // Validate that UI elements exist before binding
        if (!this.searchInput || !this.clearBtn || !this.filterToggleBtn) {
            console.error('❌ Advanced Search: UI elements not initialized, skipping event binding');
            return;
        }

        // Search input
        this.searchInput.addEventListener('input', (e) => {
            this.handleSearchInput(e.target.value);
        });

        // Clear button
        this.clearBtn.addEventListener('click', () => {
            this.clearSearch();
        });

        // Filter toggle
        this.filterToggleBtn.addEventListener('click', () => {
            this.toggleFilterPanel();
        });

        // Filter selects and inputs
        const filterSelects = this.container.querySelectorAll('.filter-select');
        filterSelects.forEach(select => {
            select.addEventListener('change', (e) => {
                this.handleFilterChange(e.target.dataset.filter, e.target.value);
            });
        });

        const filterInputs = this.container.querySelectorAll('.filter-input');
        filterInputs.forEach(input => {
            input.addEventListener('input', (e) => {
                this.handleFilterChange(e.target.dataset.filter, e.target.value);
            });
        });

        // Filter actions
        this.container.querySelector('.filter-clear-btn').addEventListener('click', () => {
            this.clearAllFilters();
        });

        this.container.querySelector('.filter-save-btn').addEventListener('click', () => {
            this.showSaveFilterDialog();
        });

        this.container.querySelector('.filter-load-btn').addEventListener('click', () => {
            this.showSavedFilters();
        });

        // Close saved filters list
        this.container.querySelector('.saved-filters-close').addEventListener('click', () => {
            this.hideSavedFilters();
        });

        // Export results
        this.container.querySelector('.export-results-btn').addEventListener('click', () => {
            this.exportResults();
        });

        // Close filter panel when clicking outside
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.advanced-search-container')) {
                if (this.state.isOpen) {
                    this.hideFilterPanel();
                }
            }
        });
    }

    // ============================
    // SEARCH FUNCTIONALITY
    // ============================

    handleSearchInput(query) {
        this.state.searchQuery = query.toLowerCase();

        // Show/hide clear button
        if (query) {
            this.clearBtn.style.display = 'block';
        } else {
            this.clearBtn.style.display = 'none';
        }

        // Debounced search
        clearTimeout(this.debounceTimer);
        this.debounceTimer = setTimeout(() => {
            this.performFilter();
        }, this.config.debounceDelay);
    }

    clearSearch() {
        this.searchInput.value = '';
        this.state.searchQuery = '';
        this.clearBtn.style.display = 'none';
        this.performFilter();
    }

    // ============================
    // FILTER FUNCTIONALITY
    // ============================

    handleFilterChange(filterName, value) {
        if (value) {
            this.state.activeFilters[filterName] = value;
        } else {
            delete this.state.activeFilters[filterName];
        }

        this.updateActiveFiltersDisplay();
        this.updateFilterBadge();
        this.performFilter();
    }

    clearAllFilters() {
        this.state.activeFilters = {};
        this.state.searchQuery = '';
        this.searchInput.value = '';
        this.clearBtn.style.display = 'none';

        // Clear all filter inputs
        const filterSelects = this.container.querySelectorAll('.filter-select');
        filterSelects.forEach(select => select.value = '');

        const filterInputs = this.container.querySelectorAll('.filter-input');
        filterInputs.forEach(input => input.value = '');

        this.updateActiveFiltersDisplay();
        this.updateFilterBadge();
        this.performFilter();
    }

    toggleFilterPanel() {
        if (this.state.isOpen) {
            this.hideFilterPanel();
        } else {
            this.showFilterPanel();
        }
    }

    showFilterPanel() {
        this.filterPanel.style.display = 'block';
        this.state.isOpen = true;
        this.filterToggleBtn.classList.add('active');
    }

    hideFilterPanel() {
        this.filterPanel.style.display = 'none';
        this.state.isOpen = false;
        this.filterToggleBtn.classList.remove('active');
        this.hideSavedFilters();
    }

    updateFilterBadge() {
        const filterCount = Object.keys(this.state.activeFilters).length;

        if (filterCount > 0) {
            this.filterBadge.textContent = filterCount;
            this.filterBadge.style.display = 'block';
        } else {
            this.filterBadge.style.display = 'none';
        }
    }

    updateActiveFiltersDisplay() {
        const filterCount = Object.keys(this.state.activeFilters).length;

        if (filterCount === 0) {
            this.activeFiltersContainer.style.display = 'none';
            return;
        }

        this.activeFiltersContainer.style.display = 'block';

        // Create filter tags
        let tagsHTML = '';
        for (const [key, value] of Object.entries(this.state.activeFilters)) {
            const displayValue = this.getFilterDisplayValue(key, value);
            tagsHTML += `
                <span class="filter-tag">
                    ${displayValue}
                    <button class="filter-tag-remove" data-filter="${key}">
                        <i class="fas fa-times"></i>
                    </button>
                </span>
            `;
        }

        this.activeFiltersTags.innerHTML = tagsHTML;

        // Bind remove events
        const removeBtns = this.activeFiltersTags.querySelectorAll('.filter-tag-remove');
        removeBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const filterName = e.currentTarget.dataset.filter;
                delete this.state.activeFilters[filterName];

                // Clear the corresponding input
                const input = this.container.querySelector(`[data-filter="${filterName}"]`);
                if (input) input.value = '';

                this.updateActiveFiltersDisplay();
                this.updateFilterBadge();
                this.performFilter();
            });
        });
    }

    getFilterDisplayValue(key, value) {
        const filterLabels = {
            router: 'Router',
            status: 'Status',
            uptime: 'Uptime',
            ip_start: 'IP From',
            ip_end: 'IP To',
        };

        const label = filterLabels[key] || key;
        return `${label}: ${value}`;
    }

    // ============================
    // FILTERING LOGIC
    // ============================

    performFilter() {
        const data = this.config.dataSource;

        // Apply search query
        let filtered = data;

        if (this.state.searchQuery) {
            filtered = filtered.filter(item => {
                return this.config.searchFields.some(field => {
                    const value = item[field];
                    if (!value) return false;
                    return value.toString().toLowerCase().includes(this.state.searchQuery);
                });
            });
        }

        // Apply filters
        for (const [filterKey, filterValue] of Object.entries(this.state.activeFilters)) {
            filtered = this.applyFilter(filtered, filterKey, filterValue);
        }

        this.state.currentResults = filtered;
        this.updateResultsSummary(filtered.length, data.length);

        // Call callback
        if (this.config.onFilter && typeof this.config.onFilter === 'function') {
            this.config.onFilter(filtered);
        }

        return filtered;
    }

    applyFilter(data, filterKey, filterValue) {
        switch (filterKey) {
            case 'router':
                return data.filter(item => item.router === filterValue);

            case 'status':
                return data.filter(item => item.status === filterValue);

            case 'uptime':
                const minUptime = parseInt(filterValue);
                return data.filter(item => {
                    const uptime = this.parseUptime(item.uptime);
                    return uptime >= minUptime;
                });

            case 'ip_start':
            case 'ip_end':
                // IP range filtering
                if (this.state.activeFilters.ip_start && this.state.activeFilters.ip_end) {
                    return data.filter(item => {
                        return this.isIPInRange(
                            item.ip_address,
                            this.state.activeFilters.ip_start,
                            this.state.activeFilters.ip_end
                        );
                    });
                }
                return data;

            default:
                return data;
        }
    }

    parseUptime(uptimeStr) {
        if (!uptimeStr) return 0;

        // Parse formats like "1h23m45s", "45m12s", "30s"
        const parts = uptimeStr.match(/(\d+)([dhms])/g);
        if (!parts) return 0;

        let totalSeconds = 0;
        parts.forEach(part => {
            const value = parseInt(part);
            const unit = part.slice(-1);

            switch (unit) {
                case 'd': totalSeconds += value * 86400; break;
                case 'h': totalSeconds += value * 3600; break;
                case 'm': totalSeconds += value * 60; break;
                case 's': totalSeconds += value; break;
            }
        });

        return totalSeconds;
    }

    isIPInRange(ip, start, end) {
        const ipToNum = (ip) => {
            return ip.split('.').reduce((acc, octet) => (acc << 8) + parseInt(octet), 0) >>> 0;
        };

        const ipNum = ipToNum(ip);
        const startNum = ipToNum(start);
        const endNum = ipToNum(end);

        return ipNum >= startNum && ipNum <= endNum;
    }

    updateResultsSummary(filteredCount, totalCount) {
        const filteredCountEl = document.getElementById('filtered-count');
        const totalCountEl = document.getElementById('total-count');

        if (filteredCountEl) filteredCountEl.textContent = filteredCount;
        if (totalCountEl) totalCountEl.textContent = totalCount;
    }

    // ============================
    // SAVED FILTERS
    // ============================

    loadSavedFilters() {
        try {
            const saved = localStorage.getItem(this.config.storageKey);
            if (saved) {
                this.state.savedFilters = JSON.parse(saved);
            }
        } catch (error) {
            console.error('Failed to load saved filters:', error);
        }
    }

    saveFilters() {
        try {
            localStorage.setItem(this.config.storageKey, JSON.stringify(this.state.savedFilters));
        } catch (error) {
            console.error('Failed to save filters:', error);
        }
    }

    showSaveFilterDialog() {
        const filterCount = Object.keys(this.state.activeFilters).length;
        if (filterCount === 0 && !this.state.searchQuery) {
            alert('No filters to save');
            return;
        }

        const name = prompt('Enter a name for this filter:');
        if (!name) return;

        const filterPreset = {
            id: Date.now(),
            name: name,
            searchQuery: this.state.searchQuery,
            filters: { ...this.state.activeFilters },
            createdAt: new Date().toISOString(),
        };

        this.state.savedFilters.push(filterPreset);
        this.saveFilters();

        alert('Filter saved successfully!');
    }

    showSavedFilters() {
        const listBody = this.savedFiltersList.querySelector('.saved-filters-body');

        if (this.state.savedFilters.length === 0) {
            listBody.innerHTML = '<p class="text-muted text-center py-3">No saved filters</p>';
        } else {
            let html = '';
            this.state.savedFilters.forEach(filter => {
                html += `
                    <div class="saved-filter-item" data-id="${filter.id}">
                        <div class="saved-filter-info">
                            <div class="saved-filter-name">${filter.name}</div>
                            <div class="saved-filter-meta">
                                ${Object.keys(filter.filters).length} filters
                                ${filter.searchQuery ? '+ search' : ''}
                            </div>
                        </div>
                        <div class="saved-filter-actions">
                            <button class="btn-icon load-filter" title="Load">
                                <i class="fas fa-upload"></i>
                            </button>
                            <button class="btn-icon delete-filter" title="Delete">
                                <i class="fas fa-trash"></i>
                            </button>
                        </div>
                    </div>
                `;
            });
            listBody.innerHTML = html;

            // Bind events
            listBody.querySelectorAll('.load-filter').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    const id = parseInt(e.target.closest('.saved-filter-item').dataset.id);
                    this.loadFilter(id);
                });
            });

            listBody.querySelectorAll('.delete-filter').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    const id = parseInt(e.target.closest('.saved-filter-item').dataset.id);
                    this.deleteFilter(id);
                });
            });
        }

        this.savedFiltersList.style.display = 'block';
    }

    hideSavedFilters() {
        this.savedFiltersList.style.display = 'none';
    }

    loadFilter(id) {
        const filter = this.state.savedFilters.find(f => f.id === id);
        if (!filter) return;

        // Load search query
        this.searchInput.value = filter.searchQuery || '';
        this.state.searchQuery = filter.searchQuery || '';

        // Load filters
        this.state.activeFilters = { ...filter.filters };

        // Update UI
        for (const [key, value] of Object.entries(filter.filters)) {
            const input = this.container.querySelector(`[data-filter="${key}"]`);
            if (input) input.value = value;
        }

        this.updateActiveFiltersDisplay();
        this.updateFilterBadge();
        this.performFilter();
        this.hideSavedFilters();
    }

    deleteFilter(id) {
        if (!confirm('Delete this saved filter?')) return;

        this.state.savedFilters = this.state.savedFilters.filter(f => f.id !== id);
        this.saveFilters();
        this.showSavedFilters();
    }

    // ============================
    // EXPORT
    // ============================

    exportResults() {
        if (typeof window.exportToCSV === 'function') {
            window.exportToCSV(this.state.currentResults, 'filtered-clients.csv');
        } else {
            alert('Export functionality not available');
        }
    }

    // ============================
    // PUBLIC API
    // ============================

    setDataSource(data) {
        this.config.dataSource = data;
        this.performFilter();
    }

    updateRouterOptions(routers) {
        const routerSelect = this.container.querySelector('[data-filter="router"]');
        if (!routerSelect) return;

        // Keep current value
        const currentValue = routerSelect.value;

        // Clear and repopulate
        routerSelect.innerHTML = '<option value="">All Routers</option>';
        routers.forEach(router => {
            const option = document.createElement('option');
            option.value = router;
            option.textContent = router;
            routerSelect.appendChild(option);
        });

        // Restore value if still valid
        if (routers.includes(currentValue)) {
            routerSelect.value = currentValue;
        }
    }

    getResults() {
        return this.state.currentResults;
    }

    reset() {
        this.clearAllFilters();
        this.hideFilterPanel();
    }

    destroy() {
        if (this.container) {
            this.container.remove();
        }
    }
}

// Export for use in other scripts
if (typeof window !== 'undefined') {
    window.AdvancedSearchFilter = AdvancedSearchFilter;
}
