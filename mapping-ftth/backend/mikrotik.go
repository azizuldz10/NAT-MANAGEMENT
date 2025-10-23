package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"fmt"
	"net"
	"time"
	"bytes"
	"encoding/binary"
)

type MikrotikAPI struct {
	Host     string
	User     string
	Password string
	Port     int
	conn     net.Conn
}

func NewMikrotikAPI(config MikrotikConfig) *MikrotikAPI {
	return &MikrotikAPI{
		Host:     config.Host,
		User:     config.User,
		Password: config.Password,
		Port:     config.Port,
	}
}

func (m *MikrotikAPI) Connect() error {
	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return err
	}
	m.conn = conn
	
	// Set read/write deadline
	m.conn.SetDeadline(time.Now().Add(30 * time.Second))

	// Login
	if err := m.login(); err != nil {
		m.conn.Close()
		return err
	}

	return nil
}

func (m *MikrotikAPI) Disconnect() {
	if m.conn != nil {
		m.conn.Close()
	}
}

func (m *MikrotikAPI) login() error {
	// Send /login command
	if err := m.writeCommand("/login", nil); err != nil {
		return err
	}

	// Read response
	_, err := m.readResponse()
	if err != nil {
		return err
	}

	// Send credentials
	attrs := map[string]string{
		"name":     m.User,
		"password": m.Password,
	}
	if err := m.writeCommand("/login", attrs); err != nil {
		return err
	}

	// Read login response
	_, err = m.readResponse()
	return err
}

func (m *MikrotikAPI) GetActivePPPoE() ([]map[string]string, error) {
	// Reset deadline for this operation
	if m.conn != nil {
		m.conn.SetDeadline(time.Now().Add(30 * time.Second))
	}
	
	if err := m.writeCommand("/ppp/active/print", nil); err != nil {
		return nil, fmt.Errorf("write command failed: %v", err)
	}

	responses, err := m.readResponse()
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}

	fmt.Printf("DEBUG GetActivePPPoE: Got %d responses\n", len(responses))

	var activeUsers []map[string]string
	for i, resp := range responses {
		fmt.Printf("DEBUG Response[%d]: %+v\n", i, resp)
		if _, ok := resp["!re"]; ok {
			user := make(map[string]string)
			if name, ok := resp["=name"]; ok {
				user["name"] = name
			}
			if addr, ok := resp["=address"]; ok {
				user["address"] = addr
			}
			if uptime, ok := resp["=uptime"]; ok {
				user["uptime"] = uptime
			}
			if callerId, ok := resp["=caller-id"]; ok {
				user["caller_id"] = callerId
			}
			// Tambah field service (profile)
			if service, ok := resp["=service"]; ok {
				user["profile"] = service
			}
			if len(user) > 0 {
				fmt.Printf("DEBUG: Adding user: %+v\n", user)
				activeUsers = append(activeUsers, user)
			}
		}
	}

	fmt.Printf("DEBUG GetActivePPPoE: Returning %d active users\n", len(activeUsers))
	return activeUsers, nil
}

// GetPPPoESecrets - Fetch all PPPoE secrets (including offline users) with profiles
func (m *MikrotikAPI) GetPPPoESecrets() ([]map[string]string, error) {
	// Reset deadline for this operation
	if m.conn != nil {
		m.conn.SetDeadline(time.Now().Add(60 * time.Second))
	}
	
	if err := m.writeCommand("/ppp/secret/print", nil); err != nil {
		return nil, fmt.Errorf("write command failed: %v", err)
	}

	responses, err := m.readResponse()
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}

	fmt.Printf("DEBUG GetPPPoESecrets: Got %d responses\n", len(responses))

	var secrets []map[string]string
	for i, resp := range responses {
		fmt.Printf("DEBUG Secret[%d]: %+v\n", i, resp)
		if _, ok := resp["!re"]; ok {
			secret := make(map[string]string)
			if name, ok := resp["=name"]; ok {
				secret["name"] = name
			}
			if service, ok := resp["=service"]; ok {
				secret["profile"] = service
			}
			if profile, ok := resp["=profile"]; ok {
				secret["profile"] = profile
			}
			if len(secret) > 0 && secret["name"] != "" {
				fmt.Printf("DEBUG: Adding secret: %+v\n", secret)
				secrets = append(secrets, secret)
			}
		}
	}

	fmt.Printf("DEBUG GetPPPoESecrets: Returning %d secrets\n", len(secrets))
	return secrets, nil
}

func (m *MikrotikAPI) writeCommand(command string, attrs map[string]string) error {
	if err := m.writeWord(command); err != nil {
		return err
	}

	for k, v := range attrs {
		if err := m.writeWord("=" + k + "=" + v); err != nil {
			return err
		}
	}

	// Write empty word to signal end
	return m.writeLen(0)
}

func (m *MikrotikAPI) writeWord(word string) error {
	if err := m.writeLen(len(word)); err != nil {
		return err
	}
	_, err := m.conn.Write([]byte(word))
	return err
}

func (m *MikrotikAPI) writeLen(l int) error {
	var buf bytes.Buffer

	if l < 0x80 {
		buf.WriteByte(byte(l))
	} else if l < 0x4000 {
		buf.WriteByte(byte(l>>8 | 0x80))
		buf.WriteByte(byte(l))
	} else if l < 0x200000 {
		buf.WriteByte(byte(l>>16 | 0xC0))
		buf.WriteByte(byte(l >> 8))
		buf.WriteByte(byte(l))
	} else if l < 0x10000000 {
		buf.WriteByte(byte(l>>24 | 0xE0))
		buf.WriteByte(byte(l >> 16))
		buf.WriteByte(byte(l >> 8))
		buf.WriteByte(byte(l))
	} else {
		buf.WriteByte(0xF0)
		binary.Write(&buf, binary.BigEndian, uint32(l))
	}

	_, err := m.conn.Write(buf.Bytes())
	return err
}

func (m *MikrotikAPI) readResponse() ([]map[string]string, error) {
	var responses []map[string]string

	for {
		word, err := m.readWord()
		if err != nil {
			return nil, err
		}

		if len(word) == 0 {
			break
		}

		response := make(map[string]string)
		response[word] = ""

		// Read attributes
		for {
			attr, err := m.readWord()
			if err != nil {
				return nil, err
			}

			if len(attr) == 0 {
				break
			}

			if attr[0] == '=' {
				parts := bytes.SplitN([]byte(attr[1:]), []byte("="), 2)
				if len(parts) == 2 {
					response["="+string(parts[0])] = string(parts[1])
				}
			}
		}

		responses = append(responses, response)

		if word == "!done" {
			break
		}
	}

	return responses, nil
}

func (m *MikrotikAPI) readWord() (string, error) {
	l, err := m.readLen()
	if err != nil {
		return "", err
	}

	if l == 0 {
		return "", nil
	}

	buf := make([]byte, l)
	_, err = m.conn.Read(buf)
	return string(buf), err
}

func (m *MikrotikAPI) readLen() (int, error) {
	buf := make([]byte, 1)
	_, err := m.conn.Read(buf)
	if err != nil {
		return 0, err
	}

	b := buf[0]
	if b == 0 {
		return 0, nil
	}

	if (b & 0x80) == 0 {
		return int(b), nil
	}

	if (b & 0xC0) == 0x80 {
		buf2 := make([]byte, 1)
		m.conn.Read(buf2)
		return int(b&^0xC0)<<8 + int(buf2[0]), nil
	}

	if (b & 0xE0) == 0xC0 {
		buf2 := make([]byte, 2)
		m.conn.Read(buf2)
		return int(b&^0xE0)<<16 + int(buf2[0])<<8 + int(buf2[1]), nil
	}

	if (b & 0xF0) == 0xE0 {
		buf2 := make([]byte, 3)
		m.conn.Read(buf2)
		return int(b&^0xF0)<<24 + int(buf2[0])<<16 + int(buf2[1])<<8 + int(buf2[2]), nil
	}

	if (b & 0xF8) == 0xF0 {
		buf2 := make([]byte, 4)
		m.conn.Read(buf2)
		return int(buf2[0])<<24 + int(buf2[1])<<16 + int(buf2[2])<<8 + int(buf2[3]), nil
	}

	return 0, fmt.Errorf("invalid length byte")
}

// Handlers
func GetMikrotikConfig(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var config MikrotikConfig
		err := db.QueryRow(`
			SELECT id, host, user, password, port FROM mikrotik_config WHERE id = 1
		`).Scan(&config.ID, &config.Host, &config.User, &config.Password, &config.Port)

		if err != nil {
			sendError(w, "Failed to fetch mikrotik config", err)
			return
		}

		// Don't send password to frontend
		config.Password = "****"
		sendSuccess(w, "Config fetched successfully", config)
	}
}

func UpdateMikrotikConfig(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var config MikrotikConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			sendError(w, "Invalid request body", err)
			return
		}

		_, err := db.Exec(`
			UPDATE mikrotik_config 
			SET host = ?, user = ?, password = ?, port = ? 
			WHERE id = 1
		`, config.Host, config.User, config.Password, config.Port)

		if err != nil {
			sendError(w, "Failed to update config", err)
			return
		}

		sendSuccess(w, "Config updated successfully", config)
	}
}

func TestMikrotikConnection(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var config MikrotikConfig
		err := db.QueryRow(`
			SELECT id, host, user, password, port FROM mikrotik_config WHERE id = 1
		`).Scan(&config.ID, &config.Host, &config.User, &config.Password, &config.Port)

		if err != nil {
			sendError(w, "Failed to fetch config", err)
			return
		}

		mikrotik := NewMikrotikAPI(config)
		if err := mikrotik.Connect(); err != nil {
			sendError(w, "Connection failed", err)
			return
		}
		defer mikrotik.Disconnect()

		sendSuccess(w, "Connection successful", map[string]string{
			"status": "connected",
			"host":   config.Host,
		})
	}
}

func GetMikrotikStatus(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var config MikrotikConfig
		err := db.QueryRow(`
			SELECT id, host, user, password, port FROM mikrotik_config WHERE id = 1
		`).Scan(&config.ID, &config.Host, &config.User, &config.Password, &config.Port)

		if err != nil {
			sendError(w, "Failed to fetch config", err)
			return
		}

		mikrotik := NewMikrotikAPI(config)
		if err := mikrotik.Connect(); err != nil {
			sendError(w, "Connection failed", err)
			return
		}
		defer mikrotik.Disconnect()

		// Step 1: Get all PPPoE secrets (untuk update profile semua user)
		// Note: Disabled for now to avoid connection conflict, profile updated from active connections only
		/*
		secrets, err := mikrotik.GetPPPoESecrets()
		if err != nil {
			fmt.Printf("WARNING: Failed to get PPPoE secrets: %v\n", err)
		} else {
			fmt.Printf("DEBUG: Found %d PPPoE secrets\n", len(secrets))
			// Update profile untuk semua user yang ada di database
			profileUpdated := 0
			for _, secret := range secrets {
				if pppoe, ok := secret["name"]; ok && pppoe != "" {
					profile := secret["profile"]
					if profile == "" {
						profile = "default"
					}
					result, err := db.Exec(`UPDATE pelanggan SET profile = ? WHERE LOWER(pppoe) = LOWER(?)`, profile, pppoe)
					if err == nil {
						if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
							profileUpdated++
							fmt.Printf("DEBUG: Updated profile for %s = %s\n", pppoe, profile)
						}
					}
				}
			}
			fmt.Printf("DEBUG: Updated profile for %d users\n", profileUpdated)
		}
		*/

		// Step 2: Get active PPPoE connections
		activeUsers, err := mikrotik.GetActivePPPoE()
		if err != nil {
			// Log error but try to continue with empty result
			fmt.Printf("ERROR GetActivePPPoE: %v\n", err)
			
			// Reset all to offline if cannot get active users
			db.Exec(`UPDATE pelanggan SET status = 'offline' WHERE pppoe != '' AND pppoe IS NOT NULL`)
			
			sendSuccess(w, "Sync completed with errors", map[string]interface{}{
				"active_connections": 0,
				"online_customers":   0,
				"error":              err.Error(),
				"users":              []map[string]string{},
			})
			return
		}

		// Debug: Log active users
		fmt.Printf("DEBUG: Found %d active connections\n", len(activeUsers))
		for i, user := range activeUsers {
			fmt.Printf("  [%d] name=%s, address=%s, profile=%s\n", i+1, user["name"], user["address"], user["profile"])
		}

		// Step 3: Reset all to offline first
		db.Exec(`UPDATE pelanggan SET status = 'offline' WHERE pppoe != '' AND pppoe IS NOT NULL`)

		// Step 4: Update online users based on active connections (status + profile)
		onlineCount := 0
		for _, user := range activeUsers {
			if pppoe, ok := user["name"]; ok && pppoe != "" {
				profile := user["profile"]
				if profile == "" {
					profile = "default"
				}
				fmt.Printf("DEBUG: Trying to match PPPOE: %s with profile: %s\n", pppoe, profile)
				result, err := db.Exec(`UPDATE pelanggan SET status = 'online', profile = ? WHERE LOWER(pppoe) = LOWER(?)`, profile, pppoe)
				if err == nil {
					if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
						onlineCount++
						fmt.Printf("DEBUG: Matched! Updated %d row(s)\n", rowsAffected)
					} else {
						fmt.Printf("DEBUG: No match found for: %s\n", pppoe)
					}
				}
			}
		}

		sendSuccess(w, "Status fetched successfully", map[string]interface{}{
			"active_connections": len(activeUsers),
			"online_customers":   onlineCount,
			"users":              activeUsers,
		})
	}
}
