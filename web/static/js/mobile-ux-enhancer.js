/**
 * Mobile UX Enhancements
 * Features:
 * - Swipe gestures (swipe-to-refresh, swipe actions)
 * - Floating Action Button (FAB)
 * - Pull-to-refresh
 * - Touch optimizations
 * - Mobile-specific interactions
 */

class MobileUXEnhancer {
    constructor(options = {}) {
        // Configuration
        this.config = {
            enablePullToRefresh: options.enablePullToRefresh !== false,
            enableSwipeGestures: options.enableSwipeGestures !== false,
            enableFAB: options.enableFAB !== false,
            refreshCallback: options.refreshCallback || null,
            fabActions: options.fabActions || [],
            swipeThreshold: options.swipeThreshold || 100,
            pullThreshold: options.pullThreshold || 80,
        };

        // State
        this.state = {
            isPulling: false,
            pullDistance: 0,
            touchStartY: 0,
            touchStartX: 0,
            isRefreshing: false,
            isFABOpen: false,
            isMobile: this.detectMobile(),
        };

        // Elements
        this.pullIndicator = null;
        this.fab = null;

        // Initialize only on mobile
        if (this.state.isMobile) {
            this.init();
        }

        console.log('✅ Mobile UX Enhancer initialized');
    }

    // ============================
    // INITIALIZATION
    // ============================

    init() {
        if (this.config.enablePullToRefresh) {
            this.setupPullToRefresh();
        }

        if (this.config.enableSwipeGestures) {
            this.setupSwipeGestures();
        }

        if (this.config.enableFAB) {
            this.setupFAB();
        }

        this.setupTouchOptimizations();
        this.setupMobileMenu();
    }

    detectMobile() {
        return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
            || window.innerWidth < 768;
    }

    // ============================
    // PULL-TO-REFRESH
    // ============================

    setupPullToRefresh() {
        // Create pull indicator
        this.pullIndicator = document.createElement('div');
        this.pullIndicator.className = 'pull-to-refresh-indicator';
        this.pullIndicator.innerHTML = `
            <div class="pull-icon">
                <i class="fas fa-arrow-down"></i>
            </div>
            <div class="pull-text">Pull to refresh</div>
        `;
        document.body.insertBefore(this.pullIndicator, document.body.firstChild);

        // Bind touch events
        let touchStartY = 0;
        let scrollTop = 0;

        document.addEventListener('touchstart', (e) => {
            scrollTop = window.pageYOffset || document.documentElement.scrollTop;
            touchStartY = e.touches[0].clientY;
        }, { passive: true });

        document.addEventListener('touchmove', (e) => {
            if (scrollTop > 0) return; // Only work at top of page

            const touchY = e.touches[0].clientY;
            const pullDistance = touchY - touchStartY;

            if (pullDistance > 0 && pullDistance < 200) {
                this.state.pullDistance = pullDistance;
                this.updatePullIndicator(pullDistance);
            }
        }, { passive: true });

        document.addEventListener('touchend', () => {
            if (this.state.pullDistance > this.config.pullThreshold) {
                this.triggerRefresh();
            }

            this.resetPullIndicator();
        });
    }

    updatePullIndicator(distance) {
        const threshold = this.config.pullThreshold;
        const progress = Math.min(distance / threshold, 1);
        const rotation = progress * 180;

        this.pullIndicator.style.transform = `translateY(${Math.min(distance, threshold)}px)`;
        this.pullIndicator.style.opacity = progress;

        const icon = this.pullIndicator.querySelector('.pull-icon');
        icon.style.transform = `rotate(${rotation}deg)`;

        if (distance > threshold) {
            this.pullIndicator.classList.add('ready');
            this.pullIndicator.querySelector('.pull-text').textContent = 'Release to refresh';
        } else {
            this.pullIndicator.classList.remove('ready');
            this.pullIndicator.querySelector('.pull-text').textContent = 'Pull to refresh';
        }

        this.state.isPulling = true;
    }

    resetPullIndicator() {
        if (!this.state.isRefreshing) {
            this.pullIndicator.style.transform = 'translateY(-100px)';
            this.pullIndicator.style.opacity = '0';
            this.pullIndicator.classList.remove('ready');
        }

        this.state.pullDistance = 0;
        this.state.isPulling = false;
    }

    async triggerRefresh() {
        if (this.state.isRefreshing) return;

        this.state.isRefreshing = true;
        this.pullIndicator.classList.add('refreshing');
        this.pullIndicator.querySelector('.pull-text').textContent = 'Refreshing...';

        const icon = this.pullIndicator.querySelector('.pull-icon i');
        icon.className = 'fas fa-spinner fa-spin';

        try {
            if (this.config.refreshCallback && typeof this.config.refreshCallback === 'function') {
                await this.config.refreshCallback();
            } else if (typeof window.refreshData === 'function') {
                await window.refreshData();
            }

            // Show success briefly
            icon.className = 'fas fa-check';
            this.pullIndicator.querySelector('.pull-text').textContent = 'Refreshed!';

            await new Promise(resolve => setTimeout(resolve, 500));
        } catch (error) {
            console.error('Refresh failed:', error);
            icon.className = 'fas fa-times';
            this.pullIndicator.querySelector('.pull-text').textContent = 'Refresh failed';

            await new Promise(resolve => setTimeout(resolve, 1000));
        } finally {
            this.state.isRefreshing = false;
            this.pullIndicator.classList.remove('refreshing');
            icon.className = 'fas fa-arrow-down';
            this.resetPullIndicator();
        }
    }

    // ============================
    // SWIPE GESTURES
    // ============================

    setupSwipeGestures() {
        let touchStartX = 0;
        let touchStartY = 0;
        let touchEndX = 0;
        let touchEndY = 0;

        document.addEventListener('touchstart', (e) => {
            touchStartX = e.changedTouches[0].screenX;
            touchStartY = e.changedTouches[0].screenY;
        }, { passive: true });

        document.addEventListener('touchend', (e) => {
            touchEndX = e.changedTouches[0].screenX;
            touchEndY = e.changedTouches[0].screenY;

            this.handleSwipeGesture(touchStartX, touchStartY, touchEndX, touchEndY);
        });
    }

    handleSwipeGesture(startX, startY, endX, endY) {
        const deltaX = endX - startX;
        const deltaY = endY - startY;
        const absDeltaX = Math.abs(deltaX);
        const absDeltaY = Math.abs(deltaY);

        // Ignore small movements
        if (absDeltaX < 50 && absDeltaY < 50) return;

        // Determine swipe direction
        if (absDeltaX > absDeltaY) {
            // Horizontal swipe
            if (deltaX > 0) {
                this.onSwipeRight();
            } else {
                this.onSwipeLeft();
            }
        } else {
            // Vertical swipe
            if (deltaY > 0) {
                this.onSwipeDown();
            } else {
                this.onSwipeUp();
            }
        }
    }

    onSwipeRight() {
        // Open sidebar on swipe right
        const sidebar = document.getElementById('sidebar');
        if (sidebar && !sidebar.classList.contains('open')) {
            const hamburgerBtn = document.getElementById('hamburgerBtn');
            if (hamburgerBtn) hamburgerBtn.click();
        }
    }

    onSwipeLeft() {
        // Close sidebar on swipe left
        const sidebar = document.getElementById('sidebar');
        if (sidebar && sidebar.classList.contains('open')) {
            const overlay = document.getElementById('sidebarOverlay');
            if (overlay) overlay.click();
        }

        // Or close FAB if open
        if (this.state.isFABOpen) {
            this.closeFAB();
        }
    }

    onSwipeDown() {
        // Could be used for additional actions
        console.log('Swipe down detected');
    }

    onSwipeUp() {
        // Hide FAB on swipe up
        if (this.fab) {
            this.fab.style.transform = 'translateY(100px)';
            setTimeout(() => {
                if (this.fab) this.fab.style.transform = '';
            }, 2000);
        }
    }

    // ============================
    // FLOATING ACTION BUTTON (FAB)
    // ============================

    setupFAB() {
        this.fab = document.createElement('div');
        this.fab.className = 'mobile-fab';
        this.fab.innerHTML = `
            <button class="fab-main" title="Quick Actions">
                <i class="fas fa-ellipsis-v"></i>
            </button>
            <div class="fab-menu">
                <button class="fab-action" data-action="refresh" title="Refresh">
                    <i class="fas fa-sync-alt"></i>
                </button>
                <button class="fab-action" data-action="search" title="Search">
                    <i class="fas fa-search"></i>
                </button>
                <button class="fab-action" data-action="filter" title="Filter">
                    <i class="fas fa-filter"></i>
                </button>
                <button class="fab-action" data-action="export" title="Export">
                    <i class="fas fa-download"></i>
                </button>
            </div>
        `;

        document.body.appendChild(this.fab);

        // Bind FAB events
        const mainBtn = this.fab.querySelector('.fab-main');
        mainBtn.addEventListener('click', () => {
            this.toggleFAB();
        });

        // Bind action buttons
        const actionBtns = this.fab.querySelectorAll('.fab-action');
        actionBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const action = e.currentTarget.dataset.action;
                this.handleFABAction(action);
                this.closeFAB();
            });
        });

        // Close FAB when clicking outside
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.mobile-fab') && this.state.isFABOpen) {
                this.closeFAB();
            }
        });
    }

    toggleFAB() {
        if (this.state.isFABOpen) {
            this.closeFAB();
        } else {
            this.openFAB();
        }
    }

    openFAB() {
        this.fab.classList.add('open');
        this.state.isFABOpen = true;

        const mainBtn = this.fab.querySelector('.fab-main i');
        mainBtn.className = 'fas fa-times';
    }

    closeFAB() {
        this.fab.classList.remove('open');
        this.state.isFABOpen = false;

        const mainBtn = this.fab.querySelector('.fab-main i');
        mainBtn.className = 'fas fa-ellipsis-v';
    }

    handleFABAction(action) {
        switch (action) {
            case 'refresh':
                if (typeof window.refreshData === 'function') {
                    window.refreshData();
                }
                break;
            case 'search':
                const searchInput = document.querySelector('.search-input');
                if (searchInput) {
                    searchInput.focus();
                    searchInput.scrollIntoView({ behavior: 'smooth', block: 'center' });
                }
                break;
            case 'filter':
                if (window.advancedSearch) {
                    window.advancedSearch.toggleFilterPanel();
                }
                break;
            case 'export':
                const exportBtn = document.querySelector('.export-results-btn');
                if (exportBtn) exportBtn.click();
                break;
        }
    }

    // ============================
    // TOUCH OPTIMIZATIONS
    // ============================

    setupTouchOptimizations() {
        // Add touch-friendly classes
        document.body.classList.add('touch-optimized');

        // Increase tap targets for small buttons
        const smallButtons = document.querySelectorAll('.btn-sm, .btn-xs');
        smallButtons.forEach(btn => {
            btn.style.minHeight = '44px';
            btn.style.minWidth = '44px';
        });

        // Add haptic feedback simulation
        this.addHapticFeedback();

        // Optimize table for touch
        this.optimizeTableForTouch();
    }

    addHapticFeedback() {
        // Add visual feedback for touch
        const interactiveElements = document.querySelectorAll('button, a, .clickable');

        interactiveElements.forEach(el => {
            el.addEventListener('touchstart', function() {
                this.style.transform = 'scale(0.95)';
                this.style.opacity = '0.8';
            }, { passive: true });

            el.addEventListener('touchend', function() {
                this.style.transform = '';
                this.style.opacity = '';
            }, { passive: true });
        });
    }

    optimizeTableForTouch() {
        const tables = document.querySelectorAll('table');

        tables.forEach(table => {
            // Add touch scroll indicator
            const wrapper = table.closest('.table-responsive');
            if (wrapper) {
                wrapper.style.position = 'relative';

                // Add scroll hint
                const scrollHint = document.createElement('div');
                scrollHint.className = 'table-scroll-hint';
                scrollHint.innerHTML = '<i class="fas fa-arrows-alt-h"></i> Scroll';
                wrapper.appendChild(scrollHint);

                // Hide hint after first scroll
                wrapper.addEventListener('scroll', function() {
                    scrollHint.style.opacity = '0';
                }, { once: true, passive: true });
            }
        });
    }

    // ============================
    // MOBILE MENU
    // ============================

    setupMobileMenu() {
        // Add quick access menu for mobile
        const pageHeader = document.querySelector('.page-header');
        if (pageHeader) {
            const mobileMenu = document.createElement('div');
            mobileMenu.className = 'mobile-quick-menu';
            mobileMenu.innerHTML = `
                <button class="mobile-menu-btn" data-action="menu">
                    <i class="fas fa-bars"></i>
                </button>
                <button class="mobile-menu-btn" data-action="refresh">
                    <i class="fas fa-sync-alt"></i>
                </button>
                <button class="mobile-menu-btn" data-action="search">
                    <i class="fas fa-search"></i>
                </button>
            `;

            pageHeader.appendChild(mobileMenu);

            // Bind events
            mobileMenu.querySelectorAll('.mobile-menu-btn').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    const action = e.currentTarget.dataset.action;

                    switch (action) {
                        case 'menu':
                            const hamburger = document.getElementById('hamburgerBtn');
                            if (hamburger) hamburger.click();
                            break;
                        case 'refresh':
                            if (typeof window.refreshData === 'function') {
                                window.refreshData();
                            }
                            break;
                        case 'search':
                            const searchInput = document.querySelector('.search-input');
                            if (searchInput) searchInput.focus();
                            break;
                    }
                });
            });
        }
    }

    // ============================
    // PUBLIC API
    // ============================

    refresh() {
        this.triggerRefresh();
    }

    showFAB() {
        if (this.fab) {
            this.fab.style.display = 'block';
        }
    }

    hideFAB() {
        if (this.fab) {
            this.fab.style.display = 'none';
        }
    }

    destroy() {
        if (this.pullIndicator) {
            this.pullIndicator.remove();
        }

        if (this.fab) {
            this.fab.remove();
        }

        document.body.classList.remove('touch-optimized');

        console.log('🗑️ Mobile UX Enhancer destroyed');
    }
}

// Make it globally available
if (typeof window !== 'undefined') {
    window.MobileUXEnhancer = MobileUXEnhancer;
}
