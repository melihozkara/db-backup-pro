package backup

import (
	"database/sql"
	"dbbackup/internal/i18n"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
)

// PostgresProvider implements BackupProvider for PostgreSQL databases
type PostgresProvider struct {
	toolPath string
}

// NewPostgresProvider creates a new PostgreSQL backup provider
func NewPostgresProvider() *PostgresProvider {
	return &PostgresProvider{
		toolPath: findExecutable("pg_dump"),
	}
}

// GetToolPath returns the current tool path
func (p *PostgresProvider) GetToolPath() string {
	return p.toolPath
}

// SetToolPath sets a custom tool path
func (p *PostgresProvider) SetToolPath(path string) {
	if path != "" {
		p.toolPath = path
	}
}

// TestConnection tests the database connection using Go native driver
func (p *PostgresProvider) TestConnection(config BackupConfig) error {
	sslMode := "disable"
	if config.SSLEnabled {
		sslMode = "require"
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=10",
		config.Host, config.Port, config.Username, config.Password, config.DatabaseName, sslMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	return nil
}

// Backup performs a PostgreSQL database backup using pg_dump
func (p *PostgresProvider) Backup(config BackupConfig, outputDir string) (*BackupResult, error) {
	if _, err := os.Stat(p.toolPath); err != nil {
		return nil, fmt.Errorf(i18n.T("errors.pgdumpNotFound"))
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

	// Build pg_dump arguments
	args := []string{
		"-h", config.Host,
		"-p", fmt.Sprintf("%d", config.Port),
		"-U", config.Username,
		"-d", config.DatabaseName,
		"-F", "p", // Plain SQL format
		"-f", outputPath,
		"--no-password",
	}

	// Create command
	cmd := exec.Command(p.toolPath, args...)
	prepareBundledCmd(cmd)

	// Set password via environment variable (PGPASSWORD)
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, "PGPASSWORD="+config.Password)

	// Add SSL mode if needed
	if config.SSLEnabled {
		cmd.Env = append(cmd.Env, "PGSSLMODE=require")
	} else {
		cmd.Env = append(cmd.Env, "PGSSLMODE=disable")
	}

	// Run pg_dump
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(outputPath) // Clean up partial dump file
		return nil, fmt.Errorf("pg_dump failed: %s - %v", string(output), err)
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

