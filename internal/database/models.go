package database

import "time"

// DatabaseConnection - Veritabani baglanti bilgileri
type DatabaseConnection struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"` // postgres, mysql, mongodb
	Host         string    `json:"host"`
	Port         int       `json:"port"`
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	DatabaseName string    `json:"database_name"`
	SSLEnabled   bool      `json:"ssl_enabled"`
	AuthDatabase string    `json:"auth_database"` // MongoDB authSource (varsayilan: admin)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StorageTarget - Yedekleme hedefi (local, ftp, sftp, s3)
type StorageTarget struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // local, ftp, sftp, s3
	Config    string    `json:"config"` // JSON olarak ayarlar
	CreatedAt time.Time `json:"created_at"`
}

// LocalStorageConfig - Local depolama ayarlari
type LocalStorageConfig struct {
	Path string `json:"path"`
}

// FTPStorageConfig - FTP depolama ayarlari
type FTPStorageConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Path     string `json:"path"`
}

// SFTPStorageConfig - SFTP depolama ayarlari
type SFTPStorageConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	Path       string `json:"path"`
}

// S3StorageConfig - S3 depolama ayarlari
type S3StorageConfig struct {
	Endpoint        string `json:"endpoint"` // MinIO icin custom endpoint
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Path            string `json:"path"`
	UseSSL          bool   `json:"use_ssl"`
}

// BackupJob - Yedekleme gorevi
type BackupJob struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	DatabaseID     int64     `json:"database_id"`
	StorageID      int64     `json:"storage_id"`
	ScheduleType   string    `json:"schedule_type"` // manual, interval, daily, weekly
	ScheduleConfig string    `json:"schedule_config"` // JSON
	Compression    string    `json:"compression"` // none, gzip, zip
	Encryption     bool      `json:"encryption"`
	EncryptionKey  string    `json:"encryption_key,omitempty"`
	RetentionDays  int       `json:"retention_days"`
	IsActive       bool      `json:"is_active"`
	CustomPrefix   string    `json:"custom_prefix"`   // Dosya adi prefix (bos ise db adi kullanilir)
	CustomFolder   string    `json:"custom_folder"`   // Alt klasor yolu (orn: dbyedekleri/testyedekler)
	FolderGrouping string    `json:"folder_grouping"` // Tarih bazli klasorleme: "", "daily", "monthly", "yearly"
	CreatedAt      time.Time `json:"created_at"`
}

// ScheduleConfig - Zamanlama ayarlari
type ScheduleConfig struct {
	// interval icin
	IntervalMinutes int `json:"interval_minutes,omitempty"`

	// daily icin
	Hour   int `json:"hour,omitempty"`
	Minute int `json:"minute,omitempty"`

	// weekly icin
	Weekdays []int `json:"weekdays,omitempty"` // 0=Pazar, 1=Pazartesi, ...
}

// BackupHistory - Yedekleme gecmisi
type BackupHistory struct {
	ID               int64      `json:"id"`
	JobID            int64      `json:"job_id"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	Status           string     `json:"status"` // running, success, failed
	FileName         string     `json:"file_name"`
	FileSize         int64      `json:"file_size"`
	StoragePath      string     `json:"storage_path"`
	ErrorMessage     string     `json:"error_message,omitempty"`
	NotificationSent bool       `json:"notification_sent"`
}

// Settings - Uygulama ayarlari
type Settings struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TelegramSettings - Telegram bildirim ayarlari
type TelegramSettings struct {
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
	Enabled  bool   `json:"enabled"`
}

// ToolPaths - Veritabani araclari yollari
type ToolPaths struct {
	PgDump    string `json:"pg_dump"`
	MysqlDump string `json:"mysqldump"`
	MongoDump string `json:"mongodump"`
}

// AppSettings - Tum uygulama ayarlari
type AppSettings struct {
	Telegram         TelegramSettings `json:"telegram"`
	ToolPaths        ToolPaths        `json:"tool_paths"`
	DefaultRetention int              `json:"default_retention"`
	Language         string           `json:"language"`
}

// HistoryFilter - Gecmis filtreleme parametreleri
type HistoryFilter struct {
	JobID  int64  `json:"job_id"`
	Status string `json:"status"`
	Limit  int    `json:"limit"`
}

// DashboardStats - Dashboard istatistikleri
type DashboardStats struct {
	TotalDatabases int `json:"total_databases"`
	TotalStorages  int `json:"total_storages"`
	TotalJobs      int `json:"total_jobs"`
	ActiveJobs     int `json:"active_jobs"`
	Last24hTotal   int `json:"last_24h_total"`
	Last24hSuccess int `json:"last_24h_success"`
	Last24hFailed  int `json:"last_24h_failed"`
}
