/**
 * ONT WiFi Information Extractor
 * Mengekstrak informasi SSID dan Password WiFi dari ONT
 * Support: Model Lama & GM220-S
 */

const { chromium } = require('playwright');
const fs = require('fs').promises;

class ONTWiFiExtractor {
    constructor(ontUrl, username = 'admin', password = 'admin', debug = false) {
        this.ontUrl = ontUrl;
        this.username = username;
        this.password = password;
        this.debug = debug;
        this.wifiInfo = {};
        this.ontModel = null;
    }

    log(message) {
        const timestamp = new Date().toLocaleTimeString('id-ID');
        console.log(`[${timestamp}] ${message}`);
    }

    async login(page) {
        this.log(`Mengakses ${this.ontUrl}...`);

        try {
            await page.goto(this.ontUrl, { timeout: 30000 });
        } catch (error) {
            throw new Error(`Tidak dapat mengakses ONT: ${error.message}`);
        }

        // Wait longer for page to fully load
        await page.waitForTimeout(3000);

        if (this.debug) {
            await page.screenshot({ path: 'debug_00_login_page.png' });
            this.log('[DEBUG] Screenshot saved: debug_00_login_page.png');
        }

        // Check for ZXHN F450 (different login flow)
        const title = await page.title();
        if (title.includes('ZXHN F450')) {
            await this.loginZTE_F450(page);
            return;
        }

        // Check if already logged in (no login form)
        const usernameExists = await page.locator('#username, #Frm_Username, input[name="username"]').count();

        if (usernameExists === 0) {
            this.log('Tidak ada form login - ONT sudah auto-login');
            return;
        }

        this.log('Mencari form login...');

        // Try to find visible username/password fields
        let usernameField, passwordField, loginButton;

        // Check if fields are directly visible
        const directUsername = page.locator('#username, #Frm_Username, input[name="username"]');
        const directPassword = page.locator('#password, #Frm_Password, input[name="password"]');

        const usernameCount = await directUsername.count();
        if (usernameCount > 0) {
            usernameField = directUsername.first();
            passwordField = directPassword.first();
            loginButton = page.getByRole('button', { name: /login|ok/i }).first();
        } else {
            throw new Error('Form login tidak ditemukan');
        }

        this.log('Melakukan login...');
        await usernameField.fill(this.username, { timeout: 10000 });
        await page.waitForTimeout(500);
        await passwordField.fill(this.password, { timeout: 10000 });
        await page.waitForTimeout(500);

        await loginButton.click();
        await page.waitForTimeout(4000);

        if (this.debug) {
            await page.screenshot({ path: 'debug_01_after_login.png' });
            this.log('[DEBUG] Screenshot saved: debug_01_after_login.png');
        }

        this.log('Login berhasil!');
    }

    async detectONTModel(page) {
        this.log('Mendeteksi model ONT...');

        try {
            // Cek page title untuk ZXHN F450
            const title = await page.title();
            if (title.includes('ZXHN F450')) {
                this.ontModel = 'ZTE_F450';
                this.log('Terdeteksi model: ZXHN F450');
                return 'ZTE_F450';
            }

            // Cek frame mainFrame (indikasi GM220-S)
            const frame = page.frame({ name: 'mainFrame' });
            if (frame) {
                const content = await page.content();
                if (content.includes('GM220-S')) {
                    this.ontModel = 'GM220-S';
                    this.log('Terdeteksi model: GM220-S');
                    return 'GM220-S';
                }
            }

            // Cek #Fstmenu (indikasi model lama)
            const fstmenu = await page.locator('#Fstmenu').count();
            if (fstmenu > 0) {
                this.ontModel = 'OLD_MODEL';
                this.log('Terdeteksi model: OLD_MODEL');
                return 'OLD_MODEL';
            }
        } catch (error) {
            this.log(`Warning: Gagal deteksi model: ${error.message}`);
        }

        // Default to old model
        this.ontModel = 'OLD_MODEL';
        return 'OLD_MODEL';
    }

    // ========== GM220-S Methods ==========

    async navigateToWLAN_GM220S(page) {
        this.log('[GM220-S] Membuka menu Network...');
        const frame = page.frame({ name: 'mainFrame' });

        if (!frame) {
            throw new Error('Frame mainFrame tidak ditemukan');
        }

        // Klik Network
        await frame.getByRole('cell', { name: '+Network', exact: true }).click();
        await page.waitForTimeout(2000);

        // Klik WLAN Radio2.4G
        this.log('[GM220-S] Membuka menu WLAN Radio2.4G...');
        await frame.getByRole('cell', { name: '+WLAN Radio2.4G', exact: true }).click();
        await page.waitForTimeout(2000);

        if (this.debug) {
            await page.screenshot({ path: 'debug_02_gm220s_wlan.png' });
            this.log('[DEBUG] Screenshot saved: debug_02_gm220s_wlan.png');
        }
    }

    async extractSSID_GM220S(page) {
        this.log('[GM220-S] Mengekstrak SSID...');
        const frame = page.frame({ name: 'mainFrame' });

        // Klik SSID Settings
        await frame.getByRole('cell', { name: 'SSID Settings', exact: true }).click();
        await page.waitForTimeout(2000);

        // Cari SSID dari input field
        const inputs = await frame.locator('input[type="text"]').all();

        for (const input of inputs) {
            try {
                const value = await input.inputValue();
                // SSID: 3-32 karakter, bukan pure number
                if (value && value.length >= 3 && value.length <= 32 && !/^\d+$/.test(value)) {
                    this.wifiInfo.ssid = value;
                    this.log(`[GM220-S] SSID ditemukan: ${value}`);
                    break;
                }
            } catch (error) {
                continue;
            }
        }

        if (this.debug) {
            await page.screenshot({ path: 'debug_03_gm220s_ssid.png' });
            this.log('[DEBUG] Screenshot saved: debug_03_gm220s_ssid.png');
        }
    }

    async extractPassword_GM220S(page) {
        this.log('[GM220-S] Mengekstrak password...');
        const frame = page.frame({ name: 'mainFrame' });

        // Klik Security
        await frame.getByRole('cell', { name: 'Security', exact: true }).click();
        await page.waitForTimeout(3000);

        // Ekstrak password dari input field (langsung terlihat)
        const inputs = await frame.locator('input[type="text"]').all();

        for (const input of inputs) {
            try {
                const value = await input.inputValue();
                // Password WiFi: minimal 8 karakter
                if (value && value.length >= 8) {
                    this.wifiInfo.password = value;
                    this.log(`[GM220-S] Password ditemukan: ${value}`);
                    break;
                }
            } catch (error) {
                continue;
            }
        }

        // Ekstrak Authentication Type
        try {
            const authOptions = await frame.locator('select option[selected]').all();
            for (const option of authOptions) {
                const text = await option.textContent();
                if (text && (text.includes('WPA') || text.includes('PSK'))) {
                    this.wifiInfo.authentication = text.trim();
                    this.log(`[GM220-S] Authentication: ${text.trim()}`);
                    break;
                }
            }
        } catch (error) {
            // Ignore
        }

        // Ekstrak Encryption Type
        try {
            const encryptOptions = await frame.locator('select option[selected]').all();
            for (const option of encryptOptions) {
                const text = await option.textContent();
                if (text && (text.includes('AES') || text.includes('TKIP'))) {
                    this.wifiInfo.encryption = text.trim();
                    this.log(`[GM220-S] Encryption: ${text.trim()}`);
                    break;
                }
            }
        } catch (error) {
            // Ignore
        }

        if (this.debug) {
            await page.screenshot({ path: 'debug_04_gm220s_password.png' });
            this.log('[DEBUG] Screenshot saved: debug_04_gm220s_password.png');
        }
    }

    // ========== ZXHN F450 Methods ==========

    async loginZTE_F450(page) {
        this.log('[ZTE_F450] Melakukan login...');

        // Fill username and password
        await page.locator('input[type="text"]').first().fill(this.username);
        await page.waitForTimeout(500);
        await page.locator('input[type="password"]').first().fill(this.password);
        await page.waitForTimeout(500);

        // Click login button
        await page.getByRole('button', { name: 'Login' }).click();
        await page.waitForTimeout(3000);

        if (this.debug) {
            await page.screenshot({ path: 'debug_01_zte_after_login.png' });
            this.log('[DEBUG] Screenshot saved: debug_01_zte_after_login.png');
        }

        this.log('[ZTE_F450] Login berhasil!');
    }

    async extractWiFi_ZTE_F450(page) {
        this.log('[ZTE_F450] Mengekstrak informasi WiFi dari dashboard...');

        // Wait for page to fully load
        await page.waitForTimeout(2000);

        // Find iframe that contains dashboard
        const frames = page.frames();
        let dashboardFrame = null;

        for (const frame of frames) {
            try {
                const wlanStatusCount = await frame.locator('text=WLAN Status').count();
                if (wlanStatusCount > 0) {
                    dashboardFrame = frame;
                    this.log('[ZTE_F450] Dashboard iframe ditemukan');
                    break;
                }
            } catch (error) {
                continue;
            }
        }

        if (!dashboardFrame) {
            this.log('[ZTE_F450] Warning: Dashboard iframe tidak ditemukan, coba main page');
            dashboardFrame = page;
        }

        if (this.debug) {
            await page.screenshot({ path: 'debug_02_zte_dashboard.png', fullPage: true });
            this.log('[DEBUG] Screenshot saved: debug_02_zte_dashboard.png');
        }

        // Extract SSID - cari row yang mengandung "SSID" dan ambil cell berikutnya
        try {
            const ssidRow = dashboardFrame.locator('tr:has-text("SSID")').first();
            const cells = await ssidRow.locator('td').all();

            if (cells.length >= 2) {
                const ssidValue = await cells[1].textContent();
                if (ssidValue && ssidValue.trim().length > 0) {
                    this.wifiInfo.ssid = ssidValue.trim();
                    this.log(`[ZTE_F450] SSID ditemukan: ${ssidValue.trim()}`);
                }
            }
        } catch (error) {
            this.log(`[ZTE_F450] Error ekstrak SSID: ${error.message}`);
        }

        // Extract Password (PSK) - cari row yang mengandung "PSK" dan ambil cell berikutnya
        try {
            const pskRow = dashboardFrame.locator('tr:has-text("PSK")').first();
            const cells = await pskRow.locator('td').all();

            if (cells.length >= 2) {
                const pskValue = await cells[1].textContent();
                if (pskValue && pskValue.trim().length >= 8) {
                    this.wifiInfo.password = pskValue.trim();
                    this.log(`[ZTE_F450] Password ditemukan: ${pskValue.trim()}`);
                }
            }
        } catch (error) {
            this.log(`[ZTE_F450] Error ekstrak password: ${error.message}`);
        }

        // Extract Admin Status
        try {
            const adminRow = dashboardFrame.locator('tr:has-text("Admin Status")').first();
            const cells = await adminRow.locator('td').all();

            if (cells.length >= 2) {
                const adminValue = await cells[1].textContent();
                if (adminValue) {
                    this.wifiInfo.admin_status = adminValue.trim();
                    this.log(`[ZTE_F450] Admin Status: ${adminValue.trim()}`);
                }
            }
        } catch (error) {
            // Ignore
        }

        // Extract Encryption Type
        try {
            const encryptRow = dashboardFrame.locator('tr:has-text("Encryption Type")').first();
            const cells = await encryptRow.locator('td').all();

            if (cells.length >= 2) {
                const encryptValue = await cells[1].textContent();
                if (encryptValue) {
                    this.wifiInfo.encryption = encryptValue.trim();
                    this.log(`[ZTE_F450] Encryption: ${encryptValue.trim()}`);
                }
            }
        } catch (error) {
            // Ignore
        }

        if (this.debug) {
            await page.screenshot({ path: 'debug_03_zte_extracted.png' });
            this.log('[DEBUG] Screenshot saved: debug_03_zte_extracted.png');
        }
    }

    // ========== Old Model Methods ==========

    async navigateToWLAN_OldModel(page) {
        this.log('[OLD] Membuka menu NETWORK...');

        // Cek apakah konten ada di iframe
        const frames = page.frames();
        let workingFrame = page;

        // Cari iframe yang berisi konten (biasanya ada menu)
        for (const frame of frames) {
            const menuCount = await frame.locator('a:has-text("NETWORK"), link:has-text("NETWORK")').count();
            if (menuCount > 0) {
                workingFrame = frame;
                this.log('[OLD] Konten ditemukan di iframe');
                break;
            }
        }

        // Coba click NETWORK
        try {
            await workingFrame.locator('#Fstmenu').getByRole('link', { name: 'NETWORK' }).click();
        } catch (error) {
            // Try alternative selector
            await workingFrame.getByRole('link', { name: 'NETWORK' }).click();
        }

        await page.waitForTimeout(2000);

        this.log('[OLD] Membuka menu WLAN SET...');

        // Wait for WLAN SET link and click
        await page.waitForTimeout(1000);

        try {
            await workingFrame.getByRole('link', { name: 'WLAN SET' }).click();
        } catch (error) {
            // Try finding in new frame after navigation
            const newFrames = page.frames();
            for (const frame of newFrames) {
                const wlanCount = await frame.locator('a:has-text("WLAN SET"), link:has-text("WLAN SET")').count();
                if (wlanCount > 0) {
                    await frame.getByRole('link', { name: 'WLAN SET' }).click();
                    break;
                }
            }
        }

        await page.waitForTimeout(2000);

        if (this.debug) {
            await page.screenshot({ path: 'debug_02_old_wlan.png' });
            this.log('[DEBUG] Screenshot saved: debug_02_old_wlan.png');
        }
    }

    async extractSSID_OldModel(page) {
        this.log('[OLD] Mengekstrak SSID...');

        try {
            // Wait for page to be ready
            await page.waitForTimeout(1000);

            // Cari iframe yang berisi content
            const frames = page.frames();
            let workingFrame = page;

            for (const frame of frames) {
                const inputCount = await frame.locator('input[type="text"]').count();
                if (inputCount > 0) {
                    workingFrame = frame;
                    this.log(`[OLD] Konten form ditemukan di iframe (${inputCount} inputs)`);
                    break;
                }
            }

            // DEBUG: Dump all input values
            if (this.debug) {
                const allInputs = await workingFrame.locator('input[type="text"]').all();
                this.log(`[DEBUG] Found ${allInputs.length} text inputs in working frame`);
                for (let i = 0; i < Math.min(allInputs.length, 10); i++) {
                    try {
                        const value = await allInputs[i].inputValue();
                        const name = await allInputs[i].getAttribute('name');
                        this.log(`[DEBUG] Input ${i}: name="${name}" value="${value}"`);
                    } catch (error) {
                        // Skip
                    }
                }
            }

            // Coba berbagai cara ekstrak SSID
            const selectors = [
                'input[name="ssid"]',
                'input[name="SSID"]',
                'td:has-text("SSID:") + td input',
                'input[type="text"]' // Fallback: cari semua text input
            ];

            for (const selector of selectors) {
                try {
                    const locator = workingFrame.locator(selector);
                    const count = await locator.count();

                    if (count > 0) {
                        // Jika banyak input, cari yang punya value
                        for (let i = 0; i < count; i++) {
                            const element = locator.nth(i);
                            const value = await element.inputValue();

                            // SSID biasanya 3-32 karakter, bukan pure number
                            if (value && value.length >= 3 && value.length <= 32 && !/^\d+$/.test(value)) {
                                // Skip jika value seperti BSSID atau IP
                                if (!value.includes(':') && !value.includes('.')) {
                                    this.wifiInfo.ssid = value;
                                    this.log(`[OLD] SSID ditemukan: ${value} (selector: ${selector})`);
                                    return;
                                }
                            }
                        }
                    }
                } catch (error) {
                    continue;
                }
            }

            this.log('[OLD] Warning: SSID tidak ditemukan di WLAN BASIC, akan dicoba dari SECURITY page');
        } catch (error) {
            this.log(`[OLD] Error ekstrak SSID: ${error.message}`);
        }
    }

    async extractPassword_OldModel(page) {
        this.log('[OLD] Membuka menu SECURITY...');

        // Cari iframe yang berisi menu
        const frames = page.frames();
        let workingFrame = page;

        for (const frame of frames) {
            const securityCount = await frame.locator('a:has-text("SECURITY"), link:has-text("SECURITY")').count();
            if (securityCount > 0) {
                workingFrame = frame;
                break;
            }
        }

        try {
            await workingFrame.locator('[id="21"]').getByRole('link', { name: 'SECURITY' }).click();
        } catch (error) {
            await workingFrame.getByRole('link', { name: 'SECURITY' }).click();
        }

        await page.waitForTimeout(2000);

        // Find new working frame for content after navigation
        const newFrames = page.frames();
        for (const frame of newFrames) {
            const inputCount = await frame.locator('input').count();
            if (inputCount > 0) {
                workingFrame = frame;
                break;
            }
        }

        // Scroll down untuk melihat Manual AP Setup section
        this.log('[OLD] Scroll down untuk melihat Manual AP Setup...');
        try {
            // Try multiple scroll methods
            await workingFrame.evaluate(() => {
                window.scrollTo(0, document.body.scrollHeight);
                // Also try scrolling parent if in iframe
                if (window.parent !== window) {
                    window.parent.scrollTo(0, document.body.scrollHeight);
                }
            });
        } catch (error) {
            this.log(`[OLD] Warning: Scroll gagal: ${error.message}`);
        }

        await page.waitForTimeout(2000);

        // Scroll sekali lagi untuk memastikan
        try {
            await workingFrame.evaluate(() => window.scrollBy(0, 1000));
        } catch (error) {
            // Ignore
        }

        await page.waitForTimeout(1000);

        if (this.debug) {
            await page.screenshot({ path: 'debug_03_old_security_before.png', fullPage: true });
            this.log('[DEBUG] Screenshot saved: debug_03_old_security_before.png');
        }

        // Ekstrak SSID dari dropdown jika belum dapat
        if (!this.wifiInfo.ssid) {
            try {
                // Try to find SSID from dropdown/select
                const ssidSelectors = [
                    'select option[selected]',
                    'select[name*="ssid"] option[selected]',
                    'td:has-text("Select SSID") + td select option[selected]'
                ];

                for (const selector of ssidSelectors) {
                    const locator = workingFrame.locator(selector);
                    const count = await locator.count();
                    if (count > 0) {
                        const ssid = await locator.first().textContent();
                        if (ssid && ssid.length >= 3 && ssid.length <= 32) {
                            this.wifiInfo.ssid = ssid.trim();
                            this.log(`[OLD] SSID ditemukan dari dropdown: ${ssid.trim()}`);
                            break;
                        }
                    }
                }
            } catch (error) {
                this.log(`[OLD] Warning: Gagal ekstrak SSID dari dropdown: ${error.message}`);
            }
        }

        // Klik "Click here to display" - coba berbagai cara
        this.log('[OLD] Mencari link "Click here to display"...');
        let displayClicked = false;

        try {
            // Method 1: getByText
            const displayLink1 = workingFrame.getByText('Click here to display', { exact: false });
            if (await displayLink1.count() > 0) {
                this.log('[OLD] Klik "Click here to display" (method 1)...');
                await displayLink1.first().click();
                await page.waitForTimeout(2000);
                displayClicked = true;
                this.log('[OLD] Berhasil klik display link');
            }
        } catch (error) {
            this.log(`[OLD] Method 1 failed: ${error.message}`);
        }

        if (!displayClicked) {
            try {
                // Method 2: Link selector
                const displayLink2 = workingFrame.locator('a:has-text("Click here to display")');
                if (await displayLink2.count() > 0) {
                    this.log('[OLD] Klik "Click here to display" (method 2)...');
                    await displayLink2.first().click();
                    await page.waitForTimeout(2000);
                    displayClicked = true;
                    this.log('[OLD] Berhasil klik display link');
                }
            } catch (error) {
                this.log(`[OLD] Method 2 failed: ${error.message}`);
            }
        }

        if (!displayClicked) {
            try {
                // Method 3: Partial text match
                const displayLink3 = workingFrame.locator('a').filter({ hasText: /display/i });
                if (await displayLink3.count() > 0) {
                    this.log('[OLD] Klik "Click here to display" (method 3)...');
                    await displayLink3.first().click();
                    await page.waitForTimeout(2000);
                    displayClicked = true;
                    this.log('[OLD] Berhasil klik display link');
                }
            } catch (error) {
                this.log(`[OLD] Method 3 failed: ${error.message}`);
            }
        }

        if (!displayClicked) {
            this.log('[OLD] Warning: Tidak bisa klik display link, password mungkin sudah terlihat atau link tidak ada');
        }

        if (this.debug) {
            await page.screenshot({ path: 'debug_04_old_security_after.png' });
            this.log('[DEBUG] Screenshot saved: debug_04_old_security_after.png');
        }

        // Ekstrak password - coba berbagai selector
        this.log('[OLD] Mengekstrak password...');
        const passwordSelectors = [
            'input[name="wlWpaPsk"]',
            'input[name*="assphrase"]',
            'input[name*="password"]',
            'td:has-text("WPA/WAPI passphrase:") + td input',
            'td:has-text("passphrase") + td input'
        ];

        for (const selector of passwordSelectors) {
            try {
                const count = await workingFrame.locator(selector).count();
                if (count > 0) {
                    const password = await workingFrame.locator(selector).first().inputValue();
                    if (password && password.length >= 8) {
                        this.wifiInfo.password = password;
                        this.log(`[OLD] Password ditemukan: ${password}`);
                        break;
                    }
                }
            } catch (error) {
                continue;
            }
        }

        // Ekstrak authentication & encryption
        try {
            // Authentication
            const authSelectors = [
                'select option[selected]:has-text("WPA")',
                'td:has-text("Network Authentication:") + td select option[selected]',
                'select[name*="uthen"] option[selected]'
            ];

            for (const selector of authSelectors) {
                const locator = workingFrame.locator(selector);
                if (await locator.count() > 0) {
                    const text = await locator.first().textContent();
                    if (text && text.includes('WPA')) {
                        this.wifiInfo.authentication = text.trim();
                        this.log(`[OLD] Authentication: ${text.trim()}`);
                        break;
                    }
                }
            }
        } catch (error) {
            // Ignore
        }

        try {
            // Encryption
            const encryptSelectors = [
                'td:has-text("WPA/WAPI Encryption:") + td select option[selected]',
                'select[name*="ncryp"] option[selected]'
            ];

            for (const selector of encryptSelectors) {
                const locator = workingFrame.locator(selector);
                if (await locator.count() > 0) {
                    const text = await locator.first().textContent();
                    if (text) {
                        this.wifiInfo.encryption = text.trim();
                        this.log(`[OLD] Encryption: ${text.trim()}`);
                        break;
                    }
                }
            }
        } catch (error) {
            // Ignore
        }

        if (!this.wifiInfo.password) {
            this.log('[OLD] Password tidak ditemukan');
        }
    }

    // ========== Main Extract Method ==========

    async extractWiFiInfo() {
        const browser = await chromium.launch({
            headless: !this.debug
        });

        const context = await browser.newContext();
        const page = await context.newPage();

        try {
            // Login
            await this.login(page);

            // Detect model
            const model = await this.detectONTModel(page);

            console.log(`\n[INFO] Menggunakan strategi ekstraksi untuk ${model}\n`);

            // Extract based on model
            if (model === 'ZTE_F450') {
                await this.extractWiFi_ZTE_F450(page);
            } else if (model === 'GM220-S') {
                await this.navigateToWLAN_GM220S(page);
                await this.extractSSID_GM220S(page);
                await this.extractPassword_GM220S(page);
            } else {
                await this.navigateToWLAN_OldModel(page);
                await this.extractSSID_OldModel(page);
                await this.extractPassword_OldModel(page);
            }

            // Add metadata
            this.wifiInfo.extracted_at = new Date().toISOString();
            this.wifiInfo.ont_url = this.ontUrl;
            this.wifiInfo.ont_model = model;

            // Print results
            console.log('\n' + '='.repeat(50));
            console.log('Informasi WiFi ONT');
            console.log('='.repeat(50));
            console.log(`ONT URL       : ${this.wifiInfo.ont_url || 'N/A'}`);
            console.log(`ONT Model     : ${this.wifiInfo.ont_model || 'N/A'}`);
            console.log(`SSID          : ${this.wifiInfo.ssid || 'N/A'}`);
            console.log(`Password      : ${this.wifiInfo.password || 'N/A'}`);
            console.log(`Authentication: ${this.wifiInfo.authentication || 'N/A'}`);
            console.log(`Encryption    : ${this.wifiInfo.encryption || 'N/A'}`);
            console.log(`Extracted At  : ${this.wifiInfo.extracted_at || 'N/A'}`);
            console.log('='.repeat(50) + '\n');

        } catch (error) {
            console.error(`[ERROR] ${error.message}`);
            this.wifiInfo.error = error.message;

            if (this.debug) {
                console.error('\n[DEBUG] Stack trace:');
                console.error(error.stack);
            }
        } finally {
            await browser.close();
        }

        return this.wifiInfo;
    }

    async saveToJSON(outputFile = 'wifi_info.json') {
        await fs.writeFile(outputFile, JSON.stringify(this.wifiInfo, null, 2));
        this.log(`Informasi disimpan ke ${outputFile}`);
    }
}

// ========== Main Function ==========

async function main() {
    const args = process.argv.slice(2);

    if (args.length < 1) {
        console.log('Usage: node ont-wifi-extractor.js <ONT_URL> [username] [password] [--debug]');
        console.log('\nContoh:');
        console.log('  node ont-wifi-extractor.js http://tunnel3.ebilling.id:20131/');
        console.log('  node ont-wifi-extractor.js http://192.168.1.1/ admin admin123');
        console.log('  node ont-wifi-extractor.js http://192.168.1.1/ admin admin --debug');
        process.exit(1);
    }

    // Parse arguments
    const ontUrl = args[0];
    let username = 'admin';
    let password = 'admin';
    let debug = args.includes('--debug');

    // Remove --debug from args
    const filteredArgs = args.filter(arg => arg !== '--debug');

    if (filteredArgs.length > 1) username = filteredArgs[1];
    if (filteredArgs.length > 2) password = filteredArgs[2];

    // Create extractor
    const extractor = new ONTWiFiExtractor(ontUrl, username, password, debug);

    if (debug) {
        console.log('\n[DEBUG MODE ENABLED]');
        console.log('- Browser akan ditampilkan (non-headless)');
        console.log('- Screenshot akan disimpan otomatis');
        console.log('='.repeat(50) + '\n');
    }

    // Extract
    const wifiInfo = await extractor.extractWiFiInfo();

    // Save to JSON
    if (wifiInfo && !wifiInfo.error) {
        await extractor.saveToJSON();
        console.log('\n[SUCCESS] Ekstraksi berhasil!');
        process.exit(0);
    } else {
        console.log('\n[FAILED] Ekstraksi gagal!');
        if (debug) {
            console.log('\nPeriksa screenshot debug_*.png untuk informasi lebih lanjut');
        }
        process.exit(1);
    }
}

// Run
if (require.main === module) {
    main().catch(error => {
        console.error('[FATAL ERROR]', error);
        process.exit(1);
    });
}

module.exports = ONTWiFiExtractor;
