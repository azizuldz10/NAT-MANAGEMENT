/**
 * ZTE ZXHN F450 WiFi Information Extractor
 * Khusus untuk model ZTE ZXHN F450 (EPON ONU)
 * Ekstraksi langsung dari dashboard
 */

const { chromium } = require('playwright');
const fs = require('fs').promises;

class ZTEF450Extractor {
    constructor(ontUrl, username = 'admin', password = 'admin', debug = false) {
        this.ontUrl = ontUrl;
        this.username = username;
        this.password = password;
        this.debug = debug;
        this.wifiInfo = {};
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

        await page.waitForTimeout(3000);

        if (this.debug) {
            await page.screenshot({ path: 'zte_debug_00_login.png' });
            this.log('[DEBUG] Screenshot: zte_debug_00_login.png');
        }

        // Check if it's ZTE ZXHN series (F450, F477V2, etc)
        const title = await page.title();
        const content = await page.content();

        // Check for ZTE ZXHN series indicators
        const isZTE = title.includes('ZXHN') || content.includes('ZTE') || content.includes('WELCOME');

        if (!isZTE) {
            throw new Error(`Bukan ZTE ZXHN series! Page title: ${title}`);
        }

        this.log(`Terdeteksi ZTE model (title: ${title || 'Empty'})`);

        this.log('Melakukan login...');

        // Check if it's the WELCOME page with Administrator/User selection
        const welcomeText = await page.locator('text=WELCOME').count();

        if (welcomeText > 0) {
            this.log('Detected WELCOME page, clicking Administrator...');

            // Click Administrator using JavaScript
            await page.evaluate(() => {
                if (typeof GetAccessAccount === 'function') {
                    GetAccessAccount('adm');
                }
            });

            await page.waitForTimeout(1000);

            if (this.debug) {
                await page.screenshot({ path: 'zte_debug_01a_after_click_admin.png' });
                this.log('[DEBUG] Screenshot: zte_debug_01a_after_click_admin.png');
            }
        }

        // Now fill password (and username if visible)
        const usernameField = page.locator('input[type="text"]').first();
        const usernameCount = await usernameField.count();

        if (usernameCount > 0) {
            await usernameField.fill(this.username);
            await page.waitForTimeout(500);
        }

        await page.locator('input[type="password"]').first().fill(this.password);
        await page.waitForTimeout(500);

        if (this.debug) {
            await page.screenshot({ path: 'zte_debug_01_before_login.png' });
            this.log('[DEBUG] Screenshot: zte_debug_01_before_login.png');
        }

        // Click login button (can be "Login" or "GO")
        try {
            await page.getByRole('button', { name: /Login|GO/i }).click();
        } catch (error) {
            // Try alternative: click by button text
            const goButton = page.locator('button:has-text("GO"), input[value="GO"]');
            if (await goButton.count() > 0) {
                await goButton.first().click();
            } else {
                throw new Error('Login button tidak ditemukan');
            }
        }
        await page.waitForTimeout(4000);

        if (this.debug) {
            await page.screenshot({ path: 'zte_debug_02_after_login.png' });
            this.log('[DEBUG] Screenshot: zte_debug_02_after_login.png');
        }

        // Check if login successful by looking for dashboard
        const currentUrl = page.url();
        if (currentUrl.includes('start.ghtml')) {
            this.log('Login berhasil!');
        } else {
            // Check for error message
            const bodyText = await page.locator('body').textContent();
            if (bodyText.includes('incorrect') || bodyText.includes('wrong')) {
                throw new Error('Login gagal - Username atau password salah');
            }
            this.log('Warning: URL tidak sesuai expected, tapi lanjut ekstraksi...');
        }
    }

    async extractWiFiInfo(page) {
        this.log('Mengekstrak informasi WiFi dari dashboard...');

        await page.waitForTimeout(2000);

        // Find iframe that contains dashboard
        const frames = page.frames();
        let dashboardFrame = null;

        this.log('Mencari dashboard iframe...');
        for (const frame of frames) {
            try {
                const wlanStatusCount = await frame.locator('text=WLAN Status').count();
                if (wlanStatusCount > 0) {
                    dashboardFrame = frame;
                    this.log('Dashboard iframe ditemukan!');
                    break;
                }
            } catch (error) {
                continue;
            }
        }

        if (!dashboardFrame) {
            throw new Error('Dashboard iframe tidak ditemukan!');
        }

        if (this.debug) {
            await page.screenshot({ path: 'zte_debug_03_dashboard.png', fullPage: true });
            this.log('[DEBUG] Screenshot: zte_debug_03_dashboard.png');
        }

        // Extract SSID
        try {
            const ssidRow = dashboardFrame.locator('tr:has-text("SSID")').first();
            const cells = await ssidRow.locator('td').all();

            if (cells.length >= 2) {
                const ssidValue = await cells[1].textContent();
                if (ssidValue && ssidValue.trim().length > 0) {
                    this.wifiInfo.ssid = ssidValue.trim();
                    this.log(`✓ SSID: ${ssidValue.trim()}`);
                }
            }
        } catch (error) {
            this.log(`✗ Error ekstrak SSID: ${error.message}`);
        }

        // Extract Password (PSK)
        try {
            const pskRow = dashboardFrame.locator('tr:has-text("PSK")').first();
            const cells = await pskRow.locator('td').all();

            if (cells.length >= 2) {
                const pskValue = await cells[1].textContent();
                if (pskValue && pskValue.trim().length >= 8) {
                    this.wifiInfo.password = pskValue.trim();
                    this.log(`✓ Password: ${pskValue.trim()}`);
                }
            }
        } catch (error) {
            this.log(`✗ Error ekstrak password: ${error.message}`);
        }

        // Extract Admin Status
        try {
            const adminRow = dashboardFrame.locator('tr:has-text("Admin Status")').first();
            const cells = await adminRow.locator('td').all();

            if (cells.length >= 2) {
                const adminValue = await cells[1].textContent();
                if (adminValue) {
                    this.wifiInfo.status = adminValue.trim();
                    this.log(`✓ Status: ${adminValue.trim()}`);
                }
            }
        } catch (error) {
            // Not critical
        }

        // Extract Encryption Type
        try {
            const encryptRow = dashboardFrame.locator('tr:has-text("Encryption Type")').first();
            const cells = await encryptRow.locator('td').all();

            if (cells.length >= 2) {
                const encryptValue = await cells[1].textContent();
                if (encryptValue) {
                    this.wifiInfo.encryption = encryptValue.trim();
                    this.log(`✓ Encryption: ${encryptValue.trim()}`);
                }
            }
        } catch (error) {
            // Not critical
        }

        if (this.debug) {
            await page.screenshot({ path: 'zte_debug_04_final.png' });
            this.log('[DEBUG] Screenshot: zte_debug_04_final.png');
        }

        // Validate we got essential info
        if (!this.wifiInfo.ssid || !this.wifiInfo.password) {
            throw new Error('Gagal ekstrak SSID atau Password!');
        }
    }

    async extract() {
        const browser = await chromium.launch({
            headless: !this.debug
        });

        const context = await browser.newContext();
        const page = await context.newPage();

        try {
            await this.login(page);
            await this.extractWiFiInfo(page);

            // Add metadata
            this.wifiInfo.extracted_at = new Date().toISOString();
            this.wifiInfo.ont_url = this.ontUrl;
            this.wifiInfo.ont_model = 'ZXHN F450';
            this.wifiInfo.credentials = {
                username: this.username,
                password_used: this.password
            };

            return this.wifiInfo;

        } catch (error) {
            this.wifiInfo.error = error.message;
            throw error;
        } finally {
            await browser.close();
        }
    }

    async saveToJSON(outputFile = 'zte_wifi_info.json') {
        await fs.writeFile(outputFile, JSON.stringify(this.wifiInfo, null, 2));
        this.log(`Informasi disimpan ke ${outputFile}`);
    }

    printResults() {
        console.log('\n' + '='.repeat(60));
        console.log('         ZTE ZXHN F450 - WiFi Information');
        console.log('='.repeat(60));
        console.log(`ONT URL        : ${this.wifiInfo.ont_url || 'N/A'}`);
        console.log(`Model          : ${this.wifiInfo.ont_model || 'N/A'}`);
        console.log(`Login Username : ${this.wifiInfo.credentials?.username || 'N/A'}`);
        console.log('─'.repeat(60));
        console.log(`SSID           : ${this.wifiInfo.ssid || 'N/A'}`);
        console.log(`Password       : ${this.wifiInfo.password || 'N/A'}`);
        console.log(`Status         : ${this.wifiInfo.status || 'N/A'}`);
        console.log(`Encryption     : ${this.wifiInfo.encryption || 'N/A'}`);
        console.log('─'.repeat(60));
        console.log(`Extracted At   : ${this.wifiInfo.extracted_at || 'N/A'}`);
        console.log('='.repeat(60) + '\n');
    }
}

// ========== Main Function ==========

async function main() {
    const args = process.argv.slice(2);

    if (args.length < 1) {
        console.log('ZTE ZXHN F450 WiFi Extractor');
        console.log('Usage: node zte-f450-extractor.js <ONT_URL> [username] [password] [--debug]');
        console.log('\nContoh:');
        console.log('  node zte-f450-extractor.js http://tunnel3.ebilling.id:15634/');
        console.log('  node zte-f450-extractor.js http://tunnel3.ebilling.id:15634/ admin admin');
        console.log('  node zte-f450-extractor.js http://tunnel3.ebilling.id:15634/ admin suportadmin --debug');
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
    const extractor = new ZTEF450Extractor(ontUrl, username, password, debug);

    if (debug) {
        console.log('\n' + '='.repeat(60));
        console.log('DEBUG MODE ENABLED');
        console.log('- Browser akan ditampilkan (non-headless)');
        console.log('- Screenshot akan disimpan otomatis (zte_debug_*.png)');
        console.log('='.repeat(60) + '\n');
    }

    console.log('\n🔍 ZTE ZXHN F450 WiFi Extractor');
    console.log(`📡 Target: ${ontUrl}`);
    console.log(`👤 Credentials: ${username}/${password}\n`);

    try {
        // Extract
        await extractor.extract();

        // Print results
        extractor.printResults();

        // Save to JSON
        await extractor.saveToJSON();

        console.log('✅ [SUCCESS] Ekstraksi berhasil!\n');
        process.exit(0);

    } catch (error) {
        console.error(`\n❌ [ERROR] ${error.message}\n`);

        if (debug) {
            console.error('Stack trace:');
            console.error(error.stack);
            console.log('\n💡 Periksa screenshot zte_debug_*.png untuk informasi lebih lanjut\n');
        }

        process.exit(1);
    }
}

// Run
if (require.main === module) {
    main().catch(error => {
        console.error('❌ [FATAL ERROR]', error);
        process.exit(1);
    });
}

module.exports = ZTEF450Extractor;
