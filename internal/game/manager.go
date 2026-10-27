package game

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"sync"
)

// RoomManager mengelola seluruh instansi room yang sedang aktif di memori server.
type RoomManager struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

var globalManager = &RoomManager{
	rooms: make(map[string]*Room),
}

// GetManager mengembalikan singleton instance RoomManager.
func GetManager() *RoomManager {
	return globalManager
}

func generateRandomCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			idx = big.NewInt(0)
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}

func (m *RoomManager) generateRoomCodeLocked() string {
	for {
		code := generateRandomCode()
		if _, exists := m.rooms[code]; !exists {
			return code
		}
	}
}

// GenerateRoomCode menghasilkan kode unik 6 karakter huruf kapital dan angka (menghindari karakter ambigu O/0, I/1).
func (m *RoomManager) GenerateRoomCode() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for {
		code := generateRandomCode()
		if _, exists := m.rooms[code]; !exists {
			return code
		}
	}
}

// CreateRoom membuat room baru dan menyimpannya di memori.
func (m *RoomManager) CreateRoom(rounds, duration, maxPlayers int) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	code := m.generateRoomCodeLocked()
	room := NewRoom(code, func(emptyCode string) {
		m.RemoveRoom(emptyCode)
	})

	if rounds >= 3 && rounds <= 10 {
		room.TotalRounds = rounds
	}
	if duration >= 30 && duration <= 180 {
		room.RoundDuration = duration
	}
	if maxPlayers >= 2 && maxPlayers <= 12 {
		room.MaxPlayers = maxPlayers
	}

	m.rooms[code] = room
	return room
}

// GetOrCreateRoom mengembalikan room yang sudah ada atau membuat room baru jika belum ada.
func (m *RoomManager) GetOrCreateRoom(code string) (*Room, bool) {
	code = strings.ToUpper(strings.TrimSpace(code))
	m.mu.Lock()
	defer m.mu.Unlock()

	if room, exists := m.rooms[code]; exists {
		return room, false
	}

	room := NewRoom(code, func(emptyCode string) {
		m.RemoveRoom(emptyCode)
	})
	m.rooms[code] = room
	return room, true
}

// GetRoom mengambil room berdasarkan kode.
func (m *RoomManager) GetRoom(code string) (*Room, bool) {
	code = strings.ToUpper(strings.TrimSpace(code))
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, exists := m.rooms[code]
	return room, exists
}

// RemoveRoom menghapus room dari memori.
func (m *RoomManager) RemoveRoom(code string) {
	code = strings.ToUpper(strings.TrimSpace(code))
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.rooms, code)
	fmt.Printf("[RoomManager] Room %s dihapus dari memori\n", code)
}
