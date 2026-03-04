package server

import (
	"io/fs"
	"net/http"
	"strings"
)

// setupRoutes configures all HTTP routes using Go 1.22+ enhanced ServeMux
func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// ==================== API ROUTES ====================

	// Databases
	mux.HandleFunc("GET /api/databases", s.handleGetDatabases)
	mux.HandleFunc("POST /api/databases", s.handleAddDatabase)
	mux.HandleFunc("GET /api/databases/{id}", s.handleGetDatabase)
	mux.HandleFunc("PUT /api/databases/{id}", s.handleUpdateDatabase)
	mux.HandleFunc("DELETE /api/databases/{id}", s.handleDeleteDatabase)
	mux.HandleFunc("POST /api/databases/test", s.handleTestDatabase)

	// Storage Targets
	mux.HandleFunc("GET /api/storage-targets", s.handleGetStorageTargets)
	mux.HandleFunc("POST /api/storage-targets", s.handleAddStorageTarget)
	mux.HandleFunc("GET /api/storage-targets/{id}", s.handleGetStorageTarget)
	mux.HandleFunc("PUT /api/storage-targets/{id}", s.handleUpdateStorageTarget)
	mux.HandleFunc("DELETE /api/storage-targets/{id}", s.handleDeleteStorageTarget)
	mux.HandleFunc("POST /api/storage-targets/test", s.handleTestStorageTarget)

	// Backup Jobs
	mux.HandleFunc("GET /api/jobs", s.handleGetJobs)
	mux.HandleFunc("POST /api/jobs", s.handleAddJob)
	mux.HandleFunc("GET /api/jobs/{id}", s.handleGetJob)
	mux.HandleFunc("PUT /api/jobs/{id}", s.handleUpdateJob)
	mux.HandleFunc("DELETE /api/jobs/{id}", s.handleDeleteJob)
	mux.HandleFunc("PUT /api/jobs/{id}/toggle", s.handleToggleJob)
	mux.HandleFunc("POST /api/jobs/{id}/run", s.handleRunJob)

	// History & Dashboard
	mux.HandleFunc("GET /api/history", s.handleGetHistory)
	mux.HandleFunc("GET /api/dashboard/stats", s.handleGetDashboardStats)

	// Settings
	mux.HandleFunc("GET /api/settings", s.handleGetSettings)
	mux.HandleFunc("PUT /api/settings", s.handleSaveSettings)
	mux.HandleFunc("PUT /api/language", s.handleSetLanguage)
	mux.HandleFunc("GET /api/translations/{lang}", s.handleGetTranslations)

	// Telegram
	mux.HandleFunc("POST /api/telegram/test-token", s.handleTestTelegramToken)
	mux.HandleFunc("POST /api/telegram/test-message", s.handleTestTelegramMessage)

	// Utility
	mux.HandleFunc("POST /api/validate-folder", s.handleValidateFolder)
	mux.HandleFunc("GET /api/server-config", s.handleGetServerConfig)
	mux.HandleFunc("PUT /api/server-config", s.handleSaveServerConfig)

	// SSE Events
	mux.Handle("GET /api/events", s.SSEHub)

	// ==================== SPA STATIC FILES ====================
	s.setupSPA(mux)

	return mux
}

// setupSPA serves the embedded frontend assets with SPA fallback
func (s *Server) setupSPA(mux *http.ServeMux) {
	// Get the dist subdirectory from the embedded FS
	distFS, err := fs.Sub(s.assets, "frontend/dist")
	if err != nil {
		// Fallback: try without subdirectory
		distFS = s.assets
	}

	fileServer := http.FileServer(http.FS(distFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// API routes are already handled above
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the static file
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists
		f, err := distFS.Open(strings.TrimPrefix(path, "/"))
		if err != nil {
			// SPA fallback: serve index.html for unknown paths (react-router)
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	})
}
