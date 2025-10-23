package main

import (
	"database/sql"
	"time"
	_ "modernc.org/sqlite"
)

func InitDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./ftth.db")
	if err != nil {
		return nil, err
	}

	// Create routers table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS routers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			parent_id TEXT,
			coordinates TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(parent_id) REFERENCES routers(id) ON DELETE RESTRICT
		)
	`)
	if err != nil {
		return nil, err
	}

	// Create pelanggan table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS pelanggan (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			odp_id TEXT NOT NULL,
			pppoe TEXT,
			whatsapp TEXT,
			profile TEXT,
			coordinates TEXT NOT NULL,
			status TEXT DEFAULT 'offline',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(odp_id) REFERENCES routers(id) ON DELETE RESTRICT
		)
	`)
	if err != nil {
		return nil, err
	}

	// Create mikrotik_config table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS mikrotik_config (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			host TEXT NOT NULL DEFAULT '192.168.88.1',
			user TEXT NOT NULL DEFAULT 'admin',
			password TEXT NOT NULL DEFAULT '',
			port INTEGER NOT NULL DEFAULT 8728
		)
	`)
	if err != nil {
		return nil, err
	}

	// Insert default mikrotik config if not exists
	_, err = db.Exec(`
		INSERT OR IGNORE INTO mikrotik_config (id, host, user, password, port)
		VALUES (1, '192.168.88.1', 'admin', '', 8728)
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

type Router struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	ParentID    *string   `json:"parent_id"`
	Coordinates string    `json:"coordinates"`
	CreatedAt   time.Time `json:"created_at"`
}

type Pelanggan struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	ODPID       string    `json:"odp_id"`
	PPPOE       string    `json:"pppoe"`
	WhatsApp    string    `json:"whatsapp"`
	Profile     string    `json:"profile"`
	Coordinates string    `json:"coordinates"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type MikrotikConfig struct {
	ID       int    `json:"id"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Port     int    `json:"port"`
}
