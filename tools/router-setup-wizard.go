package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-routeros/routeros"
	"github.com/sirupsen/logrus"
)

// RouterConfig holds router configuration
type RouterConfig struct {
	Name           string
	Host           string
	Port           int
	Username       string
	Password       string
	TunnelEndpoint string
	PublicONTURL   string
	Description    string
	Enabled        bool
}

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	printBanner()

	scanner := bufio.NewScanner(os.Stdin)
	config := &RouterConfig{
		Enabled: true, // Default enabled
	}

	// Interactive wizard
	if !askYesNo(scanner, "🎯 Apakah Anda ingin menambahkan router baru?", true) {
		fmt.Println("❌ Setup dibatalkan.")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📝 STEP 1: Basic Information")
	fmt.Println(strings.Repeat("=", 70))

	// Router Name
	config.Name = askString(scanner, "📌 Nama Router (contoh: JAKARTA-01)", "", true)

	// Description (optional)
	config.Description = askString(scanner, "📄 Deskripsi (opsional)", "", false)

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🌐 STEP 2: Network Configuration")
	fmt.Println(strings.Repeat("=", 70))

	// Host/IP
	for {
		config.Host = askString(scanner, "🔗 IP Address/Hostname", "", true)
		if isValidHost(config.Host) {
			break
		}
		fmt.Println("❌ Invalid IP address or hostname. Try again.")
	}

	// Port
	for {
		portStr := askString(scanner, "🔌 Port API (default: 8728)", "8728", false)
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			fmt.Println("❌ Invalid port number. Try again.")
			continue
		}
		config.Port = port
		break
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🔐 STEP 3: Authentication")
	fmt.Println(strings.Repeat("=", 70))

	// Username
	config.Username = askString(scanner, "👤 Username", "admin", false)

	// Password
	config.Password = askPassword(scanner, "🔑 Password")

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🔧 STEP 4: NAT Configuration")
	fmt.Println(strings.Repeat("=", 70))

	// Tunnel Endpoint
	config.TunnelEndpoint = askString(scanner, "🎯 Tunnel Endpoint (format: IP:PORT, contoh: 172.22.28.5:80)", "", true)

	// Public ONT URL
	config.PublicONTURL = askString(scanner, "🌍 Public ONT URL (contoh: http://tunnel3.ebilling.id:19701)", "", true)

	// Summary
	printSummary(config)

	if !askYesNo(scanner, "\n✅ Apakah konfigurasi sudah benar?", true) {
		fmt.Println("❌ Setup dibatalkan. Silakan jalankan ulang untuk konfigurasi baru.")
		return
	}

	// Test connection
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🔍 Testing Connection...")
	fmt.Println(strings.Repeat("=", 70))

	if testConnection(config, logger) {
		fmt.Println("\n✅ Connection test berhasil!")

		// Save configuration
		fmt.Println("\n" + strings.Repeat("=", 70))
		fmt.Println("💾 Saving Configuration...")
		fmt.Println(strings.Repeat("=", 70))

		if saveConfiguration(config) {
			fmt.Println("\n🎉 Router berhasil dikonfigurasi!")
			printNextSteps(config)
		} else {
			fmt.Println("\n⚠️  Warning: Connection berhasil tapi gagal save configuration.")
			fmt.Println("📋 Manual configuration details:")
			printManualConfig(config)
		}
	} else {
		fmt.Println("\n❌ Connection test gagal.")
		fmt.Println("\n💡 Troubleshooting:")
		fmt.Println("1. Pastikan router online dan bisa di-ping")
		fmt.Println("2. Verify IP address dan port (8728 untuk API)")
		fmt.Println("3. Check apakah API service enabled di router")
		fmt.Println("4. Verify username dan password")
		fmt.Println("5. Check firewall rules")
		fmt.Println("\n📖 Lihat docs/TROUBLESHOOTING.md untuk panduan lengkap")

		if askYesNo(scanner, "\n❓ Apakah Anda ingin save configuration meskipun test gagal?", false) {
			if saveConfiguration(config) {
				fmt.Println("\n💾 Configuration tersimpan. Silakan perbaiki masalah koneksi nanti.")
				printManualConfig(config)
			}
		}
	}
}

func printBanner() {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("🚀 NAT Management System - Router Setup Wizard")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("📖 Wizard ini akan membantu Anda menambahkan router baru")
	fmt.Println("📄 Dokumentasi lengkap: docs/ROUTER-SETUP.md")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
}

func askString(scanner *bufio.Scanner, prompt string, defaultValue string, required bool) string {
	for {
		if defaultValue != "" {
			fmt.Printf("%s [%s]: ", prompt, defaultValue)
		} else if required {
			fmt.Printf("%s (*required): ", prompt)
		} else {
			fmt.Printf("%s: ", prompt)
		}

		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			if defaultValue != "" {
				return defaultValue
			}
			if !required {
				return ""
			}
			fmt.Println("❌ Field ini wajib diisi!")
			continue
		}

		return input
	}
}

func askPassword(scanner *bufio.Scanner, prompt string) string {
	fmt.Printf("%s: ", prompt)
	scanner.Scan()
	return scanner.Text()
}

func askYesNo(scanner *bufio.Scanner, prompt string, defaultYes bool) bool {
	defaultStr := "Y/n"
	if !defaultYes {
		defaultStr = "y/N"
	}

	fmt.Printf("%s [%s]: ", prompt, defaultStr)
	scanner.Scan()
	input := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if input == "" {
		return defaultYes
	}

	return input == "y" || input == "yes"
}

func isValidHost(host string) bool {
	// Check if it's a valid IP
	if net.ParseIP(host) != nil {
		return true
	}

	// Check if it's a valid hostname
	if len(host) == 0 || len(host) > 253 {
		return false
	}

	for _, char := range host {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '.') {
			return false
		}
	}

	return true
}

func printSummary(config *RouterConfig) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("📋 Configuration Summary")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Router Name     : %s\n", config.Name)
	if config.Description != "" {
		fmt.Printf("Description     : %s\n", config.Description)
	}
	fmt.Printf("Host            : %s\n", config.Host)
	fmt.Printf("Port            : %d\n", config.Port)
	fmt.Printf("Username        : %s\n", config.Username)
	fmt.Printf("Password        : %s\n", strings.Repeat("*", len(config.Password)))
	fmt.Printf("Tunnel Endpoint : %s\n", config.TunnelEndpoint)
	fmt.Printf("Public ONT URL  : %s\n", config.PublicONTURL)
	fmt.Printf("Status          : %s\n", "Enabled")
	fmt.Println(strings.Repeat("=", 70))
}

func testConnection(config *RouterConfig, logger *logrus.Logger) bool {
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Test 1: TCP Connection
	fmt.Printf("\n🔍 Test 1: TCP Connection to %s... ", address)
	conn, err := net.DialTimeout("tcp", address, 15*time.Second)
	if err != nil {
		fmt.Printf("❌ FAILED\n")
		fmt.Printf("   Error: %v\n", err)
		return false
	}
	conn.Close()
	fmt.Printf("✅ OK\n")

	// Test 2: RouterOS API Authentication
	fmt.Printf("🔍 Test 2: RouterOS API Authentication... ")
	client, err := routeros.Dial(address, config.Username, config.Password)
	if err != nil {
		fmt.Printf("❌ FAILED\n")
		fmt.Printf("   Error: %v\n", err)
		return false
	}
	defer client.Close()
	fmt.Printf("✅ OK\n")

	// Test 3: Get Router Identity
	fmt.Printf("🔍 Test 3: Get Router Identity... ")
	identityReply, err := client.Run("/system/identity/print")
	if err != nil {
		fmt.Printf("⚠️  WARNING\n")
		fmt.Printf("   Error: %v\n", err)
	} else {
		if len(identityReply.Re) > 0 {
			identity := identityReply.Re[0].Map["name"]
			fmt.Printf("✅ OK\n")
			fmt.Printf("   Router Identity: %s\n", identity)
		} else {
			fmt.Printf("⚠️  WARNING (no identity returned)\n")
		}
	}

	// Test 4: Get System Resources
	fmt.Printf("🔍 Test 4: Get System Resources... ")
	resourceReply, err := client.Run("/system/resource/print")
	if err != nil {
		fmt.Printf("⚠️  WARNING\n")
		fmt.Printf("   Error: %v\n", err)
	} else {
		if len(resourceReply.Re) > 0 {
			res := resourceReply.Re[0].Map
			fmt.Printf("✅ OK\n")
			fmt.Printf("   Version : %s\n", res["version"])
			fmt.Printf("   Board   : %s\n", res["board-name"])
			fmt.Printf("   Platform: %s\n", res["platform"])
		} else {
			fmt.Printf("⚠️  WARNING (no resources returned)\n")
		}
	}

	return true
}

func saveConfiguration(config *RouterConfig) bool {
	fmt.Println("\n💾 IMPORTANT: Configuration auto-save not implemented yet.")
	fmt.Println("📋 Please add router manually via Web UI:")
	fmt.Println("\n1. Login to NAT Management System")
	fmt.Println("2. Go to Router Management")
	fmt.Println("3. Click 'Add Router'")
	fmt.Println("4. Use the following details:\n")
	printManualConfig(config)
	return true
}

func printManualConfig(config *RouterConfig) {
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Router Name     : %s\n", config.Name)
	fmt.Printf("Host            : %s\n", config.Host)
	fmt.Printf("Port            : %d\n", config.Port)
	fmt.Printf("Username        : %s\n", config.Username)
	fmt.Printf("Password        : %s\n", config.Password)
	fmt.Printf("Tunnel Endpoint : %s\n", config.TunnelEndpoint)
	fmt.Printf("Public ONT URL  : %s\n", config.PublicONTURL)
	if config.Description != "" {
		fmt.Printf("Description     : %s\n", config.Description)
	}
	fmt.Printf("Status          : Enabled\n")
	fmt.Println(strings.Repeat("-", 70))
}

func printNextSteps(config *RouterConfig) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎯 Next Steps")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("1. ✅ Add router configuration via Web UI (see details above)")
	fmt.Println("2. 🔐 Setup user access permissions (User Management)")
	fmt.Println("3. 🔧 Configure NAT rules di MikroTik:")
	fmt.Println("   - Buat NAT rule dengan comment 'REMOTE ONT PELANGGAN'")
	fmt.Println("   - Set destination ke tunnel endpoint")
	fmt.Println("4. 🧪 Test NAT operations via aplikasi")
	fmt.Println("5. 📖 Read docs/ROUTER-SETUP.md untuk panduan lengkap")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("\n🎉 Setup wizard selesai!")
}
