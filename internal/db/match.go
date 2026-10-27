package db

import (
	"context"
	"fmt"
)

// MatchPlayerResult merepresentasikan data statistik dan reward akhir satu pemain di pertandingan.
type MatchPlayerResult struct {
	PlayerID           string
	FinalScore         int
	RankPosition       int
	CoinsEarned        int
	RatingChange       int
	OldRating          int
	NewRating          int
	OldTier            string
	NewTier            string
	IsWinner           bool
	CorrectGuesses     int
	GuessAttempts      int
	DrawingsMade       int
	SuccessfulDrawings int
}

// RoundResult merepresentasikan ringkasan satu ronde permainan.
type RoundResult struct {
	RoundNumber         int
	DrawerPlayerID      string
	Word                string
	Category            string
	CorrectGuessesCount int
	DurationSeconds     int
}

// MatchRecordParams membungkus seluruh data pertandingan untuk disimpan permanen.
type MatchRecordParams struct {
	MatchID        string
	RoomCode       string
	TotalRounds    int
	RoundDuration  int
	WinnerPlayerID string
	Players        []MatchPlayerResult
	Rounds         []RoundResult
}

// DetermineRankTier mengonversi rating numerik menjadi tier rank kompetitif.
func DetermineRankTier(rating int) string {
	switch {
	case rating >= 2000:
		return "Diamond"
	case rating >= 1600:
		return "Platinum"
	case rating >= 1300:
		return "Gold"
	case rating >= 1100:
		return "Silver"
	default:
		return "Bronze"
	}
}

// CreateMatch mencatat pertandingan baru saat permainan dimulai.
func (w *DBWrapper) CreateMatch(ctx context.Context, matchID, roomCode string, totalRounds, roundDuration int) error {
	if w == nil || w.DB == nil {
		return nil
	}

	query := `
	INSERT INTO matches (id, room_code, total_rounds, round_duration, status, started_at)
	VALUES ($1, $2, $3, $4, 'in_progress', NOW())
	ON CONFLICT (id) DO NOTHING;
	`
	_, err := w.DB.ExecContext(ctx, query, matchID, roomCode, totalRounds, roundDuration)
	return err
}

// RecordMatch menyimpan seluruh data akhir pertandingan, statistik pemain, dan reward koin dalam satu transaksi atomik.
func (w *DBWrapper) RecordMatch(ctx context.Context, params MatchRecordParams) error {
	if w == nil || w.DB == nil {
		return nil
	}

	tx, err := w.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi record match: %w", err)
	}
	defer tx.Rollback()

	// 1. Simpan atau perbarui tabel matches
	insertMatchQuery := `
	INSERT INTO matches (id, room_code, total_rounds, round_duration, winner_player_id, status, started_at, finished_at)
	VALUES ($1, $2, $3, $4, NULLIF($5, ''), 'finished', NOW() - INTERVAL '5 minutes', NOW())
	ON CONFLICT (id) DO UPDATE SET
		winner_player_id = EXCLUDED.winner_player_id,
		status = 'finished',
		finished_at = NOW();
	`
	_, err = tx.ExecContext(ctx, insertMatchQuery,
		params.MatchID, params.RoomCode, params.TotalRounds, params.RoundDuration, params.WinnerPlayerID,
	)
	if err != nil {
		return fmt.Errorf("gagal insert matches: %w", err)
	}

	// 2. Simpan tiap partisipan & update statistik
	for _, p := range params.Players {
		// Insert match_players
		insertMP := `
		INSERT INTO match_players (match_id, player_id, final_score, rank_position, coins_earned, rating_change, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (match_id, player_id) DO NOTHING;
		`
		if _, err := tx.ExecContext(ctx, insertMP, params.MatchID, p.PlayerID, p.FinalScore, p.RankPosition, p.CoinsEarned, p.RatingChange); err != nil {
			return fmt.Errorf("gagal insert match_players (%s): %w", p.PlayerID, err)
		}

		// Update player_stats
		winIncrement := 0
		lossIncrement := 0
		if p.IsWinner {
			winIncrement = 1
		} else {
			lossIncrement = 1
		}

		newTier := DetermineRankTier(p.NewRating)
		updateStats := `
		UPDATE player_stats
		SET matches_played = matches_played + 1,
		    matches_won = matches_won + $1,
		    matches_lost = matches_lost + $2,
		    total_score = total_score + $3,
		    correct_guesses = correct_guesses + $4,
		    guess_attempts = guess_attempts + $5,
		    drawings_made = drawings_made + $6,
		    successful_drawings = successful_drawings + $7,
		    rating = $8,
		    rank_tier = $9,
		    updated_at = NOW()
		WHERE player_id = $10;
		`
		if _, err := tx.ExecContext(ctx, updateStats,
			winIncrement, lossIncrement, p.FinalScore, p.CorrectGuesses, p.GuessAttempts,
			p.DrawingsMade, p.SuccessfulDrawings, p.NewRating, newTier, p.PlayerID,
		); err != nil {
			return fmt.Errorf("gagal update player_stats (%s): %w", p.PlayerID, err)
		}

		// Insert rank_history
		insertRankHist := `
		INSERT INTO rank_history (player_id, match_id, old_rating, new_rating, rating_change, old_rank_tier, new_rank_tier, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW());
		`
		if _, err := tx.ExecContext(ctx, insertRankHist,
			p.PlayerID, params.MatchID, p.OldRating, p.NewRating, p.RatingChange, p.OldTier, newTier,
		); err != nil {
			return fmt.Errorf("gagal insert rank_history (%s): %w", p.PlayerID, err)
		}

		// Tambahkan reward koin ke wallet
		if p.CoinsEarned > 0 {
			var balanceAfter int
			updateWallet := `
			UPDATE player_wallets
			SET coins = coins + $1, updated_at = NOW()
			WHERE player_id = $2
			RETURNING coins;
			`
			err = tx.QueryRowContext(ctx, updateWallet, p.CoinsEarned, p.PlayerID).Scan(&balanceAfter)
			if err != nil {
				return fmt.Errorf("gagal update player_wallets (%s): %w", p.PlayerID, err)
			}

			// Catat mutasi koin dengan reference_id unik untuk mencegah duplikasi (idempotent)
			refID := fmt.Sprintf("match_%s_%s", params.MatchID, p.PlayerID)
			insertCoinTx := `
			INSERT INTO coin_transactions (player_id, amount, balance_after, source, match_id, reference_id, created_at)
			VALUES ($1, $2, $3, 'match_reward', $4, $5, NOW())
			ON CONFLICT (player_id, reference_id) DO NOTHING;
			`
			if _, err := tx.ExecContext(ctx, insertCoinTx, p.PlayerID, p.CoinsEarned, balanceAfter, params.MatchID, refID); err != nil {
				return fmt.Errorf("gagal catat coin_transactions (%s): %w", p.PlayerID, err)
			}
		}
	}

	// 3. Simpan detail tiap ronde
	for _, r := range params.Rounds {
		insertRound := `
		INSERT INTO rounds (match_id, round_number, drawer_player_id, word, category, correct_guesses_count, duration_seconds, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (match_id, round_number) DO NOTHING;
		`
		if _, err := tx.ExecContext(ctx, insertRound,
			params.MatchID, r.RoundNumber, r.DrawerPlayerID, r.Word, r.Category, r.CorrectGuessesCount, r.DurationSeconds,
		); err != nil {
			return fmt.Errorf("gagal insert rounds (%d): %w", r.RoundNumber, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("gagal commit transaksi record match: %w", err)
	}

	return nil
}

// LeaderboardEntry merepresentasikan satu baris pada papan peringkat global.
type LeaderboardEntry struct {
	RankPosition  int    `json:"rankPosition"`
	PlayerID      string `json:"playerId"`
	DisplayName   string `json:"displayName"`
	Avatar        string `json:"avatar"`
	Rating        int    `json:"rating"`
	RankTier      string `json:"rankTier"`
	MatchesPlayed int    `json:"matchesPlayed"`
	MatchesWon    int    `json:"matchesWon"`
	TotalScore    int    `json:"totalScore"`
}

// GetLeaderboard mengambil peringkat teratas pemain berdasarkan rating atau jumlah kemenangan.
func (w *DBWrapper) GetLeaderboard(ctx context.Context, sortBy string, limit int) ([]LeaderboardEntry, error) {
	if w == nil || w.DB == nil {
		return []LeaderboardEntry{}, nil
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	orderBy := "s.rating DESC"
	if sortBy == "wins" {
		orderBy = "s.matches_won DESC, s.rating DESC"
	} else if sortBy == "score" {
		orderBy = "s.total_score DESC"
	}

	query := fmt.Sprintf(`
	SELECT 
		p.id, p.display_name, p.avatar,
		s.rating, s.rank_tier, s.matches_played, s.matches_won, s.total_score
	FROM player_stats s
	JOIN players p ON s.player_id = p.id
	ORDER BY %s
	LIMIT $1;
	`, orderBy)

	rows, err := w.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("gagal query leaderboard: %w", err)
	}
	defer rows.Close()

	var list []LeaderboardEntry
	rank := 1
	for rows.Next() {
		var e LeaderboardEntry
		e.RankPosition = rank
		rank++

		if err := rows.Scan(
			&e.PlayerID, &e.DisplayName, &e.Avatar,
			&e.Rating, &e.RankTier, &e.MatchesPlayed, &e.MatchesWon, &e.TotalScore,
		); err != nil {
			return nil, err
		}
		list = append(list, e)
	}

	return list, nil
}
