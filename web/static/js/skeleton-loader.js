/**
 * Skeleton Loader System
 * Modern loading states with shimmer effects
 * Features:
 * - Multiple skeleton types (table, card, stats)
 * - Shimmer animation
 * - Adaptive to content structure
 * - Easy integration
 */

class SkeletonLoader {
    constructor(options = {}) {
        // Configuration
        this.config = {
            shimmerColor: options.shimmerColor || 'rgba(255, 255, 255, 0.8)',
            baseColor: options.baseColor || '#f0f0f0',
            animationDuration: options.animationDuration || '1.5s',
        };

        console.log('✅ Skeleton Loader initialized');
    }

    // ============================
    // TABLE SKELETON
    // ============================

    createTableSkeleton(rows = 5, columns = 5) {
        let html = '<tbody class="skeleton-loader">';

        for (let i = 0; i < rows; i++) {
            html += '<tr>';
            for (let j = 0; j < columns; j++) {
                html += `
                    <td>
                        <div class="skeleton-box" style="width: ${this.getRandomWidth(60, 100)}%"></div>
                    </td>
                `;
            }
            html += '</tr>';
        }

        html += '</tbody>';
        return html;
    }

    showTableSkeleton(tableSelector, rows = 5, columns = 5) {
        const table = document.querySelector(tableSelector);
        if (!table) {
            console.warn(`Table not found: ${tableSelector}`);
            return;
        }

        const tbody = table.querySelector('tbody');
        if (tbody) {
            tbody.outerHTML = this.createTableSkeleton(rows, columns);
        }
    }

    // ============================
    // CARD SKELETON
    // ============================

    createCardSkeleton(options = {}) {
        const {
            showImage = true,
            showTitle = true,
            showDescription = true,
            showActions = true,
            lines = 3
        } = options;

        return `
            <div class="skeleton-card">
                ${showImage ? '<div class="skeleton-image"></div>' : ''}
                <div class="skeleton-content">
                    ${showTitle ? '<div class="skeleton-title"></div>' : ''}
                    ${showDescription ? this.createTextLines(lines) : ''}
                </div>
                ${showActions ? '<div class="skeleton-actions"></div>' : ''}
            </div>
        `;
    }

    createCardsGrid(container, count = 6, options = {}) {
        if (typeof container === 'string') {
            container = document.querySelector(container);
        }

        if (!container) {
            console.warn('Container not found');
            return;
        }

        let html = '<div class="skeleton-grid">';
        for (let i = 0; i < count; i++) {
            html += this.createCardSkeleton(options);
        }
        html += '</div>';

        container.innerHTML = html;
    }

    // ============================
    // STATS SKELETON
    // ============================

    createStatsSkeleton(count = 4) {
        let html = '';

        for (let i = 0; i < count; i++) {
            html += `
                <div class="col-lg-3 col-md-6 col-sm-6 col-6">
                    <div class="skeleton-stat-card">
                        <div class="skeleton-stat-value"></div>
                        <div class="skeleton-stat-label"></div>
                    </div>
                </div>
            `;
        }

        return html;
    }

    showStatsSkeleton(containerSelector, count = 4) {
        const container = document.querySelector(containerSelector);
        if (!container) {
            console.warn(`Stats container not found: ${containerSelector}`);
            return;
        }

        container.innerHTML = this.createStatsSkeleton(count);
    }

    // ============================
    // NAT CONFIG SKELETON
    // ============================

    createNATConfigSkeleton(count = 3) {
        let html = '';

        for (let i = 0; i < count; i++) {
            html += `
                <div class="col-md-6 col-12">
                    <div class="skeleton-nat-card">
                        <div class="skeleton-nat-header">
                            <div class="skeleton-box" style="width: 40%; height: 20px;"></div>
                            <div class="skeleton-box" style="width: 60px; height: 30px; border-radius: 20px;"></div>
                        </div>
                        <div class="skeleton-nat-body">
                            ${this.createTextLines(4)}
                        </div>
                        <div class="skeleton-nat-footer">
                            <div class="skeleton-box" style="width: 80px; height: 32px; border-radius: 8px;"></div>
                            <div class="skeleton-box" style="width: 80px; height: 32px; border-radius: 8px;"></div>
                        </div>
                    </div>
                </div>
            `;
        }

        return html;
    }

    showNATConfigSkeleton(containerSelector, count = 3) {
        const container = document.querySelector(containerSelector);
        if (!container) {
            console.warn(`NAT config container not found: ${containerSelector}`);
            return;
        }

        container.innerHTML = this.createNATConfigSkeleton(count);
    }

    // ============================
    // SEARCH BOX SKELETON
    // ============================

    createSearchSkeleton() {
        return `
            <div class="skeleton-search">
                <div class="skeleton-box" style="width: 100%; height: 48px; border-radius: 12px;"></div>
            </div>
        `;
    }

    // ============================
    // UTILITY METHODS
    // ============================

    createTextLines(count = 3) {
        let html = '<div class="skeleton-text-lines">';

        for (let i = 0; i < count; i++) {
            const width = i === count - 1 ? this.getRandomWidth(50, 80) : this.getRandomWidth(80, 100);
            html += `<div class="skeleton-line" style="width: ${width}%"></div>`;
        }

        html += '</div>';
        return html;
    }

    getRandomWidth(min, max) {
        return Math.floor(Math.random() * (max - min + 1)) + min;
    }

    // ============================
    // SHOW/HIDE METHODS
    // ============================

    show(selector, type = 'table', options = {}) {
        switch (type) {
            case 'table':
                this.showTableSkeleton(selector, options.rows, options.columns);
                break;
            case 'stats':
                this.showStatsSkeleton(selector, options.count);
                break;
            case 'natconfig':
                this.showNATConfigSkeleton(selector, options.count);
                break;
            case 'cards':
                this.createCardsGrid(selector, options.count, options);
                break;
            default:
                console.warn(`Unknown skeleton type: ${type}`);
        }
    }

    hide(selector) {
        const container = document.querySelector(selector);
        if (container) {
            const skeletons = container.querySelectorAll('.skeleton-loader, .skeleton-grid, .skeleton-stat-card, .skeleton-nat-card, .skeleton-search');
            skeletons.forEach(skeleton => skeleton.remove());
        }
    }

    // ============================
    // LOADING OVERLAY
    // ============================

    showLoadingOverlay(containerSelector, message = 'Loading...') {
        const container = document.querySelector(containerSelector);
        if (!container) return;

        const overlay = document.createElement('div');
        overlay.className = 'skeleton-overlay';
        overlay.innerHTML = `
            <div class="skeleton-overlay-content">
                <div class="skeleton-spinner"></div>
                <p>${message}</p>
            </div>
        `;

        container.style.position = 'relative';
        container.appendChild(overlay);
    }

    hideLoadingOverlay(containerSelector) {
        const container = document.querySelector(containerSelector);
        if (!container) return;

        const overlay = container.querySelector('.skeleton-overlay');
        if (overlay) {
            overlay.remove();
        }
    }

    // ============================
    // INLINE LOADER
    // ============================

    createInlineLoader(text = 'Loading') {
        return `
            <div class="skeleton-inline-loader">
                <div class="skeleton-dots">
                    <span></span>
                    <span></span>
                    <span></span>
                </div>
                <span>${text}</span>
            </div>
        `;
    }

    // ============================
    // PUBLIC API
    // ============================

    // Show skeleton for entire page sections
    showPageSkeleton(sections = ['stats', 'natconfig', 'clients']) {
        sections.forEach(section => {
            switch (section) {
                case 'stats':
                    this.showStatsSkeleton('#statsContainer', 4);
                    break;
                case 'natconfig':
                    this.showNATConfigSkeleton('#natConfigContainer', 3);
                    break;
                case 'clients':
                    this.showTableSkeleton('#clientsTable tbody', 8, 5);
                    break;
            }
        });
    }

    hidePageSkeleton() {
        // Remove all skeleton elements
        document.querySelectorAll('.skeleton-loader, .skeleton-grid, .skeleton-stat-card, .skeleton-nat-card, .skeleton-search, .skeleton-overlay').forEach(el => {
            el.remove();
        });
    }
}

// Make it globally available
if (typeof window !== 'undefined') {
    window.SkeletonLoader = SkeletonLoader;

    // Create global instance
    window.skeletonLoader = new SkeletonLoader();
}
