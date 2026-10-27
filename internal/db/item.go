package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ItemRecord merepresentasikan item yang dapat dibeli dengan koin.
type ItemRecord struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	CostCoins       int    `json:"costCoins"`
	DurationSeconds int    `json:"durationSeconds"`
	IsActive        bool   `json:"isActive"`
}

// GetCatalogItems mengambil seluruh item yang sedang aktif di katalog.
func (w *DBWrapper) GetCatalogItems(ctx context.Context) ([]ItemRecord, error) {
	if w == nil || w.DB == nil {
		// Fallback daftar item jika database offline
		return []ItemRecord{
			{
				ID:              "oil_splash",
				Name:            "Lempar Oli",
				Description:     "Menutupi sebagian kanvas penggambar dengan cipratan oli hitam pekat selama 5 detik.",
				CostCoins:       25,
				DurationSeconds: 5,
				IsActive:        true,
			},
			{
				ID:              "ink_bomb",
				Name:            "Bom Tinta",
				Description:     "Ledakan tinta yang mengaburkan pandangan kanvas penggambar selama 6 detik.",
				CostCoins:       35,
				DurationSeconds: 6,
				IsActive:        true,
			},
			{
				ID:              "fog_screen",
				Name:            "Kabut Tebal",
				Description:     "Menutupi area gambar dengan kabut asap tebal selama 7 detik.",
				CostCoins:       30,
				DurationSeconds: 7,
				IsActive:        true,
			},
			{
				ID:              "annoying_fly",
				Name:            "Lalat Pengganggu",
				Description:     "Lalat animasi mendengung yang terbang acak mengganggu pandangan kanvas selama 8 detik.",
				CostCoins:       40,
				DurationSeconds: 8,
				IsActive:        true,
			},
			{
				ID:              "canvas_earthquake",
				Name:            "Gempa Kanvas",
				Description:     "Mengguncang kanvas gambar dengan getaran gempa dahsyat selama 5 detik.",
				CostCoins:       45,
				DurationSeconds: 5,
				IsActive:        true,
			},
			{
				ID:              "effect_cleaner",
				Name:            "Pembersih Efek",
				Description:     "Membersihkan seketika seluruh efek gangguan chaos di kanvas kamu!",
				CostCoins:       20,
				DurationSeconds: 0,
				IsActive:        true,
			},
			{
				ID:              "turbo_time",
				Name:            "Mode Turbo",
				Description:     "Menambah waktu ronde 10 detik dan menggandakan skor tebakan.",
				CostCoins:       50,
				DurationSeconds: 10,
				IsActive:        true,
			},
		}, nil
	}

	query := `
	SELECT id, name, description, cost_coins, duration_seconds, is_active
	FROM item_catalog
	WHERE is_active = TRUE
	ORDER BY cost_coins ASC;
	`

	rows, err := w.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal query katalog item: %w", err)
	}
	defer rows.Close()

	var items []ItemRecord
	for rows.Next() {
		var it ItemRecord
		if err := rows.Scan(&it.ID, &it.Name, &it.Description, &it.CostCoins, &it.DurationSeconds, &it.IsActive); err != nil {
			return nil, err
		}
		items = append(items, it)
	}

	return items, nil
}

// UseItemResult membawa hasil pemakaian item setelah divalidasi dan dipotong koinnya.
type UseItemResult struct {
	ItemID          string `json:"itemId"`
	ItemName        string `json:"itemName"`
	UserPlayerID    string `json:"userPlayerId"`
	TargetPlayerID  string `json:"targetPlayerId"`
	RemainingCoins  int    `json:"remainingCoins"`
	DurationSeconds int    `json:"durationSeconds"`
}

// UseChaosItem memproses pembelian & penggunaan item secara atomik di server.
func (w *DBWrapper) UseChaosItem(ctx context.Context, matchID string, roundNumber int, userID, targetID, itemID string) (*UseItemResult, error) {
	if w == nil || w.DB == nil {
		// Fallback jika tanpa database
		return &UseItemResult{
			ItemID:          itemID,
			ItemName:        "Lempar Oli",
			UserPlayerID:    userID,
			TargetPlayerID:  targetID,
			RemainingCoins:  100,
			DurationSeconds: 5,
		}, nil
	}

	tx, err := w.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi pemakaian item: %w", err)
	}
	defer tx.Rollback()

	// 1. Ambil detail item dari katalog
	var itemName string
	var costCoins int
	var duration int
	queryItem := `
	SELECT name, cost_coins, duration_seconds
	FROM item_catalog
	WHERE id = $1 AND is_active = TRUE;
	`
	err = tx.QueryRowContext(ctx, queryItem, itemID).Scan(&itemName, &costCoins, &duration)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item '%s' tidak ditemukan atau sedang tidak aktif", itemID)
		}
		return nil, err
	}

	// 2. Kunci wallet pengguna untuk mencegah race condition / double spend
	var currentCoins int
	queryWallet := `
	SELECT coins
	FROM player_wallets
	WHERE player_id = $1
	FOR UPDATE;
	`
	err = tx.QueryRowContext(ctx, queryWallet, userID).Scan(&currentCoins)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil saldo dompet: %w", err)
	}

	if currentCoins < costCoins {
		return nil, fmt.Errorf("koin tidak mencukupi (memerlukan %d koin, saldo kamu: %d)", costCoins, currentCoins)
	}

	// 3. Potong koin
	newBalance := currentCoins - costCoins
	updateWallet := `
	UPDATE player_wallets
	SET coins = $1, updated_at = NOW()
	WHERE player_id = $2;
	`
	if _, err := tx.ExecContext(ctx, updateWallet, newBalance, userID); err != nil {
		return nil, fmt.Errorf("gagal memotong saldo koin: %w", err)
	}

	// 4. Catat transaksi koin secara aman
	var safeMatchID sql.NullString
	if matchID != "" {
		var existingID string
		if err := tx.QueryRowContext(ctx, `SELECT id FROM matches WHERE id = $1;`, matchID).Scan(&existingID); err == nil {
			safeMatchID = sql.NullString{String: existingID, Valid: true}
		}
	}

	refID := fmt.Sprintf("use_%s_%d_%d", userID, time.Now().UnixNano(), costCoins)
	insertTx := `
	INSERT INTO coin_transactions (player_id, amount, balance_after, source, match_id, reference_id, created_at)
	VALUES ($1, $2, $3, 'item_usage', $4, $5, NOW());
	`
	if _, err := tx.ExecContext(ctx, insertTx, userID, -costCoins, newBalance, safeMatchID, refID); err != nil {
		return nil, fmt.Errorf("gagal mencatat transaksi koin: %w", err)
	}

	// 5. Catat riwayat pemakaian item
	insertUsage := `
	INSERT INTO item_usage (match_id, round_number, user_player_id, target_player_id, item_id, cost_coins, used_at)
	VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, NOW());
	`
	if _, err := tx.ExecContext(ctx, insertUsage, safeMatchID, roundNumber, userID, targetID, itemID, costCoins); err != nil {
		return nil, fmt.Errorf("gagal mencatat riwayat pemakaian item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal commit transaksi item: %w", err)
	}

	return &UseItemResult{
		ItemID:          itemID,
		ItemName:        itemName,
		UserPlayerID:    userID,
		TargetPlayerID:  targetID,
		RemainingCoins:  newBalance,
		DurationSeconds: duration,
	}, nil
}
