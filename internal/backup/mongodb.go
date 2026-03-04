package backup

import (
	"context"
	"dbbackup/internal/i18n"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDBProvider implements BackupProvider for MongoDB databases
type MongoDBProvider struct {
	toolPath string
}

// NewMongoDBProvider creates a new MongoDB backup provider
func NewMongoDBProvider() *MongoDBProvider {
	return &MongoDBProvider{
		toolPath: findExecutable("mongodump"),
	}
}

// GetToolPath returns the current tool path
func (p *MongoDBProvider) GetToolPath() string {
	return p.toolPath
}

// SetToolPath sets a custom tool path
func (p *MongoDBProvider) SetToolPath(path string) {
	if path != "" {
		p.toolPath = path
	}
}

// buildConnectionURI builds a MongoDB connection URI
func (p *MongoDBProvider) buildConnectionURI(config BackupConfig) string {
	scheme := "mongodb"
	if config.SSLEnabled {
		scheme = "mongodb+srv"
	}

	// authSource: kullanici belirlemediyse varsayilan "admin"
	authSource := config.AuthDatabase
	if authSource == "" {
		authSource = "admin"
	}

	// Build URI with authentication (URL-encode credentials for special characters)
	if config.Username != "" && config.Password != "" {
		return fmt.Sprintf("%s://%s:%s@%s:%d/%s?authSource=%s",
			scheme, url.PathEscape(config.Username), url.PathEscape(config.Password),
			config.Host, config.Port, config.DatabaseName, authSource)
	}

	return fmt.Sprintf("%s://%s:%d/%s",
		scheme, config.Host, config.Port, config.DatabaseName)
}

// TestConnection tests the database connection using Go native driver
func (p *MongoDBProvider) TestConnection(config BackupConfig) error {
	uri := p.buildConnectionURI(config)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	return nil
}

// Backup performs a MongoDB database backup using mongodump
func (p *MongoDBProvider) Backup(config BackupConfig, outputDir string) (*BackupResult, error) {
	if _, err := os.Stat(p.toolPath); err != nil {
		return nil, fmt.Errorf(i18n.T("errors.mongodumpNotFound"))
	}

	startTime := time.Now()

	// Ensure output directory exists
	if err := ensureDir(outputDir); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate output directory name for mongodump
	name := config.CustomPrefix
	if name == "" {
		name = config.DatabaseName
	}
	timestamp := time.Now().Format("20060102_150405")
	dumpDirName := fmt.Sprintf("%s_%s", name, timestamp)
	dumpPath := filepath.Join(outputDir, dumpDirName)

	// Build mongodump arguments using --uri (avoids password in process list via --password flag)
	uri := p.buildConnectionURI(config)
	args := []string{
		"--uri", uri,
		"--db", config.DatabaseName,
		"--out", dumpPath,
	}

	// Create command
	cmd := exec.Command(p.toolPath, args...)
	prepareBundledCmd(cmd)

	// Run mongodump
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up partial dump on error
		os.RemoveAll(dumpPath)
		return nil, fmt.Errorf("mongodump failed: %s - %v", string(output), err)
	}

	// Create archive from the dump directory
	archiveName := dumpDirName + ".tar.gz"
	archivePath := filepath.Join(outputDir, archiveName)

	// Create tar.gz archive
	tarCmd := exec.Command("tar", "-czf", archivePath, "-C", outputDir, dumpDirName)
	if tarOutput, tarErr := tarCmd.CombinedOutput(); tarErr != nil {
		return nil, fmt.Errorf("failed to create archive: %s - %v", string(tarOutput), tarErr)
	}

	// Remove the dump directory
	os.RemoveAll(dumpPath)

	// Get file size
	fileSize, err := getFileSize(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file size: %w", err)
	}

	finishTime := time.Now()

	return &BackupResult{
		FilePath:   archivePath,
		FileName:   archiveName,
		FileSize:   fileSize,
		Duration:   finishTime.Sub(startTime),
		StartedAt:  startTime,
		FinishedAt: finishTime,
	}, nil
}

