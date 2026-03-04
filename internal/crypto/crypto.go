package crypto

import (
	"archive/zip"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CompressGzip compresses a file using gzip
func CompressGzip(inputPath, outputPath string) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	// Set the original filename
	gzipWriter.Name = filepath.Base(inputPath)

	// Copy data
	if _, err := io.Copy(gzipWriter, inputFile); err != nil {
		return fmt.Errorf("failed to compress file: %w", err)
	}

	return nil
}

// DecompressGzip decompresses a gzip file
func DecompressGzip(inputPath, outputPath string) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(inputFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Copy data
	if _, err := io.Copy(outputFile, gzipReader); err != nil {
		return fmt.Errorf("failed to decompress file: %w", err)
	}

	return nil
}

// CompressZip compresses a file or directory using zip
func CompressZip(inputPath, outputPath string) error {
	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create zip writer
	zipWriter := zip.NewWriter(outputFile)
	defer zipWriter.Close()

	// Check if input is a file or directory
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to stat input: %w", err)
	}

	if info.IsDir() {
		// Walk directory and add files
		return filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			// Get relative path
			relPath, err := filepath.Rel(inputPath, path)
			if err != nil {
				return err
			}

			// Create zip entry
			writer, err := zipWriter.Create(relPath)
			if err != nil {
				return err
			}

			// Open and copy file
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return err
		})
	}

	// Single file
	writer, err := zipWriter.Create(filepath.Base(inputPath))
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("failed to compress file: %w", err)
	}

	return nil
}

// DecompressZip decompresses a zip file to a directory
func DecompressZip(inputPath, outputDir string) error {
	// Open zip file
	zipReader, err := zip.OpenReader(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zipReader.Close()

	// Extract files
	for _, file := range zipReader.File {
		path := filepath.Join(outputDir, file.Name)

		// Check for zip slip vulnerability
		if !strings.HasPrefix(path, filepath.Clean(outputDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// Create output file
		outFile, err := os.Create(path)
		if err != nil {
			return err
		}

		// Open zip entry
		zipFile, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		// Copy data
		_, err = io.Copy(outFile, zipFile)
		outFile.Close()
		zipFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// deriveKey derives a 32-byte key from a password using SHA-256
func deriveKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

// EncryptFile encrypts a file using AES-256-GCM
func EncryptFile(inputPath, outputPath, password string) error {
	// Check file size to prevent OOM (AES-GCM requires full file in memory)
	info, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to stat input file: %w", err)
	}
	const maxSize = 2 * 1024 * 1024 * 1024 // 2GB
	if info.Size() > maxSize {
		return fmt.Errorf("file too large for encryption (%d bytes, max 2GB)", info.Size())
	}

	// Read input file
	plaintext, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Derive key from password
	key := deriveKey(password)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data (nonce is prepended to ciphertext)
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Write output file
	if err := os.WriteFile(outputPath, ciphertext, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// DecryptFile decrypts a file using AES-256-GCM
func DecryptFile(inputPath, outputPath, password string) error {
	// Read input file
	ciphertext, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Derive key from password
	key := deriveKey(password)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check minimum size
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputPath, plaintext, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// ProcessBackup applies compression and/or encryption to a backup file
func ProcessBackup(inputPath string, compression string, encrypt bool, encryptionKey string) (string, error) {
	currentPath := inputPath

	// Apply compression if specified
	if compression != "none" && compression != "" {
		var compressedPath string
		var err error

		switch compression {
		case "gzip":
			compressedPath = inputPath + ".gz"
			err = CompressGzip(inputPath, compressedPath)
		case "zip":
			compressedPath = inputPath + ".zip"
			err = CompressZip(inputPath, compressedPath)
		default:
			return "", fmt.Errorf("unsupported compression type: %s", compression)
		}

		if err != nil {
			return "", fmt.Errorf("compression failed: %w", err)
		}

		// Remove original file
		os.Remove(inputPath)
		currentPath = compressedPath
	}

	// Apply encryption if specified
	if encrypt && encryptionKey != "" {
		encryptedPath := currentPath + ".enc"
		if err := EncryptFile(currentPath, encryptedPath, encryptionKey); err != nil {
			// Cleanup compressed file if it differs from original input
			if currentPath != inputPath {
				os.Remove(currentPath)
			}
			return "", fmt.Errorf("encryption failed: %w", err)
		}

		// Remove unencrypted file
		os.Remove(currentPath)
		currentPath = encryptedPath
	}

	return currentPath, nil
}
