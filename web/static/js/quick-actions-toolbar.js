/**
 * Quick Actions Toolbar
 * Bulk operations and selection management for client table
 * Features:
 * - Multi-select rows with checkboxes
 * - Select all/none functionality
 * - Bulk disconnect clients
 * - Bulk export selected items
 * - Bulk set as NAT target
 * - Smooth animations and feedback
 */

class QuickActionsToolbar {
    constructor(options = {}) {
        // Configuration
        this.config = {
            tableSelector: options.tableSelector || '#clientsTable',
            toolbarId: options.toolbarId || 'quickActionsToolbar',
            enableBulkDisconnect: options.enableBulkDisconnect !== false,
            enableBulkExport: options.enableBulkExport !== false,
            enableBulkNATTarget: options.enableBulkNATTarget !== false,
            maxSelections: options.maxSelections || 100,
            confirmActions: options.confirmActions !== false,
        };

        // State
        this.state = {
            selectedRows: new Set(),
            allClientsData: [],
            isProcessing: false,
        };

        // Elements
        this.toolbar = null;
        this.table = null;
        this.selectAllCheckbox = null;

        // Initialize
        this.init();

        console.log('✅ Quick Actions Toolbar initialized');
    }

    // ============================
    // INITIALIZATION
    // ============================

    init() {
        this.createToolbar();
        this.injectCheckboxColumn();
        this.createSelectAllControl();
        this.bindEvents();
    }

    createToolbar() {
        // Create toolbar element
        this.toolbar = document.createElement('div');
        this.toolbar.id = this.config.toolbarId;
        this.toolbar.className = 'quick-actions-toolbar';
        this.toolbar.innerHTML = `
            <div class="toolbar-selection-info">
                <div class="selection-count">
                    <span class="selection-count-badge" id="selectionCountBadge">0</span>
                    <span class="selection-count-text">items selected</span>
                </div>
            </div>

            <div class="toolbar-actions">
                ${this.config.enableBulkDisconnect ? `
                    <button class="toolbar-btn toolbar-btn-danger" id="bulkDisconnectBtn" title="Disconnect selected clients">
                        <i class="fas fa-plug"></i>
                        <span class="toolbar-btn-text">Disconnect</span>
                    </button>
                ` : ''}

                ${this.config.enableBulkNATTarget ? `
                    <button class="toolbar-btn toolbar-btn-primary" id="bulkNATTargetBtn" title="Set as NAT target">
                        <i class="fas fa-bullseye"></i>
                        <span class="toolbar-btn-text">Set as Target</span>
                    </button>
                ` : ''}

                ${this.config.enableBulkExport ? `
                    <button class="toolbar-btn" id="bulkExportBtn" title="Export selected">
                        <i class="fas fa-download"></i>
                        <span class="toolbar-btn-text">Export</span>
                    </button>
                ` : ''}

                <div class="toolbar-divider"></div>

                <button class="toolbar-btn" id="clearSelectionBtn" title="Clear selection">
                    <i class="fas fa-times-circle"></i>
                    <span class="toolbar-btn-text">Clear</span>
                </button>

                <button class="toolbar-btn toolbar-btn-close" id="closeToolbarBtn" title="Close toolbar">
                    <i class="fas fa-chevron-down"></i>
                </button>
            </div>
        `;

        document.body.appendChild(this.toolbar);
    }

    injectCheckboxColumn() {
        // Find table
        this.table = document.querySelector(this.config.tableSelector);
        if (!this.table) {
            console.warn(`Table not found: ${this.config.tableSelector}`);
            return;
        }

        // Add checkbox header to thead
        const thead = this.table.querySelector('thead tr');
        if (thead) {
            const checkboxHeader = document.createElement('th');
            checkboxHeader.style.width = '40px';
            checkboxHeader.id = 'checkboxHeaderCell';
            thead.insertBefore(checkboxHeader, thead.firstChild);
        }
    }

    createSelectAllControl() {
        // Create select all wrapper above table
        const tableCard = this.table.closest('.card-body');
        if (!tableCard) return;

        const selectAllWrapper = document.createElement('div');
        selectAllWrapper.className = 'select-all-wrapper';
        selectAllWrapper.innerHTML = `
            <input type="checkbox" id="selectAllCheckbox" class="select-all-checkbox">
            <label for="selectAllCheckbox" class="select-all-label">
                <i class="fas fa-check-square me-1"></i>
                Select All Clients
            </label>
        `;

        // Insert before table
        const tableResponsive = tableCard.querySelector('.table-responsive');
        if (tableResponsive) {
            tableCard.insertBefore(selectAllWrapper, tableResponsive);
        }

        this.selectAllCheckbox = document.getElementById('selectAllCheckbox');
    }

    // ============================
    // EVENT BINDING
    // ============================

    bindEvents() {
        // Select all checkbox
        if (this.selectAllCheckbox) {
            this.selectAllCheckbox.addEventListener('change', (e) => {
                this.handleSelectAll(e.target.checked);
            });
        }

        // Toolbar button events
        document.getElementById('bulkDisconnectBtn')?.addEventListener('click', () => {
            this.handleBulkDisconnect();
        });

        document.getElementById('bulkNATTargetBtn')?.addEventListener('click', () => {
            this.handleBulkNATTarget();
        });

        document.getElementById('bulkExportBtn')?.addEventListener('click', () => {
            this.handleBulkExport();
        });

        document.getElementById('clearSelectionBtn')?.addEventListener('click', () => {
            this.clearSelection();
        });

        document.getElementById('closeToolbarBtn')?.addEventListener('click', () => {
            this.hideToolbar();
            this.clearSelection();
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            // Ctrl+A or Cmd+A to select all (when focused on table)
            if ((e.ctrlKey || e.metaKey) && e.key === 'a' && this.isTableFocused()) {
                e.preventDefault();
                this.handleSelectAll(true);
            }

            // Escape to clear selection
            if (e.key === 'Escape' && this.state.selectedRows.size > 0) {
                this.clearSelection();
            }
        });
    }

    isTableFocused() {
        const activeElement = document.activeElement;
        return this.table && this.table.contains(activeElement);
    }

    // ============================
    // SELECTION MANAGEMENT
    // ============================

    addCheckboxToRow(row) {
        // Check if already has checkbox
        if (row.querySelector('.client-row-checkbox')) return;

        // Get client data from row
        const routerName = row.dataset.router;
        const username = row.dataset.username;
        const ipAddress = row.dataset.ip;
        const callerId = row.dataset.caller;

        if (!routerName || !ipAddress) return;

        // Create checkbox cell
        const checkboxCell = document.createElement('td');
        checkboxCell.innerHTML = `
            <input type="checkbox"
                   class="client-row-checkbox"
                   data-router="${routerName}"
                   data-ip="${ipAddress}"
                   data-username="${username}"
                   data-caller="${callerId}">
        `;

        // Insert as first cell
        row.insertBefore(checkboxCell, row.firstChild);

        // Make row selectable
        row.classList.add('client-row-selectable');

        // Bind checkbox event
        const checkbox = checkboxCell.querySelector('.client-row-checkbox');
        checkbox.addEventListener('change', (e) => {
            this.handleRowSelection(row, e.target.checked);
        });

        // Click row to toggle checkbox
        row.addEventListener('click', (e) => {
            // Don't trigger if clicking on buttons or other inputs
            if (e.target.tagName === 'BUTTON' || e.target.tagName === 'INPUT') return;
            if (e.target.closest('button') || e.target.closest('input')) return;

            checkbox.checked = !checkbox.checked;
            checkbox.dispatchEvent(new Event('change'));
        });
    }

    handleRowSelection(row, isSelected) {
        const checkbox = row.querySelector('.client-row-checkbox');
        if (!checkbox) return;

        const rowData = {
            router: checkbox.dataset.router,
            ip: checkbox.dataset.ip,
            username: checkbox.dataset.username,
            caller: checkbox.dataset.caller,
            row: row,
        };

        const rowId = `${rowData.router}::${rowData.ip}`;

        if (isSelected) {
            // Check max selections
            if (this.state.selectedRows.size >= this.config.maxSelections) {
                checkbox.checked = false;
                this.showError(`Maximum ${this.config.maxSelections} selections allowed`);
                return;
            }

            this.state.selectedRows.set(rowId, rowData);
            row.classList.add('client-row-selected');
        } else {
            this.state.selectedRows.delete(rowId);
            row.classList.remove('client-row-selected');
        }

        this.updateToolbar();
        this.updateSelectAllCheckbox();
    }

    handleSelectAll(selectAll) {
        const rows = this.table.querySelectorAll('tbody tr');

        rows.forEach(row => {
            const checkbox = row.querySelector('.client-row-checkbox');
            if (!checkbox) return;

            if (selectAll && this.state.selectedRows.size < this.config.maxSelections) {
                checkbox.checked = true;
                this.handleRowSelection(row, true);
            } else if (!selectAll) {
                checkbox.checked = false;
                this.handleRowSelection(row, false);
            }
        });

        this.updateToolbar();
    }

    updateSelectAllCheckbox() {
        if (!this.selectAllCheckbox) return;

        const rows = this.table.querySelectorAll('tbody tr');
        const checkboxes = this.table.querySelectorAll('.client-row-checkbox');
        const checkedCount = this.table.querySelectorAll('.client-row-checkbox:checked').length;

        if (checkedCount === 0) {
            this.selectAllCheckbox.checked = false;
            this.selectAllCheckbox.indeterminate = false;
        } else if (checkedCount === checkboxes.length) {
            this.selectAllCheckbox.checked = true;
            this.selectAllCheckbox.indeterminate = false;
        } else {
            this.selectAllCheckbox.checked = false;
            this.selectAllCheckbox.indeterminate = true;
        }
    }

    clearSelection() {
        this.state.selectedRows.clear();

        // Uncheck all checkboxes
        const checkboxes = this.table.querySelectorAll('.client-row-checkbox');
        checkboxes.forEach(checkbox => {
            checkbox.checked = false;
        });

        // Remove selected class from rows
        const selectedRows = this.table.querySelectorAll('.client-row-selected');
        selectedRows.forEach(row => {
            row.classList.remove('client-row-selected');
        });

        this.updateToolbar();
        this.updateSelectAllCheckbox();
    }

    // ============================
    // TOOLBAR MANAGEMENT
    // ============================

    updateToolbar() {
        const count = this.state.selectedRows.size;
        const badge = document.getElementById('selectionCountBadge');

        if (badge) {
            badge.textContent = count;
            badge.classList.add('updated');
            setTimeout(() => badge.classList.remove('updated'), 300);
        }

        if (count > 0) {
            this.showToolbar();
        } else {
            this.hideToolbar();
        }
    }

    showToolbar() {
        if (!this.toolbar.classList.contains('active')) {
            this.toolbar.classList.add('active');

            // Add bounce animation on first show
            if (!this.toolbar.dataset.shown) {
                this.toolbar.classList.add('show-with-bounce');
                this.toolbar.dataset.shown = 'true';
                setTimeout(() => {
                    this.toolbar.classList.remove('show-with-bounce');
                }, 500);
            }
        }
    }

    hideToolbar() {
        this.toolbar.classList.remove('active');
    }

    // ============================
    // BULK ACTIONS
    // ============================

    async handleBulkDisconnect() {
        if (this.state.isProcessing) return;
        if (this.state.selectedRows.size === 0) return;

        // Confirmation
        if (this.config.confirmActions) {
            const confirmed = await this.showConfirmDialog(
                'Bulk Disconnect',
                `Are you sure you want to disconnect ${this.state.selectedRows.size} selected clients?`,
                'warning'
            );

            if (!confirmed) return;
        }

        this.state.isProcessing = true;
        this.setButtonLoading('bulkDisconnectBtn', true);

        const results = {
            success: [],
            failed: [],
        };

        // Process each selected client
        for (const [rowId, rowData] of this.state.selectedRows) {
            try {
                const response = await fetch('/api/nat/disconnect', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        router: rowData.router,
                        username: rowData.username || '',
                        ip_address: rowData.ip,
                    }),
                });

                const result = await response.json();

                if (result.status === 'success') {
                    results.success.push(rowData);
                } else {
                    results.failed.push({ ...rowData, error: result.message });
                }
            } catch (error) {
                results.failed.push({ ...rowData, error: error.message });
            }
        }

        // Show results
        this.showBulkActionResults('Disconnect', results);

        // Clear selection and refresh data
        this.clearSelection();
        if (typeof window.loadClients === 'function') {
            await window.loadClients();
        }

        this.state.isProcessing = false;
        this.setButtonLoading('bulkDisconnectBtn', false);
    }

    async handleBulkNATTarget() {
        if (this.state.isProcessing) return;
        if (this.state.selectedRows.size === 0) return;

        // Only allow single selection for NAT target
        if (this.state.selectedRows.size > 1) {
            this.showError('Please select only ONE client to set as NAT target');
            return;
        }

        const [selectedData] = this.state.selectedRows.values();

        // Use existing selectClientForNAT function
        if (typeof window.selectClientForNAT === 'function') {
            window.selectClientForNAT(selectedData.router, selectedData.ip);
            this.clearSelection();
        } else {
            this.showError('NAT target function not available');
        }
    }

    async handleBulkExport() {
        if (this.state.selectedRows.size === 0) return;

        // Prepare data for export
        const dataToExport = Array.from(this.state.selectedRows.values()).map(rowData => ({
            router: rowData.router,
            username: rowData.username,
            ip_address: rowData.ip,
            caller_id: rowData.caller,
        }));

        // Use existing data exporter if available
        if (window.dataExporter) {
            window.dataExporter.showExportDialog(dataToExport);
        } else {
            // Fallback: download as JSON
            this.downloadJSON(dataToExport, 'selected-clients.json');
        }
    }

    // ============================
    // UI HELPERS
    // ============================

    setButtonLoading(buttonId, isLoading) {
        const button = document.getElementById(buttonId);
        if (!button) return;

        if (isLoading) {
            button.classList.add('loading');
            button.disabled = true;
            const icon = button.querySelector('i');
            if (icon) {
                icon.dataset.originalClass = icon.className;
                icon.className = 'fas fa-spinner fa-spin';
            }
        } else {
            button.classList.remove('loading');
            button.disabled = false;
            const icon = button.querySelector('i');
            if (icon && icon.dataset.originalClass) {
                icon.className = icon.dataset.originalClass;
                delete icon.dataset.originalClass;
            }
        }
    }

    showConfirmDialog(title, message, type = 'info') {
        return new Promise((resolve) => {
            const confirmed = confirm(`${title}\n\n${message}`);
            resolve(confirmed);
        });
    }

    showBulkActionResults(actionName, results) {
        const totalSuccess = results.success.length;
        const totalFailed = results.failed.length;

        let message = `${actionName} completed:\n\n`;
        message += `✅ Success: ${totalSuccess}\n`;

        if (totalFailed > 0) {
            message += `❌ Failed: ${totalFailed}\n\n`;
            message += 'Failed items:\n';
            results.failed.slice(0, 5).forEach(item => {
                message += `- ${item.router} / ${item.ip}: ${item.error}\n`;
            });

            if (totalFailed > 5) {
                message += `... and ${totalFailed - 5} more`;
            }
        }

        if (totalSuccess > 0 && totalFailed === 0) {
            this.showSuccess(message);
        } else if (totalFailed > 0) {
            this.showWarning(message);
        }
    }

    showSuccess(message) {
        if (typeof Toast !== 'undefined') {
            Toast.success('Success', message);
        } else {
            alert(message);
        }
    }

    showError(message) {
        if (typeof Toast !== 'undefined') {
            Toast.error('Error', message);
        } else {
            alert('Error: ' + message);
        }
    }

    showWarning(message) {
        if (typeof Toast !== 'undefined') {
            Toast.warning('Warning', message);
        } else {
            alert('Warning: ' + message);
        }
    }

    downloadJSON(data, filename) {
        const json = JSON.stringify(data, null, 2);
        const blob = new Blob([json], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = filename;
        link.click();
        URL.revokeObjectURL(url);
    }

    // ============================
    // PUBLIC API
    // ============================

    // Refresh checkboxes when table is updated
    refreshCheckboxes() {
        const rows = this.table.querySelectorAll('tbody tr');

        rows.forEach(row => {
            this.addCheckboxToRow(row);
        });

        this.updateSelectAllCheckbox();
    }

    // Get selected items
    getSelectedItems() {
        return Array.from(this.state.selectedRows.values());
    }

    // Get selection count
    getSelectionCount() {
        return this.state.selectedRows.size;
    }

    // Programmatically select rows
    selectRows(rowIds) {
        rowIds.forEach(rowId => {
            const row = this.table.querySelector(`tr[data-router="${rowId.router}"][data-ip="${rowId.ip}"]`);
            if (row) {
                const checkbox = row.querySelector('.client-row-checkbox');
                if (checkbox && !checkbox.checked) {
                    checkbox.checked = true;
                    this.handleRowSelection(row, true);
                }
            }
        });
    }

    // Destroy instance
    destroy() {
        // Remove toolbar
        if (this.toolbar) {
            this.toolbar.remove();
        }

        // Remove checkboxes
        const checkboxCells = this.table.querySelectorAll('td:first-child');
        checkboxCells.forEach(cell => {
            if (cell.querySelector('.client-row-checkbox')) {
                cell.remove();
            }
        });

        // Remove checkbox header
        const checkboxHeader = document.getElementById('checkboxHeaderCell');
        if (checkboxHeader) {
            checkboxHeader.remove();
        }

        // Remove select all control
        const selectAllWrapper = document.querySelector('.select-all-wrapper');
        if (selectAllWrapper) {
            selectAllWrapper.remove();
        }

        // Clear state
        this.state.selectedRows.clear();

        console.log('🗑️ Quick Actions Toolbar destroyed');
    }
}

// Make it globally available
if (typeof window !== 'undefined') {
    window.QuickActionsToolbar = QuickActionsToolbar;
}
