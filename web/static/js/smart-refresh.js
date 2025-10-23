/**
 * Smart Auto-Refresh Module
 * Features:
 * - Adaptive refresh intervals (faster when active, slower when idle)
 * - Pause on user interaction
 * - Visual indicator
 * - User control
 */

class SmartAutoRefresh {
    constructor(config = {}) {
        this.config = {
            baseInterval: config.baseInterval || 30000,        // 30 seconds (active)
            slowInterval: config.slowInterval || 90000,        // 90 seconds (idle)
            pauseOnInteraction: config.pauseOnInteraction !== false,
            pauseDuration: config.pauseDuration || 10000,      // 10 seconds pause after interaction
            refreshCallback: config.refreshCallback || null,
            onStateChange: config.onStateChange || null,
            ...config
        };

        this.state = {
            isRunning: false,
            isPaused: false,
            isIdle: false,
            lastRefresh: null,
            lastInteraction: null,
            refreshCount: 0
        };

        this.timers = {
            refresh: null,
            idleDetector: null,
            pauseTimeout: null
        };

        this.idleThreshold = 60000; // 1 minute of no interaction = idle
        this.interactionEvents = ['mousedown', 'keydown', 'scroll', 'touchstart'];

        this.boundHandleInteraction = this.handleInteraction.bind(this);
        this.init();
    }

    init() {
        // Setup interaction listeners if pause on interaction is enabled
        if (this.config.pauseOnInteraction) {
            this.setupInteractionListeners();
        }

        // Setup idle detection
        this.setupIdleDetection();
    }

    setupInteractionListeners() {
        this.interactionEvents.forEach(event => {
            document.addEventListener(event, this.boundHandleInteraction, { passive: true });
        });
    }

    setupIdleDetection() {
        // Check for idle state every 10 seconds
        this.timers.idleDetector = setInterval(() => {
            this.checkIdleState();
        }, 10000);
    }

    checkIdleState() {
        if (!this.state.lastInteraction) {
            this.state.lastInteraction = Date.now();
            return;
        }

        const timeSinceInteraction = Date.now() - this.state.lastInteraction;
        const wasIdle = this.state.isIdle;
        this.state.isIdle = timeSinceInteraction > this.idleThreshold;

        // Notify state change if idle state changed
        if (wasIdle !== this.state.isIdle) {
            console.log(`📊 Auto-refresh: ${this.state.isIdle ? 'IDLE' : 'ACTIVE'} mode`);
            this.notifyStateChange();

            // Restart refresh with new interval if running
            if (this.state.isRunning && !this.state.isPaused) {
                this.scheduleNextRefresh();
            }
        }
    }

    handleInteraction() {
        this.state.lastInteraction = Date.now();

        // If was idle, switch back to active
        if (this.state.isIdle) {
            this.state.isIdle = false;
            console.log('📊 Auto-refresh: Back to ACTIVE mode');
            this.notifyStateChange();
        }

        // Pause auto-refresh temporarily on interaction
        if (this.config.pauseOnInteraction && this.state.isRunning) {
            this.pauseTemporarily();
        }
    }

    pauseTemporarily() {
        // Clear existing refresh timer
        if (this.timers.refresh) {
            clearTimeout(this.timers.refresh);
            this.timers.refresh = null;
        }

        // Clear existing pause timeout
        if (this.timers.pauseTimeout) {
            clearTimeout(this.timers.pauseTimeout);
        }

        this.state.isPaused = true;
        this.notifyStateChange();

        // Resume after pause duration
        this.timers.pauseTimeout = setTimeout(() => {
            this.state.isPaused = false;
            this.notifyStateChange();
            this.scheduleNextRefresh();
        }, this.config.pauseDuration);
    }

    start() {
        if (this.state.isRunning) {
            console.log('⚠️ Auto-refresh already running');
            return;
        }

        console.log('✅ Starting smart auto-refresh');
        this.state.isRunning = true;
        this.state.lastInteraction = Date.now();
        this.notifyStateChange();
        this.scheduleNextRefresh();
    }

    stop() {
        console.log('⏹️ Stopping smart auto-refresh');
        this.state.isRunning = false;
        this.state.isPaused = false;

        // Clear all timers
        if (this.timers.refresh) {
            clearTimeout(this.timers.refresh);
            this.timers.refresh = null;
        }
        if (this.timers.pauseTimeout) {
            clearTimeout(this.timers.pauseTimeout);
            this.timers.pauseTimeout = null;
        }

        this.notifyStateChange();
    }

    pause() {
        if (!this.state.isRunning) return;

        console.log('⏸️ Pausing auto-refresh');
        this.state.isPaused = true;

        if (this.timers.refresh) {
            clearTimeout(this.timers.refresh);
            this.timers.refresh = null;
        }
        if (this.timers.pauseTimeout) {
            clearTimeout(this.timers.pauseTimeout);
            this.timers.pauseTimeout = null;
        }

        this.notifyStateChange();
    }

    resume() {
        if (!this.state.isRunning || !this.state.isPaused) return;

        console.log('▶️ Resuming auto-refresh');
        this.state.isPaused = false;
        this.notifyStateChange();
        this.scheduleNextRefresh();
    }

    scheduleNextRefresh() {
        // Clear existing timer
        if (this.timers.refresh) {
            clearTimeout(this.timers.refresh);
        }

        // Determine interval based on idle state
        const interval = this.state.isIdle ? this.config.slowInterval : this.config.baseInterval;

        console.log(`⏰ Next refresh in ${interval / 1000}s (${this.state.isIdle ? 'IDLE' : 'ACTIVE'} mode)`);

        this.timers.refresh = setTimeout(() => {
            this.performRefresh();
        }, interval);
    }

    async performRefresh() {
        if (!this.state.isRunning || this.state.isPaused) return;

        console.log('🔄 Performing auto-refresh...');
        this.state.lastRefresh = Date.now();
        this.state.refreshCount++;

        try {
            if (this.config.refreshCallback) {
                await this.config.refreshCallback();
                console.log(`✅ Auto-refresh #${this.state.refreshCount} completed`);
            }
        } catch (error) {
            console.error('❌ Auto-refresh failed:', error);
        }

        // Schedule next refresh
        if (this.state.isRunning) {
            this.scheduleNextRefresh();
        }
    }

    notifyStateChange() {
        if (this.config.onStateChange) {
            this.config.onStateChange(this.getStatus());
        }
    }

    getStatus() {
        return {
            isRunning: this.state.isRunning,
            isPaused: this.state.isPaused,
            isIdle: this.state.isIdle,
            lastRefresh: this.state.lastRefresh,
            refreshCount: this.state.refreshCount,
            interval: this.state.isIdle ? this.config.slowInterval : this.config.baseInterval
        };
    }

    destroy() {
        console.log('🗑️ Destroying auto-refresh instance');
        this.stop();

        // Remove interaction listeners
        this.interactionEvents.forEach(event => {
            document.removeEventListener(event, this.boundHandleInteraction);
        });

        // Clear idle detector
        if (this.timers.idleDetector) {
            clearInterval(this.timers.idleDetector);
            this.timers.idleDetector = null;
        }
    }
}

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = SmartAutoRefresh;
}
