/**
 * Auto-Refresh UI Component
 * Visual indicator and controls for SmartAutoRefresh
 */

class AutoRefreshUI {
    constructor(refreshInstance, options = {}) {
        this.refresh = refreshInstance;
        this.options = {
            position: options.position || 'top-right',
            showTimer: options.showTimer !== false,
            ...options
        };

        this.elements = {};
        this.nextRefreshTimer = null;
        this.init();
    }

    init() {
        this.createUI();
        this.attachEventListeners();
        this.startTimerUpdate();
    }

    createUI() {
        // Create main indicator
        const indicator = document.createElement('div');
        indicator.className = 'auto-refresh-indicator';
        indicator.id = 'autoRefreshIndicator';

        indicator.innerHTML = `
            <span class="auto-refresh-status-icon"></span>
            <span class="auto-refresh-text">Auto-Refresh</span>
            ${this.options.showTimer ? '<span class="auto-refresh-timer" id="autoRefreshTimer">--:--</span>' : ''}
            <i class="fas fa-chevron-down" style="font-size: 0.7rem; opacity: 0.7;"></i>
        `;

        // Create dropdown menu
        const menu = document.createElement('div');
        menu.className = 'auto-refresh-menu';
        menu.id = 'autoRefreshMenu';

        menu.innerHTML = `
            <div class="auto-refresh-menu-item" data-action="toggle">
                <i class="fas fa-pause"></i>
                <span id="toggleText">Pause</span>
            </div>
            <div class="auto-refresh-menu-divider"></div>
            <div class="auto-refresh-menu-item" data-action="refresh-now">
                <i class="fas fa-sync-alt"></i>
                <span>Refresh Now</span>
            </div>
            <div class="auto-refresh-menu-divider"></div>
            <div class="auto-refresh-menu-info">
                <strong>Status</strong>
                <div id="refreshStatusText">Active</div>
                <div style="margin-top: 6px;">
                    <small>Last refresh: <span id="lastRefreshTime">Never</span></small>
                </div>
                <div>
                    <small>Refresh count: <span id="refreshCount">0</span></small>
                </div>
            </div>
        `;

        // Append to indicator
        indicator.appendChild(menu);

        // Add to document
        document.body.appendChild(indicator);

        // Store references
        this.elements = {
            indicator,
            menu,
            timer: document.getElementById('autoRefreshTimer'),
            toggleText: document.getElementById('toggleText'),
            statusText: document.getElementById('refreshStatusText'),
            lastRefreshTime: document.getElementById('lastRefreshTime'),
            refreshCount: document.getElementById('refreshCount')
        };

        // Initial update
        this.updateUI(this.refresh.getStatus());
    }

    attachEventListeners() {
        // Toggle menu on click
        this.elements.indicator.addEventListener('click', (e) => {
            // Don't toggle if clicking menu item
            if (e.target.closest('.auto-refresh-menu-item')) return;

            this.elements.menu.classList.toggle('show');
        });

        // Handle menu actions
        this.elements.menu.addEventListener('click', (e) => {
            const menuItem = e.target.closest('.auto-refresh-menu-item');
            if (!menuItem) return;

            const action = menuItem.dataset.action;
            this.handleAction(action);

            // Close menu
            this.elements.menu.classList.remove('show');
        });

        // Close menu when clicking outside
        document.addEventListener('click', (e) => {
            if (!this.elements.indicator.contains(e.target)) {
                this.elements.menu.classList.remove('show');
            }
        });

        // Update UI on state change
        this.refresh.config.onStateChange = (status) => {
            this.updateUI(status);
        };
    }

    handleAction(action) {
        switch (action) {
            case 'toggle':
                if (this.refresh.state.isPaused) {
                    this.refresh.resume();
                } else {
                    this.refresh.pause();
                }
                break;
            case 'refresh-now':
                this.refresh.performRefresh();
                break;
        }
    }

    updateUI(status) {
        // Update indicator class
        this.elements.indicator.classList.remove('active', 'paused', 'idle', 'refreshing');

        if (status.isPaused) {
            this.elements.indicator.classList.add('paused');
        } else if (status.isIdle) {
            this.elements.indicator.classList.add('idle');
        } else if (status.isRunning) {
            this.elements.indicator.classList.add('active');
        }

        // Update toggle button text
        if (this.elements.toggleText) {
            this.elements.toggleText.textContent = status.isPaused ? 'Resume' : 'Pause';
        }

        // Update status text
        let statusText = 'Stopped';
        if (status.isRunning) {
            if (status.isPaused) {
                statusText = 'Paused';
            } else if (status.isIdle) {
                statusText = `Idle (${status.interval / 1000}s)`;
            } else {
                statusText = `Active (${status.interval / 1000}s)`;
            }
        }
        if (this.elements.statusText) {
            this.elements.statusText.textContent = statusText;
        }

        // Update last refresh time
        if (this.elements.lastRefreshTime && status.lastRefresh) {
            const time = new Date(status.lastRefresh);
            this.elements.lastRefreshTime.textContent = time.toLocaleTimeString();
        }

        // Update refresh count
        if (this.elements.refreshCount) {
            this.elements.refreshCount.textContent = status.refreshCount;
        }
    }

    startTimerUpdate() {
        if (!this.options.showTimer || !this.elements.timer) return;

        // Update timer every second
        setInterval(() => {
            if (!this.refresh.state.isRunning || this.refresh.state.isPaused) {
                this.elements.timer.textContent = '--:--';
                return;
            }

            const status = this.refresh.getStatus();
            const nextRefresh = this.refresh.state.lastRefresh
                ? this.refresh.state.lastRefresh + status.interval
                : Date.now() + status.interval;

            const timeLeft = Math.max(0, nextRefresh - Date.now());
            const seconds = Math.floor((timeLeft / 1000) % 60);
            const minutes = Math.floor(timeLeft / 1000 / 60);

            this.elements.timer.textContent = `${minutes}:${seconds.toString().padStart(2, '0')}`;
        }, 1000);
    }

    setRefreshing(isRefreshing) {
        if (isRefreshing) {
            this.elements.indicator.classList.add('refreshing');
        } else {
            this.elements.indicator.classList.remove('refreshing');
        }
    }

    destroy() {
        if (this.elements.indicator) {
            this.elements.indicator.remove();
        }
    }
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = AutoRefreshUI;
}
