package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/google/uuid"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Router Handlers
func GetRouters(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT id, name, type, parent_id, coordinates, created_at 
			FROM routers 
			ORDER BY created_at DESC
		`)
		if err != nil {
			sendError(w, "Failed to fetch routers", err)
			return
		}
		defer rows.Close()

		routers := []Router{}
		for rows.Next() {
			var router Router
			err := rows.Scan(&router.ID, &router.Name, &router.Type, &router.ParentID, &router.Coordinates, &router.CreatedAt)
			if err != nil {
				continue
			}
			routers = append(routers, router)
		}

		sendSuccess(w, "Routers fetched successfully", routers)
	}
}

func GetRouterByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var router Router
		err := db.QueryRow(`
			SELECT id, name, type, parent_id, coordinates, created_at 
			FROM routers WHERE id = ?
		`, id).Scan(&router.ID, &router.Name, &router.Type, &router.ParentID, &router.Coordinates, &router.CreatedAt)

		if err == sql.ErrNoRows {
			sendError(w, "Router not found", err)
			return
		}
		if err != nil {
			sendError(w, "Failed to fetch router", err)
			return
		}

		sendSuccess(w, "Router fetched successfully", router)
	}
}

func CreateRouter(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var router Router
		if err := json.NewDecoder(r.Body).Decode(&router); err != nil {
			sendError(w, "Invalid request body", err)
			return
		}

		router.ID = uuid.New().String()

		_, err := db.Exec(`
			INSERT INTO routers (id, name, type, parent_id, coordinates) 
			VALUES (?, ?, ?, ?, ?)
		`, router.ID, router.Name, router.Type, router.ParentID, router.Coordinates)

		if err != nil {
			sendError(w, "Failed to create router", err)
			return
		}

		sendSuccess(w, "Router created successfully", router)
	}
}

func UpdateRouter(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var router Router
		if err := json.NewDecoder(r.Body).Decode(&router); err != nil {
			sendError(w, "Invalid request body", err)
			return
		}

		result, err := db.Exec(`
			UPDATE routers 
			SET name = ?, type = ?, parent_id = ?, coordinates = ? 
			WHERE id = ?
		`, router.Name, router.Type, router.ParentID, router.Coordinates, id)

		if err != nil {
			sendError(w, "Failed to update router", err)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			sendError(w, "Router not found", nil)
			return
		}

		router.ID = id
		sendSuccess(w, "Router updated successfully", router)
	}
}

func DeleteRouter(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		// Check if router has children
		var childCount int
		err := db.QueryRow(`SELECT COUNT(*) FROM routers WHERE parent_id = ?`, id).Scan(&childCount)
		if err == nil && childCount > 0 {
			sendError(w, "Cannot delete router with children", nil)
			return
		}

		// Check if router has pelanggan (for ODP)
		var pelangganCount int
		err = db.QueryRow(`SELECT COUNT(*) FROM pelanggan WHERE odp_id = ?`, id).Scan(&pelangganCount)
		if err == nil && pelangganCount > 0 {
			sendError(w, "Cannot delete ODP with pelanggan", nil)
			return
		}

		result, err := db.Exec(`DELETE FROM routers WHERE id = ?`, id)
		if err != nil {
			sendError(w, "Failed to delete router", err)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			sendError(w, "Router not found", nil)
			return
		}

		sendSuccess(w, "Router deleted successfully", nil)
	}
}

// Pelanggan Handlers
func GetPelanggan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT id, name, odp_id, pppoe, whatsapp, COALESCE(profile, ''), coordinates, status, created_at 
			FROM pelanggan 
			ORDER BY created_at DESC
		`)
		if err != nil {
			sendError(w, "Failed to fetch pelanggan", err)
			return
		}
		defer rows.Close()

		pelanggan := []Pelanggan{}
		for rows.Next() {
			var p Pelanggan
			err := rows.Scan(&p.ID, &p.Name, &p.ODPID, &p.PPPOE, &p.WhatsApp, &p.Profile, &p.Coordinates, &p.Status, &p.CreatedAt)
			if err != nil {
				continue
			}
			pelanggan = append(pelanggan, p)
		}

		sendSuccess(w, "Pelanggan fetched successfully", pelanggan)
	}
}

func GetPelangganByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var p Pelanggan
		err := db.QueryRow(`
			SELECT id, name, odp_id, pppoe, whatsapp, COALESCE(profile, ''), coordinates, status, created_at 
			FROM pelanggan WHERE id = ?
		`, id).Scan(&p.ID, &p.Name, &p.ODPID, &p.PPPOE, &p.WhatsApp, &p.Profile, &p.Coordinates, &p.Status, &p.CreatedAt)

		if err == sql.ErrNoRows {
			sendError(w, "Pelanggan not found", err)
			return
		}
		if err != nil {
			sendError(w, "Failed to fetch pelanggan", err)
			return
		}

		sendSuccess(w, "Pelanggan fetched successfully", p)
	}
}

func CreatePelanggan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p Pelanggan
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			sendError(w, "Invalid request body", err)
			return
		}

		p.ID = uuid.New().String()
		if p.Status == "" {
			p.Status = "offline"
		}

		_, err := db.Exec(`
			INSERT INTO pelanggan (id, name, odp_id, pppoe, whatsapp, profile, coordinates, status) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, p.ID, p.Name, p.ODPID, p.PPPOE, p.WhatsApp, p.Profile, p.Coordinates, p.Status)

		if err != nil {
			sendError(w, "Failed to create pelanggan", err)
			return
		}

		sendSuccess(w, "Pelanggan created successfully", p)
	}
}

func UpdatePelanggan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var p Pelanggan
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			sendError(w, "Invalid request body", err)
			return
		}

		result, err := db.Exec(`
			UPDATE pelanggan 
			SET name = ?, odp_id = ?, pppoe = ?, whatsapp = ?, profile = ?, coordinates = ?, status = ? 
			WHERE id = ?
		`, p.Name, p.ODPID, p.PPPOE, p.WhatsApp, p.Profile, p.Coordinates, p.Status, id)

		if err != nil {
			sendError(w, "Failed to update pelanggan", err)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			sendError(w, "Pelanggan not found", nil)
			return
		}

		p.ID = id
		sendSuccess(w, "Pelanggan updated successfully", p)
	}
}

func DeletePelanggan(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		result, err := db.Exec(`DELETE FROM pelanggan WHERE id = ?`, id)
		if err != nil {
			sendError(w, "Failed to delete pelanggan", err)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			sendError(w, "Pelanggan not found", nil)
			return
		}

		sendSuccess(w, "Pelanggan deleted successfully", nil)
	}
}

// Utility functions
func sendSuccess(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func sendError(w http.ResponseWriter, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	
	errMsg := message
	if err != nil {
		errMsg = message + ": " + err.Error()
	}
	
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   errMsg,
	})
}
