/**
 * Data Export Utility
 * Supports multiple formats: CSV, Excel (XLSX), PDF
 * Features:
 * - Export with current filters applied
 * - Custom column selection
 * - Formatted output with headers
 * - Export presets
 */

class DataExporter {
    constructor(options = {}) {
        // Configuration
        this.config = {
            appName: options.appName || 'NAT Management',
            defaultFilename: options.defaultFilename || 'export',
            includeTimestamp: options.includeTimestamp !== false, // Default true
            dateFormat: options.dateFormat || 'YYYY-MM-DD_HH-mm-ss',
        };

        // Available columns for export
        this.availableColumns = [
            { key: 'router', label: 'Router', enabled: true },
            { key: 'username', label: 'Username', enabled: true },
            { key: 'ip_address', label: 'IP Address', enabled: true },
            { key: 'caller_id', label: 'Caller ID', enabled: true },
            { key: 'uptime', label: 'Uptime', enabled: true },
            { key: 'status', label: 'Status', enabled: false },
        ];

        console.log('✅ Data Exporter initialized');
    }

    // ============================
    // EXPORT TO CSV
    // ============================

    exportToCSV(data, filename = null) {
        if (!data || data.length === 0) {
            this.showError('No data to export');
            return;
        }

        // Get enabled columns
        const enabledColumns = this.availableColumns.filter(col => col.enabled);

        // Build CSV header
        const headers = enabledColumns.map(col => col.label);
        let csv = headers.join(',') + '\n';

        // Build CSV rows
        data.forEach(row => {
            const values = enabledColumns.map(col => {
                let value = row[col.key] || '';

                // Escape quotes and wrap in quotes if contains comma
                if (typeof value === 'string') {
                    value = value.replace(/"/g, '""');
                    if (value.includes(',') || value.includes('\n') || value.includes('"')) {
                        value = `"${value}"`;
                    }
                }

                return value;
            });

            csv += values.join(',') + '\n';
        });

        // Download CSV file
        const fullFilename = this.getFilename(filename || 'clients', 'csv');
        this.downloadFile(csv, fullFilename, 'text/csv;charset=utf-8;');

        console.log(`✅ Exported ${data.length} rows to CSV: ${fullFilename}`);
        return true;
    }

    // ============================
    // EXPORT TO EXCEL (XLSX)
    // ============================

    async exportToExcel(data, filename = null) {
        if (!data || data.length === 0) {
            this.showError('No data to export');
            return;
        }

        try {
            // Check if SheetJS library is available
            if (typeof XLSX === 'undefined') {
                console.warn('⚠️ SheetJS library not loaded, falling back to CSV');
                this.showWarning('Excel export requires SheetJS library. Exporting as CSV instead.');
                return this.exportToCSV(data, filename);
            }

            // Get enabled columns
            const enabledColumns = this.availableColumns.filter(col => col.enabled);

            // Build worksheet data
            const wsData = [];

            // Add header row
            const headers = enabledColumns.map(col => col.label);
            wsData.push(headers);

            // Add data rows
            data.forEach(row => {
                const values = enabledColumns.map(col => row[col.key] || '');
                wsData.push(values);
            });

            // Create worksheet
            const ws = XLSX.utils.aoa_to_sheet(wsData);

            // Set column widths
            const colWidths = enabledColumns.map(() => ({ wch: 20 }));
            ws['!cols'] = colWidths;

            // Create workbook
            const wb = XLSX.utils.book_new();
            XLSX.utils.book_append_sheet(wb, ws, 'Clients');

            // Add metadata
            wb.Props = {
                Title: this.config.appName + ' Export',
                Subject: 'Client Data',
                Author: this.config.appName,
                CreatedDate: new Date()
            };

            // Generate Excel file
            const fullFilename = this.getFilename(filename || 'clients', 'xlsx');
            XLSX.writeFile(wb, fullFilename);

            console.log(`✅ Exported ${data.length} rows to Excel: ${fullFilename}`);
            return true;
        } catch (error) {
            console.error('Excel export failed:', error);
            this.showError('Excel export failed. Falling back to CSV.');
            return this.exportToCSV(data, filename);
        }
    }

    // ============================
    // EXPORT TO PDF
    // ============================

    async exportToPDF(data, filename = null) {
        if (!data || data.length === 0) {
            this.showError('No data to export');
            return;
        }

        try {
            // Check if jsPDF library is available
            if (typeof jsPDF === 'undefined') {
                console.warn('⚠️ jsPDF library not loaded, falling back to CSV');
                this.showWarning('PDF export requires jsPDF library. Exporting as CSV instead.');
                return this.exportToCSV(data, filename);
            }

            // Initialize jsPDF
            const { jsPDF } = window.jspdf || window;
            const doc = new jsPDF('l', 'mm', 'a4'); // Landscape orientation

            // Get enabled columns
            const enabledColumns = this.availableColumns.filter(col => col.enabled);

            // Add title
            doc.setFontSize(18);
            doc.text(this.config.appName + ' - Client Report', 14, 15);

            // Add timestamp
            doc.setFontSize(10);
            doc.text('Generated: ' + new Date().toLocaleString(), 14, 22);

            // Add total count
            doc.text(`Total Clients: ${data.length}`, 14, 28);

            // Prepare table data
            const headers = [enabledColumns.map(col => col.label)];
            const body = data.map(row => {
                return enabledColumns.map(col => {
                    const value = row[col.key] || '';
                    return value.toString();
                });
            });

            // Add table using autoTable plugin
            if (doc.autoTable) {
                doc.autoTable({
                    startY: 35,
                    head: headers,
                    body: body,
                    theme: 'grid',
                    styles: {
                        fontSize: 8,
                        cellPadding: 2,
                    },
                    headStyles: {
                        fillColor: [102, 126, 234],
                        textColor: 255,
                        fontStyle: 'bold',
                    },
                    alternateRowStyles: {
                        fillColor: [245, 245, 245],
                    },
                    margin: { top: 35 },
                });
            } else {
                // Fallback if autoTable not available
                console.warn('⚠️ jsPDF autoTable plugin not loaded, falling back to CSV');
                this.showWarning('PDF export requires jsPDF autoTable plugin. Exporting as CSV instead.');
                return this.exportToCSV(data, filename);
            }

            // Save PDF
            const fullFilename = this.getFilename(filename || 'clients', 'pdf');
            doc.save(fullFilename);

            console.log(`✅ Exported ${data.length} rows to PDF: ${fullFilename}`);
            return true;
        } catch (error) {
            console.error('PDF export failed:', error);
            this.showError('PDF export failed. Falling back to CSV.');
            return this.exportToCSV(data, filename);
        }
    }

    // ============================
    // EXPORT DIALOG
    // ============================

    showExportDialog(data) {
        const dialog = document.createElement('div');
        dialog.className = 'export-modal-overlay';
        dialog.innerHTML = `
            <div class="export-modal">
                <div class="export-modal-header">
                    <h5><i class="fas fa-download"></i> Export Data</h5>
                    <button class="export-modal-close">
                        <i class="fas fa-times"></i>
                    </button>
                </div>

                <div class="export-modal-body">
                    <!-- Format Selection -->
                    <div class="export-section">
                        <label class="export-label">
                            <i class="fas fa-file-export"></i> Export Format
                        </label>
                        <div class="export-format-options">
                            <button class="export-format-btn active" data-format="csv">
                                <i class="fas fa-file-csv"></i>
                                <span class="format-name">CSV</span>
                                <span class="format-desc">Comma-separated values</span>
                            </button>
                            <button class="export-format-btn" data-format="xlsx">
                                <i class="fas fa-file-excel"></i>
                                <span class="format-name">Excel</span>
                                <span class="format-desc">Microsoft Excel format</span>
                            </button>
                            <button class="export-format-btn" data-format="pdf">
                                <i class="fas fa-file-pdf"></i>
                                <span class="format-name">PDF</span>
                                <span class="format-desc">Portable document</span>
                            </button>
                        </div>
                    </div>

                    <!-- Column Selection -->
                    <div class="export-section">
                        <label class="export-label">
                            <i class="fas fa-columns"></i> Columns to Export
                        </label>
                        <div class="export-columns">
                            ${this.availableColumns.map(col => `
                                <label class="export-column-item">
                                    <input type="checkbox"
                                           class="export-column-checkbox"
                                           data-column="${col.key}"
                                           ${col.enabled ? 'checked' : ''}>
                                    <span>${col.label}</span>
                                </label>
                            `).join('')}
                        </div>
                    </div>

                    <!-- Filename -->
                    <div class="export-section">
                        <label class="export-label">
                            <i class="fas fa-signature"></i> Filename
                        </label>
                        <div class="export-filename-group">
                            <input type="text"
                                   class="export-filename-input"
                                   value="${this.config.defaultFilename}"
                                   placeholder="Enter filename">
                            <label class="export-timestamp-checkbox">
                                <input type="checkbox" checked>
                                <span>Add timestamp</span>
                            </label>
                        </div>
                    </div>

                    <!-- Data Summary -->
                    <div class="export-summary">
                        <i class="fas fa-info-circle"></i>
                        <span>Ready to export <strong>${data.length}</strong> records</span>
                    </div>
                </div>

                <div class="export-modal-footer">
                    <button class="btn-secondary export-cancel-btn">Cancel</button>
                    <button class="btn-primary export-confirm-btn">
                        <i class="fas fa-download"></i> Export
                    </button>
                </div>
            </div>
        `;

        document.body.appendChild(dialog);

        // Bind events
        let selectedFormat = 'csv';

        // Format selection
        const formatBtns = dialog.querySelectorAll('.export-format-btn');
        formatBtns.forEach(btn => {
            btn.addEventListener('click', () => {
                formatBtns.forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                selectedFormat = btn.dataset.format;
            });
        });

        // Column checkboxes
        const columnCheckboxes = dialog.querySelectorAll('.export-column-checkbox');
        columnCheckboxes.forEach(checkbox => {
            checkbox.addEventListener('change', () => {
                const column = this.availableColumns.find(col => col.key === checkbox.dataset.column);
                if (column) {
                    column.enabled = checkbox.checked;
                }
            });
        });

        // Close handlers
        const closeDialog = () => dialog.remove();
        dialog.querySelector('.export-modal-close').addEventListener('click', closeDialog);
        dialog.querySelector('.export-cancel-btn').addEventListener('click', closeDialog);
        dialog.addEventListener('click', (e) => {
            if (e.target === dialog) closeDialog();
        });

        // Export handler
        dialog.querySelector('.export-confirm-btn').addEventListener('click', async () => {
            const filenameInput = dialog.querySelector('.export-filename-input');
            const addTimestamp = dialog.querySelector('.export-timestamp-checkbox input').checked;

            this.config.includeTimestamp = addTimestamp;
            const filename = filenameInput.value.trim() || this.config.defaultFilename;

            // Show loading state
            const confirmBtn = dialog.querySelector('.export-confirm-btn');
            confirmBtn.disabled = true;
            confirmBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Exporting...';

            try {
                // Perform export
                let success = false;
                switch (selectedFormat) {
                    case 'csv':
                        success = this.exportToCSV(data, filename);
                        break;
                    case 'xlsx':
                        success = await this.exportToExcel(data, filename);
                        break;
                    case 'pdf':
                        success = await this.exportToPDF(data, filename);
                        break;
                }

                if (success) {
                    this.showSuccess(`Data exported successfully as ${selectedFormat.toUpperCase()}`);
                    closeDialog();
                }
            } catch (error) {
                console.error('Export failed:', error);
                this.showError('Export failed. Please try again.');
                confirmBtn.disabled = false;
                confirmBtn.innerHTML = '<i class="fas fa-download"></i> Export';
            }
        });
    }

    // ============================
    // UTILITY METHODS
    // ============================

    getFilename(base, extension) {
        let filename = base;

        if (this.config.includeTimestamp) {
            const timestamp = this.formatDate(new Date());
            filename += '_' + timestamp;
        }

        filename += '.' + extension;
        return filename;
    }

    formatDate(date) {
        const pad = (n) => n.toString().padStart(2, '0');

        const year = date.getFullYear();
        const month = pad(date.getMonth() + 1);
        const day = pad(date.getDate());
        const hours = pad(date.getHours());
        const minutes = pad(date.getMinutes());
        const seconds = pad(date.getSeconds());

        return `${year}-${month}-${day}_${hours}-${minutes}-${seconds}`;
    }

    downloadFile(content, filename, mimeType) {
        const blob = new Blob([content], { type: mimeType });
        const url = URL.createObjectURL(blob);

        const link = document.createElement('a');
        link.href = url;
        link.download = filename;
        link.style.display = 'none';

        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);

        // Clean up
        setTimeout(() => URL.revokeObjectURL(url), 100);
    }

    // ============================
    // NOTIFICATIONS
    // ============================

    showSuccess(message) {
        if (window.Toast && typeof window.Toast.success === 'function') {
            window.Toast.success('Export Success', message);
        } else {
            alert(message);
        }
    }

    showError(message) {
        if (window.Toast && typeof window.Toast.error === 'function') {
            window.Toast.error('Export Error', message);
        } else {
            alert(message);
        }
    }

    showWarning(message) {
        if (window.Toast && typeof window.Toast.warning === 'function') {
            window.Toast.warning('Export Warning', message);
        } else {
            alert(message);
        }
    }

    // ============================
    // PUBLIC API
    // ============================

    setColumns(columns) {
        this.availableColumns = columns;
    }

    getColumns() {
        return this.availableColumns;
    }

    enableColumn(key) {
        const column = this.availableColumns.find(col => col.key === key);
        if (column) {
            column.enabled = true;
        }
    }

    disableColumn(key) {
        const column = this.availableColumns.find(col => col.key === key);
        if (column) {
            column.enabled = false;
        }
    }
}

// Make it globally available
if (typeof window !== 'undefined') {
    window.DataExporter = DataExporter;

    // Convenience function for quick CSV export
    window.exportToCSV = function(data, filename) {
        const exporter = new DataExporter();
        return exporter.exportToCSV(data, filename);
    };
}
