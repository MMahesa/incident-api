package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/MMahesa/incident-api/internal/api"
	"github.com/MMahesa/incident-api/internal/incidents"
)

func main() {
	addr := ":8080"
	if value := os.Getenv("PORT"); value != "" {
		addr = ":" + value
	}

	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = filepath.Join("data", "incidents.json")
	}

	store, err := incidents.NewFileStore(dataFile)
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Seed([]incidents.Incident{
		{
			Title:       "API latency spike",
			Service:     "auth-service",
			Severity:    incidents.SeverityHigh,
			Status:      incidents.StatusInvestigating,
			Description: "P95 latency increased beyond 700ms for the last 10 minutes.",
			Owner:       "noc-team",
		},
		{
			Title:       "Packet loss on uplink",
			Service:     "edge-router",
			Severity:    incidents.SeverityCritical,
			Status:      incidents.StatusMitigated,
			Description: "Intermittent packet loss detected on the primary uplink.",
			Owner:       "network-ops",
		},
	}); err != nil {
		log.Fatal(err)
	}

	handler := api.NewServer(store)
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("incident-api listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
