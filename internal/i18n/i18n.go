package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed locales/*.json
var localeFS embed.FS

var (
	currentLang  = "tr"
	translations = make(map[string]map[string]string) // lang -> flat key -> value
	mu           sync.RWMutex
)

func init() {
	// Varsayilan dili yukle
	loadLanguage("tr")
}

// loadLanguage JSON dosyasini okuyup flat map'e cevirir
func loadLanguage(lang string) error {
	data, err := localeFS.ReadFile(fmt.Sprintf("locales/%s.json", lang))
	if err != nil {
		return fmt.Errorf("locale file not found: %s", lang)
	}

	var nested map[string]interface{}
	if err := json.Unmarshal(data, &nested); err != nil {
		return fmt.Errorf("failed to parse locale file: %w", err)
	}

	flat := make(map[string]string)
	flatten("", nested, flat)

	mu.Lock()
	translations[lang] = flat
	mu.Unlock()

	return nil
}

// flatten nested JSON'u dot-notation flat map'e cevirir
func flatten(prefix string, m map[string]interface{}, result map[string]string) {
	for key, value := range m {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			result[fullKey] = v
		case map[string]interface{}:
			flatten(fullKey, v, result)
		default:
			result[fullKey] = fmt.Sprintf("%v", v)
		}
	}
}

// SetLanguage aktif dili degistirir
func SetLanguage(lang string) error {
	mu.RLock()
	_, exists := translations[lang]
	mu.RUnlock()

	if !exists {
		if err := loadLanguage(lang); err != nil {
			return err
		}
	}

	mu.Lock()
	currentLang = lang
	mu.Unlock()

	return nil
}

// T verilen key icin ceviriyi dondurur
// Opsiyonel args varsa fmt.Sprintf ile formatlar
func T(key string, args ...interface{}) string {
	mu.RLock()
	trans, ok := translations[currentLang]
	mu.RUnlock()

	if !ok {
		return key
	}

	val, ok := trans[key]
	if !ok {
		return key
	}

	if len(args) > 0 {
		return fmt.Sprintf(val, args...)
	}

	return val
}

// GetAllTranslations belirtilen dildeki tum cevirileri nested JSON olarak dondurur
// Frontend'e Wails binding ile gonderilir
func GetAllTranslations(lang string) (map[string]interface{}, error) {
	data, err := localeFS.ReadFile(fmt.Sprintf("locales/%s.json", lang))
	if err != nil {
		return nil, fmt.Errorf("locale file not found: %s", lang)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse locale file: %w", err)
	}

	return result, nil
}

// GetCurrentLanguage mevcut dili dondurur
func GetCurrentLanguage() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}
