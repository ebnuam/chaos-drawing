package game

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"drawingchaos/internal/db"
)

// Game states
const (
	StateWaiting     = "waiting"
	StateStarting    = "starting"
	StateDrawing     = "drawing"
	StateRoundResult = "round_result"
	StateGameOver    = "game_over"
)

// RoomStatePayload adalah representasi state lengkap yang dikirim ke browser.
type RoomStatePayload struct {
	RoomCode      string         `json:"roomCode"`
	State         string         `json:"state"`
	Players       []PlayerPublic `json:"players"`
	CurrentRound  int            `json:"currentRound"`
	TotalRounds   int            `json:"totalRounds"`
	RoundDuration int            `json:"roundDuration"`
	TimeRemaining int            `json:"timeRemaining"`
	DrawerID      string         `json:"drawerId"`
	DrawerName    string         `json:"drawerName"`
	SecretWord    string         `json:"secretWord,omitempty"` // HANYA diisi untuk drawer
	Hint          string         `json:"hint,omitempty"`       // Masked string untuk penebak
	Clue          string         `json:"clue,omitempty"`       // Petunjuk bantuan kata
	Category      string         `json:"category,omitempty"`
	Strokes       []Stroke       `json:"strokes,omitempty"`
}

// Room mengelola sesi satu ruangan permainan.
type Room struct {
	Code               string
	MatchID            string
	MaxPlayers         int
	TotalRounds        int
	RoundDuration      int
	State              string
	CurrentRound       int
	CurrentDrawerIndex int
	CurrentPrompt      Prompt
	CurrentHint        string
	CurrentClues       []string
	RevealedClue       string
	TimeRemaining      int
	Strokes            []Stroke
	UsedPrompts        map[string]bool
	Players            []*Player
	RoundsPlayed       []db.RoundResult
	Recorded           bool

	timerStopChan chan struct{}
	mu            sync.Mutex
	onEmpty       func(code string)
}

// NewRoom membuat room baru dengan kode unik.
func NewRoom(code string, onEmpty func(code string)) *Room {
	return &Room{
		Code:          code,
		MaxPlayers:    6,
		TotalRounds:   5,
		RoundDuration: 80,
		State:         StateWaiting,
		CurrentRound:  0,
		UsedPrompts:   make(map[string]bool),
		Players:       make([]*Player, 0),
		onEmpty:       onEmpty,
	}
}

// GetPlayerByID mencari pemain berdasarkan ID dalam room.
func (r *Room) GetPlayerByID(id string) *Player {
	for _, p := range r.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// AddPlayer menambahkan pemain ke dalam room.
func (r *Room) AddPlayer(player *Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= r.MaxPlayers {
		return fmt.Errorf("room penuh (maksimal %d pemain)", r.MaxPlayers)
	}

	// Tangani nama duplikat jika ada
	baseName := player.Name
	counter := 2
	for r.isNameTaken(player.Name) {
		player.Name = fmt.Sprintf("%s %d", baseName, counter)
		counter++
	}

	// Pemain pertama otomatis menjadi Host
	if len(r.Players) == 0 {
		player.IsHost = true
	}

	r.Players = append(r.Players, player)
	log.Printf("[Room %s] Pemain %s bergabung (total: %d)", r.Code, player.Name, len(r.Players))

	r.broadcastRoomStateLocked()
	r.broadcastChatLocked(ChatMessagePayload{
		PlayerName: "Sistem",
		Text:       fmt.Sprintf("%s bergabung ke room.", player.Name),
		IsSystem:   true,
	})

	return nil
}

func (r *Room) isNameTaken(name string) bool {
	for _, p := range r.Players {
		if strings.EqualFold(p.Name, name) {
			return true
		}
	}
	return false
}

// RemovePlayer mengeluarkan pemain dari room.
func (r *Room) RemovePlayer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	idx := -1
	var removedPlayer *Player
	for i, p := range r.Players {
		if p.ID == playerID {
			idx = i
			removedPlayer = p
			break
		}
	}

	if idx == -1 {
		return
	}

	r.Players = append(r.Players[:idx], r.Players[idx+1:]...)
	log.Printf("[Room %s] Pemain %s keluar (sisa: %d)", r.Code, removedPlayer.Name, len(r.Players))

	if len(r.Players) == 0 {
		r.stopTimerLocked()
		if r.onEmpty != nil {
			go r.onEmpty(r.Code)
		}
		return
	}

	// Jika host yang keluar, oper hak host ke pemain berikutnya
	if removedPlayer.IsHost && len(r.Players) > 0 {
		r.Players[0].IsHost = true
		r.broadcastChatLocked(ChatMessagePayload{
			PlayerName: "Sistem",
			Text:       fmt.Sprintf("%s sekarang menjadi Host.", r.Players[0].Name),
			IsSystem:   true,
		})
	}

	// Jika pemain yang sedang menggambar keluar saat ronde aktif
	if r.State == StateDrawing && idx == r.CurrentDrawerIndex {
		r.broadcastChatLocked(ChatMessagePayload{
			PlayerName: "Sistem",
			Text:       "Penggambar keluar dari permainan. Ronde dihentikan.",
			IsSystem:   true,
		})
		r.endRoundLocked()
		return
	}

	r.broadcastRoomStateLocked()
	r.broadcastChatLocked(ChatMessagePayload{
		PlayerName: "Sistem",
		Text:       fmt.Sprintf("%s meninggalkan room.", removedPlayer.Name),
		IsSystem:   true,
	})
}

// StartGame memulai permainan dari host.
func (r *Room) StartGame(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p := r.GetPlayerByID(playerID)
	if p == nil || !p.IsHost {
		return fmt.Errorf("hanya host yang dapat memulai permainan")
	}

	if len(r.Players) < 1 {
		return fmt.Errorf("perlu minimal 1 pemain untuk memulai")
	}

	// Reset skor seluruh pemain
	for _, pl := range r.Players {
		pl.Score = 0
		pl.RoundScore = 0
		pl.HasGuessedCorrectly = false
	}

	r.CurrentRound = 1
	r.CurrentDrawerIndex = 0
	r.UsedPrompts = make(map[string]bool)
	r.MatchID = fmt.Sprintf("m_%d", time.Now().UnixNano())
	r.RoundsPlayed = make([]db.RoundResult, 0)
	r.Recorded = false

	// Catat status match baru ke database secara asinkron
	if database := db.GetDB(); database != nil {
		go func(mid, code string, tr, rd int) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = database.CreateMatch(ctx, mid, code, tr, rd)
		}(r.MatchID, r.Code, r.TotalRounds, r.RoundDuration)
	}

	r.startRoundLocked()
	return nil
}

func (r *Room) startRoundLocked() {
	r.stopTimerLocked()
	r.State = StateDrawing
	r.Strokes = make([]Stroke, 0)

	// Pastikan kanvas di seluruh klien langsung dibersihkan saat pergantian ronde
	r.broadcastToAllLocked(WSMessage{
		Type: MsgTypeCanvasClear,
	})

	// Pastikan drawer valid
	if len(r.Players) == 0 {
		return
	}
	drawerIdx := r.CurrentDrawerIndex % len(r.Players)
	drawer := r.Players[drawerIdx]
	drawer.DrawingsMade++

	// Reset status tebakan tiap pemain
	for _, p := range r.Players {
		p.HasGuessedCorrectly = false
		p.RoundScore = 0
	}

	// Ambil kata rahasia dari bank kata terpusat database (atau fallback)
	database := db.GetDB()
	if database != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		rec, err := database.GetRandomWordWithClues(ctx, r.UsedPrompts)
		cancel()
		if err == nil && rec != nil {
			r.CurrentPrompt = Prompt{Word: rec.Word, Category: rec.Category}
			r.CurrentClues = rec.Clues
		} else {
			r.CurrentPrompt = GetRandomPrompt(r.UsedPrompts)
			r.CurrentClues = nil
		}
	} else {
		r.CurrentPrompt = GetRandomPrompt(r.UsedPrompts)
		r.CurrentClues = nil
	}
	r.RevealedClue = ""
	r.UsedPrompts[strings.ToLower(r.CurrentPrompt.Word)] = true
	r.CurrentHint = GenerateHint(r.CurrentPrompt.Word, 0)
	r.TimeRemaining = r.RoundDuration

	r.broadcastRoomStateLocked()
	r.broadcastChatLocked(ChatMessagePayload{
		PlayerName: "Sistem",
		Text:       fmt.Sprintf("Ronde %d dimulai! Giliran %s menggambar.", r.CurrentRound, drawer.Name),
		IsSystem:   true,
	})

	// Jalankan countdown timer di goroutine server-authoritative
	r.timerStopChan = make(chan struct{})
	go r.runRoundTimer(r.timerStopChan)
}

func (r *Room) runRoundTimer(stopChan chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			r.mu.Lock()
			if r.State != StateDrawing {
				r.mu.Unlock()
				return
			}

			r.TimeRemaining--

			// Tampilkan sebagian petunjuk huruf & clue saat sisa waktu 50% dan 25%
			halfTime := r.RoundDuration / 2
			quarterTime := r.RoundDuration / 4

			if r.TimeRemaining == halfTime {
				r.CurrentHint = GenerateHint(r.CurrentPrompt.Word, 0.25)
				if len(r.CurrentClues) > 0 {
					r.RevealedClue = fmt.Sprintf("Petunjuk 1: %s", r.CurrentClues[0])
					r.broadcastChatLocked(ChatMessagePayload{
						PlayerName: "Petunjuk",
						Text:       fmt.Sprintf("💡 %s", r.RevealedClue),
						IsSystem:   true,
					})
				}
				r.broadcastRoomStateLocked()
			} else if r.TimeRemaining == quarterTime {
				r.CurrentHint = GenerateHint(r.CurrentPrompt.Word, 0.50)
				if len(r.CurrentClues) > 1 {
					r.RevealedClue = fmt.Sprintf("Petunjuk 2: %s", r.CurrentClues[1])
					r.broadcastChatLocked(ChatMessagePayload{
						PlayerName: "Petunjuk",
						Text:       fmt.Sprintf("💡 %s", r.RevealedClue),
						IsSystem:   true,
					})
				}
				r.broadcastRoomStateLocked()
			}

			// Broadcast timer tick ke semua klien
			tickMsg := WSMessage{
				Type: MsgTypeTimerTick,
				Payload: TimerTickPayload{
					TimeRemaining: r.TimeRemaining,
				},
			}
			for _, p := range r.Players {
				p.SendMessage(tickMsg)
			}

			if r.TimeRemaining <= 0 {
				r.broadcastChatLocked(ChatMessagePayload{
					PlayerName: "Sistem",
					Text:       fmt.Sprintf("Waktu habis! Kata rahasianya adalah: %s", r.CurrentPrompt.Word),
					IsSystem:   true,
				})
				r.endRoundLocked()
				r.mu.Unlock()
				return
			}

			r.mu.Unlock()
		}
	}
}

func (r *Room) stopTimerLocked() {
	if r.timerStopChan != nil {
		close(r.timerStopChan)
		r.timerStopChan = nil
	}
}

func (r *Room) endRoundLocked() {
	r.stopTimerLocked()
	r.State = StateRoundResult

	drawer := r.currentDrawerLocked()
	drawerName := ""
	drawerID := ""
	if drawer != nil {
		drawerName = drawer.Name
		drawerID = drawer.ID
	}

	correctCount := 0
	for _, p := range r.Players {
		if p.HasGuessedCorrectly {
			correctCount++
		}
	}

	if drawerID != "" {
		r.RoundsPlayed = append(r.RoundsPlayed, db.RoundResult{
			RoundNumber:         r.CurrentRound,
			DrawerPlayerID:      drawerID,
			Word:                r.CurrentPrompt.Word,
			Category:            r.CurrentPrompt.Category,
			CorrectGuessesCount: correctCount,
			DurationSeconds:     r.RoundDuration - r.TimeRemaining,
		})
	}

	// Kirim pesan round_ended
	scores := r.getPublicPlayersLocked()
	roundEndedMsg := WSMessage{
		Type: MsgTypeRoundEnded,
		Payload: RoundEndedPayload{
			Round:      r.CurrentRound,
			SecretWord: r.CurrentPrompt.Word,
			DrawerName: drawerName,
			Scores:     scores,
		},
	}
	for _, p := range r.Players {
		p.SendMessage(roundEndedMsg)
	}

	r.broadcastRoomStateLocked()

	// Jeda 5 detik sebelum lanjut ke ronde berikutnya atau game over
	go func() {
		time.Sleep(5 * time.Second)
		r.mu.Lock()
		defer r.mu.Unlock()

		if r.State != StateRoundResult {
			return
		}

		if r.CurrentRound >= r.TotalRounds {
			r.endGameLocked()
		} else {
			r.CurrentRound++
			r.CurrentDrawerIndex++
			r.startRoundLocked()
		}
	}()
}

func (r *Room) endGameLocked() {
	r.stopTimerLocked()
	r.State = StateGameOver

	if r.Recorded {
		return
	}
	r.Recorded = true

	// Urutkan pemain berdasarkan skor tertinggi
	sortedPlayers := make([]*Player, len(r.Players))
	copy(sortedPlayers, r.Players)
	sort.Slice(sortedPlayers, func(i, j int) bool {
		return sortedPlayers[i].Score > sortedPlayers[j].Score
	})

	winnerID := ""
	if len(sortedPlayers) > 0 {
		winnerID = sortedPlayers[0].ID
	}

	var matchPlayersData []db.MatchPlayerResult
	var leaderboard []GameEndedPlayer

	for rank, pl := range sortedPlayers {
		rankPos := rank + 1
		isWin := rank == 0

		// Koin dan rating berdasarkan peringkat pertandingan
		coins := 10
		ratingDelta := -10
		switch rankPos {
		case 1:
			coins = 50
			ratingDelta = 30
		case 2:
			coins = 30
			ratingDelta = 15
		case 3:
			coins = 20
			ratingDelta = 5
		}

		oldRating := pl.Rating
		newRating := oldRating + ratingDelta
		if newRating < 100 {
			newRating = 100
		}
		newTier := db.DetermineRankTier(newRating)

		pl.Rating = newRating
		pl.RankTier = newTier
		pl.Coins += coins

		matchPlayersData = append(matchPlayersData, db.MatchPlayerResult{
			PlayerID:           pl.ID,
			FinalScore:         pl.Score,
			RankPosition:       rankPos,
			CoinsEarned:        coins,
			RatingChange:       ratingDelta,
			OldRating:          oldRating,
			NewRating:          newRating,
			OldTier:            pl.RankTier,
			NewTier:            newTier,
			IsWinner:           isWin,
			CorrectGuesses:     pl.CorrectGuesses,
			GuessAttempts:      pl.GuessAttempts,
			DrawingsMade:       pl.DrawingsMade,
			SuccessfulDrawings: pl.SuccessfulDrawings,
		})

		leaderboard = append(leaderboard, GameEndedPlayer{
			PlayerPublic: pl.ToPublic(false),
			CoinsEarned:  coins,
			RatingChange: ratingDelta,
			NewRating:    newRating,
			NewRankTier:  newTier,
		})
	}

	// Simpan ke database PostgreSQL secara asinkron
	if database := db.GetDB(); database != nil {
		go func(params db.MatchRecordParams) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if err := database.RecordMatch(ctx, params); err != nil {
				log.Printf("[Database Warning] Gagal simpan match %s: %v", params.MatchID, err)
			} else {
				log.Printf("[Database] Hasil match %s berhasil disimpan permanen.", params.MatchID)
			}
		}(db.MatchRecordParams{
			MatchID:        r.MatchID,
			RoomCode:       r.Code,
			TotalRounds:    r.TotalRounds,
			RoundDuration:  r.RoundDuration,
			WinnerPlayerID: winnerID,
			Players:        matchPlayersData,
			Rounds:         r.RoundsPlayed,
		})
	}

	r.broadcastRoomStateLocked()
	r.broadcastToAllLocked(WSMessage{
		Type: MsgTypeGameEnded,
		Payload: GameEndedPayload{
			Leaderboard: leaderboard,
		},
	})
}

// RestartGame mengembalikan room ke mode lobby
func (r *Room) RestartGame(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p := r.GetPlayerByID(playerID)
	if p == nil || !p.IsHost {
		return fmt.Errorf("hanya host yang dapat me-restart permainan")
	}

	r.stopTimerLocked()
	r.State = StateWaiting
	r.CurrentRound = 0
	r.CurrentDrawerIndex = 0
	r.Strokes = nil
	r.UsedPrompts = make(map[string]bool)

	// Bersihkan kanvas saat kembali ke lobby/restart
	r.broadcastToAllLocked(WSMessage{
		Type: MsgTypeCanvasClear,
	})

	for _, pl := range r.Players {
		pl.Score = 0
		pl.RoundScore = 0
		pl.HasGuessedCorrectly = false
	}

	r.broadcastRoomStateLocked()
	r.broadcastChatLocked(ChatMessagePayload{
		PlayerName: "Sistem",
		Text:       "Permainan di-reset. Kembali ke lobby.",
		IsSystem:   true,
	})

	return nil
}

// HandleDrawStroke menangani penambahan goresan dari penggambar.
func (r *Room) HandleDrawStroke(playerID string, stroke Stroke) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != StateDrawing {
		return fmt.Errorf("permainan belum dalam fase menggambar")
	}

	drawer := r.currentDrawerLocked()
	if drawer == nil || drawer.ID != playerID {
		return fmt.Errorf("kamu bukan penggambar pada ronde ini")
	}

	// Batasi jumlah goresan untuk mencegah payload abuse
	if len(r.Strokes) > 5000 {
		return fmt.Errorf("batas goresan kanvas tercapai")
	}

	r.Strokes = append(r.Strokes, stroke)

	// Broadcast goresan ke seluruh pemain lain
	msg := WSMessage{
		Type: MsgTypeStrokeDrawn,
		Payload: StrokePayload{
			Stroke: stroke,
		},
	}
	for _, p := range r.Players {
		if p.ID != playerID {
			p.SendMessage(msg)
		}
	}

	return nil
}

// HandleClearCanvas membersihkan kanvas dari penggambar.
func (r *Room) HandleClearCanvas(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != StateDrawing {
		return fmt.Errorf("kanvas hanya bisa dibersihkan saat fase menggambar")
	}

	drawer := r.currentDrawerLocked()
	if drawer == nil || drawer.ID != playerID {
		return fmt.Errorf("hanya penggambar yang dapat membersihkan kanvas")
	}

	r.Strokes = make([]Stroke, 0)
	r.broadcastToAllLocked(WSMessage{
		Type: MsgTypeCanvasClear,
	})

	return nil
}

// HandleUndoStroke membatalkan goresan terakhir dari penggambar.
func (r *Room) HandleUndoStroke(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != StateDrawing {
		return fmt.Errorf("undo hanya bisa dilakukan saat fase menggambar")
	}

	drawer := r.currentDrawerLocked()
	if drawer == nil || drawer.ID != playerID {
		return fmt.Errorf("hanya penggambar yang dapat melakukan undo")
	}

	if len(r.Strokes) > 0 {
		r.Strokes = r.Strokes[:len(r.Strokes)-1]
		r.broadcastToAllLocked(WSMessage{
			Type: MsgTypeStrokeUndone,
		})
	}

	return nil
}

// HandleSubmitGuess memvalidasi tebakan dari pemain penebak.
func (r *Room) HandleSubmitGuess(playerID, text string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != StateDrawing {
		return fmt.Errorf("hanya dapat menebak saat fase menggambar aktif")
	}

	player := r.GetPlayerByID(playerID)
	if player == nil {
		return fmt.Errorf("pemain tidak ditemukan")
	}

	drawer := r.currentDrawerLocked()
	if drawer != nil && drawer.ID == playerID {
		return fmt.Errorf("penggambar tidak boleh mengirim tebakan")
	}

	if player.HasGuessedCorrectly {
		return fmt.Errorf("kamu sudah berhasil menebak pada ronde ini")
	}

	cleanGuess := strings.TrimSpace(text)
	if cleanGuess == "" {
		return nil
	}

	player.GuessAttempts++

	// Periksa kecocokan tebakan dengan kata rahasia
	if IsMatchGuess(cleanGuess, r.CurrentPrompt.Word) {
		player.HasGuessedCorrectly = true
		player.CorrectGuesses++

		// Hitung skor berdasarkan sisa waktu (semakin cepat, semakin besar)
		ratio := float64(r.TimeRemaining) / float64(r.RoundDuration)
		timeBonus := int(math.Round(ratio * 150))
		guesserScore := 100 + timeBonus
		player.Score += guesserScore
		player.RoundScore = guesserScore

		// Penggambar mendapat 50 poin setiap ada penebak yang berhasil
		if drawer != nil {
			drawer.Score += 50
			drawer.RoundScore += 50
			drawer.SuccessfulDrawings++
		}

		// Kirim notifikasi tebakan benar ke penebak
		player.SendMessage(WSMessage{
			Type: MsgTypeChatMessage,
			Payload: ChatMessagePayload{
				PlayerName: "Sistem",
				Text:       fmt.Sprintf("Hebat! Tebakanmu benar (+%d poin)!", guesserScore),
				IsSystem:   true,
				IsCorrect:  true,
			},
		})

		// Broadcast ke pemain lain tanpa membocorkan kata jawaban
		for _, p := range r.Players {
			if p.ID != playerID {
				p.SendMessage(WSMessage{
					Type: MsgTypeChatMessage,
					Payload: ChatMessagePayload{
						PlayerName: "Sistem",
						Text:       fmt.Sprintf("%s berhasil menebak!", player.Name),
						IsSystem:   true,
						IsCorrect:  true,
					},
				})
			}
		}

		r.broadcastRoomStateLocked()

		// Jika seluruh penebak sudah berhasil menebak, selesaikan ronde lebih awal
		allGuessed := true
		for _, p := range r.Players {
			if drawer != nil && p.ID == drawer.ID {
				continue
			}
			if !p.HasGuessedCorrectly {
				allGuessed = false
				break
			}
		}

		if allGuessed {
			r.broadcastChatLocked(ChatMessagePayload{
				PlayerName: "Sistem",
				Text:       "Semua pemain telah menebak dengan benar! Ronde selesai.",
				IsSystem:   true,
			})
			r.endRoundLocked()
		}

		return nil
	}

	// Jika tebakan salah, kirim sebagai pesan chat biasa
	r.broadcastChatLocked(ChatMessagePayload{
		PlayerName: player.Name,
		Avatar:     player.Avatar,
		Text:       cleanGuess,
		IsSystem:   false,
		IsCorrect:  false,
	})

	return nil
}

// HandleSendChat mengirim pesan chat umum.
func (r *Room) HandleSendChat(playerID, text string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p := r.GetPlayerByID(playerID)
	if p == nil {
		return
	}

	cleanText := strings.TrimSpace(text)
	if cleanText == "" {
		return
	}

	r.broadcastChatLocked(ChatMessagePayload{
		PlayerName: p.Name,
		Avatar:     p.Avatar,
		Text:       cleanText,
		IsSystem:   false,
	})
}

// HandleSendReaction membroadcast emoji reaksi melayang.
func (r *Room) HandleSendReaction(playerID, emoji string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p := r.GetPlayerByID(playerID)
	if p == nil {
		return
	}

	r.broadcastToAllLocked(WSMessage{
		Type: MsgTypeReaction,
		Payload: ReactionBroadcastPayload{
			PlayerName: p.Name,
			Emoji:      emoji,
		},
	})
}

// HandleUseItem memproses aktivasi Chaos Item oleh pemain.
func (r *Room) HandleUseItem(playerID, itemID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != StateDrawing {
		return fmt.Errorf("item hanya dapat digunakan saat fase menggambar sedang berlangsung")
	}

	user := r.GetPlayerByID(playerID)
	if user == nil {
		return fmt.Errorf("pemain tidak ditemukan")
	}

	drawer := r.currentDrawerLocked()
	targetID := ""
	if drawer != nil {
		targetID = drawer.ID
	}

	if itemID == "effect_cleaner" {
		targetID = user.ID
	} else if itemID == "turbo_time" {
		targetID = ""
		r.TimeRemaining += 10
		if r.TimeRemaining > r.RoundDuration {
			r.RoundDuration = r.TimeRemaining
		}
	} else {
		if drawer != nil && drawer.ID == playerID {
			return fmt.Errorf("penggambar tidak dapat menggunakan item gangguan terhadap diri sendiri")
		}
	}

	database := db.GetDB()
	res, err := database.UseChaosItem(context.Background(), r.MatchID, r.CurrentRound, playerID, targetID, itemID)
	if err != nil {
		return err
	}

	user.Coins = res.RemainingCoins

	r.broadcastToAllLocked(WSMessage{
		Type: MsgTypeItemUsed,
		Payload: ItemActivatedPayload{
			UserPlayerID:    user.ID,
			UserPlayerName:  user.Name,
			TargetPlayerID:  targetID,
			ItemID:          res.ItemID,
			ItemName:        res.ItemName,
			DurationSeconds: res.DurationSeconds,
		},
	})

	var logText string
	if itemID == "effect_cleaner" {
		logText = fmt.Sprintf("%s menggunakan Pembersih Efek! Layar kembali bersih. ✨", user.Name)
	} else if itemID == "turbo_time" {
		logText = fmt.Sprintf("%s mengaktifkan Mode Turbo! +10 detik waktu tambahan! ⚡", user.Name)
	} else if drawer != nil {
		logText = fmt.Sprintf("%s melemparkan %s ke kanvas %s! 🛢️", user.Name, res.ItemName, drawer.Name)
	} else {
		logText = fmt.Sprintf("%s menggunakan %s!", user.Name, res.ItemName)
	}

	r.broadcastChatLocked(ChatMessagePayload{
		PlayerName: "Sistem",
		Text:       logText,
		IsSystem:   true,
	})

	r.broadcastRoomStateLocked()
	return nil
}

func (r *Room) currentDrawerLocked() *Player {
	if len(r.Players) == 0 {
		return nil
	}
	idx := r.CurrentDrawerIndex % len(r.Players)
	return r.Players[idx]
}

func (r *Room) getPublicPlayersLocked() []PlayerPublic {
	drawer := r.currentDrawerLocked()
	res := make([]PlayerPublic, len(r.Players))
	for i, p := range r.Players {
		isDrawing := drawer != nil && p.ID == drawer.ID
		res[i] = p.ToPublic(isDrawing)
	}
	return res
}

func (r *Room) broadcastRoomStateLocked() {
	drawer := r.currentDrawerLocked()
	drawerID := ""
	drawerName := ""
	if drawer != nil {
		drawerID = drawer.ID
		drawerName = drawer.Name
	}

	playersPublic := r.getPublicPlayersLocked()

	for _, p := range r.Players {
		isDrawer := drawer != nil && p.ID == drawer.ID

		payload := RoomStatePayload{
			RoomCode:      r.Code,
			State:         r.State,
			Players:       playersPublic,
			CurrentRound:  r.CurrentRound,
			TotalRounds:   r.TotalRounds,
			RoundDuration: r.RoundDuration,
			TimeRemaining: r.TimeRemaining,
			DrawerID:      drawerID,
			DrawerName:    drawerName,
			Category:      r.CurrentPrompt.Category,
			Clue:          r.RevealedClue,
			Strokes:       r.Strokes,
		}

		if isDrawer {
			// Penggambar mendapat kata rahasia asli
			payload.SecretWord = r.CurrentPrompt.Word
		} else {
			// Penebak HANYA mendapat petunjuk masked (_ _ _ _)
			payload.Hint = r.CurrentHint
		}

		p.SendMessage(WSMessage{
			Type:    MsgTypeRoomState,
			Payload: payload,
		})
	}
}

func (r *Room) broadcastChatLocked(chat ChatMessagePayload) {
	r.broadcastToAllLocked(WSMessage{
		Type:    MsgTypeChatMessage,
		Payload: chat,
	})
}

func (r *Room) broadcastToAllLocked(msg WSMessage) {
	for _, p := range r.Players {
		p.SendMessage(msg)
	}
}
