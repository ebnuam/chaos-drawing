package game

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 5 * time.Second
	pongWait   = 30 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 64 * 1024 // 64 KB per WS message
)

// PlayerPublic adalah data pemain yang aman dikirim ke semua klien.
type PlayerPublic struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Avatar              string `json:"avatar"`
	IsHost              bool   `json:"isHost"`
	Score               int    `json:"score"`
	HasGuessedCorrectly bool   `json:"hasGuessedCorrectly"`
	IsDrawing           bool   `json:"isDrawing"`
	Rating              int    `json:"rating"`
	RankTier            string `json:"rankTier"`
	Coins               int    `json:"coins"`
}

// Player merepresentasikan satu sesi pemain aktif dalam room.
type Player struct {
	ID                  string
	Name                string
	Avatar              string
	IsHost              bool
	Score               int
	RoundScore          int
	HasGuessedCorrectly bool
	Rating              int
	RankTier            string
	Coins               int
	CorrectGuesses      int
	GuessAttempts       int
	DrawingsMade        int
	SuccessfulDrawings  int
	Conn                *websocket.Conn
	sendChan            chan []byte
	closed              bool
	mu                  sync.Mutex
}

// NewPlayer membuat instansi Player baru.
func NewPlayer(id, name, avatar string, isHost bool, conn *websocket.Conn) *Player {
	return &Player{
		ID:       id,
		Name:     name,
		Avatar:   avatar,
		IsHost:   isHost,
		Score:    0,
		Rating:   1000,
		RankTier: "Bronze",
		Coins:    100,
		Conn:     conn,
		sendChan: make(chan []byte, 128),
	}
}

// ToPublic mengubah Player menjadi representasi aman publik.
func (p *Player) ToPublic(isDrawing bool) PlayerPublic {
	p.mu.Lock()
	defer p.mu.Unlock()
	return PlayerPublic{
		ID:                  p.ID,
		Name:                p.Name,
		Avatar:              p.Avatar,
		IsHost:              p.IsHost,
		Score:               p.Score,
		HasGuessedCorrectly: p.HasGuessedCorrectly,
		IsDrawing:           isDrawing,
		Rating:              p.Rating,
		RankTier:            p.RankTier,
		Coins:               p.Coins,
	}
}

// SendMessage mengantrekan pesan JSON ke websocket client secara thread-safe.
func (p *Player) SendMessage(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("gagal serialize pesan WS: %v", err)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}

	select {
	case p.sendChan <- data:
	default:
		log.Printf("buffer sendChan penuh untuk pemain %s, drop message", p.Name)
	}
}

// Close menutup channel pengiriman dan koneksi websocket.
func (p *Player) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.closed {
		p.closed = true
		close(p.sendChan)
		if p.Conn != nil {
			p.Conn.Close()
		}
	}
}

// WritePump memproses antrean pesan untuk dikirim ke websocket client.
func (p *Player) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.Close()
	}()

	for {
		select {
		case message, ok := <-p.sendChan:
			if !ok {
				if p.Conn != nil {
					p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
					p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				}
				return
			}

			if p.Conn == nil {
				return
			}

			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := p.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Kirim pesan lain yang sudah ada di antrean
			n := len(p.sendChan)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-p.sendChan)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			if p.Conn == nil {
				return
			}
			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
