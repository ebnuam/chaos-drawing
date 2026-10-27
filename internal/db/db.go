// Package db mengelola koneksi database PostgreSQL, migrasi skema tabel, dan transaksi data.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DBWrapper membungkus koneksi database sql.DB dengan metode helper.
type DBWrapper struct {
	*sql.DB
}

var globalDB *DBWrapper

// GetDB mengembalikan instance global database.
func GetDB() *DBWrapper {
	return globalDB
}

// Connect menginisialisasi koneksi PostgreSQL menggunakan connection pool dan driver pgx.
// Jika databaseURL kosong, fungsi ini mengembalikan nil tanpa error untuk mode graceful fallback (offline/in-memory).
func Connect(databaseURL string) (*DBWrapper, error) {
	if databaseURL == "" {
		log.Println("[Database] DATABASE_URL tidak diset. Menjalankan dalam mode memori (tanpa database).")
		return nil, nil
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi database: %w", err)
	}

	// Konfigurasi connection pooling yang optimal untuk Neon serverless
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping ke database gagal: %w", err)
	}

	log.Println("[Database] Berhasil terhubung ke PostgreSQL Neon.")
	wrapper := &DBWrapper{DB: db}
	globalDB = wrapper
	return wrapper, nil
}

// Migrate menjalankan migrasi skema tabel secara idempotent (aman dijalankan berulang kali).
func (w *DBWrapper) Migrate() error {
	if w == nil || w.DB == nil {
		return nil
	}

	schemaSQL := `
	-- 1. Tabel pemain (players)
	CREATE TABLE IF NOT EXISTS players (
		id VARCHAR(64) PRIMARY KEY,
		display_name VARCHAR(32) NOT NULL,
		avatar VARCHAR(32) NOT NULL DEFAULT 'cat-orange',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	-- 2. Statistik pemain (player_stats)
	CREATE TABLE IF NOT EXISTS player_stats (
		player_id VARCHAR(64) PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
		matches_played INT NOT NULL DEFAULT 0,
		matches_won INT NOT NULL DEFAULT 0,
		total_score INT NOT NULL DEFAULT 0,
		correct_guesses INT NOT NULL DEFAULT 0,
		drawings_made INT NOT NULL DEFAULT 0,
		successful_drawings INT NOT NULL DEFAULT 0,
		rating INT NOT NULL DEFAULT 1000,
		rank_tier VARCHAR(16) NOT NULL DEFAULT 'Bronze',
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_player_stats_rating ON player_stats(rating DESC);
	CREATE INDEX IF NOT EXISTS idx_player_stats_matches_won ON player_stats(matches_won DESC);

	-- 3. Dompet koin pemain (player_wallets)
	CREATE TABLE IF NOT EXISTS player_wallets (
		player_id VARCHAR(64) PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
		coins INT NOT NULL DEFAULT 100 CHECK (coins >= 0),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	-- 4. Pertandingan (matches)
	CREATE TABLE IF NOT EXISTS matches (
		id VARCHAR(64) PRIMARY KEY,
		room_code VARCHAR(8) NOT NULL,
		total_rounds INT NOT NULL,
		round_duration INT NOT NULL,
		winner_player_id VARCHAR(64) REFERENCES players(id) ON DELETE SET NULL,
		status VARCHAR(16) NOT NULL DEFAULT 'in_progress',
		started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		finished_at TIMESTAMPTZ
	);
	CREATE INDEX IF NOT EXISTS idx_matches_started_at ON matches(started_at DESC);

	-- 5. Partisipan pertandingan (match_players)
	CREATE TABLE IF NOT EXISTS match_players (
		id BIGSERIAL PRIMARY KEY,
		match_id VARCHAR(64) NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
		player_id VARCHAR(64) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
		final_score INT NOT NULL DEFAULT 0,
		rank_position INT NOT NULL DEFAULT 1,
		coins_earned INT NOT NULL DEFAULT 0,
		rating_change INT NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE (match_id, player_id)
	);
	CREATE INDEX IF NOT EXISTS idx_match_players_player_id ON match_players(player_id);

	-- 6. Detail ronde pertandingan (rounds)
	CREATE TABLE IF NOT EXISTS rounds (
		id BIGSERIAL PRIMARY KEY,
		match_id VARCHAR(64) NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
		round_number INT NOT NULL,
		drawer_player_id VARCHAR(64) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
		word VARCHAR(64) NOT NULL,
		category VARCHAR(32) NOT NULL,
		correct_guesses_count INT NOT NULL DEFAULT 0,
		duration_seconds INT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE (match_id, round_number)
	);

	-- 7. Katalog item chaos (item_catalog)
	CREATE TABLE IF NOT EXISTS item_catalog (
		id VARCHAR(32) PRIMARY KEY,
		name VARCHAR(64) NOT NULL,
		description TEXT NOT NULL,
		cost_coins INT NOT NULL CHECK (cost_coins >= 0),
		duration_seconds INT NOT NULL DEFAULT 5,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	-- 8. Penggunaan item (item_usage)
	CREATE TABLE IF NOT EXISTS item_usage (
		id BIGSERIAL PRIMARY KEY,
		match_id VARCHAR(64) REFERENCES matches(id) ON DELETE SET NULL,
		round_number INT,
		user_player_id VARCHAR(64) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
		target_player_id VARCHAR(64) REFERENCES players(id) ON DELETE SET NULL,
		item_id VARCHAR(32) NOT NULL REFERENCES item_catalog(id),
		cost_coins INT NOT NULL,
		used_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_item_usage_user ON item_usage(user_player_id);

	-- 9. Transaksi koin (coin_transactions)
	CREATE TABLE IF NOT EXISTS coin_transactions (
		id BIGSERIAL PRIMARY KEY,
		player_id VARCHAR(64) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
		amount INT NOT NULL,
		balance_after INT NOT NULL CHECK (balance_after >= 0),
		source VARCHAR(32) NOT NULL,
		match_id VARCHAR(64) REFERENCES matches(id) ON DELETE SET NULL,
		reference_id VARCHAR(64),
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE (player_id, reference_id)
	);
	CREATE INDEX IF NOT EXISTS idx_coin_transactions_player ON coin_transactions(player_id, created_at DESC);

	-- 10. Riwayat rank (rank_history)
	CREATE TABLE IF NOT EXISTS rank_history (
		id BIGSERIAL PRIMARY KEY,
		player_id VARCHAR(64) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
		match_id VARCHAR(64) NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
		old_rating INT NOT NULL,
		new_rating INT NOT NULL,
		rating_change INT NOT NULL,
		old_rank_tier VARCHAR(16) NOT NULL,
		new_rank_tier VARCHAR(16) NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_rank_history_player ON rank_history(player_id, created_at DESC);

	-- 11. Kolom tambahan autentikasi & akun permanen
	ALTER TABLE players ADD COLUMN IF NOT EXISTS username VARCHAR(32) UNIQUE;
	ALTER TABLE players ADD COLUMN IF NOT EXISTS email VARCHAR(128) UNIQUE;
	ALTER TABLE players ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);
	ALTER TABLE players ADD COLUMN IF NOT EXISTS is_guest BOOLEAN NOT NULL DEFAULT TRUE;
	ALTER TABLE players ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;
	ALTER TABLE players ADD COLUMN IF NOT EXISTS account_status VARCHAR(16) NOT NULL DEFAULT 'active';
	CREATE INDEX IF NOT EXISTS idx_players_username ON players(LOWER(username));
	CREATE INDEX IF NOT EXISTS idx_players_email ON players(LOWER(email));

	-- 12. Kolom tambahan statistik detail tebakan & kekalahan
	ALTER TABLE player_stats ADD COLUMN IF NOT EXISTS matches_lost INT NOT NULL DEFAULT 0;
	ALTER TABLE player_stats ADD COLUMN IF NOT EXISTS guess_attempts INT NOT NULL DEFAULT 0;

	-- 13. Sesi login pemain (sessions)
	CREATE TABLE IF NOT EXISTS sessions (
		id VARCHAR(64) PRIMARY KEY,
		player_id VARCHAR(64) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		expires_at TIMESTAMPTZ NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_sessions_player_id ON sessions(player_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

	-- 14. Bank kata terpusat (words & word_clues)
	CREATE TABLE IF NOT EXISTS words (
		id SERIAL PRIMARY KEY,
		word VARCHAR(64) UNIQUE NOT NULL,
		category VARCHAR(32) NOT NULL,
		difficulty VARCHAR(16) NOT NULL DEFAULT 'medium',
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_words_category ON words(category);

	CREATE TABLE IF NOT EXISTS word_clues (
		id BIGSERIAL PRIMARY KEY,
		word_id INT NOT NULL REFERENCES words(id) ON DELETE CASCADE,
		clue_text TEXT NOT NULL,
		clue_order INT NOT NULL DEFAULT 1,
		UNIQUE (word_id, clue_order)
	);
	CREATE INDEX IF NOT EXISTS idx_word_clues_word_id ON word_clues(word_id);

	-- 15. Hapus foreign key ketat match_id pada transaksi koin dan item_usage agar tidak terjadi race condition saat match in-progress
	ALTER TABLE coin_transactions DROP CONSTRAINT IF EXISTS coin_transactions_match_id_fkey;
	ALTER TABLE item_usage DROP CONSTRAINT IF EXISTS item_usage_match_id_fkey;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := w.DB.ExecContext(ctx, schemaSQL)
	if err != nil {
		return fmt.Errorf("eksekusi migrasi skema gagal: %w", err)
	}

	// Seed 7 katalog item chaos
	seedItemsSQL := `
	INSERT INTO item_catalog (id, name, description, cost_coins, duration_seconds, is_active)
	VALUES 
		('oil_splash', 'Lempar Oli', 'Menutupi sebagian kanvas penggambar dengan cipratan oli hitam pekat selama 5 detik.', 25, 5, TRUE),
		('ink_bomb', 'Bom Tinta', 'Ledakan tinta yang mengaburkan pandangan kanvas penggambar selama 6 detik.', 35, 6, TRUE),
		('fog_screen', 'Kabut Tebal', 'Menutupi area gambar dengan kabut asap tebal selama 7 detik.', 30, 7, TRUE),
		('annoying_fly', 'Lalat Pengganggu', 'Lalat animasi mendengung yang terbang acak mengganggu pandangan kanvas selama 8 detik.', 40, 8, TRUE),
		('canvas_earthquake', 'Gempa Kanvas', 'Mengguncang kanvas gambar dengan getaran gempa dahsyat selama 5 detik.', 45, 5, TRUE),
		('effect_cleaner', 'Pembersih Efek', 'Membersihkan seketika seluruh efek gangguan chaos di kanvas kamu!', 20, 0, TRUE),
		('turbo_time', 'Mode Turbo', 'Menambah waktu ronde 10 detik dan menggandakan skor tebakan.', 50, 10, TRUE)
	ON CONFLICT (id) DO UPDATE SET
		name = EXCLUDED.name,
		description = EXCLUDED.description,
		cost_coins = EXCLUDED.cost_coins,
		duration_seconds = EXCLUDED.duration_seconds,
		is_active = EXCLUDED.is_active;
	`
	if _, err := w.DB.ExecContext(ctx, seedItemsSQL); err != nil {
		log.Printf("[Database] Peringatan: seed katalog item gagal: %v", err)
	}

	// Seed bank kata terpusat (110 kata dalam 10 kategori)
	if err := w.SeedWords(ctx); err != nil {
		log.Printf("[Database] Peringatan: seed bank kata gagal: %v", err)
	}

	log.Println("[Database] Migrasi skema tabel & seed data berhasil dijalankan secara aman.")
	return nil
}
