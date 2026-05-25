package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	AppName           string
	AppVersion        string
	Host              string
	Port              int
	DataDir           string
	DBPath            string
	ChromeHistoryPath string
	CORSOrigins       []string
}

func Load() Config {
	home, _ := os.UserHomeDir()
	dataDir := envString("DATA_DIR", filepath.Join(home, ".six-sense"))
	port := envInt("PORT", 8000)

	return Config{
		AppName:           envString("APP_NAME", "Six Sense Web"),
		AppVersion:        envString("APP_VERSION", "0.1.0"),
		Host:              envString("HOST", "127.0.0.1"),
		Port:              port,
		DataDir:           dataDir,
		DBPath:            envString("DB_PATH", filepath.Join(dataDir, "web.db")),
		ChromeHistoryPath: envString("CHROME_HISTORY_PATH", filepath.Join(home, "Library/Application Support/Google/Chrome/Default/History")),
		CORSOrigins:       envList("CORS_ORIGINS", []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
	}
}

func envString(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envList(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
