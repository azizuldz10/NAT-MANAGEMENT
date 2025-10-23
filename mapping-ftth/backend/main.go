package main

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	db, err := InitDatabase()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	// Router endpoints (Server, OLT, ODC, ODP)
	api.HandleFunc("/routers", GetRouters(db)).Methods("GET")
	api.HandleFunc("/routers", CreateRouter(db)).Methods("POST")
	api.HandleFunc("/routers/{id}", GetRouterByID(db)).Methods("GET")
	api.HandleFunc("/routers/{id}", UpdateRouter(db)).Methods("PUT")
	api.HandleFunc("/routers/{id}", DeleteRouter(db)).Methods("DELETE")

	// Pelanggan endpoints
	api.HandleFunc("/pelanggan", GetPelanggan(db)).Methods("GET")
	api.HandleFunc("/pelanggan", CreatePelanggan(db)).Methods("POST")
	api.HandleFunc("/pelanggan/{id}", GetPelangganByID(db)).Methods("GET")
	api.HandleFunc("/pelanggan/{id}", UpdatePelanggan(db)).Methods("PUT")
	api.HandleFunc("/pelanggan/{id}", DeletePelanggan(db)).Methods("DELETE")

	// Mikrotik endpoints
	api.HandleFunc("/mikrotik/status", GetMikrotikStatus(db)).Methods("GET")
	api.HandleFunc("/mikrotik/test", TestMikrotikConnection(db)).Methods("GET")
	api.HandleFunc("/mikrotik/config", GetMikrotikConfig(db)).Methods("GET")
	api.HandleFunc("/mikrotik/config", UpdateMikrotikConfig(db)).Methods("PUT")

	// Utility endpoints
	api.HandleFunc("/stats", GetStatistics(db)).Methods("GET")
	api.HandleFunc("/export", ExportData(db)).Methods("GET")
	api.HandleFunc("/import", ImportData(db)).Methods("POST")

	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../.")))

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Println("Backend server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
