// Package auth menangani hashing password, manajemen sesi berbasis cookie HttpOnly,
// registrasi, login, migrasi/penautan akun guest, dan perlindungan rate-limit.
package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"drawingchaos/internal/db"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
	ErrRateLimit  = errors.New("terlalu banyak percobaan autentikasi, silakan coba beberapa saat lagi")
)

// SessionRecord menyimpan informasi sesi pengguna.
type SessionRecord struct {
	ID        string    `json:"id"`
	PlayerID  string    `json:"playerId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// UserProfile merepresentasikan data profil lengkap pengguna yang telah terautentikasi.
type UserProfile struct {
	ID          string                `json:"id"`
	Username    string                `json:"username"`
	Email       string                `json:"email"`
	DisplayName string                `json:"displayName"`
	Avatar      string                `json:"avatar"`
	IsGuest     bool                  `json:"isGuest"`
	Stats       *db.PlayerStatsRecord `json:"stats,omitempty"`
}

// RateLimiter sederhana berbasis sliding window per IP.
type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

var GlobalAuthLimiter = NewRateLimiter(15, 1*time.Minute)

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow memeriksa apakah IP diizinkan melakukan request auth.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Bersihkan riwayat lama
	var valid []time.Time
	for _, t := range rl.attempts[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.attempts[ip] = valid
		return false
	}

	valid = append(valid, now)
	rl.attempts[ip] = valid
	return true
}

// HashPassword melakukan enkripsi password menggunakan bcrypt dengan cost standar (10).
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword memverifikasi kecocokan plaintext password dengan bcrypt hash.
func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidateUsername memverifikasi aturan penamaan username (3-32 karakter, alfanumerik & underscore).
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 32 {
		return errors.New("username harus berukuran antara 3 hingga 32 karakter")
	}
	if !usernameRegex.MatchString(username) {
		return errors.New("username hanya boleh berisi huruf, angka, dan underscore (_)")
	}
	return nil
}

// ValidateEmail memverifikasi keabsahan format email.
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email tidak boleh kosong")
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return errors.New("format email tidak valid")
	}
	return nil
}

// ValidatePassword memeriksa panjang minimal password.
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("password minimal harus 6 karakter")
	}
	if len(password) > 72 {
		return errors.New("password maksimal 72 karakter")
	}
	return nil
}

// GenerateSessionToken membuat string acak kriptografis 32-byte (64 hex characters).
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("gagal menghasilkan token sesi: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession mencatat token sesi baru di PostgreSQL dengan masa kedaluwarsa 30 hari.
func CreateSession(ctx context.Context, database *db.DBWrapper, playerID string, duration time.Duration) (string, error) {
	if database == nil || database.DB == nil {
		// Jika offline/tanpa database, buat token dummy
		return GenerateSessionToken()
	}

	token, err := GenerateSessionToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(duration)
	query := `
	INSERT INTO sessions (id, player_id, created_at, expires_at)
	VALUES ($1, $2, NOW(), $3);
	`
	_, err = database.DB.ExecContext(ctx, query, token, playerID, expiresAt)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan sesi: %w", err)
	}

	return token, nil
}

// ValidateSession memverifikasi apakah token sesi aktif dan belum kedaluwarsa.
func ValidateSession(ctx context.Context, database *db.DBWrapper, token string) (*SessionRecord, error) {
	if database == nil || database.DB == nil {
		return nil, errors.New("database tidak tersedia")
	}

	query := `
	SELECT id, player_id, created_at, expires_at
	FROM sessions
	WHERE id = $1 AND expires_at > NOW();
	`
	var s SessionRecord
	err := database.DB.QueryRowContext(ctx, query, token).Scan(
		&s.ID, &s.PlayerID, &s.CreatedAt, &s.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// RevokeSession menghapus token sesi aktif saat pengguna logout.
func RevokeSession(ctx context.Context, database *db.DBWrapper, token string) error {
	if database == nil || database.DB == nil {
		return nil
	}

	query := `DELETE FROM sessions WHERE id = $1;`
	_, err := database.DB.ExecContext(ctx, query, token)
	return err
}

// RegisterParams parameter pendaftaran akun baru atau penautan akun guest.
type RegisterParams struct {
	Username    string
	Email       string
	Password    string
	Avatar      string
	LinkGuestID string // Jika diisi, meng-upgrade akun guest tersebut menjadi akun permanen
}

// Register mendaftarkan akun permanen baru atau meningkatkan (upgrade) akun guest yang ada.
func Register(ctx context.Context, database *db.DBWrapper, params RegisterParams) (*UserProfile, string, error) {
	if database == nil || database.DB == nil {
		return nil, "", errors.New("database tidak terhubung")
	}

	params.Username = strings.TrimSpace(params.Username)
	params.Email = strings.TrimSpace(strings.ToLower(params.Email))
	if params.Avatar == "" {
		params.Avatar = "cat-orange"
	}

	if err := ValidateUsername(params.Username); err != nil {
		return nil, "", err
	}
	if err := ValidateEmail(params.Email); err != nil {
		return nil, "", err
	}
	if err := ValidatePassword(params.Password); err != nil {
		return nil, "", err
	}

	// 1. Periksa keunikan username dan email
	var existsID string
	err := database.DB.QueryRowContext(ctx, `SELECT id FROM players WHERE LOWER(username) = LOWER($1) LIMIT 1;`, params.Username).Scan(&existsID)
	if err == nil {
		return nil, "", errors.New("username sudah digunakan oleh pemain lain")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, "", fmt.Errorf("gagal cek username: %w", err)
	}

	err = database.DB.QueryRowContext(ctx, `SELECT id FROM players WHERE LOWER(email) = LOWER($1) LIMIT 1;`, params.Email).Scan(&existsID)
	if err == nil {
		return nil, "", errors.New("email sudah terdaftar, silakan gunakan email lain atau login")
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, "", fmt.Errorf("gagal cek email: %w", err)
	}

	// 2. Hash password
	pwdHash, err := HashPassword(params.Password)
	if err != nil {
		return nil, "", fmt.Errorf("gagal mengenkripsi password: %w", err)
	}

	// Periksa terlebih dahulu apakah guest ID dapat ditautkan (sebelum buka transaksi)
	isGuestLinkable := false
	if params.LinkGuestID != "" {
		var isGuest bool
		if err := database.DB.QueryRowContext(ctx, `SELECT is_guest FROM players WHERE id = $1;`, params.LinkGuestID).Scan(&isGuest); err == nil && isGuest {
			isGuestLinkable = true
		}
	}

	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, "", fmt.Errorf("gagal memulai transaksi pendaftaran: %w", err)
	}
	defer tx.Rollback()

	playerID := ""

	if isGuestLinkable {
		upgradeQuery := `
		UPDATE players
		SET username = $1,
		    email = $2,
		    password_hash = $3,
		    display_name = $1,
		    avatar = COALESCE(NULLIF($4, ''), avatar),
		    is_guest = FALSE,
		    last_login_at = NOW(),
		    updated_at = NOW()
		WHERE id = $5;
		`
		if _, err := tx.ExecContext(ctx, upgradeQuery, params.Username, params.Email, pwdHash, params.Avatar, params.LinkGuestID); err != nil {
			return nil, "", fmt.Errorf("gagal menautkan akun guest: %w", err)
		}
		playerID = params.LinkGuestID
	} else {
		playerID = fmt.Sprintf("u_%d", time.Now().UnixNano())
		insertPlayer := `
		INSERT INTO players (id, username, email, password_hash, display_name, avatar, is_guest, last_login_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, FALSE, NOW(), NOW(), NOW());
		`
		if _, err := tx.ExecContext(ctx, insertPlayer, playerID, params.Username, params.Email, pwdHash, params.Username, params.Avatar); err != nil {
			return nil, "", fmt.Errorf("gagal membuat akun pemain: %w", err)
		}

		// Inisialisasi statistik default
		insertStats := `
		INSERT INTO player_stats (player_id, rating, rank_tier)
		VALUES ($1, 1000, 'Bronze')
		ON CONFLICT (player_id) DO NOTHING;
		`
		if _, err := tx.ExecContext(ctx, insertStats, playerID); err != nil {
			return nil, "", fmt.Errorf("gagal inisialisasi stats: %w", err)
		}

		// Inisialisasi saldo 100 koin awal
		insertWallet := `
		INSERT INTO player_wallets (player_id, coins)
		VALUES ($1, 100)
		ON CONFLICT (player_id) DO NOTHING;
		`
		if _, err := tx.ExecContext(ctx, insertWallet, playerID); err != nil {
			return nil, "", fmt.Errorf("gagal inisialisasi dompet: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, "", fmt.Errorf("gagal commit pendaftaran: %w", err)
	}

	// 5. Buat sesi baru (30 hari)
	sessionToken, err := CreateSession(ctx, database, playerID, 30*24*time.Hour)
	if err != nil {
		return nil, "", fmt.Errorf("gagal membuat sesi: %w", err)
	}

	stats, _ := database.GetPlayerStats(ctx, playerID)

	profile := &UserProfile{
		ID:          playerID,
		Username:    params.Username,
		Email:       params.Email,
		DisplayName: params.Username,
		Avatar:      params.Avatar,
		IsGuest:     false,
		Stats:       stats,
	}

	return profile, sessionToken, nil
}

// Login memverifikasi kredensial username/email dan password, lalu membuat sesi baru.
func Login(ctx context.Context, database *db.DBWrapper, identifier, password string) (*UserProfile, string, error) {
	if database == nil || database.DB == nil {
		return nil, "", errors.New("database tidak terhubung")
	}

	identifier = strings.TrimSpace(identifier)
	if identifier == "" || password == "" {
		return nil, "", errors.New("username/email dan password wajib diisi")
	}

	query := `
	SELECT id, username, email, display_name, avatar, password_hash, is_guest, account_status
	FROM players
	WHERE (LOWER(username) = LOWER($1) OR LOWER(email) = LOWER($1))
	LIMIT 1;
	`
	var (
		id            string
		username      sql.NullString
		email         sql.NullString
		displayName   string
		avatar        string
		passwordHash  sql.NullString
		isGuest       bool
		accountStatus string
	)

	err := database.DB.QueryRowContext(ctx, query, identifier).Scan(
		&id, &username, &email, &displayName, &avatar, &passwordHash, &isGuest, &accountStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", errors.New("kredensial login tidak valid")
		}
		return nil, "", fmt.Errorf("gagal memeriksa akun: %w", err)
	}

	if accountStatus != "active" {
		return nil, "", errors.New("akun Anda sedang dinonaktifkan")
	}

	if !passwordHash.Valid || passwordHash.String == "" {
		return nil, "", errors.New("kredensial login tidak valid")
	}

	if !CheckPassword(passwordHash.String, password) {
		return nil, "", errors.New("kredensial login tidak valid")
	}

	// Update last_login_at
	_, _ = database.DB.ExecContext(ctx, `UPDATE players SET last_login_at = NOW(), updated_at = NOW() WHERE id = $1;`, id)

	// Buat sesi baru (30 hari)
	sessionToken, err := CreateSession(ctx, database, id, 30*24*time.Hour)
	if err != nil {
		return nil, "", fmt.Errorf("gagal membuat sesi: %w", err)
	}

	stats, _ := database.GetPlayerStats(ctx, id)

	profile := &UserProfile{
		ID:          id,
		Username:    username.String,
		Email:       email.String,
		DisplayName: displayName,
		Avatar:      avatar,
		IsGuest:     isGuest,
		Stats:       stats,
	}

	return profile, sessionToken, nil
}

// GetUserBySession mengambil profil pengguna berdasarkan token sesi yang valid.
func GetUserBySession(ctx context.Context, database *db.DBWrapper, sessionToken string) (*UserProfile, error) {
	if database == nil || database.DB == nil {
		return nil, errors.New("database tidak tersedia")
	}

	session, err := ValidateSession(ctx, database, sessionToken)
	if err != nil {
		return nil, err
	}

	query := `
	SELECT id, username, email, display_name, avatar, is_guest
	FROM players
	WHERE id = $1;
	`
	var (
		id          string
		username    sql.NullString
		email       sql.NullString
		displayName string
		avatar      string
		isGuest     bool
	)
	err = database.DB.QueryRowContext(ctx, query, session.PlayerID).Scan(
		&id, &username, &email, &displayName, &avatar, &isGuest,
	)
	if err != nil {
		return nil, err
	}

	stats, _ := database.GetPlayerStats(ctx, id)

	return &UserProfile{
		ID:          id,
		Username:    username.String,
		Email:       email.String,
		DisplayName: displayName,
		Avatar:      avatar,
		IsGuest:     isGuest,
		Stats:       stats,
	}, nil
}
