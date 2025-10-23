// ONT WiFi Extractor - Client-side JavaScript

const socket = io();

// DOM Elements
const extractionForm = document.getElementById('extractionForm');
const useCustomCredentialsCheckbox = document.getElementById('useCustomCredentials');
const credentialsSection = document.getElementById('credentialsSection');
const debugModeCheckbox = document.getElementById('debugMode');

const statusSection = document.getElementById('statusSection');
const logsSection = document.getElementById('logsSection');
const resultsSection = document.getElementById('resultsSection');
const errorSection = document.getElementById('errorSection');

const logsContainer = document.getElementById('logsContainer');
const statusText = document.getElementById('statusText');

const extractBtn = document.getElementById('extractBtn');
const copyBtn = document.getElementById('copyBtn');
const newExtractionBtn = document.getElementById('newExtractionBtn');
const retryBtn = document.getElementById('retryBtn');

// Toggle credentials section
useCustomCredentialsCheckbox.addEventListener('change', (e) => {
    credentialsSection.style.display = e.target.checked ? 'block' : 'none';
});

// Form submission
extractionForm.addEventListener('submit', (e) => {
    e.preventDefault();

    const ontUrl = document.getElementById('ontUrl').value.trim();
    const username = useCustomCredentialsCheckbox.checked ? document.getElementById('username').value.trim() : null;
    const password = useCustomCredentialsCheckbox.checked ? document.getElementById('password').value.trim() : null;
    const debug = debugModeCheckbox.checked;

    if (!ontUrl) {
        alert('Mohon masukkan URL ONT!');
        return;
    }

    startExtraction(ontUrl, username, password, debug);
});

// Start extraction
function startExtraction(url, username, password, debug) {
    // Hide all result sections
    resultsSection.style.display = 'none';
    errorSection.style.display = 'none';

    // Show status and logs
    statusSection.style.display = 'block';
    logsSection.style.display = 'block';

    // Clear previous logs
    logsContainer.innerHTML = '';

    // Disable form
    extractBtn.disabled = true;
    extractBtn.textContent = '⏳ Processing...';

    // Update status
    statusText.textContent = 'Menghubungkan ke ONT...';

    // Emit extraction request
    socket.emit('start-extraction', {
        url,
        username,
        password,
        debug
    });
}

// Socket.IO event listeners
socket.on('log', (data) => {
    addLog(data.message, data.type || 'info');
});

socket.on('status', (data) => {
    updateStatus(data.status);
});

socket.on('result', (data) => {
    handleResult(data);
});

socket.on('connect', () => {
    console.log('Connected to server');
});

socket.on('disconnect', () => {
    console.log('Disconnected from server');
});

// Add log entry
function addLog(message, type = 'info') {
    const logEntry = document.createElement('div');
    logEntry.className = `log-entry ${type}`;

    const timestamp = new Date().toLocaleTimeString('id-ID');
    logEntry.textContent = `[${timestamp}] ${message}`;

    logsContainer.appendChild(logEntry);
    logsContainer.scrollTop = logsContainer.scrollHeight;
}

// Update status text
function updateStatus(status) {
    const statusMessages = {
        'detecting': '🔍 Mendeteksi model ONT...',
        'logging-in': '🔐 Mencoba login...',
        'extracting': '📡 Mengekstrak informasi WiFi...',
        'success': '✅ Extraction berhasil!',
        'error': '❌ Extraction gagal!'
    };

    statusText.textContent = statusMessages[status] || status;
}

// Handle extraction result
function handleResult(data) {
    // Enable form
    extractBtn.disabled = false;
    extractBtn.textContent = '🚀 Start Extraction';

    // Hide status
    statusSection.style.display = 'none';

    if (data.success) {
        // Show results
        resultsSection.style.display = 'block';
        errorSection.style.display = 'none';

        // Populate results
        document.getElementById('resultModel').textContent = data.model || 'N/A';
        document.getElementById('resultSSID').textContent = data.data.ssid || 'N/A';
        document.getElementById('resultPassword').textContent = data.data.password || 'N/A';
        document.getElementById('resultSecurity').textContent = data.data.security || 'N/A';
        document.getElementById('resultEncryption').textContent = data.data.encryption || 'N/A';

        // Show credentials if available
        if (data.credentials) {
            document.getElementById('credentialsRow').style.display = 'flex';
            document.getElementById('resultCredentials').textContent =
                `${data.credentials.username} / ${data.credentials.password}`;
        } else {
            document.getElementById('credentialsRow').style.display = 'none';
        }

        addLog('✅ Extraction berhasil!', 'success');
    } else {
        // Show error
        resultsSection.style.display = 'none';
        errorSection.style.display = 'block';

        document.getElementById('errorMessage').textContent = data.error || 'Unknown error';

        addLog(`❌ Error: ${data.error}`, 'error');
    }
}

// Copy password to clipboard
copyBtn.addEventListener('click', () => {
    const password = document.getElementById('resultPassword').textContent;

    if (password && password !== 'N/A') {
        navigator.clipboard.writeText(password).then(() => {
            const originalText = copyBtn.textContent;
            copyBtn.textContent = '✅ Copied!';

            setTimeout(() => {
                copyBtn.textContent = originalText;
            }, 2000);
        }).catch(err => {
            alert('Failed to copy password: ' + err);
        });
    }
});

// New extraction button
newExtractionBtn.addEventListener('click', () => {
    resultsSection.style.display = 'none';
    errorSection.style.display = 'none';
    logsSection.style.display = 'none';

    // Clear form
    document.getElementById('ontUrl').value = '';
    document.getElementById('username').value = '';
    document.getElementById('password').value = '';
    useCustomCredentialsCheckbox.checked = false;
    debugModeCheckbox.checked = false;
    credentialsSection.style.display = 'none';

    // Scroll to top
    window.scrollTo({ top: 0, behavior: 'smooth' });
});

// Retry button
retryBtn.addEventListener('click', () => {
    errorSection.style.display = 'none';

    // Retry with same data
    extractionForm.dispatchEvent(new Event('submit'));
});
