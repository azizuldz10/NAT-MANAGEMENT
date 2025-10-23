package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type Statistics struct {
	ServerCount    int `json:"server_count"`
	OLTCount       int `json:"olt_count"`
	ODCCount       int `json:"odc_count"`
	ODPCount       int `json:"odp_count"`
	PelangganCount int `json:"pelanggan_count"`
}

type ExportDataModel struct {
	Routers   []Router    `json:"routers"`
	Pelanggan []Pelanggan `json:"pelanggan"`
}

func GetStatistics(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := Statistics{}

		db.QueryRow(`SELECT COUNT(*) FROM routers WHERE type = 'server'`).Scan(&stats.ServerCount)
		db.QueryRow(`SELECT COUNT(*) FROM routers WHERE type = 'olt'`).Scan(&stats.OLTCount)
		db.QueryRow(`SELECT COUNT(*) FROM routers WHERE type = 'odc'`).Scan(&stats.ODCCount)
		db.QueryRow(`SELECT COUNT(*) FROM routers WHERE type = 'odp'`).Scan(&stats.ODPCount)
		db.QueryRow(`SELECT COUNT(*) FROM pelanggan`).Scan(&stats.PelangganCount)

		sendSuccess(w, "Statistics fetched successfully", stats)
	}
}

func ExportData(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exportData := ExportDataModel{}

		// Get all routers
		rows, err := db.Query(`
			SELECT id, name, type, parent_id, coordinates, created_at 
			FROM routers 
			ORDER BY created_at ASC
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var router Router
				rows.Scan(&router.ID, &router.Name, &router.Type, &router.ParentID, &router.Coordinates, &router.CreatedAt)
				exportData.Routers = append(exportData.Routers, router)
			}
		}

		// Get all pelanggan
		rows2, err := db.Query(`
			SELECT id, name, odp_id, pppoe, whatsapp, COALESCE(profile, ''), coordinates, status, created_at 
			FROM pelanggan 
			ORDER BY created_at ASC
		`)
		if err == nil {
			defer rows2.Close()
			for rows2.Next() {
				var p Pelanggan
				rows2.Scan(&p.ID, &p.Name, &p.ODPID, &p.PPPOE, &p.WhatsApp, &p.Profile, &p.Coordinates, &p.Status, &p.CreatedAt)
				exportData.Pelanggan = append(exportData.Pelanggan, p)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=ftth-export.json")
		json.NewEncoder(w).Encode(exportData)
	}
}

func ImportData(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var importData ExportDataModel
		if err := json.NewDecoder(r.Body).Decode(&importData); err != nil {
			sendError(w, "Invalid JSON data", err)
			return
		}

		// Begin transaction
		tx, err := db.Begin()
		if err != nil {
			sendError(w, "Failed to start transaction", err)
			return
		}

		// Clear existing data
		tx.Exec(`DELETE FROM pelanggan`)
		tx.Exec(`DELETE FROM routers`)

		// Insert routers
		for _, router := range importData.Routers {
			_, err := tx.Exec(`
				INSERT INTO routers (id, name, type, parent_id, coordinates, created_at) 
				VALUES (?, ?, ?, ?, ?, ?)
			`, router.ID, router.Name, router.Type, router.ParentID, router.Coordinates, router.CreatedAt)
			
			if err != nil {
				tx.Rollback()
				sendError(w, "Failed to import router", err)
				return
			}
		}

		// Insert pelanggan
		for _, p := range importData.Pelanggan {
			_, err := tx.Exec(`
				INSERT INTO pelanggan (id, name, odp_id, pppoe, whatsapp, profile, coordinates, status, created_at) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, p.ID, p.Name, p.ODPID, p.PPPOE, p.WhatsApp, p.Profile, p.Coordinates, p.Status, p.CreatedAt)
			
			if err != nil {
				tx.Rollback()
				sendError(w, "Failed to import pelanggan", err)
				return
			}
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			sendError(w, "Failed to commit transaction", err)
			return
		}

		sendSuccess(w, "Data imported successfully", map[string]int{
			"routers":   len(importData.Routers),
			"pelanggan": len(importData.Pelanggan),
		})
	}
}
