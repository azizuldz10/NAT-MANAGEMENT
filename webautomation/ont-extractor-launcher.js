/**
 * ONT WiFi Extractor Launcher (Universal)
 * Auto-detect model ONT dan route ke extractor yang sesuai
 *
 * Supported Models:
 * - Fiberhome GM220-S
 * - AccesGo / OLD_MODEL (dengan menu NETWORK)
 * - ZTE ZXHN F450 (dengan iframe dashboard)
 * - ZTE ZXHN F477V2 (dengan icon menu)
 */

const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');
const ONTWiFiExtractor = require('./ont-wifi-extractor.js');
const ZTEF450Extractor = require('./zte-f450-extractor.js');
const ZTEF477V2Extractor = require('./zte-f477v2-extractor.js');

class ONTLauncher {
    constructor(ontUrl, username = null, password = null, debug = false) {
        this.ontUrl = ontUrl;
        this.username = username;
        this.password = password;
        this.debug = debug;
        this.credentialTemplate = this.loadCredentialTemplate();
        this.successfulCredentials = null;
    }

    loadCredentialTemplate() {
        try {
            const templatePath = path.join(__dirname, 'ont-credentials-template.json');
            const templateData = fs.readFileSync(templatePath, 'utf8');
            return JSON.parse(templateData);
        } catch (error) {
            this.log(`⚠ Warning: Could not load credential template: ${error.message}`);
            return null;
        }
    }

    getCredentialsForModel(model) {
        if (!this.credentialTemplate) {
            return [{ username: 'admin', password: 'admin' }];
        }

        const vendorMap = this.credentialTemplate.modelMapping;
        const vendor = vendorMap[model] || 'Generic';
        const credentials = this.credentialTemplate.credentials[vendor] || this.credentialTemplate.credentials['Generic'];

        return credentials.sort((a, b) => a.priority - b.priority);
    }

    log(message) {
        const timestamp = new Date().toLocaleTimeString('id-ID');
        console.log(`[${timestamp}] ${message}`);
    }

    async detectModel() {
        this.log('🔍 Detecting ONT model...');

        const browser = await chromium.launch({
            headless: !this.debug
        });

        const context = await browser.newContext();
        const page = await context.newPage();

        let detectedModel = null;

        try {
            await page.goto(this.ontUrl, { timeout: 30000, waitUntil: 'domcontentloaded' });
            await page.waitForTimeout(2000);

            const pageInfo = await page.evaluate(() => {
                return {
                    title: document.title,
                    bodyText: document.body.textContent.substring(0, 1000),
                    hasWelcome: document.body.textContent.includes('WELCOME'),
                    hasAdministrator: document.body.textContent.includes('Administrator'),
                    hasZTE: document.body.textContent.includes('ZTE'),
                    frameCount: document.querySelectorAll('iframe').length,
                    hasMainFrame: !!document.querySelector('iframe[name="mainFrame"]'),
                    hasFstmenu: !!document.querySelector('#Fstmenu'),
                    hasWlimgadm: !!document.querySelector('#wlimgadm')
                };
            });

            if (this.debug) {
                console.log('\n[DEBUG] Page Info:', JSON.stringify(pageInfo, null, 2));
            }

            // Detection Logic

            // 1. Check for ZTE ZXHN F477V2 (WELCOME page with icon menu after login)
            if (pageInfo.hasWelcome && pageInfo.hasWlimgadm) {
                this.log('✓ Detected: ZTE ZXHN F477V2 (WELCOME interface)');
                detectedModel = 'F477V2';
            }
            // 2. Check for ZTE ZXHN F450 (standard login with dashboard iframe)
            else if (pageInfo.title && pageInfo.title.includes('ZXHN F450')) {
                this.log('✓ Detected: ZTE ZXHN F450');
                detectedModel = 'F450';
            }
            // 3. Check for GM220-S (mainFrame iframe)
            else if (pageInfo.hasMainFrame) {
                this.log('✓ Detected: Fiberhome GM220-S');
                detectedModel = 'GM220-S';
            }
            // 4. Check for OLD_MODEL (Fstmenu)
            else if (pageInfo.hasFstmenu) {
                this.log('✓ Detected: AccesGo / OLD_MODEL');
                detectedModel = 'OLD_MODEL';
            }
            // 5. Additional ZTE detection (fallback)
            else if (pageInfo.hasZTE && pageInfo.hasWelcome) {
                this.log('✓ Detected: ZTE model (likely F477V2)');
                detectedModel = 'F477V2';
            }
            // Default fallback
            else {
                this.log('⚠ Model tidak terdeteksi otomatis, menggunakan OLD_MODEL as fallback');
                detectedModel = 'OLD_MODEL';
            }

        } catch (error) {
            this.log(`❌ Error during detection: ${error.message}`);
            this.log('Using OLD_MODEL as fallback...');
            detectedModel = 'OLD_MODEL';
        } finally {
            await browser.close();
        }

        return detectedModel;
    }

    async tryCredentials(model, credentialsList) {
        this.log(`🔑 Trying credentials for ${model}...`);

        for (let i = 0; i < credentialsList.length; i++) {
            const cred = credentialsList[i];
            const username = cred.username || 'admin';
            const password = cred.password || 'admin';

            // Skip empty credentials for models that don't need login
            if (!cred.username && !cred.password && model !== 'OLD_MODEL') {
                continue;
            }

            this.log(`  [${i + 1}/${credentialsList.length}] Trying: ${username} / ${password.replace(/./g, '*')} (${cred.note})`);

            let extractor;
            try {
                // Create extractor based on model
                switch (model) {
                    case 'F477V2':
                        extractor = new ZTEF477V2Extractor(this.ontUrl, username, password, this.debug);
                        break;
                    case 'F450':
                        extractor = new ZTEF450Extractor(this.ontUrl, username, password, this.debug);
                        break;
                    case 'GM220-S':
                    case 'OLD_MODEL':
                    default:
                        extractor = new ONTWiFiExtractor(this.ontUrl, username, password, this.debug);
                        break;
                }

                // Try extraction
                let result;
                if (model === 'F477V2' || model === 'F450') {
                    result = await extractor.extract();
                } else {
                    result = await extractor.extractWiFiInfo();
                }

                // Check if extraction was successful
                if (result && result.ssid) {
                    this.log(`  ✅ SUCCESS with ${username} / ${password}`);
                    this.successfulCredentials = { username, password };
                    return { extractor, result };
                } else {
                    this.log(`  ❌ Failed - No data extracted`);
                }

            } catch (error) {
                this.log(`  ❌ Failed: ${error.message}`);
                // Continue to next credentials
            }
        }

        // If all credentials failed
        throw new Error('All credentials failed. Could not extract WiFi information.');
    }

    async extract() {
        console.log('\n' + '='.repeat(70));
        console.log('🚀 ONT WiFi Extractor Launcher (Universal + Auto-Credentials)');
        console.log('='.repeat(70));
        console.log(`📡 Target URL: ${this.ontUrl}`);
        console.log(`🐛 Debug Mode: ${this.debug ? 'ENABLED' : 'DISABLED'}`);
        console.log('='.repeat(70) + '\n');

        // Step 1: Detect model
        const model = await this.detectModel();

        if (!model) {
            throw new Error('Failed to detect ONT model');
        }

        console.log(`\n📋 Detected Model: ${model}\n`);

        // Step 2: Determine credentials to try
        let credentialsList;

        if (this.username && this.password) {
            // User provided specific credentials - use them only
            this.log('🔑 Using user-provided credentials');
            credentialsList = [{
                username: this.username,
                password: this.password,
                note: 'User provided',
                priority: 1
            }];
        } else {
            // Use template credentials for detected model
            this.log('🔑 Using template credentials for detected model');
            credentialsList = this.getCredentialsForModel(model);
            this.log(`   Found ${credentialsList.length} credential(s) to try\n`);
        }

        // Step 3: Try credentials until one works
        let extractionResult;

        try {
            extractionResult = await this.tryCredentials(model, credentialsList);
        } catch (error) {
            return {
                success: false,
                model: model,
                error: error.message
            };
        }

        // Step 4: Save results
        const { extractor, result } = extractionResult;

        try {
            if (model === 'F477V2' || model === 'F450') {
                await extractor.printResults();
            }
            await extractor.saveToJSON();
        } catch (error) {
            this.log(`⚠ Warning: Could not save results: ${error.message}`);
        }

        return {
            success: true,
            model: model,
            data: result,
            credentials: this.successfulCredentials
        };
    }
}

// ========== Main Function ==========

async function main() {
    const args = process.argv.slice(2);

    if (args.length < 1 || args.includes('--help') || args.includes('-h')) {
        console.log('\n' + '='.repeat(70));
        console.log('ONT WiFi Extractor Launcher - Universal Auto-Detection + Auto-Credentials');
        console.log('='.repeat(70));
        console.log('\n📖 Usage:');
        console.log('  node ont-extractor-launcher.js <ONT_URL> [--debug]');
        console.log('  node ont-extractor-launcher.js <ONT_URL> <username> <password> [--debug]');
        console.log('\n📝 Examples:');
        console.log('  # SIMPLE - Hanya URL (auto-detect model + auto-try credentials)');
        console.log('  node ont-extractor-launcher.js http://192.168.1.1/');
        console.log('  node ont-extractor-launcher.js http://tunnel3.ebilling.id:15634/');
        console.log('');
        console.log('  # Dengan debug mode (lihat browser + save screenshots)');
        console.log('  node ont-extractor-launcher.js http://192.168.1.1/ --debug');
        console.log('');
        console.log('  # Dengan kredensial spesifik (skip auto-try)');
        console.log('  node ont-extractor-launcher.js http://192.168.1.1/ admin mypassword');
        console.log('');
        console.log('  # NPM shortcuts');
        console.log('  npm run extract http://192.168.1.1/');
        console.log('  npm run extract:debug http://192.168.1.1/');
        console.log('\n🔧 Supported Models:');
        console.log('  ✓ Fiberhome GM220-S (auto-detected)');
        console.log('  ✓ AccesGo / OLD_MODEL (auto-detected)');
        console.log('  ✓ ZTE ZXHN F450 (auto-detected)');
        console.log('  ✓ ZTE ZXHN F477V2 (auto-detected)');
        console.log('\n🔑 Auto-Credentials:');
        console.log('  Script otomatis mencoba kredensial umum untuk setiap model:');
        console.log('  - ZTE models: admin/suportadmin, admin/admin, user/user, admin/Zte521');
        console.log('  - Fiberhome: admin/admin, telecomadmin/admintelecom, user/user');
        console.log('  - Huawei: telecomadmin/admintelecom, root/admin, admin/admin');
        console.log('  - Generic fallback: admin/admin, user/user, admin/password');
        console.log('\n💡 Tips:');
        console.log('  - Cukup input URL saja, credentials otomatis dicoba');
        console.log('  - Gunakan --debug untuk troubleshooting');
        console.log('  - Script akan auto-detect model ONT');
        console.log('  - Output disimpan ke JSON file');
        console.log('  - Credentials yang berhasil akan ditampilkan di akhir');
        console.log('='.repeat(70) + '\n');
        process.exit(args.includes('--help') || args.includes('-h') ? 0 : 1);
    }

    // Parse arguments
    const ontUrl = args[0];
    let username = null;
    let password = null;
    let debug = args.includes('--debug');

    // Remove --debug from args
    const filteredArgs = args.filter(arg => arg !== '--debug');

    // Only use provided credentials if both username AND password are given
    if (filteredArgs.length > 2) {
        username = filteredArgs[1];
        password = filteredArgs[2];
    }

    // Create launcher
    const launcher = new ONTLauncher(ontUrl, username, password, debug);

    try {
        const result = await launcher.extract();

        if (result.success) {
            console.log('\n' + '='.repeat(70));
            console.log('✅ EXTRACTION SUCCESS!');
            console.log('='.repeat(70));
            console.log(`📱 Model: ${result.model}`);
            console.log(`📶 SSID: ${result.data.ssid || 'N/A'}`);
            console.log(`🔒 Password: ${result.data.password || 'N/A'}`);
            if (result.credentials) {
                console.log(`🔑 Successful Login: ${result.credentials.username} / ${result.credentials.password}`);
            }
            console.log('='.repeat(70) + '\n');
            process.exit(0);
        } else {
            console.error('\n' + '='.repeat(70));
            console.error('❌ EXTRACTION FAILED!');
            console.error('='.repeat(70));
            console.error(`Model: ${result.model}`);
            console.error(`Error: ${result.error}`);
            console.error('='.repeat(70) + '\n');

            if (debug) {
                console.log('💡 Check debug screenshots for more information\n');
            }

            process.exit(1);
        }

    } catch (error) {
        console.error('\n❌ [FATAL ERROR]', error.message);
        if (debug) {
            console.error('\nStack trace:');
            console.error(error.stack);
        }
        process.exit(1);
    }
}

// Run
if (require.main === module) {
    main().catch(error => {
        console.error('❌ [CRITICAL ERROR]', error);
        process.exit(1);
    });
}

module.exports = ONTLauncher;
