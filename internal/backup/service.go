package backup

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"dbbackup/internal/crypto"
	"dbbackup/internal/database"
	"dbbackup/internal/notification"
	"dbbackup/internal/storage"
)

// ProgressCallback is called to report backup progress stages
type ProgressCallback func(eventName string, data map[string]interface{})

// BackupService handles backup operations
type BackupService struct {
	repo       *database.Repository
	onProgress ProgressCallback
}

// NewBackupService creates a new backup service
func NewBackupService(repo *database.Repository) *BackupService {
	return &BackupService{repo: repo}
}

// SetProgressCallback sets the callback for progress events
func (s *BackupService) SetProgressCallback(fn ProgressCallback) {
	s.onProgress = fn
}

func (s *BackupService) emitProgress(jobID int64, jobName string, stage string) {
	if s.onProgress != nil {
		s.onProgress("backup:progress", map[string]interface{}{
			"job_id":   jobID,
			"job_name": jobName,
			"stage":    stage,
		})
	}
}

// serviceResult contains the result of a backup operation (internal)
type serviceResult struct {
	FileName    string
	FileSize    int64
	StoragePath string
}

// ExecuteBackup performs a complete backup operation
func (s *BackupService) ExecuteBackup(job database.BackupJob, db database.DatabaseConnection, st database.StorageTarget) (*database.BackupHistory, error) {
	startTime := time.Now()

	// Create backup history record
	history := &database.BackupHistory{
		JobID:     job.ID,
		StartedAt: startTime,
		Status:    "running",
	}

	// Insert initial history record
	historyID, err := s.repo.CreateBackupHistory(*history)
	if err != nil {
		return nil, fmt.Errorf("failed to create history record: %w", err)
	}
	history.ID = historyID

	// Perform the backup
	result, err := s.performBackup(job, db, st)
	if err != nil {
		// Update history with failure
		history.Status = "failed"
		history.ErrorMessage = err.Error()
		now := time.Now()
		history.CompletedAt = &now
		s.repo.UpdateBackupHistory(*history)

		// Telegram bildirimi gonder (async)
		go notification.NotifyBackupResult(job, *history)

		return history, err
	}

	// Update history with success
	history.Status = "success"
	history.FileName = result.FileName
	history.FileSize = result.FileSize
	history.StoragePath = result.StoragePath
	now := time.Now()
	history.CompletedAt = &now
	s.repo.UpdateBackupHistory(*history)

	// Telegram bildirimi gonder (async)
	go notification.NotifyBackupResult(job, *history)

	return history, nil
}

// performBackup performs the actual backup operation
func (s *BackupService) performBackup(job database.BackupJob, db database.DatabaseConnection, st database.StorageTarget) (*serviceResult, error) {
	// Get backup provider
	provider, err := GetProvider(db.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup provider: %w", err)
	}

	// Ayarlar'dan kullanici tanimli tool path'i oku ve uygula
	settingKey := map[string]string{"postgres": "pg_dump_path", "mysql": "mysqldump_path", "mongodb": "mongodump_path"}
	if key, ok := settingKey[db.Type]; ok {
		if customPath, err := database.GetSetting(key); err == nil && customPath != "" {
			provider.SetToolPath(customPath)
		}
	}

	// Create temp directory for backup
	tempDir := filepath.Join(os.TempDir(), "dbbackup", fmt.Sprintf("job_%d_%d", job.ID, time.Now().UnixNano()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Cleanup temp directory

	// Perform database backup
	backupConfig := BackupConfig{
		Type:         db.Type,
		Host:         db.Host,
		Port:         db.Port,
		Username:     db.Username,
		Password:     db.Password,
		DatabaseName: db.DatabaseName,
		SSLEnabled:   db.SSLEnabled,
		AuthDatabase: db.AuthDatabase,
		CustomPrefix: job.CustomPrefix,
	}

	s.emitProgress(job.ID, job.Name, "dumping")

	backupRes, err := provider.Backup(backupConfig, tempDir)
	if err != nil {
		return nil, fmt.Errorf("backup failed: %w", err)
	}

	// Apply compression and encryption
	finalPath := backupRes.FilePath
	if job.Compression != "none" || job.Encryption {
		s.emitProgress(job.ID, job.Name, "processing")
		finalPath, err = crypto.ProcessBackup(backupRes.FilePath, job.Compression, job.Encryption, job.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("post-processing failed: %w", err)
		}
	}

	// Get final file info
	fileInfo, err := os.Stat(finalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Get storage provider
	storageProvider, err := storage.GetProvider(st.Type, st.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage provider: %w", err)
	}

	// Generate remote path (path.Join kullan - her zaman "/" ayirici, FTP/SFTP/S3 uyumlu)
	now := time.Now()
	fileName := filepath.Base(finalPath)

	// Tarih bazli alt klasor
	var dateFolder string
	switch job.FolderGrouping {
	case "daily":
		dateFolder = now.Format("2006-01-02")
	case "monthly":
		dateFolder = now.Format("2006-01")
	case "yearly":
		dateFolder = now.Format("2006")
	}

	// Remote path olustur: custom_folder + tarih klasoru + dosya adi
	var remotePath string
	switch {
	case job.CustomFolder != "" && dateFolder != "":
		remotePath = path.Join(job.CustomFolder, dateFolder, fileName)
	case job.CustomFolder != "":
		remotePath = path.Join(job.CustomFolder, fileName)
	case dateFolder != "":
		remotePath = path.Join(dateFolder, fileName)
	default:
		remotePath = fileName
	}

	s.emitProgress(job.ID, job.Name, "uploading")

	// Upload to storage
	if err := storageProvider.Upload(finalPath, remotePath); err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	return &serviceResult{
		FileName:    filepath.Base(finalPath),
		FileSize:    fileInfo.Size(),
		StoragePath: remotePath,
	}, nil
}

// TestDatabaseConnection tests a database connection
func (s *BackupService) TestDatabaseConnection(db database.DatabaseConnection) error {
	provider, err := GetProvider(db.Type)
	if err != nil {
		return err
	}

	config := BackupConfig{
		Type:         db.Type,
		Host:         db.Host,
		Port:         db.Port,
		Username:     db.Username,
		Password:     db.Password,
		DatabaseName: db.DatabaseName,
		SSLEnabled:   db.SSLEnabled,
		AuthDatabase: db.AuthDatabase,
	}

	return provider.TestConnection(config)
}

// TestStorageConnection tests a storage connection
func (s *BackupService) TestStorageConnection(st database.StorageTarget) error {
	provider, err := storage.GetProvider(st.Type, st.Config)
	if err != nil {
		return err
	}

	return provider.TestConnection()
}

// CleanupOldBackups removes backups older than retention days
func (s *BackupService) CleanupOldBackups(job database.BackupJob, st database.StorageTarget) error {
	if job.RetentionDays <= 0 {
		return nil
	}

	storageProvider, err := storage.GetProvider(st.Type, st.Config)
	if err != nil {
		return fmt.Errorf("failed to get storage provider: %w", err)
	}

	// List files in custom folder or root
	listPrefix := job.CustomFolder
	files, err := storageProvider.List(listPrefix)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	// If folder grouping is active, also scan subdirectories (1 level deep)
	var allFiles []storage.StorageFile
	for _, file := range files {
		if file.IsDir {
			subFiles, err := storageProvider.List(path.Join(listPrefix, file.Name))
			if err == nil {
				allFiles = append(allFiles, subFiles...)
			}
		} else {
			allFiles = append(allFiles, file)
		}
	}

	// Calculate cutoff date
	cutoff := time.Now().AddDate(0, 0, -job.RetentionDays)

	// Delete old files
	for _, file := range allFiles {
		if !file.IsDir && file.ModifiedAt.Before(cutoff) {
			if err := storageProvider.Delete(file.Path); err != nil {
				// Log error but continue with other files
				fmt.Printf("Failed to delete old backup %s: %v\n", file.Path, err)
			}
		}
	}

	return nil
}
