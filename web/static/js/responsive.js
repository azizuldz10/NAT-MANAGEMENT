/**
 * PPPoE Monitor - Responsive Utilities & Theme JavaScript
 * Handles responsiveness, animations, and UI interactions
 */

/* Global API fetch wrapper: ensures cookies are included for CORS and same-origin.
   Usage: apiFetch(url, { method, headers, body, ... })
*/
window.apiFetch = function(url, options) {
    const opts = options || {};
    // Preserve existing headers; ensure credentials are included
    const merged = {
        ...opts,
        credentials: 'include'
    };
    return fetch(url, merged);
};

class ResponsiveManager {
    constructor() {
        this.breakpoints = {
            xs: 576,
            sm: 768,
            md: 992,
            lg: 1200
        };
        
        this.currentBreakpoint = this.getCurrentBreakpoint();
        this.previousBreakpoint = null;
        
        this.init();
    }
    
    init() {
        this.setupResizeHandler();
        this.setupScrollHandler();
        this.setupAnimationObserver();
        this.setupTooltips();
        this.setupMobileOptimizations();
        this.initializeCounters();
        
        // Initial setup
        this.handleResize();
        this.optimizeForDevice();
        
        console.log('✅ ResponsiveManager initialized');
    }
    
    // ============================
    // BREAKPOINT MANAGEMENT
    // ============================
    
    getCurrentBreakpoint() {
        const width = window.innerWidth;
        if (width < this.breakpoints.xs) return 'xs';
        if (width < this.breakpoints.sm) return 'sm';
        if (width < this.breakpoints.md) return 'md';
        if (width < this.breakpoints.lg) return 'lg';
        return 'xl';
    }
    
    isBreakpoint(bp) {
        return this.currentBreakpoint === bp;
    }
    
    isMobile() {
        return this.currentBreakpoint === 'xs' || this.currentBreakpoint === 'sm';
    }
    
    isTablet() {
        return this.currentBreakpoint === 'md';
    }
    
    isDesktop() {
        return this.currentBreakpoint === 'lg' || this.currentBreakpoint === 'xl';
    }
    
    // ============================
    // EVENT HANDLERS
    // ============================
    
    setupResizeHandler() {
        let resizeTimeout;
        
        window.addEventListener('resize', () => {
            clearTimeout(resizeTimeout);
            resizeTimeout = setTimeout(() => {
                this.handleResize();
            }, 150);
        });
    }
    
    handleResize() {
        this.previousBreakpoint = this.currentBreakpoint;
        this.currentBreakpoint = this.getCurrentBreakpoint();
        
        if (this.previousBreakpoint !== this.currentBreakpoint) {
            this.onBreakpointChange();
        }
        
        this.updateMobileLayout();
        this.adjustTableResponsiveness();
        this.updateNavigationStyle();
        this.optimizeCardLayout();
        
        // Dispatch custom event
        window.dispatchEvent(new CustomEvent('responsive:breakpointChange', {
            detail: {
                current: this.currentBreakpoint,
                previous: this.previousBreakpoint,
                isMobile: this.isMobile(),
                isTablet: this.isTablet(),
                isDesktop: this.isDesktop()
            }
        }));
    }
    
    onBreakpointChange() {
        console.log(`📱 Breakpoint changed: ${this.previousBreakpoint} → ${this.currentBreakpoint}`);
        
        // Update body class for CSS targeting
        document.body.className = document.body.className.replace(/\bbp-\w+/g, '');
        document.body.classList.add(`bp-${this.currentBreakpoint}`);
        
        // Trigger re-render of dynamic components
        this.triggerComponentUpdates();
    }
    
    setupScrollHandler() {
        let scrollTimeout;
        
        window.addEventListener('scroll', () => {
            clearTimeout(scrollTimeout);
            scrollTimeout = setTimeout(() => {
                this.handleScroll();
            }, 100);
        });
    }
    
    handleScroll() {
        const scrollY = window.scrollY;
        const header = document.querySelector('.page-header');
        
        if (header) {
            if (scrollY > 50) {
                header.classList.add('scrolled');
                header.style.backdropFilter = 'blur(20px)';
                header.style.background = 'rgba(255, 255, 255, 0.95)';
            } else {
                header.classList.remove('scrolled');
                header.style.backdropFilter = 'blur(10px)';
                header.style.background = 'white';
            }
        }
    }
    
    // ============================
    // MOBILE OPTIMIZATIONS
    // ============================
    
    setupMobileOptimizations() {
        if (this.isMobileDevice()) {
            // Disable hover effects on mobile
            document.body.classList.add('touch-device');
            
            // Optimize touch interactions
            this.optimizeTouchInteractions();
            
            // Reduce animations on slower devices
            this.optimizeAnimations();
        }
    }
    
    isMobileDevice() {
        return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
    }
    
    optimizeTouchInteractions() {
        // Add touch-friendly classes
        const buttons = document.querySelectorAll('.btn');
        buttons.forEach(btn => {
            btn.style.minHeight = '44px'; // iOS recommended touch target
            btn.style.minWidth = '44px';
        });
        
        // Improve table row selection on mobile
        const tableRows = document.querySelectorAll('.table tbody tr');
        tableRows.forEach(row => {
            row.addEventListener('touchstart', function() {
                this.style.backgroundColor = 'rgba(102, 126, 234, 0.1)';
            });
            
            row.addEventListener('touchend', function() {
                setTimeout(() => {
                    this.style.backgroundColor = '';
                }, 150);
            });
        });
    }
    
    updateMobileLayout() {
        const isMobile = this.isMobile();
        
        // Update router badges for mobile
        if (typeof this.updateRouterBadges === 'function') {
            this.updateRouterBadges(isMobile);
        }
        
        // Adjust table columns
        this.toggleMobileColumns(isMobile);
        
        // Update form layouts
        this.adjustFormLayouts(isMobile);
    }
    
    toggleMobileColumns(isMobile) {
        const mobileHiddenCols = document.querySelectorAll('.d-none-mobile');
        mobileHiddenCols.forEach(col => {
            col.style.display = isMobile ? 'none' : '';
        });
    }
    
    adjustFormLayouts(isMobile) {
        const formRows = document.querySelectorAll('.row.g-3, .row.g-2');
        formRows.forEach(row => {
            if (isMobile) {
                row.classList.add('mobile-form');
            } else {
                row.classList.remove('mobile-form');
            }
        });
    }
    
    // ============================
    // TABLE RESPONSIVENESS
    // ============================
    
    adjustTableResponsiveness() {
        const tables = document.querySelectorAll('.table-responsive');
        
        tables.forEach(tableContainer => {
            const table = tableContainer.querySelector('table');
            if (!table) return;
            
            if (this.isMobile()) {
                this.enableMobileTableMode(table);
            } else {
                this.disableMobileTableMode(table);
            }
        });
    }
    
    enableMobileTableMode(table) {
        table.classList.add('mobile-table');
        
        // Reduce padding for mobile
        const cells = table.querySelectorAll('th, td');
        cells.forEach(cell => {
            cell.style.padding = '0.375rem 0.25rem';
            cell.style.fontSize = '0.8rem';
        });
    }
    
    disableMobileTableMode(table) {
        table.classList.remove('mobile-table');
        
        // Restore normal padding
        const cells = table.querySelectorAll('th, td');
        cells.forEach(cell => {
            cell.style.padding = '';
            cell.style.fontSize = '';
        });
    }
    
    // ============================
    // NAVIGATION & CARDS
    // ============================
    
    updateNavigationStyle() {
        const navPills = document.querySelector('.nav-pills');
        if (!navPills) return;
        
        if (this.isMobile()) {
            navPills.classList.add('mobile-nav');
            navPills.style.flexDirection = 'column';
        } else {
            navPills.classList.remove('mobile-nav');
            navPills.style.flexDirection = '';
        }
    }
    
    optimizeCardLayout() {
        const routerCards = document.querySelectorAll('.router-card');
        
        routerCards.forEach(card => {
            if (this.isMobile()) {
                card.style.padding = '8px';
                card.style.marginBottom = '8px';
            } else if (this.isTablet()) {
                card.style.padding = '12px';
                card.style.marginBottom = '12px';
            } else {
                card.style.padding = '';
                card.style.marginBottom = '';
            }
        });
    }
    
    // ============================
    // ANIMATIONS & EFFECTS
    // ============================
    
    setupAnimationObserver() {
        if ('IntersectionObserver' in window) {
            const observer = new IntersectionObserver((entries) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        entry.target.classList.add('animate-in');
                        observer.unobserve(entry.target);
                    }
                });
            }, {
                threshold: 0.1,
                rootMargin: '0px 0px -50px 0px'
            });
            
            // Observe cards and stats
            const animatableElements = document.querySelectorAll('.stats-card, .router-card, .card');
            animatableElements.forEach(el => {
                el.classList.add('animate-ready');
                observer.observe(el);
            });
        }
    }
    
    optimizeAnimations() {
        // Reduce animations on slower devices
        const hasSlowCPU = navigator.hardwareConcurrency && navigator.hardwareConcurrency < 4;
        const hasSlowConnection = navigator.connection && navigator.connection.effectiveType === 'slow-2g';
        
        if (hasSlowCPU || hasSlowConnection) {
            document.body.classList.add('reduce-animations');
        }
        
        // Respect user preference
        if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
            document.body.classList.add('reduce-animations');
        }
    }
    
    // ============================
    // TOOLTIPS & INTERACTIONS
    // ============================
    
    setupTooltips() {
        // Simple tooltip implementation
        const elementsWithTooltips = document.querySelectorAll('[title]');
        
        elementsWithTooltips.forEach(element => {
            if (this.isMobileDevice()) {
                // Convert titles to data attributes for mobile
                element.dataset.tooltip = element.title;
                element.removeAttribute('title');
                
                element.addEventListener('touchstart', this.showMobileTooltip.bind(this));
            }
        });
    }
    
    showMobileTooltip(event) {
        const element = event.target;
        const tooltipText = element.dataset.tooltip;
        
        if (!tooltipText) return;
        
        // Create temporary tooltip
        const tooltip = document.createElement('div');
        tooltip.className = 'mobile-tooltip';
        tooltip.textContent = tooltipText;
        tooltip.style.cssText = `
            position: fixed;
            bottom: 20px;
            left: 50%;
            transform: translateX(-50%);
            background: rgba(0,0,0,0.8);
            color: white;
            padding: 8px 12px;
            border-radius: 6px;
            font-size: 0.8rem;
            z-index: 9999;
            pointer-events: none;
        `;
        
        document.body.appendChild(tooltip);
        
        setTimeout(() => {
            if (tooltip.parentNode) {
                tooltip.parentNode.removeChild(tooltip);
            }
        }, 2000);
    }
    
    // ============================
    // COUNTERS & ANIMATIONS
    // ============================
    
    initializeCounters() {
        const counters = document.querySelectorAll('[data-counter]');
        
        counters.forEach(counter => {
            if (counter && counter.dataset) {
                this.animateCounter(counter);
            }
        });
    }
    
    animateCounter(element) {
        if (!element || !element.dataset) return;
        
        const target = parseInt(element.dataset.counter) || parseInt(element.textContent) || 0;
        const duration = 2000;
        const step = target / (duration / 16);
        let current = 0;
        
        const timer = setInterval(() => {
            current += step;
            if (current >= target) {
                current = target;
                clearInterval(timer);
            }
            if (element) {
                element.textContent = Math.floor(current);
            }
        }, 16);
    }
    
    // ============================
    // UTILITY METHODS
    // ============================
    
    triggerComponentUpdates() {
        // Trigger updates for components that need re-rendering
        const updateEvents = ['table:update', 'nav:update', 'cards:update'];
        
        updateEvents.forEach(eventName => {
            window.dispatchEvent(new CustomEvent(eventName, {
                detail: { breakpoint: this.currentBreakpoint }
            }));
        });
    }
    
    optimizeForDevice() {
        // Device-specific optimizations
        const isRetina = window.devicePixelRatio > 1;
        const isHighDPI = window.devicePixelRatio > 2;
        
        if (isHighDPI) {
            document.body.classList.add('high-dpi');
        } else if (isRetina) {
            document.body.classList.add('retina');
        }
        
        // Memory optimization for low-end devices
        if (navigator.deviceMemory && navigator.deviceMemory < 4) {
            document.body.classList.add('low-memory');
        }
    }
    
    // ============================
    // PUBLIC API
    // ============================
    
    updateRouterBadges(isMobile) {
        // This method can be overridden by specific pages
        try {
            const badges = document.querySelectorAll('.badge');
            badges.forEach(badge => {
                if (!badge || !badge.dataset) return;
                
                const originalText = badge.dataset.originalText || badge.textContent;
                if (badge.dataset) {
                    badge.dataset.originalText = originalText;
                }
                
                if (isMobile && originalText && originalText.length > 8) {
                    // Abbreviate router names on mobile
                    const abbreviations = {
                        'DARUSSALAM': 'DAR',
                        'SAMSAT': 'SAM',
                        'LANE1': 'L1',
                        'LANE2': 'L2',
                        'BT JAYA/PK JAYA': 'BJ',
                        'LANE4': 'L4'
                    };
                    
                    badge.textContent = abbreviations[originalText] || originalText.substring(0, 4);
                } else {
                    badge.textContent = originalText;
                }
            });
        } catch (error) {
            console.warn('updateRouterBadges error:', error);
        }
    }
    
    // Method to be called when data is updated
    onDataUpdate() {
        try {
            this.handleResize();
            this.initializeCounters();
        } catch (error) {
            console.warn('ResponsiveManager onDataUpdate error:', error);
        }
    }
    
    // Method to refresh responsive layout
    refresh() {
        try {
            this.handleResize();
            this.optimizeForDevice();
        } catch (error) {
            console.warn('ResponsiveManager refresh error:', error);
        }
    }
}

// ============================
// THEME UTILITIES
// ============================

class ThemeUtils {
    static formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
    
    static formatPackets(packets) {
        if (packets === 0) return '0 pkts';
        if (packets < 1000) return packets + ' pkts';
        if (packets < 1000000) return (packets / 1000).toFixed(1) + 'K pkts';
        return (packets / 1000000).toFixed(1) + 'M pkts';
    }
    
    static truncateText(text, maxLength) {
        if (!text || text.length <= maxLength) return text;
        return text.substring(0, maxLength - 3) + '...';
    }
    
    static addLoadingState(element) {
        if (!element) return;
        
        element.classList.add('loading-state');
        element.style.opacity = '0.6';
        element.style.pointerEvents = 'none';
        
        const spinner = document.createElement('div');
        spinner.className = 'loading-spinner me-2';
        element.insertBefore(spinner, element.firstChild);
    }
    
    static removeLoadingState(element) {
        if (!element) return;
        
        element.classList.remove('loading-state');
        element.style.opacity = '';
        element.style.pointerEvents = '';
        
        const spinner = element.querySelector('.loading-spinner');
        if (spinner) {
            spinner.remove();
        }
    }
    
    static showToast(message, type = 'info', duration = 3000) {
        const toast = document.createElement('div');
        toast.className = `alert alert-${type} toast-notification`;
        toast.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 9999;
            min-width: 300px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            border: none;
            border-radius: 8px;
            animation: slideInRight 0.3s ease;
        `;
        toast.innerHTML = `
            <div class="d-flex align-items-center">
                <i class="fas fa-${type === 'success' ? 'check' : type === 'danger' ? 'times' : 'info'}-circle me-2"></i>
                <span>${message}</span>
                <button type="button" class="btn-close ms-auto" onclick="this.parentElement.parentElement.remove()"></button>
            </div>
        `;
        
        document.body.appendChild(toast);
        
        setTimeout(() => {
            if (toast.parentNode) {
                toast.style.animation = 'slideOutRight 0.3s ease';
                setTimeout(() => {
                    if (toast.parentNode) {
                        toast.parentNode.removeChild(toast);
                    }
                }, 300);
            }
        }, duration);
    }
}

// ============================
// CSS ANIMATIONS
// ============================

const animationCSS = `
@keyframes slideInRight {
    from {
        transform: translateX(100%);
        opacity: 0;
    }
    to {
        transform: translateX(0);
        opacity: 1;
    }
}

@keyframes slideOutRight {
    from {
        transform: translateX(0);
        opacity: 1;
    }
    to {
        transform: translateX(100%);
        opacity: 0;
    }
}

.animate-ready {
    opacity: 0;
    transform: translateY(20px);
    transition: all 0.6s ease;
}

.animate-in {
    opacity: 1;
    transform: translateY(0);
}

.reduce-animations * {
    animation-duration: 0.1s !important;
    transition-duration: 0.1s !important;
}

.touch-device .router-card:hover {
    transform: none;
}

.mobile-table th,
.mobile-table td {
    font-size: 0.8rem !important;
    padding: 0.375rem 0.25rem !important;
}

.mobile-nav .nav-link {
    margin: 0.125rem 0 !important;
    text-align: center;
}

.loading-state {
    position: relative;
}

.mobile-tooltip {
    animation: fadeInUp 0.3s ease;
}

@keyframes fadeInUp {
    from {
        opacity: 0;
        transform: translate(-50%, 10px);
    }
    to {
        opacity: 1;
        transform: translate(-50%, 0);
    }
}
`;

// Inject CSS
const style = document.createElement('style');
style.textContent = animationCSS;
document.head.appendChild(style);

// ============================
// INITIALIZATION
// ============================

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    // Initialize responsive manager
    window.responsiveManager = new ResponsiveManager();
    
    // Make ThemeUtils globally available
    window.ThemeUtils = ThemeUtils;
    
    // Expose responsive API
    window.responsive = {
        isMobile: () => window.responsiveManager.isMobile(),
        isTablet: () => window.responsiveManager.isTablet(),
        isDesktop: () => window.responsiveManager.isDesktop(),
        getCurrentBreakpoint: () => window.responsiveManager.getCurrentBreakpoint(),
        refresh: () => window.responsiveManager.refresh(),
        onDataUpdate: () => window.responsiveManager.onDataUpdate()
    };
    
    console.log('🎨 Theme and responsive utilities loaded');
});

// Handle page visibility changes
document.addEventListener('visibilitychange', () => {
    if (!document.hidden && window.responsiveManager) {
        // Refresh layout when page becomes visible
        setTimeout(() => {
            window.responsiveManager.refresh();
        }, 100);
    }
});
