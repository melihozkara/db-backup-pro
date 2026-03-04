package database

import (
	"database/sql"
	"fmt"
	"time"
)

// Repository provides database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateBackupHistory creates a new backup history record and returns the ID
func (r *Repository) CreateBackupHistory(h BackupHistory) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO backup_history (job_id, started_at, status, file_name, storage_path)
		VALUES (?, ?, ?, ?, ?)
	`, h.JobID, h.StartedAt, h.Status, h.FileName, h.StoragePath)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateBackupHistory updates a backup history record
func (r *Repository) UpdateBackupHistory(h BackupHistory) error {
	notificationSent := 0
	if h.NotificationSent {
		notificationSent = 1
	}

	_, err := r.db.Exec(`
		UPDATE backup_history
		SET completed_at = ?, status = ?, file_name = ?, file_size = ?, storage_path = ?, error_message = ?, notification_sent = ?
		WHERE id = ?
	`, h.CompletedAt, h.Status, h.FileName, h.FileSize, h.StoragePath, h.ErrorMessage, notificationSent, h.ID)
	return err
}

// ==================== DATABASE CONNECTIONS ====================

// GetAllDatabases - Tum veritabani baglantilarini getirir
func GetAllDatabases() ([]DatabaseConnection, error) {
	rows, err := DB.Query(`
		SELECT id, name, type, host, port, username, password, database_name, ssl_enabled, auth_database, created_at, updated_at
		FROM database_connections
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []DatabaseConnection
	for rows.Next() {
		var db DatabaseConnection
		var sslEnabled int
		var authDatabase sql.NullString
		err := rows.Scan(&db.ID, &db.Name, &db.Type, &db.Host, &db.Port, &db.Username, &db.Password, &db.DatabaseName, &sslEnabled, &authDatabase, &db.CreatedAt, &db.UpdatedAt)
		if err != nil {
			return nil, err
		}
		db.SSLEnabled = sslEnabled == 1
		if authDatabase.Valid {
			db.AuthDatabase = authDatabase.String
		}
		databases = append(databases, db)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return databases, nil
}

// GetDatabaseByID - ID ile veritabani baglantisini getirir
func GetDatabaseByID(id int64) (*DatabaseConnection, error) {
	var db DatabaseConnection
	var sslEnabled int
	var authDatabase sql.NullString
	err := DB.QueryRow(`
		SELECT id, name, type, host, port, username, password, database_name, ssl_enabled, auth_database, created_at, updated_at
		FROM database_connections
		WHERE id = ?
	`, id).Scan(&db.ID, &db.Name, &db.Type, &db.Host, &db.Port, &db.Username, &db.Password, &db.DatabaseName, &sslEnabled, &authDatabase, &db.CreatedAt, &db.UpdatedAt)
	if err != nil {
		return nil, err
	}
	db.SSLEnabled = sslEnabled == 1
	if authDatabase.Valid {
		db.AuthDatabase = authDatabase.String
	}
	return &db, nil
}

// CreateDatabase - Yeni veritabani baglantisi olusturur
func CreateDatabase(db *DatabaseConnection) error {
	sslEnabled := 0
	if db.SSLEnabled {
		sslEnabled = 1
	}

	result, err := DB.Exec(`
		INSERT INTO database_connections (name, type, host, port, username, password, database_name, ssl_enabled, auth_database)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, db.Name, db.Type, db.Host, db.Port, db.Username, db.Password, db.DatabaseName, sslEnabled, db.AuthDatabase)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	db.ID = id
	db.CreatedAt = time.Now()
	db.UpdatedAt = time.Now()

	return nil
}

// UpdateDatabase - Veritabani baglantisini gunceller
func UpdateDatabase(db *DatabaseConnection) error {
	sslEnabled := 0
	if db.SSLEnabled {
		sslEnabled = 1
	}

	_, err := DB.Exec(`
		UPDATE database_connections
		SET name = ?, type = ?, host = ?, port = ?, username = ?, password = ?, database_name = ?, ssl_enabled = ?, auth_database = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, db.Name, db.Type, db.Host, db.Port, db.Username, db.Password, db.DatabaseName, sslEnabled, db.AuthDatabase, db.ID)
	return err
}

// DeleteDatabase - Veritabani baglantisini siler
func DeleteDatabase(id int64) error {
	_, err := DB.Exec(`DELETE FROM database_connections WHERE id = ?`, id)
	return err
}

// ==================== STORAGE TARGETS ====================

// GetAllStorageTargets - Tum depolama hedeflerini getirir
func GetAllStorageTargets() ([]StorageTarget, error) {
	rows, err := DB.Query(`
		SELECT id, name, type, config, created_at
		FROM storage_targets
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []StorageTarget
	for rows.Next() {
		var st StorageTarget
		err := rows.Scan(&st.ID, &st.Name, &st.Type, &st.Config, &st.CreatedAt)
		if err != nil {
			return nil, err
		}
		targets = append(targets, st)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}

// GetStorageTargetByID - ID ile depolama hedefini getirir
func GetStorageTargetByID(id int64) (*StorageTarget, error) {
	var st StorageTarget
	err := DB.QueryRow(`
		SELECT id, name, type, config, created_at
		FROM storage_targets
		WHERE id = ?
	`, id).Scan(&st.ID, &st.Name, &st.Type, &st.Config, &st.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// CreateStorageTarget - Yeni depolama hedefi olusturur
func CreateStorageTarget(st *StorageTarget) error {
	result, err := DB.Exec(`
		INSERT INTO storage_targets (name, type, config)
		VALUES (?, ?, ?)
	`, st.Name, st.Type, st.Config)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	st.ID = id
	st.CreatedAt = time.Now()

	return nil
}

// UpdateStorageTarget - Depolama hedefini gunceller
func UpdateStorageTarget(st *StorageTarget) error {
	_, err := DB.Exec(`
		UPDATE storage_targets
		SET name = ?, type = ?, config = ?
		WHERE id = ?
	`, st.Name, st.Type, st.Config, st.ID)
	return err
}

// DeleteStorageTarget - Depolama hedefini siler
func DeleteStorageTarget(id int64) error {
	_, err := DB.Exec(`DELETE FROM storage_targets WHERE id = ?`, id)
	return err
}

// ==================== BACKUP JOBS ====================

// GetAllBackupJobs - Tum yedekleme gorevlerini getirir
func GetAllBackupJobs() ([]BackupJob, error) {
	rows, err := DB.Query(`
		SELECT id, name, database_id, storage_id, schedule_type, schedule_config, compression, encryption, encryption_key, retention_days, is_active, custom_prefix, custom_folder, folder_grouping, created_at
		FROM backup_jobs
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []BackupJob
	for rows.Next() {
		var job BackupJob
		var encryption, isActive int
		var scheduleConfig, encryptionKey, customPrefix, customFolder, folderGrouping sql.NullString
		err := rows.Scan(&job.ID, &job.Name, &job.DatabaseID, &job.StorageID, &job.ScheduleType, &scheduleConfig, &job.Compression, &encryption, &encryptionKey, &job.RetentionDays, &isActive, &customPrefix, &customFolder, &folderGrouping, &job.CreatedAt)
		if err != nil {
			return nil, err
		}
		job.Encryption = encryption == 1
		job.IsActive = isActive == 1
		if scheduleConfig.Valid {
			job.ScheduleConfig = scheduleConfig.String
		}
		if encryptionKey.Valid {
			job.EncryptionKey = encryptionKey.String
		}
		if customPrefix.Valid {
			job.CustomPrefix = customPrefix.String
		}
		if customFolder.Valid {
			job.CustomFolder = customFolder.String
		}
		if folderGrouping.Valid {
			job.FolderGrouping = folderGrouping.String
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

// GetActiveBackupJobs - Aktif yedekleme gorevlerini getirir
func GetActiveBackupJobs() ([]BackupJob, error) {
	rows, err := DB.Query(`
		SELECT id, name, database_id, storage_id, schedule_type, schedule_config, compression, encryption, encryption_key, retention_days, is_active, custom_prefix, custom_folder, folder_grouping, created_at
		FROM backup_jobs
		WHERE is_active = 1
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []BackupJob
	for rows.Next() {
		var job BackupJob
		var encryption, isActive int
		var scheduleConfig, encryptionKey, customPrefix, customFolder, folderGrouping sql.NullString
		err := rows.Scan(&job.ID, &job.Name, &job.DatabaseID, &job.StorageID, &job.ScheduleType, &scheduleConfig, &job.Compression, &encryption, &encryptionKey, &job.RetentionDays, &isActive, &customPrefix, &customFolder, &folderGrouping, &job.CreatedAt)
		if err != nil {
			return nil, err
		}
		job.Encryption = encryption == 1
		job.IsActive = isActive == 1
		if scheduleConfig.Valid {
			job.ScheduleConfig = scheduleConfig.String
		}
		if encryptionKey.Valid {
			job.EncryptionKey = encryptionKey.String
		}
		if customPrefix.Valid {
			job.CustomPrefix = customPrefix.String
		}
		if customFolder.Valid {
			job.CustomFolder = customFolder.String
		}
		if folderGrouping.Valid {
			job.FolderGrouping = folderGrouping.String
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

// GetBackupJobByID - ID ile yedekleme gorevini getirir
func GetBackupJobByID(id int64) (*BackupJob, error) {
	var job BackupJob
	var encryption, isActive int
	var scheduleConfig, encryptionKey, customPrefix, customFolder, folderGrouping sql.NullString
	err := DB.QueryRow(`
		SELECT id, name, database_id, storage_id, schedule_type, schedule_config, compression, encryption, encryption_key, retention_days, is_active, custom_prefix, custom_folder, folder_grouping, created_at
		FROM backup_jobs
		WHERE id = ?
	`, id).Scan(&job.ID, &job.Name, &job.DatabaseID, &job.StorageID, &job.ScheduleType, &scheduleConfig, &job.Compression, &encryption, &encryptionKey, &job.RetentionDays, &isActive, &customPrefix, &customFolder, &folderGrouping, &job.CreatedAt)
	if err != nil {
		return nil, err
	}
	job.Encryption = encryption == 1
	job.IsActive = isActive == 1
	if scheduleConfig.Valid {
		job.ScheduleConfig = scheduleConfig.String
	}
	if encryptionKey.Valid {
		job.EncryptionKey = encryptionKey.String
	}
	if customPrefix.Valid {
		job.CustomPrefix = customPrefix.String
	}
	if customFolder.Valid {
		job.CustomFolder = customFolder.String
	}
	if folderGrouping.Valid {
		job.FolderGrouping = folderGrouping.String
	}
	return &job, nil
}

// CreateBackupJob - Yeni yedekleme gorevi olusturur
func CreateBackupJob(job *BackupJob) error {
	encryption := 0
	if job.Encryption {
		encryption = 1
	}
	isActive := 0
	if job.IsActive {
		isActive = 1
	}

	result, err := DB.Exec(`
		INSERT INTO backup_jobs (name, database_id, storage_id, schedule_type, schedule_config, compression, encryption, encryption_key, retention_days, is_active, custom_prefix, custom_folder, folder_grouping)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, job.Name, job.DatabaseID, job.StorageID, job.ScheduleType, job.ScheduleConfig, job.Compression, encryption, job.EncryptionKey, job.RetentionDays, isActive, job.CustomPrefix, job.CustomFolder, job.FolderGrouping)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	job.ID = id
	job.CreatedAt = time.Now()

	return nil
}

// UpdateBackupJob - Yedekleme gorevini gunceller
func UpdateBackupJob(job *BackupJob) error {
	encryption := 0
	if job.Encryption {
		encryption = 1
	}
	isActive := 0
	if job.IsActive {
		isActive = 1
	}

	_, err := DB.Exec(`
		UPDATE backup_jobs
		SET name = ?, database_id = ?, storage_id = ?, schedule_type = ?, schedule_config = ?, compression = ?, encryption = ?, encryption_key = ?, retention_days = ?, is_active = ?, custom_prefix = ?, custom_folder = ?, folder_grouping = ?
		WHERE id = ?
	`, job.Name, job.DatabaseID, job.StorageID, job.ScheduleType, job.ScheduleConfig, job.Compression, encryption, job.EncryptionKey, job.RetentionDays, isActive, job.CustomPrefix, job.CustomFolder, job.FolderGrouping, job.ID)
	return err
}

// ToggleBackupJobActive - Yedekleme gorevinin aktiflik durumunu degistirir
func ToggleBackupJobActive(id int64, active bool) error {
	isActive := 0
	if active {
		isActive = 1
	}
	_, err := DB.Exec(`UPDATE backup_jobs SET is_active = ? WHERE id = ?`, isActive, id)
	return err
}

// DeleteBackupJob - Yedekleme gorevini siler
func DeleteBackupJob(id int64) error {
	_, err := DB.Exec(`DELETE FROM backup_jobs WHERE id = ?`, id)
	return err
}

// ==================== BACKUP HISTORY ====================

// GetBackupHistory - Yedekleme gecmisini getirir (filtreleme ile)
func GetBackupHistory(jobID int64, status string, limit int) ([]BackupHistory, error) {
	query := `
		SELECT id, job_id, started_at, completed_at, status, file_name, file_size, storage_path, error_message, notification_sent
		FROM backup_history
		WHERE 1=1
	`
	args := []interface{}{}

	if jobID > 0 {
		query += ` AND job_id = ?`
		args = append(args, jobID)
	}
	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}

	query += ` ORDER BY started_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []BackupHistory
	for rows.Next() {
		var h BackupHistory
		var completedAt sql.NullTime
		var notificationSent int
		err := rows.Scan(&h.ID, &h.JobID, &h.StartedAt, &completedAt, &h.Status, &h.FileName, &h.FileSize, &h.StoragePath, &h.ErrorMessage, &notificationSent)
		if err != nil {
			return nil, err
		}
		if completedAt.Valid {
			h.CompletedAt = &completedAt.Time
		}
		h.NotificationSent = notificationSent == 1
		history = append(history, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

// CreateBackupHistory - Yeni yedekleme gecmisi olusturur
func CreateBackupHistory(h *BackupHistory) error {
	result, err := DB.Exec(`
		INSERT INTO backup_history (job_id, status, file_name, storage_path)
		VALUES (?, ?, ?, ?)
	`, h.JobID, h.Status, h.FileName, h.StoragePath)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	h.ID = id
	h.StartedAt = time.Now()

	return nil
}

// UpdateBackupHistory - Yedekleme gecmisini gunceller
func UpdateBackupHistory(h *BackupHistory) error {
	notificationSent := 0
	if h.NotificationSent {
		notificationSent = 1
	}

	_, err := DB.Exec(`
		UPDATE backup_history
		SET completed_at = ?, status = ?, file_name = ?, file_size = ?, storage_path = ?, error_message = ?, notification_sent = ?
		WHERE id = ?
	`, h.CompletedAt, h.Status, h.FileName, h.FileSize, h.StoragePath, h.ErrorMessage, notificationSent, h.ID)
	return err
}

// GetRecentBackupStats - Son 24 saat icin yedekleme istatistiklerini getirir
func GetRecentBackupStats() (total int, success int, failed int, err error) {
	err = DB.QueryRow(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
		FROM backup_history
		WHERE started_at >= datetime('now', '-24 hours')
	`).Scan(&total, &success, &failed)
	return
}

// CleanOldBackupHistory - Eski yedekleme gecmisini temizler
func CleanOldBackupHistory(retentionDays int) error {
	_, err := DB.Exec(`
		DELETE FROM backup_history
		WHERE started_at < datetime('now', ? || ' days')
	`, -retentionDays)
	return err
}
