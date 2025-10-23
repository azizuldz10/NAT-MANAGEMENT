/**
 * Toast Notification System
 * Provides beautiful, accessible toast notifications with queue management
 *
 * Usage:
 *   Toast.success('Operation completed!');
 *   Toast.error('Something went wrong', 'Please try again');
 *   Toast.warning('Warning message', 'Optional description');
 *   Toast.info('Information', 'Details here');
 *
 * Advanced:
 *   Toast.show({
 *     type: 'success',
 *     title: 'Success!',
 *     message: 'Your data has been saved.',
 *     duration: 5000,
 *     actions: [
 *       { text: 'Undo', onClick: () => undoAction() },
 *       { text: 'View', onClick: () => viewItem() }
 *     ]
 *   });
 */

class ToastNotification {
    constructor() {
        this.container = null;
        this.toasts = [];
        this.maxToasts = 5;
        this.defaultDuration = 5000; // 5 seconds
        this.init();
    }

    init() {
        // Create container if it doesn't exist
        if (!this.container) {
            this.container = document.createElement('div');
            this.container.className = 'toast-container';
            this.container.setAttribute('role', 'region');
            this.container.setAttribute('aria-label', 'Notifications');
            this.container.setAttribute('aria-live', 'polite');
            document.body.appendChild(this.container);
        }
    }

    /**
     * Show a toast notification
     * @param {Object} options - Toast options
     * @param {string} options.type - Toast type (success, error, warning, info)
     * @param {string} options.title - Toast title
     * @param {string} [options.message] - Optional message
     * @param {number} [options.duration] - Duration in ms (0 = no auto-dismiss)
     * @param {Array} [options.actions] - Action buttons
     * @param {boolean} [options.dismissible] - Show close button (default: true)
     */
    show(options) {
        const {
            type = 'info',
            title,
            message = '',
            duration = this.defaultDuration,
            actions = [],
            dismissible = true
        } = options;

        // Limit number of toasts
        if (this.toasts.length >= this.maxToasts) {
            this.remove(this.toasts[0]);
        }

        // Create toast element
        const toast = this.createToast({
            type,
            title,
            message,
            duration,
            actions,
            dismissible
        });

        // Add to container and track
        this.container.appendChild(toast.element);
        this.toasts.push(toast);

        // Trigger animation
        requestAnimationFrame(() => {
            toast.element.classList.add('show');
        });

        // Auto-dismiss if duration > 0
        if (duration > 0) {
            this.startAutoDismiss(toast, duration);
        }

        return toast;
    }

    createToast(options) {
        const { type, title, message, duration, actions, dismissible } = options;

        // Create toast element
        const element = document.createElement('div');
        element.className = `toast ${type}`;
        element.setAttribute('role', 'alert');
        element.setAttribute('aria-live', type === 'error' ? 'assertive' : 'polite');

        // Icon
        const icon = this.getIcon(type);
        const iconHTML = `<div class="toast-icon">${icon}</div>`;

        // Content
        const contentHTML = `
            <div class="toast-content">
                <p class="toast-title">${this.escapeHtml(title)}</p>
                ${message ? `<p class="toast-message">${this.escapeHtml(message)}</p>` : ''}
                ${actions.length > 0 ? this.createActionsHTML(actions) : ''}
            </div>
        `;

        // Close button
        const closeHTML = dismissible ? `
            <button class="toast-close" aria-label="Close notification" type="button">
                ×
            </button>
        ` : '';

        // Progress bar
        const progressHTML = duration > 0 ? '<div class="toast-progress"></div>' : '';

        element.innerHTML = iconHTML + contentHTML + closeHTML + progressHTML;

        // Event listeners
        if (dismissible) {
            const closeBtn = element.querySelector('.toast-close');
            closeBtn.addEventListener('click', () => this.remove({ element }));
        }

        // Action button listeners
        if (actions.length > 0) {
            const actionBtns = element.querySelectorAll('.toast-action-btn');
            actionBtns.forEach((btn, index) => {
                btn.addEventListener('click', () => {
                    if (actions[index].onClick) {
                        actions[index].onClick();
                    }
                    if (actions[index].dismissOnClick !== false) {
                        this.remove({ element });
                    }
                });
            });
        }

        return { element, type, duration };
    }

    createActionsHTML(actions) {
        const actionsHTML = actions.map(action => {
            const className = action.primary ? 'toast-action-btn primary' : 'toast-action-btn';
            return `<button class="${className}" type="button">${this.escapeHtml(action.text)}</button>`;
        }).join('');

        return `<div class="toast-actions">${actionsHTML}</div>`;
    }

    startAutoDismiss(toast, duration) {
        const progressBar = toast.element.querySelector('.toast-progress');

        if (progressBar) {
            progressBar.style.width = '100%';
            progressBar.style.transitionDuration = `${duration}ms`;

            requestAnimationFrame(() => {
                progressBar.style.width = '0%';
            });
        }

        toast.timeout = setTimeout(() => {
            this.remove(toast);
        }, duration);

        // Pause on hover
        toast.element.addEventListener('mouseenter', () => {
            if (toast.timeout) {
                clearTimeout(toast.timeout);
                if (progressBar) {
                    const computedWidth = window.getComputedStyle(progressBar).width;
                    progressBar.style.width = computedWidth;
                    progressBar.style.transitionDuration = '0s';
                }
            }
        });

        // Resume on leave
        toast.element.addEventListener('mouseleave', () => {
            if (progressBar) {
                const currentWidth = parseFloat(window.getComputedStyle(progressBar).width);
                const totalWidth = toast.element.offsetWidth;
                const remainingRatio = currentWidth / totalWidth;
                const remainingDuration = duration * remainingRatio;

                progressBar.style.transitionDuration = `${remainingDuration}ms`;
                progressBar.style.width = '0%';

                toast.timeout = setTimeout(() => {
                    this.remove(toast);
                }, remainingDuration);
            }
        });
    }

    remove(toast) {
        if (!toast || !toast.element) return;

        // Clear timeout
        if (toast.timeout) {
            clearTimeout(toast.timeout);
        }

        // Trigger exit animation
        toast.element.classList.remove('show');
        toast.element.classList.add('hide');

        // Remove from DOM after animation
        setTimeout(() => {
            if (toast.element && toast.element.parentNode) {
                toast.element.parentNode.removeChild(toast.element);
            }
            // Remove from tracking array
            this.toasts = this.toasts.filter(t => t !== toast);
        }, 300);
    }

    getIcon(type) {
        const icons = {
            success: '✓',
            error: '✕',
            warning: '⚠',
            info: 'ℹ'
        };
        return icons[type] || icons.info;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Convenience methods
    success(title, message, duration) {
        return this.show({ type: 'success', title, message, duration });
    }

    error(title, message, duration) {
        return this.show({ type: 'error', title, message, duration: duration || 7000 });
    }

    warning(title, message, duration) {
        return this.show({ type: 'warning', title, message, duration });
    }

    info(title, message, duration) {
        return this.show({ type: 'info', title, message, duration });
    }

    // Clear all toasts
    clearAll() {
        this.toasts.forEach(toast => this.remove(toast));
    }
}

// Create global instance
const Toast = new ToastNotification();

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = Toast;
}
