package main

import (
	"log"
	"net/http"

	"github.com/codergrounds/lancache-unifi/internal/config"
	"github.com/codergrounds/lancache-unifi/internal/scheduler"
	"github.com/codergrounds/lancache-unifi/internal/sync"
	"github.com/codergrounds/lancache-unifi/internal/unifi"
	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("lancache-unifi: ")

	// Attempt to load .env file locally if present
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] configuration error: %v", err)
	}

	if cfg.DryRun {
		log.Println("[INFO] running in DRY-RUN mode — no changes will be made")
	}

	// Create clients
	unifiClient := unifi.NewClient(cfg.UniFiHost, cfg.UniFiSite, cfg.UniFiAPIKey)
	httpClient := &http.Client{}

	// Create syncer
	syncer := sync.NewSyncer(
		unifiClient,
		httpClient,
		cfg.LancacheIP,
		cfg.ServiceAllowlist,
		cfg.ServiceBlocklist,
		cfg.DryRun,
		cfg.TTL,
	)

	// Run initial sync immediately on startup
	log.Println("[INFO] running initial sync on startup")
	if err := syncer.Run(); err != nil {
		log.Printf("[ERROR] initial sync failed: %v", err)
	}

	// Start scheduler for recurring syncs
	sched := scheduler.New(cfg.CronSchedule)
	log.Printf("[INFO] scheduled sync with expression: %s", sched.Schedule())

	if err := sched.Start(syncer.Run); err != nil {
		log.Fatalf("[FATAL] scheduler error: %v", err)
	}
}
