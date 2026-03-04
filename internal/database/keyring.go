package database

import (
	"crypto/rand"
	"dbbackup/internal/i18n"
	"encoding/hex"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "DBBackupPro"
	keyringUser    = "encryption-key"
)

// GetOrCreateEncryptionKey retrieves the encryption key from OS Keychain.
// If no key exists, generates a new 32-byte random key and stores it.
func GetOrCreateEncryptionKey() (string, error) {
	// Keychain'den anahtari almaya calis
	key, err := keyring.Get(keyringService, keyringUser)
	if err == nil && key != "" {
		return key, nil
	}

	// Anahtar yok, yeni olustur (32 byte = AES-256)
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf(i18n.T("errors.randomKeyFailed")+": %w", err)
	}

	key = hex.EncodeToString(keyBytes)

	// Keychain'e kaydet
	if err := keyring.Set(keyringService, keyringUser, key); err != nil {
		return "", fmt.Errorf(i18n.T("errors.keychainSaveFailed")+": %w", err)
	}

	return key, nil
}
