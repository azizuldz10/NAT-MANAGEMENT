/**
 * Keyboard Shortcuts Manager
 * Power user features with customizable key bindings
 * Features:
 * - Global shortcuts for common actions
 * - Context-aware shortcuts
 * - Visual help overlay
 * - Mac/Windows compatibility
 * - Customizable bindings
 */

class KeyboardShortcuts {
    constructor(options = {}) {
        // Configuration
        this.config = {
            enabled: options.enabled !== false,
            showHelp: options.showHelp !== false,
            customBindings: options.customBindings || {},
        };

        // Detect OS for modifier keys
        this.isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
        this.modKey = this.isMac ? 'metaKey' : 'ctrlKey';
        this.modKeyName = this.isMac ? '⌘' : 'Ctrl';

        // Default key bindings
        this.shortcuts = {
            // Navigation & Refresh
            'refresh': {
                key: 'r',
                mod: true,
                description: 'Refresh data',
                handler: () => this.triggerRefresh(),
                category: 'Navigation'
            },
            'focusSearch': {
                key: 'f',
                mod: true,
                description: 'Focus search box',
                handler: () => this.focusSearch(),
                category: 'Navigation'
            },
            'toggleFilter': {
                key: 'k',
                mod: true,
                description: 'Toggle advanced filters',
                handler: () => this.toggleFilters(),
                category: 'Navigation'
            },

            // Data Operations
            'export': {
                key: 'e',
                mod: true,
                description: 'Export data',
                handler: () => this.triggerExport(),
                category: 'Data'
            },
            'clearFilters': {
                key: 'x',
                mod: true,
                description: 'Clear all filters',
                handler: () => this.clearFilters(),
                category: 'Data'
            },

            // Auto-Refresh Controls
            'pauseRefresh': {
                key: 'p',
                mod: true,
                description: 'Pause/Resume auto-refresh',
                handler: () => this.toggleAutoRefresh(),
                category: 'Auto-Refresh'
            },
            'refreshNow': {
                key: 'n',
                mod: true,
                shift: true,
                description: 'Refresh immediately',
                handler: () => this.refreshNow(),
                category: 'Auto-Refresh'
            },

            // UI Controls
            'toggleSidebar': {
                key: 'b',
                mod: true,
                description: 'Toggle sidebar',
                handler: () => this.toggleSidebar(),
                category: 'UI'
            },
            'showHelp': {
                key: '?',
                mod: false,
                shift: true,
                description: 'Show keyboard shortcuts',
                handler: () => this.showHelpOverlay(),
                category: 'Help'
            },
            'closeModal': {
                key: 'Escape',
                mod: false,
                description: 'Close modals/dialogs',
                handler: () => this.closeTopModal(),
                category: 'UI'
            },

            // Quick Actions
            'selectAll': {
                key: 'a',
                mod: true,
                shift: true,
                description: 'Select all visible items',
                handler: () => this.selectAll(),
                category: 'Selection'
            }
        };

        // Merge custom bindings
        Object.assign(this.shortcuts, this.config.customBindings);

        // State
        this.helpVisible = false;
        this.helpOverlay = null;

        // Initialize
        if (this.config.enabled) {
            this.bindEvents();
            this.createHelpButton();
        }

        console.log('✅ Keyboard Shortcuts initialized');
    }

    // ============================
    // EVENT BINDING
    // ============================

    bindEvents() {
        document.addEventListener('keydown', (e) => {
            this.handleKeyPress(e);
        });
    }

    handleKeyPress(e) {
        // Don't trigger shortcuts when typing in inputs
        if (this.isInputFocused()) {
            // Allow Escape to work in inputs
            if (e.key === 'Escape') {
                e.target.blur();
                return;
            }
            // Allow some shortcuts in search box
            if (e.target.classList.contains('search-input') && e.key === 'Escape') {
                this.clearFilters();
                return;
            }
            return;
        }

        // Find matching shortcut
        for (const [name, shortcut] of Object.entries(this.shortcuts)) {
            if (this.matchesShortcut(e, shortcut)) {
                e.preventDefault();
                console.log(`⌨️ Shortcut triggered: ${name}`);

                try {
                    shortcut.handler();
                } catch (error) {
                    console.error(`Shortcut handler failed for ${name}:`, error);
                }

                break;
            }
        }
    }

    matchesShortcut(event, shortcut) {
        // Check if key matches
        if (event.key.toLowerCase() !== shortcut.key.toLowerCase()) {
            return false;
        }

        // Check modifier keys
        const modifierPressed = shortcut.mod ? event[this.modKey] : true;
        const shiftPressed = shortcut.shift ? event.shiftKey : !event.shiftKey;

        return modifierPressed && shiftPressed;
    }

    isInputFocused() {
        const activeElement = document.activeElement;
        if (!activeElement) return false;

        const tagName = activeElement.tagName.toLowerCase();
        return (
            tagName === 'input' ||
            tagName === 'textarea' ||
            tagName === 'select' ||
            activeElement.isContentEditable
        );
    }

    // ============================
    // SHORTCUT HANDLERS
    // ============================

    triggerRefresh() {
        if (typeof window.refreshData === 'function') {
            window.refreshData();
            this.showToast('Refreshing data...', 'info');
        }
    }

    focusSearch() {
        const searchInput = document.querySelector('.search-input, #searchInput');
        if (searchInput) {
            searchInput.focus();
            searchInput.select();
            this.showToast('Search focused', 'info');
        }
    }

    toggleFilters() {
        if (window.advancedSearch) {
            window.advancedSearch.toggleFilterPanel();
        } else {
            // Fallback: try to find filter toggle button
            const filterBtn = document.querySelector('.search-filter-toggle-btn');
            if (filterBtn) {
                filterBtn.click();
            }
        }
    }

    triggerExport() {
        const exportBtn = document.querySelector('.export-results-btn');
        if (exportBtn) {
            exportBtn.click();
            this.showToast('Opening export dialog...', 'info');
        }
    }

    clearFilters() {
        if (window.advancedSearch) {
            window.advancedSearch.clearAllFilters();
            this.showToast('Filters cleared', 'success');
        }
    }

    toggleAutoRefresh() {
        if (window.smartRefresh) {
            window.smartRefresh.toggle();
            const state = window.smartRefresh.getState();
            this.showToast(
                state.isPaused ? 'Auto-refresh paused' : 'Auto-refresh resumed',
                'info'
            );
        }
    }

    refreshNow() {
        if (window.smartRefresh) {
            window.smartRefresh.refreshNow();
            this.showToast('Refreshing now...', 'info');
        } else if (typeof window.refreshData === 'function') {
            window.refreshData();
        }
    }

    toggleSidebar() {
        const hamburgerBtn = document.getElementById('hamburgerBtn');
        if (hamburgerBtn) {
            hamburgerBtn.click();
        }
    }

    closeTopModal() {
        // Close export modal
        const exportModal = document.querySelector('.export-modal-overlay');
        if (exportModal) {
            exportModal.remove();
            return;
        }

        // Close help overlay
        if (this.helpVisible) {
            this.hideHelpOverlay();
            return;
        }

        // Close Bootstrap modals
        const modals = document.querySelectorAll('.modal.show');
        if (modals.length > 0) {
            const lastModal = modals[modals.length - 1];
            const closeBtn = lastModal.querySelector('.btn-close, [data-bs-dismiss="modal"]');
            if (closeBtn) {
                closeBtn.click();
            }
        }
    }

    selectAll() {
        this.showToast('Select all feature - Coming soon!', 'info');
        // This will be implemented with the Quick Actions Toolbar
    }

    // ============================
    // HELP OVERLAY
    // ============================

    createHelpButton() {
        const helpBtn = document.createElement('button');
        helpBtn.className = 'keyboard-shortcuts-help-btn';
        helpBtn.innerHTML = '<i class="fas fa-keyboard"></i>';
        helpBtn.title = 'Keyboard Shortcuts (Shift + ?)';
        helpBtn.addEventListener('click', () => this.showHelpOverlay());

        document.body.appendChild(helpBtn);
    }

    showHelpOverlay() {
        if (this.helpVisible) return;

        this.helpVisible = true;

        // Group shortcuts by category
        const categories = {};
        for (const [name, shortcut] of Object.entries(this.shortcuts)) {
            const category = shortcut.category || 'Other';
            if (!categories[category]) {
                categories[category] = [];
            }
            categories[category].push({ name, ...shortcut });
        }

        // Create overlay
        this.helpOverlay = document.createElement('div');
        this.helpOverlay.className = 'keyboard-shortcuts-overlay';
        this.helpOverlay.innerHTML = `
            <div class="keyboard-shortcuts-modal">
                <div class="keyboard-shortcuts-header">
                    <h3><i class="fas fa-keyboard"></i> Keyboard Shortcuts</h3>
                    <button class="keyboard-shortcuts-close">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <div class="keyboard-shortcuts-body">
                    ${Object.entries(categories).map(([category, shortcuts]) => `
                        <div class="shortcut-category">
                            <h4>${category}</h4>
                            <div class="shortcut-list">
                                ${shortcuts.map(shortcut => `
                                    <div class="shortcut-item">
                                        <span class="shortcut-description">${shortcut.description}</span>
                                        <div class="shortcut-keys">
                                            ${shortcut.mod ? `<kbd>${this.modKeyName}</kbd>` : ''}
                                            ${shortcut.shift ? '<kbd>Shift</kbd>' : ''}
                                            <kbd>${this.formatKey(shortcut.key)}</kbd>
                                        </div>
                                    </div>
                                `).join('')}
                            </div>
                        </div>
                    `).join('')}
                </div>
                <div class="keyboard-shortcuts-footer">
                    <small>Press <kbd>Esc</kbd> or click outside to close</small>
                </div>
            </div>
        `;

        document.body.appendChild(this.helpOverlay);

        // Bind close events
        this.helpOverlay.querySelector('.keyboard-shortcuts-close').addEventListener('click', () => {
            this.hideHelpOverlay();
        });

        this.helpOverlay.addEventListener('click', (e) => {
            if (e.target === this.helpOverlay) {
                this.hideHelpOverlay();
            }
        });
    }

    hideHelpOverlay() {
        if (this.helpOverlay) {
            this.helpOverlay.remove();
            this.helpOverlay = null;
        }
        this.helpVisible = false;
    }

    formatKey(key) {
        // Format special keys for display
        const keyMap = {
            'Escape': 'Esc',
            ' ': 'Space',
            'ArrowUp': '↑',
            'ArrowDown': '↓',
            'ArrowLeft': '←',
            'ArrowRight': '→'
        };

        return keyMap[key] || key.toUpperCase();
    }

    // ============================
    // UTILITIES
    // ============================

    showToast(message, type = 'info') {
        if (window.Toast && typeof window.Toast[type] === 'function') {
            window.Toast[type]('Shortcut', message);
        }
    }

    // ============================
    // PUBLIC API
    // ============================

    addShortcut(name, shortcut) {
        this.shortcuts[name] = shortcut;
    }

    removeShortcut(name) {
        delete this.shortcuts[name];
    }

    updateShortcut(name, updates) {
        if (this.shortcuts[name]) {
            Object.assign(this.shortcuts[name], updates);
        }
    }

    enable() {
        this.config.enabled = true;
        console.log('⌨️ Keyboard shortcuts enabled');
    }

    disable() {
        this.config.enabled = false;
        console.log('⌨️ Keyboard shortcuts disabled');
    }

    getShortcuts() {
        return this.shortcuts;
    }

    destroy() {
        if (this.helpOverlay) {
            this.helpOverlay.remove();
        }

        const helpBtn = document.querySelector('.keyboard-shortcuts-help-btn');
        if (helpBtn) {
            helpBtn.remove();
        }

        console.log('🗑️ Keyboard Shortcuts destroyed');
    }
}

// Make it globally available
if (typeof window !== 'undefined') {
    window.KeyboardShortcuts = KeyboardShortcuts;
}
