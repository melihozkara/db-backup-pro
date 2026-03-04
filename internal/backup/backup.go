package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// BackupResult represents the result of a backup operation
type BackupResult struct {
	FilePath   string
	FileName   string
	FileSize   int64
	Duration   time.Duration
	StartedAt  time.Time
	FinishedAt time.Time
}

// BackupConfig contains database connection info for backup
type BackupConfig struct {
	Type         string
	Host         string
	Port         int
	Username     string
	Password     string
	DatabaseName string
	SSLEnabled   bool
	AuthDatabase string // MongoDB authSource (varsayilan: admin)
	CustomPrefix string // Kullanicinin belirledigi dosya adi prefix'i
}

// BackupProvider defines the interface for database backup providers
// NOT: Bu uygulama veritabanina ASLA mudahale etmez. Sadece read-only backup alir.
type BackupProvider interface {
	Backup(config BackupConfig, outputDir string) (*BackupResult, error)
	TestConnection(config BackupConfig) error
	GetToolPath() string
	SetToolPath(path string)
}

// GetProvider returns the appropriate backup provider for the database type
func GetProvider(dbType string) (BackupProvider, error) {
	switch dbType {
	case "postgres":
		return NewPostgresProvider(), nil
	case "mysql":
		return NewMySQLProvider(), nil
	case "mongodb":
		return NewMongoDBProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// getBundledToolsDir returns the path to bundled tools directory next to the executable
func getBundledToolsDir() string {
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}
	execDir := filepath.Dir(execPath)

	// macOS .app bundle: Contents/MacOS/app -> Contents/Resources/tools/
	if runtime.GOOS == "darwin" {
		resourcesDir := filepath.Join(filepath.Dir(execDir), "Resources", "tools")
		if info, err := os.Stat(resourcesDir); err == nil && info.IsDir() {
			return resourcesDir
		}
	}

	// Fallback: tools/ directory next to executable
	toolsDir := filepath.Join(execDir, "tools")
	if info, err := os.Stat(toolsDir); err == nil && info.IsDir() {
		return toolsDir
	}

	return ""
}

// findExecutable searches for an executable: bundled tools first, then PATH, then common locations
func findExecutable(name string) string {
	exeName := name
	if runtime.GOOS == "windows" {
		exeName = name + ".exe"
	}

	// 1. Bundled tools directory (shipped with app)
	if toolsDir := getBundledToolsDir(); toolsDir != "" {
		bundled := filepath.Join(toolsDir, exeName)
		if _, err := os.Stat(bundled); err == nil {
			return bundled
		}
	}

	// 2. PATH
	path, err := exec.LookPath(name)
	if err == nil {
		return path
	}

	// 3. Common installation paths by OS
	switch runtime.GOOS {
	case "darwin":
		for _, p := range []string{
			"/usr/local/bin/" + name,
			"/opt/homebrew/bin/" + name,
			"/opt/homebrew/opt/mysql-client/bin/" + name,
			"/opt/homebrew/opt/mysql/bin/" + name,
			"/opt/homebrew/opt/libpq/bin/" + name,
			"/opt/homebrew/opt/postgresql@16/bin/" + name,
			"/opt/homebrew/opt/postgresql@15/bin/" + name,
			"/opt/homebrew/opt/mongodb-database-tools/bin/" + name,
			"/usr/local/opt/mysql-client/bin/" + name,
			"/usr/local/opt/libpq/bin/" + name,
			"/usr/bin/" + name,
		} {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	case "linux":
		for _, p := range []string{
			"/usr/bin/" + name,
			"/usr/local/bin/" + name,
		} {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	case "windows":
		// 3a. Glob-based search: finds any installed version automatically
		globs := []string{
			"C:\\Program Files\\MySQL\\MySQL Server *\\bin\\" + exeName,
			"C:\\Program Files\\MariaDB*\\bin\\" + exeName,
			"C:\\Program Files\\PostgreSQL\\*\\bin\\" + exeName,
			"C:\\Program Files\\MongoDB\\Server\\*\\bin\\" + exeName,
			"C:\\Program Files\\MongoDB\\Tools\\*\\bin\\" + exeName,
			"C:\\Program Files (x86)\\MySQL\\MySQL Server *\\bin\\" + exeName,
			"C:\\Program Files (x86)\\MariaDB*\\bin\\" + exeName,
		}
		for _, pattern := range globs {
			matches, _ := filepath.Glob(pattern)
			if len(matches) > 0 {
				// Return the last match (typically the latest version)
				return matches[len(matches)-1]
			}
		}
		// 3b. Other common Windows locations
		for _, p := range []string{
			"C:\\xampp\\mysql\\bin\\" + exeName,
			"C:\\wamp64\\bin\\mysql\\bin\\" + exeName,
			"C:\\laragon\\bin\\mysql\\bin\\" + exeName,
		} {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	return name
}

// prepareBundledCmd sets up library paths for bundled tools so they can find their shared libraries
func prepareBundledCmd(cmd *exec.Cmd) {
	toolsDir := getBundledToolsDir()
	if toolsDir == "" {
		return
	}

	libDir := filepath.Join(toolsDir, "lib")
	if _, err := os.Stat(libDir); err != nil {
		return
	}

	// Only modify env if the tool is from our bundle
	toolPath := cmd.Path
	if !strings.HasPrefix(toolPath, toolsDir) {
		return
	}

	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}

	switch runtime.GOOS {
	case "darwin":
		cmd.Env = append(cmd.Env, "DYLD_LIBRARY_PATH="+libDir)
	case "linux":
		cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH="+libDir)
	case "windows":
		// Windows: DLLs are found via PATH
		for i, env := range cmd.Env {
			if len(env) > 5 && (env[:5] == "PATH=" || env[:5] == "Path=") {
				cmd.Env[i] = env + ";" + libDir
				return
			}
		}
		cmd.Env = append(cmd.Env, "PATH="+libDir)
	}
}

// generateBackupFileName creates a unique backup filename
// name: custom_prefix (kullanici verdiyse) veya database_name
func generateBackupFileName(name, ext string) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s%s", name, timestamp, ext)
}

// getFileSize returns the size of a file
func getFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ensureDir creates directory if it doesn't exist
func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// getTempDir returns a temporary directory for backup operations
func getTempDir() string {
	dir := filepath.Join(os.TempDir(), "dbbackup")
	ensureDir(dir)
	return dir
}
