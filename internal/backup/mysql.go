package backup

import (
	"database/sql"
	"dbbackup/internal/i18n"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLProvider implements BackupProvider for MySQL databases
type MySQLProvider struct {
	toolPath string
}

// NewMySQLProvider creates a new MySQL backup provider
func NewMySQLProvider() *MySQLProvider {
	return &MySQLProvider{
		toolPath: findExecutable("mysqldump"),
	}
}

// GetToolPath returns the current tool path
func (p *MySQLProvider) GetToolPath() string {
	return p.toolPath
}

// SetToolPath sets a custom tool path
func (p *MySQLProvider) SetToolPath(path string) {
	if path != "" {
		p.toolPath = path
	}
}

// TestConnection tests the database connection using Go native driver
func (p *MySQLProvider) TestConnection(config BackupConfig) error {
	tlsParam := "false"
	if config.SSLEnabled {
		tlsParam = "true"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=10s&tls=%s",
		config.Username, config.Password, config.Host, config.Port, config.DatabaseName, tlsParam)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	return nil
}

// Backup performs a MySQL database backup using mysqldump
func (p *MySQLProvider) Backup(config BackupConfig, outputDir string) (*BackupResult, error) {
	if _, err := os.Stat(p.toolPath); err != nil {
		return nil, fmt.Errorf(i18n.T("errors.mysqldumpNotFound"))
	}

	startTime := time.Now()

	// Ensure output directory exists
	if err := ensureDir(outputDir); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate output filename
	name := config.CustomPrefix
	if name == "" {
		name = config.DatabaseName
	}
	fileName := generateBackupFileName(name, ".sql")
	outputPath := filepath.Join(outputDir, fileName)

	// Build mysqldump arguments
	args := []string{
		"-h", config.Host,
		"-P", fmt.Sprintf("%d", config.Port),
		"-u", config.Username,
		"--protocol=tcp",
		"--single-transaction",
		"--routines",
		"--triggers",
		"--result-file=" + outputPath,
		config.DatabaseName,
	}

	if config.SSLEnabled {
		args = append(args, "--ssl-mode=REQUIRED")
	}

	// Create command
	cmd := exec.Command(p.toolPath, args...)
	prepareBundledCmd(cmd)

	// Sifreyi MYSQL_PWD env ile gecir (komut satirinda gozukmez, prompt acmaz)
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, "MYSQL_PWD="+config.Password)

	// Run mysqldump
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up partial file on error
		os.Remove(outputPath)
		return nil, fmt.Errorf("mysqldump failed: %s - %v", string(output), err)
	}

	// Get file size
	fileSize, err := getFileSize(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}

	finishTime := time.Now()

	return &BackupResult{
		FilePath:   outputPath,
		FileName:   fileName,
		FileSize:   fileSize,
		Duration:   finishTime.Sub(startTime),
		StartedAt:  startTime,
		FinishedAt: finishTime,
	}, nil
}

