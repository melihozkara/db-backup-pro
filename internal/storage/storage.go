package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// StorageFile represents a file in storage
type StorageFile struct {
	Name         string
	Path         string
	Size         int64
	ModifiedAt   time.Time
	IsDir        bool
}

// StorageProvider defines the interface for storage providers
type StorageProvider interface {
	Upload(localPath string, remotePath string) error
	Download(remotePath string, localPath string) error
	Delete(remotePath string) error
	List(prefix string) ([]StorageFile, error)
	TestConnection() error
	GetType() string
}

// LocalConfig contains configuration for local storage
type LocalConfig struct {
	Path string `json:"path"`
}

// FTPConfig contains configuration for FTP storage
type FTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Path     string `json:"path"`
}

// SFTPConfig contains configuration for SFTP storage
type SFTPConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	Path       string `json:"path"`
}

// S3Config contains configuration for S3/MinIO storage
type S3Config struct {
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Path            string `json:"path"`
	UseSSL          bool   `json:"use_ssl"`
}

// GetProvider returns the appropriate storage provider for the given type and config
func GetProvider(storageType string, configJSON string) (StorageProvider, error) {
	switch storageType {
	case "local":
		var config LocalConfig
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return nil, fmt.Errorf("invalid local config: %w", err)
		}
		return NewLocalProvider(config), nil
	case "ftp":
		var config FTPConfig
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return nil, fmt.Errorf("invalid FTP config: %w", err)
		}
		return NewFTPProvider(config), nil
	case "sftp":
		var config SFTPConfig
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return nil, fmt.Errorf("invalid SFTP config: %w", err)
		}
		return NewSFTPProvider(config), nil
	case "s3":
		var config S3Config
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return nil, fmt.Errorf("invalid S3 config: %w", err)
		}
		return NewS3Provider(config), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// ensureDir creates directory if it doesn't exist
func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
