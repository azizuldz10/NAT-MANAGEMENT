/**
 * ZTE ZXHN F477V2 WiFi Information Extractor
 * Khusus untuk model ZTE ZXHN F477V2
 * Interface berbeda dari F450 - menggunakan icon menu dan iframe navigation
 */

const { chromium } = require('playwright');
const fs = require('fs').promises;

class ZTEF477V2Extractor {
    constructor(ontUrl, username = 'admin', password = 'suportadmin', debug = false) {
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

        await page.waitForTimeout(2000);

        if (this.debug) {
            await page.screenshot({ path: 'f477v2_01_welcome.png' });
            this.log('[DEBUG] Screenshot: f477v2_01_welcome.png');
        }

        this.log('Klik Administrator icon...');

        // Click Administrator icon using JavaScript
        await page.evaluate(() => {
            const adminIcon = document.getElementById('wlimgadm');
            if (adminIcon) {
                adminIcon.click();
            } else if (typeof GetAccessAccount === 'function') {
                GetAccessAccount('adm');
            }
        });

        await page.waitForTimeout(1000);

        if (this.debug) {
            await page.screenshot({ path: 'f477v2_02_password_prompt.png' });
            this.log('[DEBUG] Screenshot: f477v2_02_password_prompt.png');
        }

        this.log('Memasukkan password...');

        // Fill password
        await page.locator('#Frm_Password').fill(this.password);
        await page.waitForTimeout(500);

        // Click GO button
        await page.locator('#LoginId').click();
        await page.waitForTimeout(4000);

        if (this.debug) {
            await page.screenshot({ path: 'f477v2_03_main_menu.png' });
            this.log('[DEBUG] Screenshot: f477v2_03_main_menu.png');
        }

        this.log('Login berhasil!');
    }

    async navigateToWLAN(page) {
        this.log('Navigasi ke BasConfig menu...');

        // Click BasConfig menu (menu1)
        await page.evaluate(() => {
            const basConfigLink = document.querySelector('#menu1 a');
            if (basConfigLink) {
                basConfigLink.click();
            } else if (typeof enterFrame === 'function') {
                enterFrame('basic');
            }
        });

        await page.waitForTimeout(2000);

        if (this.debug) {
            await page.screenshot({ path: 'f477v2_04_basconfig.png' });
            this.log('[DEBUG] Screenshot: f477v2_04_basconfig.png');
        }

        this.log('Navigasi ke WLAN menu...');

        // Click WLAN tab in iframe
        await page.evaluate(() => {
            const frames = document.querySelectorAll('iframe');
            for (let frame of frames) {
                try {
                    const frameDoc = frame.contentDocument;
                    const wlanLink = Array.from(frameDoc.querySelectorAll('a'))
                        .find(a => a.textContent.trim() === 'WLAN');
                    if (wlanLink) {
                        wlanLink.click();
                        return true;
                    }
                } catch (e) {}
            }
            return false;
        });

        await page.waitForTimeout(2000);

        if (this.debug) {
            await page.screenshot({ path: 'f477v2_05_wlan_menu.png' });
            this.log('[DEBUG] Screenshot: f477v2_05_wlan_menu.png');
        }
    }

    async extractSSID(page) {
        this.log('Membuka SSID settings...');

        // Click SSID settings
        const clicked = await page.evaluate(() => {
            const frames = document.querySelectorAll('iframe');
            for (let frame of frames) {
                try {
                    const frameDoc = frame.contentDocument;
                    const ssidCell = Array.from(frameDoc.querySelectorAll('td'))
                        .find(td => td.textContent.trim() === 'SSID settings');
                    if (ssidCell) {
                        ssidCell.click();
                        return true;
                    }
                } catch (e) {}
            }
            return false;
        });

        if (!clicked) {
            throw new Error('SSID settings button tidak ditemukan');
        }

        await page.waitForTimeout(2000);

        if (this.debug) {
            await page.screenshot({ path: 'f477v2_06_ssid_settings.png' });
            this.log('[DEBUG] Screenshot: f477v2_06_ssid_settings.png');
        }

        // Extract SSID
        const ssid = await page.evaluate(() => {
            const frames = document.querySelectorAll('iframe');
            for (let frame of frames) {
                try {
                    const frameDoc = frame.contentDocument;
                    const essidInput = frameDoc.querySelector('#ESSID');
                    if (essidInput) {
                        return essidInput.value;
                    }
                } catch (e) {}
            }
            return null;
        });

        if (ssid) {
            this.wifiInfo.ssid = ssid;
            this.log(`✓ SSID: ${ssid}`);
        } else {
            this.log('✗ SSID tidak ditemukan');
        }
    }

    async extractPassword(page) {
        this.log('Membuka Security Settings...');

        // Wait a bit first
        await page.waitForTimeout(1000);

        // Try multiple methods to click Security Settings
        let clicked = false;

        // Method 1: Click in left menu if visible
        clicked = await page.evaluate(() => {
            const frames = document.querySelectorAll('iframe');
            for (let frame of frames) {
                try {
                    const frameDoc = frame.contentDocument;

                    // Try finding in sidebar menu first
                    const menuLinks = frameDoc.querySelectorAll('.subMenus a, .menu a, a');
                    for (let link of menuLinks) {
                        if (link.textContent.trim() === 'Security Settings') {
                            link.click();
                            return true;
                        }
                    }
                } catch (e) {}
            }
            return false;
        });

        if (!clicked) {
            // Method 2: Click table cell
            clicked = await page.evaluate(() => {
                const frames = document.querySelectorAll('iframe');
                for (let frame of frames) {
                    try {
                        const frameDoc = frame.contentDocument;
                        const securityCell = Array.from(frameDoc.querySelectorAll('td, div, span'))
                            .find(el => el.textContent.trim() === 'Security Settings');
                        if (securityCell) {
                            securityCell.click();
                            return true;
                        }
                    } catch (e) {}
                }
                return false;
            });
        }

        if (!clicked) {
            this.log('Warning: Security Settings click may have failed, trying to extract anyway...');
        }

        await page.waitForTimeout(3000);

        if (this.debug) {
            await page.screenshot({ path: 'f477v2_07_security_settings.png' });
            this.log('[DEBUG] Screenshot: f477v2_07_security_settings.png');
        }

        // Extract all security credentials - try multiple selectors
        const credentials = await page.evaluate(() => {
            const frames = document.querySelectorAll('iframe');
            for (let frame of frames) {
                try {
                    const doc = frame.contentDocument;

                    // Try direct selectors first
                    let password = doc.querySelector('#KeyPassphrase')?.value;
                    let ssid = doc.querySelector('#ESSID')?.value;

                    // If not found, try alternative selectors
                    if (!password) {
                        password = doc.querySelector('#Frm_KeyPassphrase')?.value;
                    }
                    if (!password) {
                        password = doc.querySelector('input[name*="assphrase"]')?.value;
                    }
                    if (!password) {
                        password = doc.querySelector('input[type="password"]')?.value;
                    }

                    if (password || ssid) {
                        return {
                            password: password,
                            ssid: ssid,
                            security: doc.querySelector('#BeaconType')?.value,
                            encryption: doc.querySelector('#WPAEncryptType')?.value,
                            authentication: doc.querySelector('#Frm_Authentication')?.value
                        };
                    }
                } catch (e) {}
            }
            return null;
        });

        if (credentials) {
            if (credentials.password) {
                this.wifiInfo.password = credentials.password;
                this.log(`✓ Password: ${credentials.password}`);
            } else {
                this.log('✗ Password tidak ditemukan di page ini');
            }
            if (credentials.ssid && !this.wifiInfo.ssid) {
                this.wifiInfo.ssid = credentials.ssid;
            }
            if (credentials.security) {
                this.wifiInfo.security = credentials.security;
                this.log(`✓ Security: ${credentials.security}`);
            }
            if (credentials.encryption) {
                this.wifiInfo.encryption = credentials.encryption;
                this.log(`✓ Encryption: ${credentials.encryption}`);
            }
            if (credentials.authentication) {
                this.wifiInfo.authentication = credentials.authentication;
            }
        } else {
            this.log('✗ Security credentials tidak ditemukan');
        }

        if (this.debug) {
            await page.screenshot({ path: 'f477v2_08_final.png', fullPage: true });
            this.log('[DEBUG] Screenshot: f477v2_08_final.png');
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
            await this.navigateToWLAN(page);
            await this.extractSSID(page);
            await this.extractPassword(page);

            // Validate we got essential info
            if (!this.wifiInfo.ssid || !this.wifiInfo.password) {
                throw new Error('Gagal ekstrak SSID atau Password!');
            }

            // Add metadata
            this.wifiInfo.extracted_at = new Date().toISOString();
            this.wifiInfo.ont_url = this.ontUrl;
            this.wifiInfo.ont_model = 'ZXHN F477V2';
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

    async saveToJSON(outputFile = 'zte_f477v2_wifi_info.json') {
        await fs.writeFile(outputFile, JSON.stringify(this.wifiInfo, null, 2));
        this.log(`Informasi disimpan ke ${outputFile}`);
    }

    printResults() {
        console.log('\n' + '='.repeat(60));
        console.log('         ZTE ZXHN F477V2 - WiFi Information');
        console.log('='.repeat(60));
        console.log(`ONT URL        : ${this.wifiInfo.ont_url || 'N/A'}`);
        console.log(`Model          : ${this.wifiInfo.ont_model || 'N/A'}`);
        console.log(`Login Username : ${this.wifiInfo.credentials?.username || 'N/A'}`);
        console.log('─'.repeat(60));
        console.log(`SSID           : ${this.wifiInfo.ssid || 'N/A'}`);
        console.log(`Password       : ${this.wifiInfo.password || 'N/A'}`);
        console.log(`Security       : ${this.wifiInfo.security || 'N/A'}`);
        console.log(`Encryption     : ${this.wifiInfo.encryption || 'N/A'}`);
        console.log(`Authentication : ${this.wifiInfo.authentication || 'N/A'}`);
        console.log('─'.repeat(60));
        console.log(`Extracted At   : ${this.wifiInfo.extracted_at || 'N/A'}`);
        console.log('='.repeat(60) + '\n');
    }
}

// ========== Main Function ==========

async function main() {
    const args = process.argv.slice(2);

    if (args.length < 1) {
        console.log('ZTE ZXHN F477V2 WiFi Extractor');
        console.log('Usage: node zte-f477v2-extractor.js <ONT_URL> [username] [password] [--debug]');
        console.log('\nContoh:');
        console.log('  node zte-f477v2-extractor.js http://tunnel3.ebilling.id:15634/');
        console.log('  node zte-f477v2-extractor.js http://tunnel3.ebilling.id:15634/ admin suportadmin');
        console.log('  node zte-f477v2-extractor.js http://tunnel3.ebilling.id:15634/ admin suportadmin --debug');
        process.exit(1);
    }

    // Parse arguments
    const ontUrl = args[0];
    let username = 'admin';
    let password = 'suportadmin';
    let debug = args.includes('--debug');

    // Remove --debug from args
    const filteredArgs = args.filter(arg => arg !== '--debug');

    if (filteredArgs.length > 1) username = filteredArgs[1];
    if (filteredArgs.length > 2) password = filteredArgs[2];

    // Create extractor
    const extractor = new ZTEF477V2Extractor(ontUrl, username, password, debug);

    if (debug) {
        console.log('\n' + '='.repeat(60));
        console.log('DEBUG MODE ENABLED');
        console.log('- Browser akan ditampilkan (non-headless)');
        console.log('- Screenshot akan disimpan otomatis (f477v2_*.png)');
        console.log('='.repeat(60) + '\n');
    }

    console.log('\n🔍 ZTE ZXHN F477V2 WiFi Extractor');
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
            console.log('\n💡 Periksa screenshot f477v2_*.png untuk informasi lebih lanjut\n');
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

module.exports = ZTEF477V2Extractor;
