/**
 * Smart Auto-Refresh System
 * Features:
 * - Adaptive refresh intervals based on user activity
 * - Pauses automatically during user interactions
 * - Manual pause/resume controls
 * - Visual indicator with countdown
 * - Configurable intervals
 */

class SmartAutoRefresh {
    constructor(options = {}) {
        // Configuration
        this.config = {
            defaultInterval: options.defaultInterval || 90000, // 90 seconds
            fastInterval: options.fastInterval || 30000, // 30 seconds when active
            slowInterval: options.slowInterval || 180000, // 3 minutes when idle
            pauseOnInteraction: options.pauseOnInteraction !== false, // Default true
            resumeDelay: options.resumeDelay || 5000, // Resume after 5s of inactivity
            refreshCallback: options.refreshCallback || null,
            autoStart: options.autoStart !== false, // Default true
            showIndicator: options.showIndicator !== false, // Default true
        };

        // State
        this.state = {
            isRunning: false,
            isPaused: false,
            isUserActive: false,
            currentInterval: this.config.defaultInterval,
            lastRefreshTime: null,
            nextRefreshTime: null,
            refreshCount: 0,
            mode: 'normal', // 'normal', 'fast', 'slow', 'paused'
        };

        // Timers
        this.refreshTimer = null;
        this.inactivityTimer = null;
        this.countdownTimer = null;

        // UI Elements
        this.indicator = null;
        this.menu = null;

        // Initialize
        if (this.config.showIndicator) {
            this.createIndicator();
        }

        this.bindEvents();

        if (this.config.autoStart) {
            this.start();
        }

        console.log('✅ Smart Auto-Refresh initialized');
    }

    // ============================
    // LIFECYCLE METHODS
    // ============================

    start() {
        if (this.state.isRunning) {
            console.log('⚠️ Auto-refresh already running');
            return;
        }

        this.state.isRunning = true;
        this.state.isPaused = false;
        this.state.lastRefreshTime = Date.now();

        this.scheduleNextRefresh();
        this.updateIndicator();

        console.log('▶️ Auto-refresh started');
    }

    stop() {
        if (!this.state.isRunning) {
            console.log('⚠️ Auto-refresh already stopped');
            return;
        }

        this.state.isRunning = false;
        this.clearTimers();
        this.updateIndicator();

        console.log('⏸️ Auto-refresh stopped');
    }

    pause() {
        if (!this.state.isRunning || this.state.isPaused) {
            return;
        }

        this.state.isPaused = true;
        this.state.mode = 'paused';
        this.clearRefreshTimer();
        this.updateIndicator();

        console.log('⏸️ Auto-refresh paused');
    }

    resume() {
        if (!this.state.isRunning || !this.state.isPaused) {
            return;
        }

        this.state.isPaused = false;
        this.scheduleNextRefresh();
        this.updateIndicator();

        console.log('▶️ Auto-refresh resumed');
    }

    toggle() {
        if (this.state.isPaused) {
            this.resume();
        } else {
            this.pause();
        }
    }

    // ============================
    // REFRESH LOGIC
    // ============================

    scheduleNextRefresh() {
        this.clearRefreshTimer();

        if (!this.state.isRunning || this.state.isPaused) {
            return;
        }

        // Determine interval based on mode
        const interval = this.getCurrentInterval();
        this.state.currentInterval = interval;
        this.state.nextRefreshTime = Date.now() + interval;

        // Schedule refresh
        this.refreshTimer = setTimeout(() => {
            this.performRefresh();
        }, interval);

        // Start countdown
        this.startCountdown();

        console.log(`⏰ Next refresh in ${interval / 1000}s`);
    }

    async performRefresh() {
        if (!this.state.isRunning || this.state.isPaused) {
            return;
        }

        console.log('🔄 Performing auto-refresh...');

        this.state.refreshCount++;
        this.state.lastRefreshTime = Date.now();

        // Update indicator to show refreshing state
        this.setRefreshingState(true);

        try {
            // Call the refresh callback
            if (this.config.refreshCallback && typeof this.config.refreshCallback === 'function') {
                await this.config.refreshCallback();
            }

            console.log(`✅ Auto-refresh completed (${this.state.refreshCount} total)`);
        } catch (error) {
            console.error('❌ Auto-refresh failed:', error);
        } finally {
            this.setRefreshingState(false);

            // Schedule next refresh
            this.scheduleNextRefresh();
        }
    }

    refreshNow() {
        console.log('🔄 Manual refresh triggered');
        this.performRefresh();
    }

    // ============================
    // ADAPTIVE BEHAVIOR
    // ============================

    getCurrentInterval() {
        // Pause on interaction
        if (this.config.pauseOnInteraction && this.state.isUserActive) {
            return this.config.defaultInterval;
        }

        // Adaptive intervals based on mode
        switch (this.state.mode) {
            case 'fast':
                return this.config.fastInterval;
            case 'slow':
                return this.config.slowInterval;
            case 'normal':
            default:
                return this.config.defaultInterval;
        }
    }

    setMode(mode) {
        if (this.state.mode === mode) {
            return;
        }

        this.state.mode = mode;
        console.log(`🔄 Auto-refresh mode changed to: ${mode}`);

        // Reschedule with new interval
        if (this.state.isRunning && !this.state.isPaused) {
            this.scheduleNextRefresh();
        }

        this.updateIndicator();
    }

    // ============================
    // USER INTERACTION DETECTION
    // ============================

    bindEvents() {
        // Detect user interactions
        const interactionEvents = ['mousedown', 'keydown', 'touchstart', 'scroll'];

        interactionEvents.forEach(event => {
            document.addEventListener(event, () => {
                this.onUserInteraction();
            }, { passive: true });
        });

        // Detect form focus
        document.addEventListener('focusin', (e) => {
            if (e.target.matches('input, textarea, select')) {
                this.onFormFocus();
            }
        });

        document.addEventListener('focusout', (e) => {
            if (e.target.matches('input, textarea, select')) {
                this.onFormBlur();
            }
        });

        // Page visibility
        document.addEventListener('visibilitychange', () => {
            if (document.hidden) {
                this.onPageHidden();
            } else {
                this.onPageVisible();
            }
        });
    }

    onUserInteraction() {
        if (!this.config.pauseOnInteraction) {
            return;
        }

        this.state.isUserActive = true;

        // Temporarily pause refresh
        if (!this.state.isPaused) {
            this.pause();
        }

        // Clear existing inactivity timer
        clearTimeout(this.inactivityTimer);

        // Set timer to resume after inactivity
        this.inactivityTimer = setTimeout(() => {
            this.state.isUserActive = false;
            this.resume();
        }, this.config.resumeDelay);
    }

    onFormFocus() {
        console.log('📝 Form focused - pausing auto-refresh');
        this.pause();
    }

    onFormBlur() {
        console.log('📝 Form blurred - resuming auto-refresh');

        // Resume after a short delay
        setTimeout(() => {
            if (!this.state.isUserActive) {
                this.resume();
            }
        }, 2000);
    }

    onPageHidden() {
        console.log('👁️ Page hidden - slowing refresh');
        this.setMode('slow');
    }

    onPageVisible() {
        console.log('👁️ Page visible - normal refresh');
        this.setMode('normal');
    }

    // ============================
    // TIMER MANAGEMENT
    // ============================

    clearRefreshTimer() {
        if (this.refreshTimer) {
            clearTimeout(this.refreshTimer);
            this.refreshTimer = null;
        }
    }

    clearCountdownTimer() {
        if (this.countdownTimer) {
            clearInterval(this.countdownTimer);
            this.countdownTimer = null;
        }
    }

    clearTimers() {
        this.clearRefreshTimer();
        this.clearCountdownTimer();

        if (this.inactivityTimer) {
            clearTimeout(this.inactivityTimer);
            this.inactivityTimer = null;
        }
    }

    startCountdown() {
        this.clearCountdownTimer();

        this.countdownTimer = setInterval(() => {
            this.updateCountdown();
        }, 1000);
    }

    updateCountdown() {
        if (!this.state.nextRefreshTime) {
            return;
        }

        const timeLeft = this.state.nextRefreshTime - Date.now();

        if (timeLeft <= 0) {
            this.clearCountdownTimer();
            return;
        }

        this.updateIndicatorTimer(timeLeft);
    }

    // ============================
    // UI METHODS
    // ============================

    createIndicator() {
        // Create indicator element
        this.indicator = document.createElement('div');
        this.indicator.className = 'auto-refresh-indicator';
        this.indicator.innerHTML = `
            <div class="auto-refresh-status-icon"></div>
            <div class="auto-refresh-text">Auto-Refresh</div>
            <div class="auto-refresh-timer">--:--</div>
        `;

        // Create menu
        this.menu = document.createElement('div');
        this.menu.className = 'auto-refresh-menu';
        this.menu.innerHTML = `
            <div class="auto-refresh-menu-item" data-action="toggle">
                <i class="fas fa-pause"></i>
                <span>Pause/Resume</span>
            </div>
            <div class="auto-refresh-menu-item" data-action="refresh">
                <i class="fas fa-sync-alt"></i>
                <span>Refresh Now</span>
            </div>
            <div class="auto-refresh-menu-divider"></div>
            <div class="auto-refresh-menu-item" data-action="fast">
                <i class="fas fa-forward"></i>
                <span>Fast Mode (30s)</span>
            </div>
            <div class="auto-refresh-menu-item" data-action="normal">
                <i class="fas fa-play"></i>
                <span>Normal Mode (90s)</span>
            </div>
            <div class="auto-refresh-menu-item" data-action="slow">
                <i class="fas fa-backward"></i>
                <span>Slow Mode (3min)</span>
            </div>
            <div class="auto-refresh-menu-divider"></div>
            <div class="auto-refresh-menu-info">
                <strong>Status</strong>
                <span class="status-text">Active</span>
            </div>
        `;

        // Append indicator
        this.indicator.appendChild(this.menu);
        document.body.appendChild(this.indicator);

        // Bind events
        this.indicator.addEventListener('click', (e) => {
            if (e.target.closest('.auto-refresh-menu-item')) {
                const action = e.target.closest('.auto-refresh-menu-item').dataset.action;
                this.handleMenuAction(action);
                this.hideMenu();
            } else if (!e.target.closest('.auto-refresh-menu')) {
                this.toggleMenu();
            }
        });

        // Close menu when clicking outside
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.auto-refresh-indicator')) {
                this.hideMenu();
            }
        });
    }

    toggleMenu() {
        this.menu.classList.toggle('show');
    }

    hideMenu() {
        this.menu.classList.remove('show');
    }

    handleMenuAction(action) {
        switch (action) {
            case 'toggle':
                this.toggle();
                break;
            case 'refresh':
                this.refreshNow();
                break;
            case 'fast':
                this.setMode('fast');
                break;
            case 'normal':
                this.setMode('normal');
                break;
            case 'slow':
                this.setMode('slow');
                break;
        }
    }

    updateIndicator() {
        if (!this.indicator) {
            return;
        }

        // Update class
        this.indicator.className = 'auto-refresh-indicator';

        if (!this.state.isRunning) {
            this.indicator.classList.add('idle');
        } else if (this.state.isPaused) {
            this.indicator.classList.add('paused');
        } else {
            this.indicator.classList.add('active');
        }

        // Update status text in menu
        const statusText = this.menu.querySelector('.status-text');
        if (statusText) {
            let status = 'Active';
            if (!this.state.isRunning) {
                status = 'Stopped';
            } else if (this.state.isPaused) {
                status = 'Paused (User Active)';
            } else {
                status = `Active (${this.state.mode})`;
            }
            statusText.textContent = status;
        }
    }

    updateIndicatorTimer(timeLeft) {
        if (!this.indicator) {
            return;
        }

        const timerElement = this.indicator.querySelector('.auto-refresh-timer');
        if (!timerElement) {
            return;
        }

        const seconds = Math.floor(timeLeft / 1000);
        const minutes = Math.floor(seconds / 60);
        const remainingSeconds = seconds % 60;

        timerElement.textContent = `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
    }

    setRefreshingState(isRefreshing) {
        if (!this.indicator) {
            return;
        }

        if (isRefreshing) {
            this.indicator.classList.add('refreshing');
        } else {
            this.indicator.classList.remove('refreshing');
        }
    }

    // ============================
    // PUBLIC API
    // ============================

    getState() {
        return { ...this.state };
    }

    getStats() {
        return {
            refreshCount: this.state.refreshCount,
            lastRefreshTime: this.state.lastRefreshTime,
            nextRefreshTime: this.state.nextRefreshTime,
            currentInterval: this.state.currentInterval,
            mode: this.state.mode,
            isRunning: this.state.isRunning,
            isPaused: this.state.isPaused,
        };
    }

    setInterval(interval) {
        this.config.defaultInterval = interval;

        if (this.state.isRunning && !this.state.isPaused && this.state.mode === 'normal') {
            this.scheduleNextRefresh();
        }
    }

    destroy() {
        this.stop();
        this.clearTimers();

        if (this.indicator) {
            this.indicator.remove();
        }

        console.log('🗑️ Smart Auto-Refresh destroyed');
    }
}

// Export for use in other scripts
if (typeof window !== 'undefined') {
    window.SmartAutoRefresh = SmartAutoRefresh;
}
