package game

// Tipe pesan WebSocket
const (
	// Client -> Server
	MsgTypeStartGame    = "start_game"
	MsgTypeDrawStroke   = "draw_stroke"
	MsgTypeClearCanvas  = "clear_canvas"
	MsgTypeUndoStroke   = "undo_stroke"
	MsgTypeSubmitGuess  = "submit_guess"
	MsgTypeSendChat     = "send_chat"
	MsgTypeSendReaction = "send_reaction"
	MsgTypeUseItem      = "use_item"
	MsgTypeRestartGame  = "restart_game"

	// Server -> Client
	MsgTypeRoomState    = "room_state"
	MsgTypeStrokeDrawn  = "stroke_drawn"
	MsgTypeCanvasClear  = "canvas_cleared"
	MsgTypeStrokeUndone = "stroke_undone"
	MsgTypeChatMessage  = "chat_message"
	MsgTypeTimerTick    = "timer_tick"
	MsgTypeRoundEnded   = "round_ended"
	MsgTypeGameEnded    = "game_ended"
	MsgTypeReaction     = "reaction_broadcast"
	MsgTypeItemUsed     = "item_activated"
	MsgTypeError        = "error"
)

// Point merepresentasikan titik koordinat pada kanvas (proporsional 1200x800).
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Stroke merepresentasikan satu goresan kuas/pensil.
type Stroke struct {
	Color    string  `json:"color"`
	Size     int     `json:"size"`
	IsEraser bool    `json:"isEraser"`
	Points   []Point `json:"points"`
}

// WSMessage adalah struktur pembungkus pesan WebSocket.
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// StrokePayload membawa data goresan baru.
type StrokePayload struct {
	Stroke Stroke `json:"stroke"`
}

// GuessPayload membawa tebakan atau pesan teks dari pemain.
type GuessPayload struct {
	Text string `json:"text"`
}

// ReactionPayload membawa emoji reaksi dari pemain.
type ReactionPayload struct {
	Emoji string `json:"emoji"`
}

// UseItemPayload adalah permintaan penggunaan item dari client.
type UseItemPayload struct {
	ItemID string `json:"itemId"`
}

// ItemActivatedPayload dibroadcast ke seluruh pemain saat ada item chaos yang aktif.
type ItemActivatedPayload struct {
	UserPlayerID    string `json:"userPlayerId"`
	UserPlayerName  string `json:"userPlayerName"`
	TargetPlayerID  string `json:"targetPlayerId"`
	ItemID          string `json:"itemId"`
	ItemName        string `json:"itemName"`
	DurationSeconds int    `json:"durationSeconds"`
}

// ChatMessagePayload adalah pesan chat/tebakan yang dibroadcast ke pemain.
type ChatMessagePayload struct {
	PlayerName string `json:"playerName"`
	Avatar     string `json:"avatar"`
	Text       string `json:"text"`
	IsSystem   bool   `json:"isSystem"`
	IsCorrect  bool   `json:"isCorrect"`
}

// TimerTickPayload adalah sinkronisasi sisa detik timer ronde.
type TimerTickPayload struct {
	TimeRemaining int `json:"timeRemaining"`
}

// RoundEndedPayload membawa ringkasan hasil ronde.
type RoundEndedPayload struct {
	Round      int            `json:"round"`
	SecretWord string         `json:"secretWord"`
	DrawerName string         `json:"drawerName"`
	Scores     []PlayerPublic `json:"scores"`
}

// GameEndedPlayer memuat data leaderboard akhir pertandingan.
type GameEndedPlayer struct {
	PlayerPublic
	CoinsEarned  int    `json:"coinsEarned"`
	RatingChange int    `json:"ratingChange"`
	NewRating    int    `json:"newRating"`
	NewRankTier  string `json:"newRankTier"`
}

// GameEndedPayload membawa hasil akhir pertandingan (leaderboard).
type GameEndedPayload struct {
	Leaderboard []GameEndedPlayer `json:"leaderboard"`
}

// ReactionBroadcastPayload membroadcast reaksi emoji melayang.
type ReactionBroadcastPayload struct {
	PlayerName string `json:"playerName"`
	Emoji      string `json:"emoji"`
}
