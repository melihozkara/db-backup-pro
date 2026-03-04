package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// LocalProvider implements StorageProvider for local filesystem storage
type LocalProvider struct {
	config LocalConfig
}

// NewLocalProvider creates a new local storage provider
func NewLocalProvider(config LocalConfig) *LocalProvider {
	return &LocalProvider{
		config: config,
	}
}

// GetType returns the storage type
func (p *LocalProvider) GetType() string {
	return "local"
}

// TestConnection tests if the local path is accessible
func (p *LocalProvider) TestConnection() error {
	// Check if path exists
	info, err := os.Stat(p.config.Path)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create the directory
			if err := ensureDir(p.config.Path); err != nil {
				return fmt.Errorf("cannot create directory: %w", err)
			}
			return nil
		}
		return fmt.Errorf("cannot access path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	// Test write permission by creating a temp file
	testFile := filepath.Join(p.config.Path, ".dbbackup_test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("cannot write to directory: %w", err)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}

// Upload copies a file from local path to storage
func (p *LocalProvider) Upload(localPath string, remotePath string) error {
	// Ensure the destination directory exists
	destPath := filepath.Join(p.config.Path, remotePath)
	destDir := filepath.Dir(destPath)

	if err := ensureDir(destDir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return copyFile(localPath, destPath)
}

// Download copies a file from storage to local path
func (p *LocalProvider) Download(remotePath string, localPath string) error {
	srcPath := filepath.Join(p.config.Path, remotePath)
	return copyFile(srcPath, localPath)
}

// Delete removes a file from storage
func (p *LocalProvider) Delete(remotePath string) error {
	fullPath := filepath.Join(p.config.Path, remotePath)
	return os.Remove(fullPath)
}

// List returns files in the storage with the given prefix
func (p *LocalProvider) List(prefix string) ([]StorageFile, error) {
	searchPath := filepath.Join(p.config.Path, prefix)

	// If prefix points to a directory, list its contents
	info, err := os.Stat(searchPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []StorageFile{}, nil
		}
		return nil, err
	}

	var files []StorageFile

	if info.IsDir() {
		entries, err := os.ReadDir(searchPath)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			entryInfo, err := entry.Info()
			if err != nil {
				continue
			}

			files = append(files, StorageFile{
				Name:       entry.Name(),
				Path:       filepath.Join(prefix, entry.Name()),
				Size:       entryInfo.Size(),
				ModifiedAt: entryInfo.ModTime(),
				IsDir:      entry.IsDir(),
			})
		}
	} else {
		// Single file
		files = append(files, StorageFile{
			Name:       info.Name(),
			Path:       prefix,
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
			IsDir:      false,
		})
	}

	return files, nil
}
