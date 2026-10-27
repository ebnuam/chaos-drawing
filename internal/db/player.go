package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// PlayerRecord merepresentasikan data tabel players.
type PlayerRecord struct {
	ID          string    `json:"id"`
	Username    string    `json:"username,omitempty"`
	Email       string    `json:"email,omitempty"`
	DisplayName string    `json:"displayName"`
	Avatar      string    `json:"avatar"`
	IsGuest     bool      `json:"isGuest"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// MatchHistoryEntry merepresentasikan riwayat pertandingan satu pemain.
type MatchHistoryEntry struct {
	MatchID      string    `json:"matchId"`
	RoomCode     string    `json:"roomCode"`
	FinalScore   int       `json:"finalScore"`
	RankPosition int       `json:"rankPosition"`
	CoinsEarned  int       `json:"coinsEarned"`
	RatingChange int       `json:"ratingChange"`
	CreatedAt    time.Time `json:"createdAt"`
}

// PlayerStatsRecord merepresentasikan statistik kumulatif pemain.
type PlayerStatsRecord struct {
	PlayerID           string    `json:"playerId"`
	MatchesPlayed      int       `json:"matchesPlayed"`
	MatchesWon         int       `json:"matchesWon"`
	MatchesLost        int       `json:"matchesLost"`
	TotalScore         int       `json:"totalScore"`
	CorrectGuesses     int       `json:"correctGuesses"`
	GuessAttempts      int       `json:"guessAttempts"`
	GuessAccuracy      float64   `json:"guessAccuracy"`
	DrawingsMade       int       `json:"drawingsMade"`
	SuccessfulDrawings int       `json:"successfulDrawings"`
	Rating             int       `json:"rating"`
	RankTier           string    `json:"rankTier"`
	Coins              int       `json:"coins"`
	WinRate            float64   `json:"winRate"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// PlayerProfileResponse membungkus seluruh data profil, statistik, dan riwayat pertandingan.
type PlayerProfileResponse struct {
	Player        PlayerRecord        `json:"player"`
	Stats         PlayerStatsRecord   `json:"stats"`
	RecentMatches []MatchHistoryEntry `json:"recentMatches"`
}

// EnsurePlayer mendaftarkan atau memperbarui pemain di database serta membuat wallet & stats awal jika belum ada.
func (w *DBWrapper) EnsurePlayer(ctx context.Context, id, displayName, avatar string) (*PlayerRecord, error) {
	if w == nil || w.DB == nil {
		return nil, nil
	}

	tx, err := w.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi player: %w", err)
	}
	defer tx.Rollback()

	// 1. Upsert pemain
	queryPlayer := `
	INSERT INTO players (id, display_name, avatar, last_active_at, updated_at)
	VALUES ($1, $2, $3, NOW(), NOW())
	ON CONFLICT (id) DO UPDATE SET
		display_name = CASE WHEN players.username IS NOT NULL THEN players.username ELSE EXCLUDED.display_name END,
		avatar = EXCLUDED.avatar,
		last_active_at = NOW(),
		updated_at = NOW()
	RETURNING id, COALESCE(username, ''), COALESCE(email, ''), display_name, avatar, is_guest, created_at, updated_at;
	`
	var p PlayerRecord
	err = tx.QueryRowContext(ctx, queryPlayer, id, displayName, avatar).Scan(
		&p.ID, &p.Username, &p.Email, &p.DisplayName, &p.Avatar, &p.IsGuest, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal upsert pemain: %w", err)
	}

	// 2. Inisialisasi statistik awal jika belum ada
	queryStats := `
	INSERT INTO player_stats (player_id, rating, rank_tier)
	VALUES ($1, 1000, 'Bronze')
	ON CONFLICT (player_id) DO NOTHING;
	`
	if _, err := tx.ExecContext(ctx, queryStats, id); err != nil {
		return nil, fmt.Errorf("gagal inisialisasi stats: %w", err)
	}

	// 3. Inisialisasi wallet dengan 100 koin awal jika belum ada
	queryWallet := `
	INSERT INTO player_wallets (player_id, coins)
	VALUES ($1, 100)
	ON CONFLICT (player_id) DO NOTHING;
	`
	if _, err := tx.ExecContext(ctx, queryWallet, id); err != nil {
		return nil, fmt.Errorf("gagal inisialisasi wallet: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal commit player registration: %w", err)
	}

	return &p, nil
}

// GetPlayerStats mengambil profil lengkap dan statistik pemain dari PostgreSQL.
func (w *DBWrapper) GetPlayerStats(ctx context.Context, playerID string) (*PlayerStatsRecord, error) {
	if w == nil || w.DB == nil {
		return nil, sql.ErrNoRows
	}

	query := `
	SELECT 
		s.player_id, s.matches_played, s.matches_won, s.matches_lost, s.total_score,
		s.correct_guesses, s.guess_attempts, s.drawings_made, s.successful_drawings,
		s.rating, s.rank_tier, COALESCE(w.coins, 0), s.updated_at
	FROM player_stats s
	LEFT JOIN player_wallets w ON s.player_id = w.player_id
	WHERE s.player_id = $1;
	`

	var s PlayerStatsRecord
	err := w.DB.QueryRowContext(ctx, query, playerID).Scan(
		&s.PlayerID, &s.MatchesPlayed, &s.MatchesWon, &s.MatchesLost, &s.TotalScore,
		&s.CorrectGuesses, &s.GuessAttempts, &s.DrawingsMade, &s.SuccessfulDrawings,
		&s.Rating, &s.RankTier, &s.Coins, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if s.MatchesPlayed > 0 {
		s.WinRate = float64(s.MatchesWon) / float64(s.MatchesPlayed) * 100
	}
	if s.GuessAttempts > 0 {
		s.GuessAccuracy = float64(s.CorrectGuesses) / float64(s.GuessAttempts) * 100
	}

	return &s, nil
}

// GetPlayerProfile mengambil profil lengkap, statistik, dan riwayat pertandingan pemain.
func (w *DBWrapper) GetPlayerProfile(ctx context.Context, playerID string) (*PlayerProfileResponse, error) {
	if w == nil || w.DB == nil {
		return nil, sql.ErrNoRows
	}

	queryPlayer := `
	SELECT id, COALESCE(username, ''), COALESCE(email, ''), display_name, avatar, is_guest, created_at, updated_at
	FROM players
	WHERE id = $1;
	`
	var p PlayerRecord
	err := w.DB.QueryRowContext(ctx, queryPlayer, playerID).Scan(
		&p.ID, &p.Username, &p.Email, &p.DisplayName, &p.Avatar, &p.IsGuest, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	stats, err := w.GetPlayerStats(ctx, playerID)
	if err != nil {
		stats = &PlayerStatsRecord{
			PlayerID: playerID,
			Rating:   1000,
			RankTier: "Bronze",
			Coins:    100,
		}
	}

	// Ambil 10 pertandingan terakhir
	queryMatches := `
	SELECT mp.match_id, COALESCE(m.room_code, ''), mp.final_score, mp.rank_position, mp.coins_earned, mp.rating_change, mp.created_at
	FROM match_players mp
	LEFT JOIN matches m ON mp.match_id = m.id
	WHERE mp.player_id = $1
	ORDER BY mp.created_at DESC
	LIMIT 10;
	`
	rows, err := w.DB.QueryContext(ctx, queryMatches, playerID)
	var recentMatches []MatchHistoryEntry
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var m MatchHistoryEntry
			if err := rows.Scan(&m.MatchID, &m.RoomCode, &m.FinalScore, &m.RankPosition, &m.CoinsEarned, &m.RatingChange, &m.CreatedAt); err == nil {
				recentMatches = append(recentMatches, m)
			}
		}
	}

	return &PlayerProfileResponse{
		Player:        p,
		Stats:         *stats,
		RecentMatches: recentMatches,
	}, nil
}
