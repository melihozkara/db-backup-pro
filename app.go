package main

import (
	"context"
	"dbbackup/internal/backup"
	"dbbackup/internal/config"
	"dbbackup/internal/database"
	"dbbackup/internal/i18n"
	"dbbackup/internal/notification"
	"dbbackup/internal/scheduler"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	backupService  *backup.BackupService
	scheduler      *scheduler.Scheduler
	eventEmitter   func(string, map[string]interface{})
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
	}
}

// SetEventEmitter sets the event emitter function (used by web mode to set SSE broadcaster)
func (a *App) SetEventEmitter(fn func(string, map[string]interface{})) {
	a.eventEmitter = fn
}

// emitEvent sends an event through the configured emitter
func (a *App) emitEvent(eventName string, data map[string]interface{}) {
	if a.eventEmitter != nil {
		a.eventEmitter(eventName, data)
	}
}

// initServices initializes all backend services (mode-independent)
func (a *App) initServices() error {
	// Veritabanini baslat
	if err := database.InitDB(); err != nil {
		fmt.Println(i18n.T("errors.databaseInitFailed")+":", err)
		return err
	}

	// Kaydedilmis dili yukle
	if lang, err := database.GetSetting("language"); err == nil && lang != "" {
		i18n.SetLanguage(lang)
	}

	// Backup service'i baslat
	repo := database.NewRepository(database.GetDB())
	a.backupService = backup.NewBackupService(repo)

	// Scheduler'i baslat
	var err error
	a.scheduler, err = scheduler.NewScheduler(a.backupService)
	if err != nil {
		fmt.Println(i18n.T("errors.schedulerInitFailed")+":", err)
		return err
	}

	// Aktif gorevleri yukle
	if err := a.scheduler.LoadJobs(); err != nil {
		fmt.Println(i18n.T("errors.jobsLoadFailed")+":", err)
	}

	// Event callback'lerini bagla (emitEvent uzerinden)
	a.scheduler.SetEventCallback(a.emitEvent)
	a.backupService.SetProgressCallback(a.emitEvent)

	// Scheduler'i baslat
	a.scheduler.Start()
	fmt.Println("Scheduler başlatıldı")

	return nil
}

// startup is called when the app starts (Wails mode)
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Wails modunda eventEmitter'i Wails runtime'a bagla
	a.eventEmitter = func(eventName string, data map[string]interface{}) {
		runtime.EventsEmit(a.ctx, eventName, data)
	}

	a.initServices()
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
	database.CloseDB()
}

// Shutdown is the exported version for web mode graceful shutdown
func (a *App) Shutdown() {
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
	database.CloseDB()
}

// ==================== DATABASE CONNECTIONS ====================

func (a *App) GetDatabases() ([]database.DatabaseConnection, error) {
	return database.GetAllDatabases()
}

func (a *App) GetDatabase(id int64) (*database.DatabaseConnection, error) {
	return database.GetDatabaseByID(id)
}

func (a *App) AddDatabase(db database.DatabaseConnection) error {
	return database.CreateDatabase(&db)
}

func (a *App) UpdateDatabase(db database.DatabaseConnection) error {
	return database.UpdateDatabase(&db)
}

func (a *App) DeleteDatabase(id int64) error {
	return database.DeleteDatabase(id)
}

func (a *App) TestDatabaseConnection(db database.DatabaseConnection) error {
	return a.backupService.TestDatabaseConnection(db)
}

// ==================== STORAGE TARGETS ====================

func (a *App) GetStorageTargets() ([]database.StorageTarget, error) {
	return database.GetAllStorageTargets()
}

func (a *App) GetStorageTarget(id int64) (*database.StorageTarget, error) {
	return database.GetStorageTargetByID(id)
}

func (a *App) AddStorageTarget(st database.StorageTarget) (int64, error) {
	if err := database.CreateStorageTarget(&st); err != nil {
		return 0, err
	}
	return st.ID, nil
}

func (a *App) UpdateStorageTarget(st database.StorageTarget) error {
	return database.UpdateStorageTarget(&st)
}

func (a *App) DeleteStorageTarget(id int64) error {
	return database.DeleteStorageTarget(id)
}

func (a *App) TestStorageConnection(st database.StorageTarget) error {
	return a.backupService.TestStorageConnection(st)
}

// ==================== BACKUP JOBS ====================

func (a *App) GetBackupJobs() ([]database.BackupJob, error) {
	return database.GetAllBackupJobs()
}

func (a *App) GetBackupJob(id int64) (*database.BackupJob, error) {
	return database.GetBackupJobByID(id)
}

func (a *App) AddBackupJob(job database.BackupJob) error {
	if err := database.CreateBackupJob(&job); err != nil {
		return err
	}

	if a.scheduler != nil && job.IsActive {
		if err := a.scheduler.AddJob(job); err != nil {
			fmt.Println(i18n.T("errors.jobScheduleAddFailed")+":", err)
		}
	}

	return nil
}

func (a *App) UpdateBackupJob(job database.BackupJob) error {
	if err := database.UpdateBackupJob(&job); err != nil {
		return err
	}

	if a.scheduler != nil {
		if err := a.scheduler.UpdateJob(job); err != nil {
			fmt.Println(i18n.T("errors.schedulerUpdateFailed")+":", err)
		}
	}

	return nil
}

func (a *App) DeleteBackupJob(id int64) error {
	if a.scheduler != nil {
		if err := a.scheduler.RemoveJob(id); err != nil {
			fmt.Println(i18n.T("errors.jobScheduleRemoveFailed")+":", err)
		}
	}

	return database.DeleteBackupJob(id)
}

func (a *App) ToggleBackupJobActive(id int64, active bool) error {
	if err := database.ToggleBackupJobActive(id, active); err != nil {
		return err
	}

	if a.scheduler != nil {
		job, err := database.GetBackupJobByID(id)
		if err == nil {
			if err := a.scheduler.UpdateJob(*job); err != nil {
				fmt.Println(i18n.T("errors.schedulerUpdateFailed")+":", err)
			}
		}
	}

	return nil
}

func (a *App) RunBackupNow(jobID int64) error {
	eventData := map[string]interface{}{
		"job_id": jobID,
	}

	job, err := database.GetBackupJobByID(jobID)
	if err != nil {
		eventData["error"] = i18n.T("errors.jobNotFound")
		a.emitEvent("backup:failed", eventData)
		return fmt.Errorf(i18n.T("errors.jobNotFound")+": %w", err)
	}
	eventData["job_name"] = job.Name

	db, err := database.GetDatabaseByID(int64(job.DatabaseID))
	if err != nil {
		eventData["error"] = i18n.T("errors.databaseNotFound")
		a.emitEvent("backup:failed", eventData)
		return fmt.Errorf(i18n.T("errors.databaseNotFound")+": %w", err)
	}

	st, err := database.GetStorageTargetByID(int64(job.StorageID))
	if err != nil {
		eventData["error"] = i18n.T("errors.storageNotFound")
		a.emitEvent("backup:failed", eventData)
		return fmt.Errorf(i18n.T("errors.storageNotFound")+": %w", err)
	}

	a.emitEvent("backup:started", eventData)

	_, err = a.backupService.ExecuteBackup(*job, *db, *st)
	if err != nil {
		eventData["error"] = err.Error()
		a.emitEvent("backup:failed", eventData)
		return err
	}

	a.emitEvent("backup:completed", eventData)

	go a.backupService.CleanupOldBackups(*job, *st)

	return nil
}

// ==================== BACKUP HISTORY ====================

func (a *App) GetBackupHistory(filter database.HistoryFilter) ([]database.BackupHistory, error) {
	return database.GetBackupHistory(filter.JobID, filter.Status, filter.Limit)
}

func (a *App) GetDashboardStats() (*database.DashboardStats, error) {
	stats := &database.DashboardStats{}

	dbs, err := database.GetAllDatabases()
	if err != nil {
		return nil, err
	}
	stats.TotalDatabases = len(dbs)

	storages, err := database.GetAllStorageTargets()
	if err != nil {
		return nil, err
	}
	stats.TotalStorages = len(storages)

	jobs, err := database.GetAllBackupJobs()
	if err != nil {
		return nil, err
	}
	stats.TotalJobs = len(jobs)
	for _, job := range jobs {
		if job.IsActive {
			stats.ActiveJobs++
		}
	}

	total, success, failed, err := database.GetRecentBackupStats()
	if err != nil {
		return nil, err
	}
	stats.Last24hTotal = total
	stats.Last24hSuccess = success
	stats.Last24hFailed = failed

	return stats, nil
}

// ==================== SETTINGS ====================

func (a *App) GetSettings() (*database.AppSettings, error) {
	settings := &database.AppSettings{}

	if v, err := database.GetSetting("telegram_bot_token"); err == nil {
		settings.Telegram.BotToken = v
	}
	if v, err := database.GetSetting("telegram_chat_id"); err == nil {
		settings.Telegram.ChatID = v
	}
	if v, err := database.GetSetting("telegram_enabled"); err == nil {
		settings.Telegram.Enabled = v == "true"
	}

	if v, err := database.GetSetting("pg_dump_path"); err == nil {
		settings.ToolPaths.PgDump = v
	}
	if v, err := database.GetSetting("mysqldump_path"); err == nil {
		settings.ToolPaths.MysqlDump = v
	}
	if v, err := database.GetSetting("mongodump_path"); err == nil {
		settings.ToolPaths.MongoDump = v
	}

	if v, err := database.GetSetting("default_retention"); err == nil {
		fmt.Sscanf(v, "%d", &settings.DefaultRetention)
	}
	if v, err := database.GetSetting("language"); err == nil {
		settings.Language = v
	}

	return settings, nil
}

func (a *App) SaveSettings(settings database.AppSettings) error {
	if err := database.SetSetting("telegram_bot_token", settings.Telegram.BotToken); err != nil {
		return err
	}
	if err := database.SetSetting("telegram_chat_id", settings.Telegram.ChatID); err != nil {
		return err
	}
	enabled := "false"
	if settings.Telegram.Enabled {
		enabled = "true"
	}
	if err := database.SetSetting("telegram_enabled", enabled); err != nil {
		return err
	}

	if err := database.SetSetting("pg_dump_path", settings.ToolPaths.PgDump); err != nil {
		return err
	}
	if err := database.SetSetting("mysqldump_path", settings.ToolPaths.MysqlDump); err != nil {
		return err
	}
	if err := database.SetSetting("mongodump_path", settings.ToolPaths.MongoDump); err != nil {
		return err
	}

	if err := database.SetSetting("default_retention", fmt.Sprintf("%d", settings.DefaultRetention)); err != nil {
		return err
	}
	if err := database.SetSetting("language", settings.Language); err != nil {
		return err
	}

	i18n.SetLanguage(settings.Language)

	return nil
}

func (a *App) SetAppLanguage(lang string) error {
	if err := database.SetSetting("language", lang); err != nil {
		return err
	}
	i18n.SetLanguage(lang)
	return nil
}

func (a *App) GetTranslations(lang string) (map[string]interface{}, error) {
	return i18n.GetAllTranslations(lang)
}

func (a *App) TestTelegramConnection(botToken string) error {
	notifier, err := notification.NewTelegramNotifier(botToken, "", false)
	if err != nil {
		return err
	}
	return notifier.TestConnection()
}

func (a *App) SendTestTelegramMessage(botToken, chatID string) error {
	notifier, err := notification.NewTelegramNotifier(botToken, chatID, true)
	if err != nil {
		return err
	}
	return notifier.SendTestMessage()
}

// ==================== HELPER METHODS ====================

func (a *App) ParseStorageConfig(configJSON string, storageType string) (interface{}, error) {
	switch storageType {
	case "local":
		var cfg database.LocalStorageConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	case "ftp":
		var cfg database.FTPStorageConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	case "sftp":
		var cfg database.SFTPStorageConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	case "s3":
		var cfg database.S3StorageConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	default:
		return nil, fmt.Errorf(i18n.T("errors.unknownStorageType"), storageType)
	}
}

func (a *App) SelectFolder() (string, error) {
	folder, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: i18n.T("dialogs.selectBackupFolder"),
	})
	if err != nil {
		return "", err
	}
	return folder, nil
}

func (a *App) SelectFile(title string, filters []string) (string, error) {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
	if err != nil {
		return "", err
	}
	return file, nil
}

// ValidateFolder checks if a folder path exists (for web mode)
func (a *App) ValidateFolder(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", path)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}
	return nil
}

// ==================== SERVER CONFIG ====================

func (a *App) GetServerConfig() (*config.ServerConfig, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &cfg.Server, nil
}

func (a *App) SaveServerConfig(serverCfg config.ServerConfig) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.Server = serverCfg
	return config.Save(cfg)
}
