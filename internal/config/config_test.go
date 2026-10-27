package config_test

import (
	"os"
	"testing"

	"drawingchaos/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("DATABASE_URL")

	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Errorf("expected default Port '8080', got '%s'", cfg.Port)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("expected default AppEnv 'development', got '%s'", cfg.AppEnv)
	}
	if cfg.DatabaseURL != "" {
		t.Errorf("expected empty DatabaseURL, got '%s'", cfg.DatabaseURL)
	}
}

func TestLoadCustomEnv(t *testing.T) {
	os.Setenv("PORT", "3000")
	os.Setenv("APP_ENV", "PRODUCTION")
	os.Setenv("DATABASE_URL", "postgres://test")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("APP_ENV")
		os.Unsetenv("DATABASE_URL")
	}()

	cfg := config.Load()

	if cfg.Port != "3000" {
		t.Errorf("expected Port '3000', got '%s'", cfg.Port)
	}
	if cfg.AppEnv != "production" {
		t.Errorf("expected lowercase AppEnv 'production', got '%s'", cfg.AppEnv)
	}
	if cfg.DatabaseURL != "postgres://test" {
		t.Errorf("expected DatabaseURL 'postgres://test', got '%s'", cfg.DatabaseURL)
	}
}
