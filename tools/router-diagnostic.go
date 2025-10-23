package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/go-routeros/routeros"
	"github.com/sirupsen/logrus"
)

// RouterDiagnostic performs comprehensive router connectivity diagnostics
type RouterDiagnostic struct {
	Host     string
	Port     int
	Username string
	Password string
	Logger   *logrus.Logger
}

// DiagnosticResult stores the results of diagnostic tests
type DiagnosticResult struct {
	TestName    string
	Status      string
	Message     string
	Duration    time.Duration
	Error       error
	Suggestions []string
}

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	// Parse command line arguments
	if len(os.Args) < 5 {
		printUsage()
		os.Exit(1)
	}

	host := os.Args[1]
	port := parsePort(os.Args[2])
	username := os.Args[3]
	password := os.Args[4]

	if port == 0 {
		logger.Errorf("❌ Invalid port: %s", os.Args[2])
		os.Exit(1)
	}

	diagnostic := &RouterDiagnostic{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Logger:   logger,
	}

	logger.Info("🔍 ========================================")
	logger.Info("🔍 Router Connection Diagnostic Tool")
	logger.Info("🔍 ========================================")
	logger.Infof("🎯 Target: %s:%d", host, port)
	logger.Infof("👤 Username: %s", username)
	logger.Info("🔍 ========================================\n")

	// Run all diagnostic tests
	results := diagnostic.RunAllTests()

	// Print summary
	diagnostic.PrintSummary(results)
}

// RunAllTests executes all diagnostic tests
func (rd *RouterDiagnostic) RunAllTests() []DiagnosticResult {
	var results []DiagnosticResult

	rd.Logger.Info("📋 Running diagnostic tests...\n")

	// Test 1: DNS Resolution
	results = append(results, rd.TestDNSResolution())

	// Test 2: ICMP Ping (optional, may fail due to firewall)
	results = append(results, rd.TestPing())

	// Test 3: TCP Connection (short timeout)
	results = append(results, rd.TestTCPConnection(5*time.Second))

	// Test 4: TCP Connection (medium timeout)
	results = append(results, rd.TestTCPConnection(15*time.Second))

	// Test 5: TCP Connection (long timeout)
	results = append(results, rd.TestTCPConnection(30*time.Second))

	// Test 6: RouterOS API Connection
	results = append(results, rd.TestRouterOSAPI(5*time.Second))

	// Test 7: RouterOS API Connection (longer timeout)
	results = append(results, rd.TestRouterOSAPI(15*time.Second))

	// Test 8: Get Router Identity
	results = append(results, rd.TestRouterIdentity())

	// Test 9: Get System Resources
	results = append(results, rd.TestSystemResources())

	return results
}

// TestDNSResolution tests if the host can be resolved
func (rd *RouterDiagnostic) TestDNSResolution() DiagnosticResult {
	result := DiagnosticResult{
		TestName: "DNS Resolution",
		Status:   "RUNNING",
	}

	start := time.Now()
	rd.Logger.Infof("🔍 Test 1: DNS Resolution for %s", rd.Host)

	// Check if it's already an IP address
	if net.ParseIP(rd.Host) != nil {
		result.Duration = time.Since(start)
		result.Status = "PASS"
		result.Message = fmt.Sprintf("Host is already an IP address: %s", rd.Host)
		rd.Logger.Infof("   ✅ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
		return result
	}

	// Try to resolve hostname
	ips, err := net.LookupIP(rd.Host)
	result.Duration = time.Since(start)

	if err != nil {
		result.Status = "FAIL"
		result.Error = err
		result.Message = fmt.Sprintf("DNS resolution failed: %v", err)
		result.Suggestions = []string{
			"Check if hostname is correct",
			"Try using IP address instead of hostname",
			"Check DNS server configuration",
			"Verify network connectivity",
		}
		rd.Logger.Errorf("   ❌ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
		return result
	}

	result.Status = "PASS"
	result.Message = fmt.Sprintf("Resolved to %d IP(s): %v", len(ips), ips)
	rd.Logger.Infof("   ✅ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
	return result
}

// TestPing tests ICMP ping to the host
func (rd *RouterDiagnostic) TestPing() DiagnosticResult {
	result := DiagnosticResult{
		TestName: "ICMP Ping Test",
		Status:   "SKIP",
		Message:  "ICMP ping not implemented (may be blocked by firewall anyway)",
	}
	rd.Logger.Infof("🔍 Test 2: %s", result.Message)
	rd.Logger.Info("   ⏭️  SKIPPED\n")
	return result
}

// TestTCPConnection tests TCP connection to the router
func (rd *RouterDiagnostic) TestTCPConnection(timeout time.Duration) DiagnosticResult {
	result := DiagnosticResult{
		TestName: fmt.Sprintf("TCP Connection (timeout: %v)", timeout),
		Status:   "RUNNING",
	}

	start := time.Now()
	rd.Logger.Infof("🔍 Test: TCP Connection to %s:%d (timeout: %v)", rd.Host, rd.Port, timeout)

	address := fmt.Sprintf("%s:%d", rd.Host, rd.Port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	result.Duration = time.Since(start)

	if err != nil {
		result.Status = "FAIL"
		result.Error = err
		result.Message = fmt.Sprintf("TCP connection failed: %v", err)

		// Provide specific suggestions based on error
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			result.Suggestions = []string{
				"Connection timed out - router may be offline or unreachable",
				fmt.Sprintf("Check if router is online at %s", rd.Host),
				fmt.Sprintf("Verify port %d is correct (standard RouterOS API port is 8728)", rd.Port),
				"Check firewall rules blocking the connection",
				"Try increasing timeout duration",
			}
		} else {
			result.Suggestions = []string{
				"Check if router is online and reachable",
				"Verify IP address and port are correct",
				"Check network routing and firewall rules",
				"Try pinging the router first",
			}
		}

		rd.Logger.Errorf("   ❌ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
		return result
	}
	defer conn.Close()

	result.Status = "PASS"
	result.Message = fmt.Sprintf("TCP connection successful to %s", address)
	rd.Logger.Infof("   ✅ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
	return result
}

// TestRouterOSAPI tests RouterOS API connection
func (rd *RouterDiagnostic) TestRouterOSAPI(timeout time.Duration) DiagnosticResult {
	result := DiagnosticResult{
		TestName: fmt.Sprintf("RouterOS API Connection (timeout: %v)", timeout),
		Status:   "RUNNING",
	}

	start := time.Now()
	rd.Logger.Infof("🔍 Test: RouterOS API Connection (timeout: %v)", timeout)

	// Create custom dialer with timeout
	address := fmt.Sprintf("%s:%d", rd.Host, rd.Port)

	// Note: go-routeros library doesn't support custom timeout directly
	// We need to use DialTimeout from net package first to check connectivity
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		result.Duration = time.Since(start)
		result.Status = "FAIL"
		result.Error = err
		result.Message = fmt.Sprintf("Failed to establish TCP connection: %v", err)
		result.Suggestions = []string{
			"TCP connection failed before API authentication",
			"See TCP connection test results above for more details",
		}
		rd.Logger.Errorf("   ❌ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
		return result
	}
	conn.Close()

	// Now try RouterOS API connection
	client, err := routeros.Dial(address, rd.Username, rd.Password)
	result.Duration = time.Since(start)

	if err != nil {
		result.Status = "FAIL"
		result.Error = err
		result.Message = fmt.Sprintf("RouterOS API authentication failed: %v", err)
		result.Suggestions = []string{
			"Check username and password are correct",
			"Verify API service is enabled on MikroTik (IP -> Services -> API)",
			fmt.Sprintf("Confirm port %d is the correct API port", rd.Port),
			"Check if user has API access permissions",
			"Try connecting with Winbox to verify credentials",
		}
		rd.Logger.Errorf("   ❌ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
		return result
	}
	defer client.Close()

	result.Status = "PASS"
	result.Message = "RouterOS API authentication successful"
	rd.Logger.Infof("   ✅ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
	return result
}

// TestRouterIdentity gets the router identity
func (rd *RouterDiagnostic) TestRouterIdentity() DiagnosticResult {
	result := DiagnosticResult{
		TestName: "Get Router Identity",
		Status:   "RUNNING",
	}

	start := time.Now()
	rd.Logger.Info("🔍 Test: Get Router Identity")

	client, err := routeros.Dial(fmt.Sprintf("%s:%d", rd.Host, rd.Port), rd.Username, rd.Password)
	if err != nil {
		result.Duration = time.Since(start)
		result.Status = "FAIL"
		result.Error = err
		result.Message = "Failed to connect to router"
		rd.Logger.Errorf("   ❌ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
		return result
	}
	defer client.Close()

	reply, err := client.Run("/system/identity/print")
	result.Duration = time.Since(start)

	if err != nil {
		result.Status = "FAIL"
		result.Error = err
		result.Message = fmt.Sprintf("Failed to get identity: %v", err)
		result.Suggestions = []string{
			"RouterOS API connected but command execution failed",
			"Check user permissions for system commands",
		}
		rd.Logger.Errorf("   ❌ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
		return result
	}

	if len(reply.Re) > 0 {
		identity := reply.Re[0].Map["name"]
		result.Status = "PASS"
		result.Message = fmt.Sprintf("Router Identity: %s", identity)
		rd.Logger.Infof("   ✅ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
	} else {
		result.Status = "WARN"
		result.Message = "No identity information returned"
		rd.Logger.Warnf("   ⚠️  %s (%.2fs)\n", result.Message, result.Duration.Seconds())
	}

	return result
}

// TestSystemResources gets system resource information
func (rd *RouterDiagnostic) TestSystemResources() DiagnosticResult {
	result := DiagnosticResult{
		TestName: "Get System Resources",
		Status:   "RUNNING",
	}

	start := time.Now()
	rd.Logger.Info("🔍 Test: Get System Resources")

	client, err := routeros.Dial(fmt.Sprintf("%s:%d", rd.Host, rd.Port), rd.Username, rd.Password)
	if err != nil {
		result.Duration = time.Since(start)
		result.Status = "FAIL"
		result.Error = err
		result.Message = "Failed to connect to router"
		rd.Logger.Errorf("   ❌ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
		return result
	}
	defer client.Close()

	reply, err := client.Run("/system/resource/print")
	result.Duration = time.Since(start)

	if err != nil {
		result.Status = "FAIL"
		result.Error = err
		result.Message = fmt.Sprintf("Failed to get system resources: %v", err)
		rd.Logger.Errorf("   ❌ %s (%.2fs)\n", result.Message, result.Duration.Seconds())
		return result
	}

	if len(reply.Re) > 0 {
		res := reply.Re[0].Map
		version := res["version"]
		board := res["board-name"]
		platform := res["platform"]

		result.Status = "PASS"
		result.Message = fmt.Sprintf("Version: %s, Board: %s, Platform: %s", version, board, platform)
		rd.Logger.Infof("   ✅ %s (%.2fs)\n", result.Message, result.Duration.Seconds())

		// Print additional details
		rd.Logger.Info("   📊 System Information:")
		rd.Logger.Infof("      - Version: %s", version)
		rd.Logger.Infof("      - Board: %s", board)
		rd.Logger.Infof("      - Platform: %s", platform)
		rd.Logger.Infof("      - Architecture: %s", res["architecture-name"])
		rd.Logger.Infof("      - CPU: %s", res["cpu"])
		rd.Logger.Infof("      - CPU Count: %s", res["cpu-count"])
		rd.Logger.Infof("      - Uptime: %s\n", res["uptime"])
	} else {
		result.Status = "WARN"
		result.Message = "No system resource information returned"
		rd.Logger.Warnf("   ⚠️  %s (%.2fs)\n", result.Message, result.Duration.Seconds())
	}

	return result
}

// PrintSummary prints the diagnostic summary
func (rd *RouterDiagnostic) PrintSummary(results []DiagnosticResult) {
	rd.Logger.Info("🔍 ========================================")
	rd.Logger.Info("📊 DIAGNOSTIC SUMMARY")
	rd.Logger.Info("🔍 ========================================\n")

	passCount := 0
	failCount := 0
	skipCount := 0
	warnCount := 0

	for _, result := range results {
		switch result.Status {
		case "PASS":
			passCount++
			rd.Logger.Infof("✅ %s: %s", result.TestName, result.Status)
		case "FAIL":
			failCount++
			rd.Logger.Errorf("❌ %s: %s", result.TestName, result.Status)
			if len(result.Suggestions) > 0 {
				rd.Logger.Error("   Suggestions:")
				for _, suggestion := range result.Suggestions {
					rd.Logger.Errorf("   - %s", suggestion)
				}
			}
		case "SKIP":
			skipCount++
			rd.Logger.Infof("⏭️  %s: %s", result.TestName, result.Status)
		case "WARN":
			warnCount++
			rd.Logger.Warnf("⚠️  %s: %s", result.TestName, result.Status)
		}
	}

	rd.Logger.Info("\n🔍 ========================================")
	rd.Logger.Infof("📈 Results: %d Passed, %d Failed, %d Warnings, %d Skipped",
		passCount, failCount, warnCount, skipCount)
	rd.Logger.Info("🔍 ========================================\n")

	if failCount > 0 {
		rd.Logger.Error("❌ DIAGNOSIS: Connection issues detected!")
		rd.Logger.Error("\n💡 RECOMMENDED ACTIONS:")
		rd.Logger.Error("1. Review failed tests above for specific issues")
		rd.Logger.Error("2. Follow the suggestions provided for each failed test")
		rd.Logger.Error("3. Verify router is powered on and network cable connected")
		rd.Logger.Error("4. Check if you can connect using Winbox/WebFig")
		rd.Logger.Error("5. Verify firewall rules on both sides")
		rd.Logger.Error("6. Check if API service is enabled: IP -> Services -> API")
		rd.Logger.Errorf("7. Verify port %d is the correct API port (default: 8728)", rd.Port)
	} else if passCount > 0 {
		rd.Logger.Info("✅ DIAGNOSIS: All critical tests passed!")
		rd.Logger.Info("🎉 Router connection is healthy and ready to use.")
	}

	rd.Logger.Info("\n🔍 ========================================\n")
}

// Helper functions

func printUsage() {
	fmt.Println("🔍 Router Connection Diagnostic Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  router-diagnostic <host> <port> <username> <password>")
	fmt.Println("\nExample:")
	fmt.Println("  router-diagnostic 160.19.144.8 8728 admin password123")
	fmt.Println("  router-diagnostic 192.168.1.1 8728 admin \"\"")
	fmt.Println("\nCommon Ports:")
	fmt.Println("  8728 - RouterOS API (default)")
	fmt.Println("  8729 - RouterOS API-SSL")
	fmt.Println("  80   - HTTP/WebFig")
	fmt.Println("  8291 - Winbox")
}

func parsePort(portStr string) int {
	var port int
	_, err := fmt.Sscanf(portStr, "%d", &port)
	if err != nil || port < 1 || port > 65535 {
		return 0
	}
	return port
}
