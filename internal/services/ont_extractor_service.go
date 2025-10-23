package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"nat-management-app/internal/models"

	"github.com/sirupsen/logrus"
)

// ONTExtractorService handles ONT WiFi information extraction via webautomation tools
type ONTExtractorService struct {
	logger           *logrus.Logger
	webautomationDir string
	nodeCommand      string
	defaultTimeout   time.Duration
}

// NewONTExtractorService creates a new ONT extractor service instance
func NewONTExtractorService(logger *logrus.Logger) *ONTExtractorService {
	// Get absolute path to webautomation directory (relative to executable)
	execPath, err := os.Executable()
	if err != nil {
		logger.Warnf("Could not determine executable path: %v", err)
	}

	execDir := filepath.Dir(execPath)
	webautomationDir := filepath.Join(execDir, "webautomation")

	// Fallback: try relative path from current working directory
	if _, err := os.Stat(webautomationDir); os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		webautomationDir = filepath.Join(cwd, "webautomation")
		logger.Infof("Using webautomation directory: %s", webautomationDir)
	}

	// Determine node command (node or nodejs)
	nodeCmd := "node"
	if _, err := exec.LookPath("node"); err != nil {
		// Try nodejs (some Linux distros use this)
		if _, err := exec.LookPath("nodejs"); err == nil {
			nodeCmd = "nodejs"
		} else {
			logger.Warnf("Node.js not found in PATH. ONT extraction will not work!")
		}
	}

	return &ONTExtractorService{
		logger:           logger,
		webautomationDir: webautomationDir,
		nodeCommand:      nodeCmd,
		defaultTimeout:   90 * time.Second, // 90s timeout for extraction
	}
}

// ExtractWiFiInfo extracts WiFi information from an ONT device
func (oes *ONTExtractorService) ExtractWiFiInfo(ontURL, username, password string, debug bool) (*models.ONTWiFiInfo, error) {
	oes.logger.Infof("🔍 Starting WiFi extraction for ONT: %s", ontURL)

	// Validate inputs
	if strings.TrimSpace(ontURL) == "" {
		return nil, fmt.Errorf("ONT URL cannot be empty")
	}
	if strings.TrimSpace(username) == "" {
		username = "admin" // Default username
	}
	if strings.TrimSpace(password) == "" {
		password = "admin" // Default password
	}

	// Build command
	launcherScript := filepath.Join(oes.webautomationDir, "ont-extractor-launcher.js")

	// Check if launcher script exists
	if _, err := os.Stat(launcherScript); os.IsNotExist(err) {
		return nil, fmt.Errorf("ONT extractor launcher not found: %s", launcherScript)
	}

	// Check if webautomation directory exists
	if _, err := os.Stat(oes.webautomationDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("webautomation directory not found: %s", oes.webautomationDir)
	}

	// IMPORTANT: Delete old JSON files before extraction to prevent stale data
	jsonFilesToClean := []string{
		filepath.Join(oes.webautomationDir, "zte_f477v2_wifi_info.json"),
		filepath.Join(oes.webautomationDir, "zte_wifi_info.json"),
		filepath.Join(oes.webautomationDir, "wifi_info.json"),
	}
	for _, jsonFile := range jsonFilesToClean {
		if _, err := os.Stat(jsonFile); err == nil {
			oes.logger.Infof("🗑️  Removing old JSON file: %s", jsonFile)
			if err := os.Remove(jsonFile); err != nil {
				oes.logger.Warnf("Failed to remove old JSON file %s: %v", jsonFile, err)
			}
		}
	}

	// Prepare arguments
	args := []string{launcherScript, ontURL, username, password}
	if debug {
		args = append(args, "--debug")
	}

	// Execute command with timeout
	oes.logger.Infof("Executing: %s %s", oes.nodeCommand, strings.Join(args, " "))
	oes.logger.Infof("Working directory: %s", oes.webautomationDir)

	cmd := exec.Command(oes.nodeCommand, args...)
	cmd.Dir = oes.webautomationDir // Set working directory to webautomation folder

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		oes.logger.Errorf("❌ WiFi extraction failed: %v", err)
		oes.logger.Errorf("Command output: %s", outputStr)
		return nil, fmt.Errorf("extraction failed: %v (output: %s)", err, outputStr)
	}

	oes.logger.Infof("Extraction output: %s", outputStr)

	// Parse output - look for JSON files created by extractors
	wifiInfo, parseErr := oes.parseExtractionOutput(outputStr, ontURL)
	if parseErr != nil {
		oes.logger.Errorf("Failed to parse extraction output: %v", parseErr)
		return nil, fmt.Errorf("failed to parse extraction results: %v", parseErr)
	}

	oes.logger.Infof("✅ Successfully extracted WiFi info for ONT: %s", ontURL)
	return wifiInfo, nil
}

// parseExtractionOutput parses the output from webautomation tool
func (oes *ONTExtractorService) parseExtractionOutput(output, ontURL string) (*models.ONTWiFiInfo, error) {
	// Try to read JSON output files created by extractors
	// Priority: zte_f477v2_wifi_info.json > zte_wifi_info.json > wifi_info.json

	jsonFiles := []string{
		filepath.Join(oes.webautomationDir, "zte_f477v2_wifi_info.json"),
		filepath.Join(oes.webautomationDir, "zte_wifi_info.json"),
		filepath.Join(oes.webautomationDir, "wifi_info.json"),
	}

	var wifiData map[string]interface{}
	var usedFile string

	for _, jsonFile := range jsonFiles {
		data, err := os.ReadFile(jsonFile)
		if err != nil {
			continue // File doesn't exist or can't be read, try next
		}

		if err := json.Unmarshal(data, &wifiData); err != nil {
			oes.logger.Warnf("Failed to parse %s: %v", jsonFile, err)
			continue
		}

		usedFile = jsonFile
		break
	}

	if wifiData == nil {
		// Fallback: try to parse from console output
		return oes.parseConsoleOutput(output, ontURL)
	}

	oes.logger.Debugf("Parsed WiFi info from: %s", usedFile)

	// Extract fields from JSON
	wifiInfo := &models.ONTWiFiInfo{
		SSID:           getString(wifiData, "ssid"),
		Password:       getString(wifiData, "password"),
		Security:       getString(wifiData, "security"),
		Encryption:     getString(wifiData, "encryption"),
		Authentication: getString(wifiData, "authentication"),
		ONTURL:         ontURL,
		ONTModel:       getString(wifiData, "ont_model"),
		ExtractedAt:    time.Now(),
	}

	// Validate extracted data
	if wifiInfo.SSID == "" || wifiInfo.Password == "" {
		return nil, fmt.Errorf("incomplete WiFi info: SSID or password is empty")
	}

	return wifiInfo, nil
}

// parseConsoleOutput attempts to parse WiFi info from console output (fallback)
func (oes *ONTExtractorService) parseConsoleOutput(output, ontURL string) (*models.ONTWiFiInfo, error) {
	// Look for patterns in console output
	lines := strings.Split(output, "\n")

	wifiInfo := &models.ONTWiFiInfo{
		ONTURL:      ontURL,
		ExtractedAt: time.Now(),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "SSID") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				wifiInfo.SSID = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Password") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				wifiInfo.Password = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Model") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				wifiInfo.ONTModel = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Security") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				wifiInfo.Security = strings.TrimSpace(parts[1])
			}
		}
	}

	if wifiInfo.SSID == "" || wifiInfo.Password == "" {
		return nil, fmt.Errorf("could not extract WiFi credentials from output")
	}

	return wifiInfo, nil
}

// ExtractWiFiInfoFromNATConfig extracts WiFi info using NAT configuration
func (oes *ONTExtractorService) ExtractWiFiInfoFromNATConfig(natConfig models.ONTConfig, username, password string) (*models.ONTWiFiInfo, error) {
	if natConfig.PublicONTURL == "" {
		return nil, fmt.Errorf("no public ONT URL available in NAT config")
	}

	return oes.ExtractWiFiInfo(natConfig.PublicONTURL, username, password, false)
}

// CheckWebautomationAvailability checks if webautomation tools are available
func (oes *ONTExtractorService) CheckWebautomationAvailability() error {
	// Check Node.js
	if _, err := exec.LookPath(oes.nodeCommand); err != nil {
		return fmt.Errorf("Node.js not found. Please install Node.js to use WiFi extraction features")
	}

	// Check webautomation directory
	if _, err := os.Stat(oes.webautomationDir); os.IsNotExist(err) {
		return fmt.Errorf("webautomation directory not found: %s", oes.webautomationDir)
	}

	// Check launcher script
	launcherScript := filepath.Join(oes.webautomationDir, "ont-extractor-launcher.js")
	if _, err := os.Stat(launcherScript); os.IsNotExist(err) {
		return fmt.Errorf("ONT extractor launcher not found: %s", launcherScript)
	}

	// Check package.json
	packageJSON := filepath.Join(oes.webautomationDir, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		return fmt.Errorf("package.json not found. Please run 'npm install' in webautomation directory")
	}

	// Check node_modules
	nodeModules := filepath.Join(oes.webautomationDir, "node_modules")
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		return fmt.Errorf("node_modules not found. Please run 'npm install' in webautomation directory")
	}

	oes.logger.Info("✅ Webautomation tools are available and ready")
	return nil
}

// GetSupportedModels returns list of supported ONT models
func (oes *ONTExtractorService) GetSupportedModels() []string {
	return []string{
		"Fiberhome GM220-S",
		"AccesGo / OLD_MODEL",
		"ZTE ZXHN F450",
		"ZTE ZXHN F477V2",
	}
}

// Helper function to safely get string from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
