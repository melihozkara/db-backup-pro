package storage

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPProvider implements StorageProvider for SFTP storage
type SFTPProvider struct {
	config SFTPConfig
}

// NewSFTPProvider creates a new SFTP storage provider
func NewSFTPProvider(config SFTPConfig) *SFTPProvider {
	if config.Port == 0 {
		config.Port = 22
	}
	return &SFTPProvider{
		config: config,
	}
}

// GetType returns the storage type
func (p *SFTPProvider) GetType() string {
	return "sftp"
}

// connect establishes an SFTP connection
func (p *SFTPProvider) connect() (*sftp.Client, *ssh.Client, error) {
	// Build SSH auth methods
	var authMethods []ssh.AuthMethod

	// Password auth
	if p.config.Password != "" {
		authMethods = append(authMethods, ssh.Password(p.config.Password))
	}

	// Private key auth
	if p.config.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(p.config.PrivateKey))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, nil, fmt.Errorf("no authentication method provided")
	}

	// SSH config
	sshConfig := &ssh.ClientConfig{
		User:            p.config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to SSH server
	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	// Create SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return sftpClient, sshClient, nil
}

// TestConnection tests the SFTP connection
func (p *SFTPProvider) TestConnection() error {
	sftpClient, sshClient, err := p.connect()
	if err != nil {
		return err
	}
	defer sftpClient.Close()
	defer sshClient.Close()

	// Check if directory exists or create it
	if p.config.Path != "" && p.config.Path != "/" {
		info, err := sftpClient.Stat(p.config.Path)
		if err != nil {
			// Try to create the directory
			if err := sftpClient.MkdirAll(p.config.Path); err != nil {
				return fmt.Errorf("cannot access or create directory: %w", err)
			}
		} else if !info.IsDir() {
			return fmt.Errorf("path is not a directory")
		}
	}

	return nil
}

// Upload uploads a file to SFTP server
func (p *SFTPProvider) Upload(localPath string, remotePath string) error {
	sftpClient, sshClient, err := p.connect()
	if err != nil {
		return err
	}
	defer sftpClient.Close()
	defer sshClient.Close()

	// Open local file
	srcFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer srcFile.Close()

	// Build full remote path
	fullPath := path.Join(p.config.Path, remotePath)
	remoteDir := path.Dir(fullPath)

	// Create remote directory if needed
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Create remote file
	dstFile, err := sftpClient.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer dstFile.Close()

	// Copy data
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// Download downloads a file from SFTP server
func (p *SFTPProvider) Download(remotePath string, localPath string) error {
	sftpClient, sshClient, err := p.connect()
	if err != nil {
		return err
	}
	defer sftpClient.Close()
	defer sshClient.Close()

	// Build full remote path
	fullPath := path.Join(p.config.Path, remotePath)

	// Open remote file
	srcFile, err := sftpClient.Open(fullPath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer srcFile.Close()

	// Create local file
	dstFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer dstFile.Close()

	// Copy data
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// Delete removes a file from SFTP server
func (p *SFTPProvider) Delete(remotePath string) error {
	sftpClient, sshClient, err := p.connect()
	if err != nil {
		return err
	}
	defer sftpClient.Close()
	defer sshClient.Close()

	fullPath := path.Join(p.config.Path, remotePath)
	return sftpClient.Remove(fullPath)
}

// List returns files in the SFTP directory
func (p *SFTPProvider) List(prefix string) ([]StorageFile, error) {
	sftpClient, sshClient, err := p.connect()
	if err != nil {
		return nil, err
	}
	defer sftpClient.Close()
	defer sshClient.Close()

	searchPath := path.Join(p.config.Path, prefix)
	entries, err := sftpClient.ReadDir(searchPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	var files []StorageFile
	for _, entry := range entries {
		files = append(files, StorageFile{
			Name:       entry.Name(),
			Path:       path.Join(prefix, entry.Name()),
			Size:       entry.Size(),
			ModifiedAt: entry.ModTime(),
			IsDir:      entry.IsDir(),
		})
	}

	return files, nil
}
