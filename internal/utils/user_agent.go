package utils

import (
	"strings"

	"nat-management-app/internal/models"
)

// ParseUserAgent extracts device information from user agent string
func ParseUserAgent(userAgent string) *models.DeviceInfo {
	if userAgent == "" {
		return nil
	}

	deviceInfo := &models.DeviceInfo{
		IsMobile: false,
	}

	ua := strings.ToLower(userAgent)

	// Detect browser
	deviceInfo.Browser = detectBrowser(ua)

	// Detect OS
	deviceInfo.OS = detectOS(ua)

	// Detect device type
	deviceInfo.DeviceType, deviceInfo.IsMobile = detectDeviceType(ua)

	return deviceInfo
}

// detectBrowser detects browser from user agent
func detectBrowser(ua string) string {
	browsers := []struct {
		name    string
		pattern string
	}{
		{"Edge", "edg/"},
		{"Chrome", "chrome/"},
		{"Firefox", "firefox/"},
		{"Safari", "safari/"},
		{"Opera", "opr/"},
		{"Internet Explorer", "msie"},
		{"Internet Explorer 11", "trident/"},
	}

	for _, browser := range browsers {
		if strings.Contains(ua, browser.pattern) {
			// Extract version if possible
			if idx := strings.Index(ua, browser.pattern); idx != -1 {
				versionStart := idx + len(browser.pattern)
				versionEnd := versionStart
				for versionEnd < len(ua) && (ua[versionEnd] >= '0' && ua[versionEnd] <= '9' || ua[versionEnd] == '.') {
					versionEnd++
				}
				if versionEnd > versionStart {
					version := ua[versionStart:versionEnd]
					// Get major version only
					if dotIdx := strings.Index(version, "."); dotIdx != -1 {
						version = version[:dotIdx]
					}
					return browser.name + " " + version
				}
			}
			return browser.name
		}
	}

	return "Unknown"
}

// detectOS detects operating system from user agent
func detectOS(ua string) string {
	oses := []struct {
		name    string
		pattern string
	}{
		{"Windows 11", "windows nt 10.0"},
		{"Windows 10", "windows nt 10.0"},
		{"Windows 8.1", "windows nt 6.3"},
		{"Windows 8", "windows nt 6.2"},
		{"Windows 7", "windows nt 6.1"},
		{"Mac OS X", "mac os x"},
		{"macOS", "macintosh"},
		{"iOS", "iphone"},
		{"iOS", "ipad"},
		{"Android", "android"},
		{"Linux", "linux"},
		{"Ubuntu", "ubuntu"},
		{"Chrome OS", "cros"},
	}

	for _, os := range oses {
		if strings.Contains(ua, os.pattern) {
			// Try to extract version for some OSes
			if os.name == "Android" {
				if idx := strings.Index(ua, "android "); idx != -1 {
					versionStart := idx + 8
					versionEnd := versionStart
					for versionEnd < len(ua) && (ua[versionEnd] >= '0' && ua[versionEnd] <= '9' || ua[versionEnd] == '.') {
						versionEnd++
					}
					if versionEnd > versionStart {
						version := ua[versionStart:versionEnd]
						return "Android " + version
					}
				}
			}
			return os.name
		}
	}

	return "Unknown"
}

// detectDeviceType detects device type from user agent
func detectDeviceType(ua string) (string, bool) {
	// Mobile indicators
	mobilePatterns := []string{
		"mobile", "android", "iphone", "ipod", "blackberry",
		"windows phone", "opera mini", "iemobile",
	}

	for _, pattern := range mobilePatterns {
		if strings.Contains(ua, pattern) {
			// Check if it's a tablet
			if strings.Contains(ua, "ipad") || strings.Contains(ua, "tablet") {
				return "tablet", true
			}
			return "mobile", true
		}
	}

	// Tablet specific
	if strings.Contains(ua, "ipad") || strings.Contains(ua, "tablet") {
		return "tablet", true
	}

	return "desktop", false
}
