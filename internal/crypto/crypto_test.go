package crypto

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// ==================== CompressGzip + DecompressGzip ====================

func TestCompressDecompressGzipRoundtrip(t *testing.T) {
	dir := t.TempDir()
	original := []byte("Hello, this is a gzip roundtrip test with some repeated data aaaaaa bbbbbb cccccc")

	inputPath := filepath.Join(dir, "input.txt")
	compressedPath := filepath.Join(dir, "input.txt.gz")
	outputPath := filepath.Join(dir, "output.txt")

	if err := os.WriteFile(inputPath, original, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Compress
	if err := CompressGzip(inputPath, compressedPath); err != nil {
		t.Fatalf("CompressGzip failed: %v", err)
	}

	// Verify compressed file exists and is different from original
	compressedData, err := os.ReadFile(compressedPath)
	if err != nil {
		t.Fatalf("failed to read compressed file: %v", err)
	}
	if bytes.Equal(compressedData, original) {
		t.Error("compressed data should be different from original")
	}

	// Decompress
	if err := DecompressGzip(compressedPath, outputPath); err != nil {
		t.Fatalf("DecompressGzip failed: %v", err)
	}

	// Verify roundtrip
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !bytes.Equal(original, output) {
		t.Errorf("roundtrip mismatch: got %q, want %q", output, original)
	}
}

func TestCompressGzipEmptyFile(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "empty.txt")
	compressedPath := filepath.Join(dir, "empty.txt.gz")
	outputPath := filepath.Join(dir, "output.txt")

	if err := os.WriteFile(inputPath, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	if err := CompressGzip(inputPath, compressedPath); err != nil {
		t.Fatalf("CompressGzip failed on empty file: %v", err)
	}

	if err := DecompressGzip(compressedPath, outputPath); err != nil {
		t.Fatalf("DecompressGzip failed on empty file: %v", err)
	}

	output, _ := os.ReadFile(outputPath)
	if len(output) != 0 {
		t.Errorf("expected empty output, got %d bytes", len(output))
	}
}

func TestCompressGzipNonexistentInput(t *testing.T) {
	dir := t.TempDir()
	err := CompressGzip(filepath.Join(dir, "nonexistent"), filepath.Join(dir, "out.gz"))
	if err == nil {
		t.Error("expected error for nonexistent input file")
	}
}

func TestDecompressGzipInvalidFile(t *testing.T) {
	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.gz")
	if err := os.WriteFile(badFile, []byte("not gzip data"), 0644); err != nil {
		t.Fatal(err)
	}
	err := DecompressGzip(badFile, filepath.Join(dir, "out.txt"))
	if err == nil {
		t.Error("expected error for invalid gzip data")
	}
}

// ==================== CompressZip + DecompressZip (single file) ====================

func TestCompressDecompressZipSingleFileRoundtrip(t *testing.T) {
	dir := t.TempDir()
	original := []byte("This is a zip single-file roundtrip test.")

	inputPath := filepath.Join(dir, "testfile.txt")
	zipPath := filepath.Join(dir, "archive.zip")
	outputDir := filepath.Join(dir, "extracted")

	if err := os.WriteFile(inputPath, original, 0644); err != nil {
		t.Fatalf("failed to write input file: %v", err)
	}

	// Compress
	if err := CompressZip(inputPath, zipPath); err != nil {
		t.Fatalf("CompressZip failed: %v", err)
	}

	// Verify zip file exists
	info, err := os.Stat(zipPath)
	if err != nil {
		t.Fatalf("zip file does not exist: %v", err)
	}
	if info.Size() == 0 {
		t.Error("zip file is empty")
	}

	// Decompress
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := DecompressZip(zipPath, outputDir); err != nil {
		t.Fatalf("DecompressZip failed: %v", err)
	}

	// Verify roundtrip - the file should be extracted with its base name
	extractedPath := filepath.Join(outputDir, "testfile.txt")
	output, err := os.ReadFile(extractedPath)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if !bytes.Equal(original, output) {
		t.Errorf("roundtrip mismatch: got %q, want %q", output, original)
	}
}

// ==================== CompressZip + DecompressZip (directory) ====================

func TestCompressDecompressZipDirectoryRoundtrip(t *testing.T) {
	dir := t.TempDir()

	// Create a directory structure with multiple files
	srcDir := filepath.Join(dir, "source")
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	files := map[string][]byte{
		"file1.txt":        []byte("Content of file 1"),
		"file2.txt":        []byte("Content of file 2"),
		"subdir/file3.txt": []byte("Content of file 3 in subdir"),
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(srcDir, name), content, 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	zipPath := filepath.Join(dir, "directory.zip")
	outputDir := filepath.Join(dir, "extracted")

	// Compress directory
	if err := CompressZip(srcDir, zipPath); err != nil {
		t.Fatalf("CompressZip (directory) failed: %v", err)
	}

	// Decompress
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := DecompressZip(zipPath, outputDir); err != nil {
		t.Fatalf("DecompressZip failed: %v", err)
	}

	// Verify all files
	for name, expectedContent := range files {
		extractedPath := filepath.Join(outputDir, name)
		output, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Errorf("failed to read extracted %s: %v", name, err)
			continue
		}
		if !bytes.Equal(expectedContent, output) {
			t.Errorf("content mismatch for %s: got %q, want %q", name, output, expectedContent)
		}
	}
}

// ==================== EncryptFile + DecryptFile ====================

func TestEncryptDecryptFileRoundtrip(t *testing.T) {
	dir := t.TempDir()
	original := []byte("Secret data that must survive encryption roundtrip.")
	password := "test-password-123!"

	inputPath := filepath.Join(dir, "plaintext.dat")
	encryptedPath := filepath.Join(dir, "encrypted.dat")
	decryptedPath := filepath.Join(dir, "decrypted.dat")

	if err := os.WriteFile(inputPath, original, 0644); err != nil {
		t.Fatal(err)
	}

	// Encrypt
	if err := EncryptFile(inputPath, encryptedPath, password); err != nil {
		t.Fatalf("EncryptFile failed: %v", err)
	}

	// Verify encrypted file exists and differs from original
	encryptedData, err := os.ReadFile(encryptedPath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(encryptedData, original) {
		t.Error("encrypted data should differ from plaintext")
	}

	// Decrypt
	if err := DecryptFile(encryptedPath, decryptedPath, password); err != nil {
		t.Fatalf("DecryptFile failed: %v", err)
	}

	// Verify roundtrip
	output, err := os.ReadFile(decryptedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, output) {
		t.Errorf("roundtrip mismatch: got %q, want %q", output, original)
	}
}

func TestEncryptFileProducesDifferentCiphertexts(t *testing.T) {
	// Due to random nonce, encrypting the same file twice should yield different ciphertexts
	dir := t.TempDir()
	data := []byte("same data")
	password := "same-password"

	inputPath := filepath.Join(dir, "input.dat")
	enc1 := filepath.Join(dir, "enc1.dat")
	enc2 := filepath.Join(dir, "enc2.dat")

	if err := os.WriteFile(inputPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	if err := EncryptFile(inputPath, enc1, password); err != nil {
		t.Fatal(err)
	}
	if err := EncryptFile(inputPath, enc2, password); err != nil {
		t.Fatal(err)
	}

	c1, _ := os.ReadFile(enc1)
	c2, _ := os.ReadFile(enc2)
	if bytes.Equal(c1, c2) {
		t.Error("encrypting the same file twice should produce different ciphertexts (random nonce)")
	}
}

// ==================== DecryptFile with wrong password ====================

func TestDecryptFileWrongPassword(t *testing.T) {
	dir := t.TempDir()
	original := []byte("Data encrypted with correct password")
	correctPassword := "correct-password"
	wrongPassword := "wrong-password"

	inputPath := filepath.Join(dir, "plaintext.dat")
	encryptedPath := filepath.Join(dir, "encrypted.dat")
	decryptedPath := filepath.Join(dir, "decrypted.dat")

	if err := os.WriteFile(inputPath, original, 0644); err != nil {
		t.Fatal(err)
	}

	if err := EncryptFile(inputPath, encryptedPath, correctPassword); err != nil {
		t.Fatal(err)
	}

	err := DecryptFile(encryptedPath, decryptedPath, wrongPassword)
	if err == nil {
		t.Error("expected error when decrypting with wrong password")
	}
}

func TestDecryptFileTruncatedCiphertext(t *testing.T) {
	dir := t.TempDir()
	// Write a file that is too short to contain even a nonce
	shortPath := filepath.Join(dir, "short.enc")
	if err := os.WriteFile(shortPath, []byte("short"), 0644); err != nil {
		t.Fatal(err)
	}

	err := DecryptFile(shortPath, filepath.Join(dir, "out.dat"), "password")
	if err == nil {
		t.Error("expected error for truncated ciphertext")
	}
}

// ==================== ProcessBackup with gzip compression ====================

func TestProcessBackupWithGzip(t *testing.T) {
	dir := t.TempDir()
	data := []byte("Backup data for gzip processing test")

	inputPath := filepath.Join(dir, "backup.sql")
	if err := os.WriteFile(inputPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	resultPath, err := ProcessBackup(inputPath, "gzip", false, "")
	if err != nil {
		t.Fatalf("ProcessBackup with gzip failed: %v", err)
	}

	expectedPath := inputPath + ".gz"
	if resultPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, resultPath)
	}

	if _, err := os.Stat(resultPath); err != nil {
		t.Errorf("result file does not exist: %v", err)
	}

	// Verify original was removed
	if _, err := os.Stat(inputPath); !os.IsNotExist(err) {
		t.Error("original file should have been removed after compression")
	}

	// Verify we can decompress and get the original data
	decompressedPath := filepath.Join(dir, "decompressed.sql")
	if err := DecompressGzip(resultPath, decompressedPath); err != nil {
		t.Fatalf("DecompressGzip failed: %v", err)
	}
	output, _ := os.ReadFile(decompressedPath)
	if !bytes.Equal(data, output) {
		t.Errorf("gzip roundtrip content mismatch")
	}
}

// ==================== ProcessBackup with zip compression ====================

func TestProcessBackupWithZip(t *testing.T) {
	dir := t.TempDir()
	data := []byte("Backup data for zip processing test")

	inputPath := filepath.Join(dir, "backup.sql")
	if err := os.WriteFile(inputPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	resultPath, err := ProcessBackup(inputPath, "zip", false, "")
	if err != nil {
		t.Fatalf("ProcessBackup with zip failed: %v", err)
	}

	expectedPath := inputPath + ".zip"
	if resultPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, resultPath)
	}

	if _, err := os.Stat(resultPath); err != nil {
		t.Errorf("result file does not exist: %v", err)
	}

	// Verify original was removed
	if _, err := os.Stat(inputPath); !os.IsNotExist(err) {
		t.Error("original file should have been removed after compression")
	}
}

// ==================== ProcessBackup with encryption ====================

func TestProcessBackupWithEncryption(t *testing.T) {
	dir := t.TempDir()
	data := []byte("Backup data for encryption processing test")
	password := "encryption-test-key"

	inputPath := filepath.Join(dir, "backup.sql")
	if err := os.WriteFile(inputPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	resultPath, err := ProcessBackup(inputPath, "none", true, password)
	if err != nil {
		t.Fatalf("ProcessBackup with encryption failed: %v", err)
	}

	expectedPath := inputPath + ".enc"
	if resultPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, resultPath)
	}

	if _, err := os.Stat(resultPath); err != nil {
		t.Errorf("result file does not exist: %v", err)
	}

	// Verify original was removed
	if _, err := os.Stat(inputPath); !os.IsNotExist(err) {
		t.Error("original file should have been removed after encryption")
	}

	// Verify we can decrypt and get back the original
	decryptedPath := filepath.Join(dir, "decrypted.sql")
	if err := DecryptFile(resultPath, decryptedPath, password); err != nil {
		t.Fatalf("DecryptFile failed: %v", err)
	}
	output, _ := os.ReadFile(decryptedPath)
	if !bytes.Equal(data, output) {
		t.Errorf("encryption roundtrip content mismatch")
	}
}

// ==================== ProcessBackup with compression + encryption ====================

func TestProcessBackupWithCompressionAndEncryption(t *testing.T) {
	dir := t.TempDir()
	data := []byte("Backup data for combined compression+encryption test")
	password := "combo-test-key"

	inputPath := filepath.Join(dir, "backup.sql")
	if err := os.WriteFile(inputPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	resultPath, err := ProcessBackup(inputPath, "gzip", true, password)
	if err != nil {
		t.Fatalf("ProcessBackup with gzip+encryption failed: %v", err)
	}

	expectedPath := inputPath + ".gz.enc"
	if resultPath != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, resultPath)
	}

	if _, err := os.Stat(resultPath); err != nil {
		t.Errorf("result file does not exist: %v", err)
	}

	// Verify all intermediate files are cleaned up
	if _, err := os.Stat(inputPath); !os.IsNotExist(err) {
		t.Error("original file should have been removed")
	}
	if _, err := os.Stat(inputPath + ".gz"); !os.IsNotExist(err) {
		t.Error("intermediate gzip file should have been removed")
	}

	// Full roundtrip: decrypt then decompress
	decryptedPath := filepath.Join(dir, "decrypted.gz")
	if err := DecryptFile(resultPath, decryptedPath, password); err != nil {
		t.Fatal(err)
	}
	decompressedPath := filepath.Join(dir, "decompressed.sql")
	if err := DecompressGzip(decryptedPath, decompressedPath); err != nil {
		t.Fatal(err)
	}
	output, _ := os.ReadFile(decompressedPath)
	if !bytes.Equal(data, output) {
		t.Errorf("compression+encryption roundtrip content mismatch")
	}
}

func TestProcessBackupNoCompressionNoEncryption(t *testing.T) {
	dir := t.TempDir()
	data := []byte("No processing needed")

	inputPath := filepath.Join(dir, "backup.sql")
	if err := os.WriteFile(inputPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	resultPath, err := ProcessBackup(inputPath, "none", false, "")
	if err != nil {
		t.Fatalf("ProcessBackup with no processing failed: %v", err)
	}

	if resultPath != inputPath {
		t.Errorf("expected path %s (unchanged), got %s", inputPath, resultPath)
	}

	output, _ := os.ReadFile(resultPath)
	if !bytes.Equal(data, output) {
		t.Errorf("file content should be unchanged")
	}
}

func TestProcessBackupUnsupportedCompression(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "backup.sql")
	if err := os.WriteFile(inputPath, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ProcessBackup(inputPath, "lz4", false, "")
	if err == nil {
		t.Error("expected error for unsupported compression type")
	}
}

// ==================== Zip slip attack prevention ====================

func TestDecompressZipSlipPrevention(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "malicious.zip")
	outputDir := filepath.Join(dir, "output")

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a zip archive containing a path-traversal entry like "../../evil.txt"
	if err := createZipWithEntry(zipPath, "../../evil.txt", []byte("malicious content")); err != nil {
		t.Fatalf("failed to create test zip: %v", err)
	}

	err := DecompressZip(zipPath, outputDir)
	if err == nil {
		// Even if no error, verify no file was created outside output directory
		escapedPath := filepath.Join(dir, "evil.txt")
		if _, statErr := os.Stat(escapedPath); statErr == nil {
			t.Error("zip slip attack succeeded - file was created outside output directory")
		}
	}
	// If err != nil, the zip slip was properly prevented (expected behavior)
}

// createZipWithEntry creates a zip file with a single entry having the given name.
// This is a test helper to craft zip files with arbitrary entry names (including
// path-traversal names like "../../evil.txt").
func createZipWithEntry(zipPath, entryName string, content []byte) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	entry, err := w.Create(entryName)
	if err != nil {
		return err
	}

	_, err = entry.Write(content)
	return err
}
