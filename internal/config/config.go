package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	APIHost            string
	APIUser            string
	APIPass            string
	APIPort            string
	WebPort            string
	UserPassword       string
	AdminPassword      string
	SuperAdminPassword string
	// DataDir مسیری که settings.json و بکاپ روی آن ذخیره می‌شود
	// در کانتینر میکروتیک می‌تواند به یک mount اشاره کند (مثلاً /data)
	DataDir      string
	SettingsFile string
}

func Load() *Config {
	dataDir := getEnv("DATA_DIR", "/data")
	settingsFile := getEnv("SETTINGS_FILE", filepath.Join(dataDir, "settings.json"))
	return &Config{
		APIHost:            getEnv("API_HOST", "192.168.88.1"),
		APIUser:            getEnv("API_USER", "admin"),
		APIPass:            getEnv("API_PASS", ""),
		APIPort:            getEnv("API_PORT", "8728"),
		WebPort:            getEnv("WEB_PORT", "5000"),
		UserPassword:       getEnv("WEB_USER_PASSWORD", "123"),
		AdminPassword:      getEnv("WEB_ADMIN_PASSWORD", "123456"),
		SuperAdminPassword: getEnv("WEB_SUPERADMIN_PASSWORD", "123456789"),
		DataDir:            dataDir,
		SettingsFile:       settingsFile,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
