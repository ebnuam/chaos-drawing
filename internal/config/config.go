// Package config membaca konfigurasi aplikasi dari environment variable dan file .env.
package config

import (
	"bufio"
	"os"
	"strings"
)

// Config menyimpan seluruh konfigurasi yang dibutuhkan server.
type Config struct {
	// Port adalah port HTTP server, contoh "8080".
	Port string
	// AppEnv menentukan mode Gin ("development", "test", atau "production").
	AppEnv string
	// DatabaseURL adalah connection string PostgreSQL (contoh: Neon DB).
	DatabaseURL string
}

func init() {
	loadDotEnv(".env")
}

// loadDotEnv membaca file .env sederhana jika file tersebut ada dan mengisi environment variable yang belum diset.
func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		// Coba di parent directory bila dijalankan dari subfolder
		file, err = os.Open("../" + filepath)
		if err != nil {
			return
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)

			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

// Load membaca konfigurasi dari environment dengan nilai default yang aman.
func Load() Config {
	return Config{
		Port:        envOrDefault("PORT", "8080"),
		AppEnv:      strings.ToLower(envOrDefault("APP_ENV", "development")),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
