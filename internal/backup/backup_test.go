package backup

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ==================== GetProvider returns correct provider types ====================

func TestGetProviderPostgres(t *testing.T) {
	provider, err := GetProvider("postgres")
	if err != nil {
		t.Fatalf("GetProvider(postgres) returned error: %v", err)
	}
	if provider == nil {
		t.Fatal("GetProvider(postgres) returned nil provider")
	}

	// Verify it is a PostgresProvider
	if _, ok := provider.(*PostgresProvider); !ok {
		t.Errorf("expected *PostgresProvider, got %T", provider)
	}
}

func TestGetProviderMySQL(t *testing.T) {
	provider, err := GetProvider("mysql")
	if err != nil {
		t.Fatalf("GetProvider(mysql) returned error: %v", err)
	}
	if provider == nil {
		t.Fatal("GetProvider(mysql) returned nil provider")
	}

	if _, ok := provider.(*MySQLProvider); !ok {
		t.Errorf("expected *MySQLProvider, got %T", provider)
	}
}

func TestGetProviderMongoDB(t *testing.T) {
	provider, err := GetProvider("mongodb")
	if err != nil {
		t.Fatalf("GetProvider(mongodb) returned error: %v", err)
	}
	if provider == nil {
		t.Fatal("GetProvider(mongodb) returned nil provider")
	}

	if _, ok := provider.(*MongoDBProvider); !ok {
		t.Errorf("expected *MongoDBProvider, got %T", provider)
	}
}

// ==================== GetProvider with unsupported type ====================

func TestGetProviderUnsupportedType(t *testing.T) {
	unsupported := []string{"sqlite", "redis", "cassandra", "", "POSTGRES", "MySQL"}

	for _, dbType := range unsupported {
		provider, err := GetProvider(dbType)
		if err == nil {
			t.Errorf("GetProvider(%q) should return error, got provider: %v", dbType, provider)
		}
		if provider != nil {
			t.Errorf("GetProvider(%q) should return nil provider on error", dbType)
		}
	}
}

// ==================== generateBackupFileName format ====================

func TestGenerateBackupFileName(t *testing.T) {
	tests := []struct {
		name string
		ext  string
	}{
		{"mydb", ".sql"},
		{"testdb", ".sql"},
		{"appdb", ".bson"},
		{"production", ".dump"},
	}

	for _, tt := range tests {
		fname := generateBackupFileName(tt.name, tt.ext)

		// Check name prefix
		if !strings.HasPrefix(fname, tt.name+"_") {
			t.Errorf("filename %q should start with %q_", fname, tt.name)
		}

		// Check extension
		if !strings.HasSuffix(fname, tt.ext) {
			t.Errorf("filename %q should end with %q", fname, tt.ext)
		}

		// Check format: name_YYYYMMDD_HHMMSS.ext
		pattern := regexp.MustCompile(`^` + regexp.QuoteMeta(tt.name) + `_\d{8}_\d{6}` + regexp.QuoteMeta(tt.ext) + `$`)
		if !pattern.MatchString(fname) {
			t.Errorf("filename %q does not match expected pattern name_YYYYMMDD_HHMMSS.ext", fname)
		}
	}
}

func TestGenerateBackupFileNameUniqueness(t *testing.T) {
	// Two calls in quick succession should produce the same name (same second),
	// but the function itself is deterministic per timestamp.
	name1 := generateBackupFileName("test_db", ".sql")
	name2 := generateBackupFileName("test_db", ".sql")

	// They should be equal if called within the same second
	// (this is not guaranteed but is overwhelmingly likely in a test)
	if name1 != name2 {
		t.Logf("Note: filenames differ (called across second boundary): %q vs %q", name1, name2)
	}
}

// ==================== findExecutable ====================

func TestFindExecutableLs(t *testing.T) {
	// "ls" should be found on macOS and Linux
	path := findExecutable("ls")
	if path == "" {
		t.Skip("ls not found - skipping (Windows?)")
	}

	// The result should be an absolute path or the name itself
	if path == "ls" {
		// findExecutable returns the name if not found anywhere
		// On macOS/Linux, ls should be found
		t.Log("findExecutable returned bare 'ls' - it may not be in common paths")
	}

	// If it returned a path, verify the file exists
	if filepath.IsAbs(path) {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("findExecutable returned %q but file does not exist: %v", path, err)
		}
	}
}

func TestFindExecutableEcho(t *testing.T) {
	path := findExecutable("echo")
	if path == "" {
		t.Skip("echo not found")
	}

	// On macOS/Linux, echo should exist
	if filepath.IsAbs(path) {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("findExecutable returned %q but file does not exist: %v", path, err)
		}
	}
}

func TestFindExecutableNonexistent(t *testing.T) {
	// A completely nonexistent binary should return the name itself as fallback
	path := findExecutable("this_binary_definitely_does_not_exist_xyz")
	if path != "this_binary_definitely_does_not_exist_xyz" {
		t.Errorf("expected fallback to bare name, got %q", path)
	}
}

// ==================== ensureDir creates directories ====================

func TestEnsureDirCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	newDir := filepath.Join(dir, "a", "b", "c")

	// Verify it doesn't exist yet
	if _, err := os.Stat(newDir); !os.IsNotExist(err) {
		t.Fatal("directory should not exist before ensureDir")
	}

	err := ensureDir(newDir)
	if err != nil {
		t.Fatalf("ensureDir failed: %v", err)
	}

	// Verify it exists now
	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatalf("directory does not exist after ensureDir: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected a directory, got a file")
	}
}

func TestEnsureDirExistingDirectory(t *testing.T) {
	dir := t.TempDir()

	// Calling ensureDir on an existing directory should not fail
	err := ensureDir(dir)
	if err != nil {
		t.Fatalf("ensureDir on existing directory failed: %v", err)
	}
}

func TestEnsureDirNestedCreation(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "level1", "level2", "level3", "level4")

	err := ensureDir(nested)
	if err != nil {
		t.Fatalf("ensureDir (nested) failed: %v", err)
	}

	if _, err := os.Stat(nested); err != nil {
		t.Errorf("nested directory not created: %v", err)
	}
}

// ==================== getFileSize ====================

func TestGetFileSize(t *testing.T) {
	dir := t.TempDir()
	content := []byte("Hello, this is exactly 39 bytes of data")
	filePath := filepath.Join(dir, "testfile.txt")

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatal(err)
	}

	size, err := getFileSize(filePath)
	if err != nil {
		t.Fatalf("getFileSize failed: %v", err)
	}

	expectedSize := int64(len(content))
	if size != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, size)
	}
}

func TestGetFileSizeEmptyFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "empty.txt")

	if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	size, err := getFileSize(filePath)
	if err != nil {
		t.Fatalf("getFileSize failed: %v", err)
	}

	if size != 0 {
		t.Errorf("expected size 0 for empty file, got %d", size)
	}
}

func TestGetFileSizeNonexistent(t *testing.T) {
	dir := t.TempDir()
	_, err := getFileSize(filepath.Join(dir, "nonexistent.txt"))
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestGetFileSizeLargeFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "large.bin")

	// Create a 1MB file
	data := make([]byte, 1024*1024)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatal(err)
	}

	size, err := getFileSize(filePath)
	if err != nil {
		t.Fatalf("getFileSize failed: %v", err)
	}

	if size != 1024*1024 {
		t.Errorf("expected size %d, got %d", 1024*1024, size)
	}
}

// ==================== BackupProvider interface compliance ====================

func TestProviderImplementsInterface(t *testing.T) {
	// Compile-time checks that all providers implement BackupProvider
	var _ BackupProvider = (*PostgresProvider)(nil)
	var _ BackupProvider = (*MySQLProvider)(nil)
	var _ BackupProvider = (*MongoDBProvider)(nil)
}

// ==================== getTempDir ====================

func TestGetTempDir(t *testing.T) {
	dir := getTempDir()
	if dir == "" {
		t.Error("getTempDir returned empty string")
	}

	// Verify the directory exists (getTempDir calls ensureDir internally)
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("getTempDir directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("getTempDir did not return a directory path")
	}

	// Should contain "dbbackup" in the path
	if !strings.Contains(dir, "dbbackup") {
		t.Errorf("expected temp dir to contain 'dbbackup', got %q", dir)
	}
}

// ==================== BackupResult / BackupConfig structs ====================

func TestBackupResultStruct(t *testing.T) {
	result := BackupResult{
		FilePath: "/tmp/backup.sql",
		FileName: "backup.sql",
		FileSize: 12345,
	}

	if result.FilePath != "/tmp/backup.sql" {
		t.Errorf("unexpected FilePath: %q", result.FilePath)
	}
	if result.FileName != "backup.sql" {
		t.Errorf("unexpected FileName: %q", result.FileName)
	}
	if result.FileSize != 12345 {
		t.Errorf("unexpected FileSize: %d", result.FileSize)
	}
}

func TestBackupConfigStruct(t *testing.T) {
	config := BackupConfig{
		Type:         "postgres",
		Host:         "localhost",
		Port:         5432,
		Username:     "admin",
		Password:     "secret",
		DatabaseName: "mydb",
		SSLEnabled:   true,
	}

	if config.Type != "postgres" {
		t.Errorf("unexpected Type: %q", config.Type)
	}
	if config.Port != 5432 {
		t.Errorf("unexpected Port: %d", config.Port)
	}
	if !config.SSLEnabled {
		t.Error("expected SSLEnabled to be true")
	}
}
