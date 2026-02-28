package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	httpAdapter "github.com/gjcourt/drift/internal/adapters/http"
	"github.com/gjcourt/drift/internal/adapters/storage/sqlite"
	"github.com/gjcourt/drift/internal/services"
)

func main() {
	// Configuration from env with sensible defaults.
	addr := envOr("DRIFT_ADDR", ":8080")
	dbPath := envOr("DRIFT_DB", "drift.db")

	// Resolve template and static directories relative to this source file at build time.
	// In production, override with DRIFT_TMPL_DIR and DRIFT_STATIC_DIR.
	_, file, _, _ := runtime.Caller(0)
	sourceRoot := filepath.Join(filepath.Dir(file), "..", "..")
	defaultTmplDir := filepath.Join(sourceRoot, "internal", "adapters", "http", "templates")
	defaultStaticDir := filepath.Join(sourceRoot, "web", "static")
	tmplDir := envOr("DRIFT_TMPL_DIR", defaultTmplDir)
	staticDir := envOr("DRIFT_STATIC_DIR", defaultStaticDir)

	// Open SQLite store (implements all three repository interfaces).
	store, err := sqlite.New(dbPath)
	if err != nil {
		slog.Error("open database", "err", err)
		os.Exit(1)
	}

	// Wire services.
	ingestionSvc := services.NewIngestionService(store)
	resultsSvc := services.NewResultsService(store, store)
	simSvc := services.NewSimulationService(store, store, store)

	// Build HTTP handler.
	handler := httpAdapter.New(ingestionSvc, resultsSvc, simSvc, tmplDir, staticDir)

	slog.Info("Drift starting", "addr", addr, "db", dbPath)
	if err := http.ListenAndServe(addr, handler); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
