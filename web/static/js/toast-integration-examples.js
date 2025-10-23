/**
 * Toast Integration Examples
 * Shows how to integrate the toast notification system throughout the application
 */

// ============================================================================
// BASIC USAGE
// ============================================================================

// Simple success toast
Toast.success('Data saved successfully!');

// Success with description
Toast.success('User Created', 'The new user has been added to the system');

// Error toast
Toast.error('Operation Failed', 'Please try again');

// Warning toast
Toast.warning('Limited Access', 'You can only view 10 items at a time');

// Info toast
Toast.info('New Feature Available', 'Check out our new dashboard!');

// ============================================================================
// API INTEGRATION WITH AUTOMATIC ERROR HANDLING
// ============================================================================

// GET request with success toast
API.get('/api/users', { showSuccessToast: true, successMessage: 'Users loaded!' })
    .then(data => {
        console.log('Users:', data);
    });

// POST request with automatic error handling
API.post('/api/users', {
    username: 'john_doe',
    email: 'john@example.com'
}, {
    showSuccessToast: true,
    successMessage: 'User created successfully!'
})
    .then(data => {
        console.log('Created user:', data);
    })
    .catch(err => {
        // Error is automatically shown as toast
        console.error('Failed to create user:', err);
    });

// PUT request
API.put('/api/users/123', {
    full_name: 'John Doe Updated'
}, {
    showSuccessToast: true
})
    .then(data => {
        console.log('Updated:', data);
    });

// DELETE request
API.delete('/api/users/123', {
    showSuccessToast: true,
    successMessage: 'User deleted successfully'
})
    .then(() => {
        console.log('Deleted successfully');
    });

// ============================================================================
// ADVANCED TOAST USAGE
// ============================================================================

// Toast with custom duration (10 seconds)
Toast.success('Important Message', 'This will stay for 10 seconds', 10000);

// Toast that never auto-dismisses (duration: 0)
Toast.error('Critical Error', 'Action required before continuing', 0);

// Toast with action buttons
Toast.show({
    type: 'warning',
    title: 'Unsaved Changes',
    message: 'You have unsaved changes. What would you like to do?',
    duration: 0, // Don't auto-dismiss
    actions: [
        {
            text: 'Save',
            primary: true,
            onClick: () => {
                saveChanges();
                Toast.success('Changes saved!');
            }
        },
        {
            text: 'Discard',
            onClick: () => {
                discardChanges();
                Toast.info('Changes discarded');
            }
        }
    ]
});

// Toast with redirect action
Toast.show({
    type: 'error',
    title: 'Session Expired',
    message: 'Your session has expired. Please log in again.',
    duration: 0,
    actions: [
        {
            text: 'Login',
            primary: true,
            onClick: () => {
                window.location.href = '/login';
            }
        }
    ]
});

// ============================================================================
// REPLACING OLD MODAL-BASED ERRORS WITH TOASTS
// ============================================================================

// OLD WAY (with modals)
function showSuccess_OLD(message) {
    const modal = document.getElementById('successModal');
    const msg = document.getElementById('successMessage');
    msg.textContent = message;
    openModal(modal);
}

function showError_OLD(message) {
    const modal = document.getElementById('errorModal');
    const msg = document.getElementById('errorMessage');
    msg.textContent = message;
    openModal(modal);
}

// NEW WAY (with toasts) - much simpler!
function showSuccess(message) {
    Toast.success('Success', message);
}

function showError(message) {
    Toast.error('Error', message);
}

// ============================================================================
// FORM SUBMISSION WITH VALIDATION ERRORS
// ============================================================================

async function submitForm(formData) {
    try {
        const result = await API.post('/api/users', formData);
        Toast.success('User Created', 'The user has been added successfully');
        // Reload or update UI
    } catch (error) {
        // API automatically shows error toast with field-level validation
        // For additional handling:
        if (error.fields) {
            // Highlight invalid fields in the form
            error.fields.forEach(field => {
                const input = document.getElementById(field.field);
                if (input) {
                    input.classList.add('is-invalid');
                }
            });
        }
    }
}

// ============================================================================
// NAT UPDATE WITH TOAST FEEDBACK
// ============================================================================

async function updateNATRule(router, ip, port) {
    try {
        const result = await API.post('/api/nat/update', {
            router,
            ip,
            port
        }, {
            showSuccessToast: true,
            successMessage: `NAT rule updated for ${router}`
        });

        // Reload configurations
        await loadNATConfigurations();

        return result;
    } catch (error) {
        // Error automatically shown as toast
        console.error('NAT update failed:', error);
    }
}

// ============================================================================
// ROUTER CONNECTION TEST WITH TOAST
// ============================================================================

async function testRouterConnection(routerName) {
    try {
        Toast.info('Testing Connection', `Testing ${routerName}...`, 3000);

        const result = await API.get('/api/nat/test');

        if (result.status === 'success' && result.data[routerName]) {
            const status = result.data[routerName];

            if (status.status === 'connected') {
                Toast.success(`${routerName} Online`, `Connection successful!`);
            } else {
                Toast.error(`${routerName} Offline`, status.message || 'Connection failed');
            }
        }
    } catch (error) {
        // Error automatically shown
        console.error('Connection test failed:', error);
    }
}

// ============================================================================
// BATCH OPERATIONS WITH PROGRESS TOAST
// ============================================================================

async function deleteMultipleUsers(userIds) {
    const progressToast = Toast.show({
        type: 'info',
        title: 'Deleting Users',
        message: `Deleting ${userIds.length} users...`,
        duration: 0,
        dismissible: false
    });

    let successCount = 0;
    let failCount = 0;

    for (const userId of userIds) {
        try {
            await API.delete(`/api/users/${userId}`, { showErrorToast: false });
            successCount++;
        } catch (error) {
            failCount++;
        }
    }

    // Remove progress toast
    Toast.remove(progressToast);

    // Show result
    if (failCount === 0) {
        Toast.success('Success', `All ${successCount} users deleted successfully`);
    } else if (successCount === 0) {
        Toast.error('Failed', `Failed to delete all ${failCount} users`);
    } else {
        Toast.warning('Partial Success', `${successCount} deleted, ${failCount} failed`);
    }
}

// ============================================================================
// CLEAR ALL TOASTS
// ============================================================================

// Clear all active toasts (useful when navigating away or resetting state)
function clearAllNotifications() {
    Toast.clearAll();
}
