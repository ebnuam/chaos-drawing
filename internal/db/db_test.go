package db_test

import (
	"testing"

	"drawingchaos/internal/db"
)

func TestConnectEmptyURL(t *testing.T) {
	wrapper, err := db.Connect("")
	if err != nil {
		t.Fatalf("expected no error for empty url, got: %v", err)
	}
	if wrapper != nil {
		t.Errorf("expected nil wrapper for empty url")
	}

	// Migrate on nil wrapper should return nil without panic
	if err := wrapper.Migrate(); err != nil {
		t.Errorf("expected no error migrating nil wrapper, got: %v", err)
	}
}
