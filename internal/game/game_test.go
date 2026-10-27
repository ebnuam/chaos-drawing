package game_test

import (
	"strings"
	"testing"

	"drawingchaos/internal/game"
)

func TestPromptAndHints(t *testing.T) {
	used := make(map[string]bool)
	p := game.GetRandomPrompt(used)

	if p.Word == "" {
		t.Errorf("expected non-empty prompt word")
	}
	if p.Category == "" {
		t.Errorf("expected non-empty category")
	}

	hint := game.GenerateHint("Kucing Oyen", 0)
	// Petunjuk awal harus menyembunyikan semua huruf
	if strings.Contains(hint, "K") || strings.Contains(hint, "O") {
		t.Errorf("expected hint to hide letters initially, got: %s", hint)
	}
	if !strings.Contains(hint, "_") {
		t.Errorf("expected hint to have underscores, got: %s", hint)
	}

	// Normalisasi tebakan
	if !game.IsMatchGuess("kucing", "Kucing") {
		t.Errorf("expected 'kucing' to match 'Kucing'")
	}
	if !game.IsMatchGuess("  kucing!  ", "Kucing") {
		t.Errorf("expected '  kucing!  ' to match 'Kucing'")
	}
	if !game.IsMatchGuess("nasi goreng", "Nasi Goreng") {
		t.Errorf("expected 'nasi goreng' to match 'Nasi Goreng'")
	}
	if game.IsMatchGuess("anjing", "Kucing") {
		t.Errorf("expected 'anjing' NOT to match 'Kucing'")
	}
}

func TestRoomLifecycleAndScoring(t *testing.T) {
	manager := game.GetManager()
	room := manager.CreateRoom(3, 60, 4)

	if len(room.Code) != 6 {
		t.Errorf("expected 6-char room code, got: %s", room.Code)
	}

	// Buat dua pemain
	p1 := game.NewPlayer("p1", "Bima", "cat-orange", false, nil)
	p2 := game.NewPlayer("p2", "Dian", "dino-green", false, nil)

	// Pemain 1 bergabung (harus otomatis jadi Host)
	if err := room.AddPlayer(p1); err != nil {
		t.Fatalf("gagal menambah p1: %v", err)
	}
	if !p1.IsHost {
		t.Errorf("expected p1 to be host")
	}

	// Pemain 2 bergabung
	if err := room.AddPlayer(p2); err != nil {
		t.Fatalf("gagal menambah p2: %v", err)
	}
	if p2.IsHost {
		t.Errorf("expected p2 not to be host")
	}

	// Hanya host yang bisa memulai permainan
	if err := room.StartGame("p2"); err == nil {
		t.Errorf("expected error when non-host starts game")
	}

	// Host memulai permainan
	if err := room.StartGame("p1"); err != nil {
		t.Fatalf("expected host to start game successfully, got: %v", err)
	}

	if room.State != game.StateDrawing {
		t.Errorf("expected room state 'drawing', got: %s", room.State)
	}

	// Drawer p1 mencoba menggambar goresan
	stroke := game.Stroke{
		Color:    "#1a1526",
		Size:     8,
		IsEraser: false,
		Points: []game.Point{
			{X: 10, Y: 10},
			{X: 20, Y: 20},
		},
	}
	if err := room.HandleDrawStroke("p1", stroke); err != nil {
		t.Fatalf("drawer p1 failed to draw: %v", err)
	}

	// Guesser p2 TIDAK boleh menggambar
	if err := room.HandleDrawStroke("p2", stroke); err == nil {
		t.Errorf("expected guesser p2 to be rejected from drawing")
	}

	// Drawer p1 TIDAK boleh menebak
	if err := room.HandleSubmitGuess("p1", room.CurrentPrompt.Word); err == nil {
		t.Errorf("expected drawer p1 to be rejected from guessing")
	}

	// Guesser p2 menebak dengan benar
	initialP1Score := p1.Score
	initialP2Score := p2.Score

	if err := room.HandleSubmitGuess("p2", room.CurrentPrompt.Word); err != nil {
		t.Fatalf("p2 correct guess failed: %v", err)
	}

	if !p2.HasGuessedCorrectly {
		t.Errorf("expected p2 HasGuessedCorrectly to be true")
	}
	if p2.Score <= initialP2Score {
		t.Errorf("expected p2 score to increase, got: %d", p2.Score)
	}
	if p1.Score <= initialP1Score {
		t.Errorf("expected drawer p1 score to increase when someone guessed, got: %d", p1.Score)
	}

	// p2 tidak boleh menebak lagi di ronde yang sama
	if err := room.HandleSubmitGuess("p2", room.CurrentPrompt.Word); err == nil {
		t.Errorf("expected second guess by p2 to be rejected")
	}
}
