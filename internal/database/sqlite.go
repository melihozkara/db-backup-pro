package database

import (
	"database/sql"
	"dbbackup/internal/i18n"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mutecomm/go-sqlcipher/v4"
)

// DB - Global veritabani baglantisi
var DB *sql.DB

// InitDB - Veritabani baglantisini baslatir ve tablolari olusturur
func InitDB() error {
	// Uygulama veri dizinini al
	dataDir, err := getDataDir()
	if err != nil {
		return fmt.Errorf(i18n.T("errors.dataDirFailed")+": %w", err)
	}

	// Dizin yoksa olustur
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf(i18n.T("errors.dataDirCreateFailed")+": %w", err)
	}

	dbPath := filepath.Join(dataDir, "dbbackup.db")
	encryptedDBPath := filepath.Join(dataDir, "dbbackup_encrypted.db")

	// Sifreleme anahtarini al veya olustur (OS Keychain)
	encryptionKey, err := GetOrCreateEncryptionKey()
	if err != nil {
		return fmt.Errorf(i18n.T("errors.encryptionKeyFailed")+": %w", err)
	}

	// Eski sifresiz DB varsa ve sifreli DB yoksa → migrate et
	if fileExists(dbPath) && !fileExists(encryptedDBPath) {
		if isUnencryptedSQLite(dbPath) {
			fmt.Println(i18n.T("errors.migrationDetected"))
			if err := migrateToEncrypted(dbPath, encryptedDBPath, encryptionKey); err != nil {
				return fmt.Errorf(i18n.T("errors.migrationFailed")+": %w", err)
			}
			// Eski sifresiz DB'yi sil
			os.Remove(dbPath)
			fmt.Println(i18n.T("errors.migrationComplete"))
		}
	}

	// Sifreli DB'yi ac
	db, err := openEncryptedDB(encryptedDBPath, encryptionKey)
	if err != nil {
		return fmt.Errorf(i18n.T("errors.encryptedDbOpenFailed")+": %w", err)
	}

	DB = db

	// Tablolari olustur
	if err := createTables(); err != nil {
		return fmt.Errorf(i18n.T("errors.tablesCreateFailed")+": %w", err)
	}

	return nil
}

// openEncryptedDB - Sifreli veritabanini acar
func openEncryptedDB(dbPath, key string) (*sql.DB, error) {
	// go-sqlcipher DSN formatinda anahtar gecilir — her connection otomatik sifrelenir
	dsn := fmt.Sprintf("%s?_pragma_key=x'%s'&_pragma_cipher_page_size=4096&_pragma_foreign_keys=on", dbPath, key)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf(i18n.T("errors.dbOpenFailed")+": %w", err)
	}

	// Sifrelemenin calistigini dogrula
	var result string
	if err := db.QueryRow("SELECT count(*) FROM sqlite_master").Scan(&result); err != nil {
		db.Close()
		return nil, fmt.Errorf(i18n.T("errors.encryptionVerifyFailed")+": %w", err)
	}

	return db, nil
}

// migrateToEncrypted - Sifresiz DB'yi sifreli DB'ye tasir
func migrateToEncrypted(oldPath, newPath, key string) error {
	// Eski sifresiz DB'yi ac
	oldDB, err := sql.Open("sqlite3", oldPath)
	if err != nil {
		return fmt.Errorf(i18n.T("errors.oldDbOpenFailed")+": %w", err)
	}
	defer oldDB.Close()

	// Tum islemler ayni connection uzerinde calismali
	oldDB.SetMaxOpenConns(1)

	// Sifresiz DB oldugunu belirt (bos key = sifresiz)
	if _, err := oldDB.Exec(`PRAGMA key = ''`); err != nil {
		return fmt.Errorf(i18n.T("errors.oldDbPragmaFailed")+": %w", err)
	}

	// Eski DB'nin okunabilir oldugunu dogrula
	if _, err := oldDB.Exec(`SELECT count(*) FROM sqlite_master`); err != nil {
		return fmt.Errorf(i18n.T("errors.oldDbReadFailed")+": %w", err)
	}

	// sqlcipher_export ile sifreli kopyasini olustur
	attachSQL := fmt.Sprintf(`ATTACH DATABASE '%s' AS encrypted KEY "x'%s'"`, newPath, key)
	if _, err := oldDB.Exec(attachSQL); err != nil {
		return fmt.Errorf(i18n.T("errors.encryptedDbAttachFailed")+": %w", err)
	}

	// Sifreli DB icin cipher_page_size ayarla
	if _, err := oldDB.Exec(`PRAGMA encrypted.cipher_page_size = 4096`); err != nil {
		return fmt.Errorf(i18n.T("errors.cipherPagesizeFailed")+": %w", err)
	}

	if _, err := oldDB.Exec(`SELECT sqlcipher_export('encrypted')`); err != nil {
		return fmt.Errorf(i18n.T("errors.dataMigrationFailed")+": %w", err)
	}

	if _, err := oldDB.Exec(`DETACH DATABASE encrypted`); err != nil {
		return fmt.Errorf(i18n.T("errors.encryptedDbDetachFailed")+": %w", err)
	}

	return nil
}

// isUnencryptedSQLite - Dosyanin sifresiz SQLite dosyasi olup olmadigini kontrol eder
func isUnencryptedSQLite(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// SQLite dosyalari "SQLite format 3\000" ile baslar (16 byte)
	header := make([]byte, 16)
	n, err := f.Read(header)
	if err != nil || n < 16 {
		return false
	}

	return string(header) == "SQLite format 3\000"
}

// fileExists - Dosyanin var olup olmadigini kontrol eder
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// getDataDir - Isletim sistemine gore veri dizinini dondurur
func getDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Platform bagimsiz dizin
	return filepath.Join(homeDir, ".dbbackup"), nil
}

// createTables - Veritabani tablolarini olusturur
func createTables() error {
	queries := []string{
		// database_connections tablosu
		`CREATE TABLE IF NOT EXISTS database_connections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('postgres', 'mysql', 'mongodb')),
			host TEXT NOT NULL,
			port INTEGER NOT NULL,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			database_name TEXT NOT NULL,
			ssl_enabled INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// storage_targets tablosu
		`CREATE TABLE IF NOT EXISTS storage_targets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('local', 'ftp', 'sftp', 's3')),
			config TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// backup_jobs tablosu
		`CREATE TABLE IF NOT EXISTS backup_jobs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			database_id INTEGER NOT NULL,
			storage_id INTEGER NOT NULL,
			schedule_type TEXT NOT NULL CHECK(schedule_type IN ('manual', 'interval', 'daily', 'weekly')),
			schedule_config TEXT,
			compression TEXT DEFAULT 'gzip' CHECK(compression IN ('none', 'gzip', 'zip')),
			encryption INTEGER DEFAULT 0,
			encryption_key TEXT,
			retention_days INTEGER DEFAULT 7,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (database_id) REFERENCES database_connections(id) ON DELETE CASCADE,
			FOREIGN KEY (storage_id) REFERENCES storage_targets(id) ON DELETE CASCADE
		)`,

		// backup_history tablosu
		`CREATE TABLE IF NOT EXISTS backup_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			job_id INTEGER NOT NULL,
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			status TEXT NOT NULL DEFAULT 'running' CHECK(status IN ('running', 'success', 'failed')),
			file_name TEXT,
			file_size INTEGER DEFAULT 0,
			storage_path TEXT,
			error_message TEXT,
			notification_sent INTEGER DEFAULT 0,
			FOREIGN KEY (job_id) REFERENCES backup_jobs(id) ON DELETE CASCADE
		)`,

		// settings tablosu
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,

		// Index'ler
		`CREATE INDEX IF NOT EXISTS idx_backup_history_job_id ON backup_history(job_id)`,
		`CREATE INDEX IF NOT EXISTS idx_backup_history_status ON backup_history(status)`,
		`CREATE INDEX IF NOT EXISTS idx_backup_jobs_database_id ON backup_jobs(database_id)`,
		`CREATE INDEX IF NOT EXISTS idx_backup_jobs_storage_id ON backup_jobs(storage_id)`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return fmt.Errorf(i18n.T("errors.queryFailed")+": %s, hata: %w", query, err)
		}
	}

	// Migration: backup_jobs tablosuna yeni kolonlar ekle (varsa hata verir, yoksay)
	migrations := []string{
		`ALTER TABLE backup_jobs ADD COLUMN custom_prefix TEXT DEFAULT ''`,
		`ALTER TABLE backup_jobs ADD COLUMN custom_folder TEXT DEFAULT ''`,
		`ALTER TABLE backup_jobs ADD COLUMN folder_grouping TEXT DEFAULT ''`,
		`ALTER TABLE database_connections ADD COLUMN auth_database TEXT DEFAULT ''`,
	}
	for _, m := range migrations {
		DB.Exec(m) // Kolon zaten varsa hata doner, yoksay
	}

	// Varsayilan ayarlari ekle
	if err := insertDefaultSettings(); err != nil {
		return err
	}

	return nil
}

// insertDefaultSettings - Varsayilan ayarlari ekler
func insertDefaultSettings() error {
	defaults := map[string]string{
		"telegram_bot_token": "",
		"telegram_chat_id":   "",
		"telegram_enabled":   "false",
		"pg_dump_path":       "",
		"mysqldump_path":     "",
		"mongodump_path":     "",
		"default_retention":  "7",
		"language":           "tr",
	}

	for key, value := range defaults {
		_, err := DB.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)`, key, value)
		if err != nil {
			return fmt.Errorf(i18n.T("errors.defaultSettingsFailed")+": %w", err)
		}
	}

	return nil
}

// CloseDB - Veritabani baglantisini kapatir
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

// GetDB - Veritabani baglantisini dondurur
func GetDB() *sql.DB {
	return DB
}

// GetSetting - Ayar degerini getirir
func GetSetting(key string) (string, error) {
	var value string
	err := DB.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetSetting - Ayar degerini kaydeder
func SetSetting(key, value string) error {
	_, err := DB.Exec(`INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)`, key, value)
	return err
}
