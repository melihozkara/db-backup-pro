package storage

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

// mkdirAll recursively creates directories on FTP server
func mkdirAll(conn *ftp.ServerConn, dir string) {
	parts := strings.Split(dir, "/")
	current := ""
	for _, part := range parts {
		if part == "" {
			current = "/"
			continue
		}
		if current == "" || current == "/" {
			current = current + part
		} else {
			current = current + "/" + part
		}
		conn.MakeDir(current) // Hata varsa (zaten var) yoksay
	}
}

// FTPProvider implements StorageProvider for FTP storage
type FTPProvider struct {
	config FTPConfig
}

// NewFTPProvider creates a new FTP storage provider
func NewFTPProvider(config FTPConfig) *FTPProvider {
	if config.Port == 0 {
		config.Port = 21
	}
	return &FTPProvider{
		config: config,
	}
}

// GetType returns the storage type
func (p *FTPProvider) GetType() string {
	return "ftp"
}

// connect establishes an FTP connection
func (p *FTPProvider) connect() (*ftp.ServerConn, error) {
	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	conn, err := ftp.Dial(addr, ftp.DialWithTimeout(30*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to FTP server: %w", err)
	}

	if err := conn.Login(p.config.Username, p.config.Password); err != nil {
		conn.Quit()
		return nil, fmt.Errorf("FTP login failed: %w", err)
	}

	return conn, nil
}

// TestConnection tests the FTP connection
func (p *FTPProvider) TestConnection() error {
	conn, err := p.connect()
	if err != nil {
		return err
	}
	defer conn.Quit()

	// Try to change to the target directory
	if p.config.Path != "" && p.config.Path != "/" {
		if err := conn.ChangeDir(p.config.Path); err != nil {
			// Try to create the directory
			if err := conn.MakeDir(p.config.Path); err != nil {
				return fmt.Errorf("cannot access or create directory: %w", err)
			}
		}
	}

	return nil
}

// Upload uploads a file to FTP server
func (p *FTPProvider) Upload(localPath string, remotePath string) error {
	conn, err := p.connect()
	if err != nil {
		return err
	}
	defer conn.Quit()

	// Open local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Build full remote path
	fullPath := path.Join(p.config.Path, remotePath)
	remoteDir := path.Dir(fullPath)

	// Create remote directory tree if needed
	if remoteDir != "" && remoteDir != "/" {
		mkdirAll(conn, remoteDir)
	}

	// Upload file
	if err := conn.Stor(fullPath, file); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// Download downloads a file from FTP server
func (p *FTPProvider) Download(remotePath string, localPath string) error {
	conn, err := p.connect()
	if err != nil {
		return err
	}
	defer conn.Quit()

	// Build full remote path
	fullPath := path.Join(p.config.Path, remotePath)

	// Retrieve file
	resp, err := conn.Retr(fullPath)
	if err != nil {
		return fmt.Errorf("failed to retrieve file: %w", err)
	}
	defer resp.Close()

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, resp); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// Delete removes a file from FTP server
func (p *FTPProvider) Delete(remotePath string) error {
	conn, err := p.connect()
	if err != nil {
		return err
	}
	defer conn.Quit()

	fullPath := path.Join(p.config.Path, remotePath)
	return conn.Delete(fullPath)
}

// List returns files in the FTP directory
func (p *FTPProvider) List(prefix string) ([]StorageFile, error) {
	conn, err := p.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Quit()

	searchPath := path.Join(p.config.Path, prefix)
	entries, err := conn.List(searchPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	var files []StorageFile
	for _, entry := range entries {
		if entry.Name == "." || entry.Name == ".." {
			continue
		}

		files = append(files, StorageFile{
			Name:       entry.Name,
			Path:       path.Join(prefix, entry.Name),
			Size:       int64(entry.Size),
			ModifiedAt: entry.Time,
			IsDir:      entry.Type == ftp.EntryTypeFolder,
		})
	}

	return files, nil
}
