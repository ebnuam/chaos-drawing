package auth

import (
	"testing"
	"time"
)

func TestPasswordHashingAndVerification(t *testing.T) {
	password := "SecretP@ss123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Gagal hash password: %v", err)
	}

	if hash == password {
		t.Fatalf("Hash tidak boleh sama dengan plaintext password")
	}

	if !CheckPassword(hash, password) {
		t.Errorf("CheckPassword harus bernilai true untuk password yang cocok")
	}

	if CheckPassword(hash, "WrongPassword") {
		t.Errorf("CheckPassword harus bernilai false untuk password yang salah")
	}
}

func TestValidationRules(t *testing.T) {
	// Username tests
	validUsernames := []string{"player1", "user_game", "Budi_123", "abc"}
	for _, u := range validUsernames {
		if err := ValidateUsername(u); err != nil {
			t.Errorf("Username '%s' seharusnya valid, tapi gagal: %v", u, err)
		}
	}

	invalidUsernames := []string{"", "ab", "user name", "user@name!", "this_username_is_way_too_long_to_be_allowed_12345"}
	for _, u := range invalidUsernames {
		if err := ValidateUsername(u); err == nil {
			t.Errorf("Username '%s' seharusnya tidak valid", u)
		}
	}

	// Email tests
	validEmails := []string{"user@example.com", "pemain.game@domain.co.id", "test+1@sub.domain.org"}
	for _, e := range validEmails {
		if err := ValidateEmail(e); err != nil {
			t.Errorf("Email '%s' seharusnya valid, tapi gagal: %v", e, err)
		}
	}

	invalidEmails := []string{"", "user@", "plainaddress", "@domain.com"}
	for _, e := range invalidEmails {
		if err := ValidateEmail(e); err == nil {
			t.Errorf("Email '%s' seharusnya tidak valid", e)
		}
	}

	// Password tests
	if err := ValidatePassword("12345"); err == nil {
		t.Errorf("Password kurang dari 6 karakter seharusnya gagal")
	}
	if err := ValidatePassword("123456"); err != nil {
		t.Errorf("Password 6 karakter seharusnya valid: %v", err)
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(3, 100*time.Millisecond)
	ip := "192.168.1.1"

	for i := 0; i < 3; i++ {
		if !limiter.Allow(ip) {
			t.Errorf("Percobaan ke-%d seharusnya diizinkan", i+1)
		}
	}

	if limiter.Allow(ip) {
		t.Errorf("Percobaan ke-4 seharusnya diblokir oleh rate limiter")
	}

	// Setelah window berlalu
	time.Sleep(120 * time.Millisecond)
	if !limiter.Allow(ip) {
		t.Errorf("Setelah window lewat, seharusnya diizinkan kembali")
	}
}

func TestSessionTokenGeneration(t *testing.T) {
	token1, err1 := GenerateSessionToken()
	token2, err2 := GenerateSessionToken()

	if err1 != nil || err2 != nil {
		t.Fatalf("Gagal generate token: %v, %v", err1, err2)
	}

	if len(token1) != 64 || len(token2) != 64 {
		t.Errorf("Panjang token sesi harus 64 karakter hex (32 bytes)")
	}

	if token1 == token2 {
		t.Errorf("Token sesi harus unik dan acak")
	}
}

