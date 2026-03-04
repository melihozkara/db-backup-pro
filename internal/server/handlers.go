package server

import (
	"dbbackup/internal/config"
	"dbbackup/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// AppAPI defines all methods the HTTP server needs from the App
type AppAPI interface {
	// Databases
	GetDatabases() ([]database.DatabaseConnection, error)
	GetDatabase(id int64) (*database.DatabaseConnection, error)
	AddDatabase(db database.DatabaseConnection) error
	UpdateDatabase(db database.DatabaseConnection) error
	DeleteDatabase(id int64) error
	TestDatabaseConnection(db database.DatabaseConnection) error

	// Storage Targets
	GetStorageTargets() ([]database.StorageTarget, error)
	GetStorageTarget(id int64) (*database.StorageTarget, error)
	AddStorageTarget(st database.StorageTarget) (int64, error)
	UpdateStorageTarget(st database.StorageTarget) error
	DeleteStorageTarget(id int64) error
	TestStorageConnection(st database.StorageTarget) error

	// Backup Jobs
	GetBackupJobs() ([]database.BackupJob, error)
	GetBackupJob(id int64) (*database.BackupJob, error)
	AddBackupJob(job database.BackupJob) error
	UpdateBackupJob(job database.BackupJob) error
	DeleteBackupJob(id int64) error
	ToggleBackupJobActive(id int64, active bool) error
	RunBackupNow(jobID int64) error

	// History & Dashboard
	GetBackupHistory(filter database.HistoryFilter) ([]database.BackupHistory, error)
	GetDashboardStats() (*database.DashboardStats, error)

	// Settings
	GetSettings() (*database.AppSettings, error)
	SaveSettings(settings database.AppSettings) error
	SetAppLanguage(lang string) error
	GetTranslations(lang string) (map[string]interface{}, error)

	// Telegram
	TestTelegramConnection(botToken string) error
	SendTestTelegramMessage(botToken, chatID string) error

	// Utility
	ValidateFolder(path string) error
	GetServerConfig() (*config.ServerConfig, error)
	SaveServerConfig(serverCfg config.ServerConfig) error
}

// getApp returns the typed app from the server
func (s *Server) getApp() AppAPI {
	return s.app
}

// ==================== JSON HELPERS ====================

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parseID(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")
	return strconv.ParseInt(idStr, 10, 64)
}

func decodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// ==================== DATABASE HANDLERS ====================

func (s *Server) handleGetDatabases(w http.ResponseWriter, r *http.Request) {
	data, err := s.getApp().GetDatabases()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleGetDatabase(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	data, err := s.getApp().GetDatabase(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleAddDatabase(w http.ResponseWriter, r *http.Request) {
	var db database.DatabaseConnection
	if err := decodeJSON(r, &db); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.getApp().AddDatabase(db); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (s *Server) handleUpdateDatabase(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var db database.DatabaseConnection
	if err := decodeJSON(r, &db); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	db.ID = id
	if err := s.getApp().UpdateDatabase(db); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDeleteDatabase(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := s.getApp().DeleteDatabase(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleTestDatabase(w http.ResponseWriter, r *http.Request) {
	var db database.DatabaseConnection
	if err := decodeJSON(r, &db); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.getApp().TestDatabaseConnection(db); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ==================== STORAGE TARGET HANDLERS ====================

func (s *Server) handleGetStorageTargets(w http.ResponseWriter, r *http.Request) {
	data, err := s.getApp().GetStorageTargets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleGetStorageTarget(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	data, err := s.getApp().GetStorageTarget(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleAddStorageTarget(w http.ResponseWriter, r *http.Request) {
	var st database.StorageTarget
	if err := decodeJSON(r, &st); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, err := s.getApp().AddStorageTarget(st)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (s *Server) handleUpdateStorageTarget(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var st database.StorageTarget
	if err := decodeJSON(r, &st); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	st.ID = id
	if err := s.getApp().UpdateStorageTarget(st); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDeleteStorageTarget(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := s.getApp().DeleteStorageTarget(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleTestStorageTarget(w http.ResponseWriter, r *http.Request) {
	var st database.StorageTarget
	if err := decodeJSON(r, &st); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.getApp().TestStorageConnection(st); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ==================== BACKUP JOB HANDLERS ====================

func (s *Server) handleGetJobs(w http.ResponseWriter, r *http.Request) {
	data, err := s.getApp().GetBackupJobs()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	data, err := s.getApp().GetBackupJob(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleAddJob(w http.ResponseWriter, r *http.Request) {
	var job database.BackupJob
	if err := decodeJSON(r, &job); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.getApp().AddBackupJob(job); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (s *Server) handleUpdateJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var job database.BackupJob
	if err := decodeJSON(r, &job); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	job.ID = id
	if err := s.getApp().UpdateBackupJob(job); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDeleteJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := s.getApp().DeleteBackupJob(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleToggleJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Active bool `json:"active"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.getApp().ToggleBackupJobActive(id, body.Active); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleRunJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	// Run in goroutine so the HTTP response returns immediately
	go func() {
		if err := s.getApp().RunBackupNow(id); err != nil {
			fmt.Printf("Backup job %d failed: %s\n", id, err.Error())
		}
	}()
	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// ==================== HISTORY & DASHBOARD HANDLERS ====================

func (s *Server) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := database.HistoryFilter{}

	if v := q.Get("job_id"); v != "" {
		filter.JobID, _ = strconv.ParseInt(v, 10, 64)
	}
	filter.Status = q.Get("status")
	if v := q.Get("limit"); v != "" {
		filter.Limit, _ = strconv.Atoi(v)
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	data, err := s.getApp().GetBackupHistory(filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleGetDashboardStats(w http.ResponseWriter, r *http.Request) {
	data, err := s.getApp().GetDashboardStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

// ==================== SETTINGS HANDLERS ====================

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	data, err := s.getApp().GetSettings()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleSaveSettings(w http.ResponseWriter, r *http.Request) {
	var settings database.AppSettings
	if err := decodeJSON(r, &settings); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.getApp().SaveSettings(settings); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSetLanguage(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Language string `json:"language"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.getApp().SetAppLanguage(body.Language); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleGetTranslations(w http.ResponseWriter, r *http.Request) {
	lang := r.PathValue("lang")
	data, err := s.getApp().GetTranslations(lang)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

// ==================== TELEGRAM HANDLERS ====================

func (s *Server) handleTestTelegramToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BotToken string `json:"bot_token"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.getApp().TestTelegramConnection(body.BotToken); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleTestTelegramMessage(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BotToken string `json:"bot_token"`
		ChatID   string `json:"chat_id"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.getApp().SendTestTelegramMessage(body.BotToken, body.ChatID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ==================== UTILITY HANDLERS ====================

func (s *Server) handleValidateFolder(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := s.getApp().ValidateFolder(body.Path); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleGetServerConfig(w http.ResponseWriter, r *http.Request) {
	data, err := s.getApp().GetServerConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleSaveServerConfig(w http.ResponseWriter, r *http.Request) {
	var cfg config.ServerConfig
	if err := decodeJSON(r, &cfg); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.getApp().SaveServerConfig(cfg); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
