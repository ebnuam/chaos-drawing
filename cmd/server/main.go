// Drawing Chaos Online — Entry point HTTP & WebSocket server.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"drawingchaos/internal/auth"
	"drawingchaos/internal/config"
	"drawingchaos/internal/db"
	"drawingchaos/internal/game"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		// Izinkan localhost dan 127.0.0.1
		if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") ||
			strings.HasPrefix(origin, "https://localhost") || strings.HasPrefix(origin, "https://127.0.0.1") {
			return true
		}
		// Izinkan jika mengandung Host header
		host := r.Host
		if strings.Contains(origin, host) {
			return true
		}
		return false
	},
}

type CreateRoomRequest struct {
	Rounds     int `json:"rounds"`
	Duration   int `json:"duration"`
	MaxPlayers int `json:"maxPlayers"`
}

type RegisterRequestBody struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Avatar      string `json:"avatar"`
	LinkGuestID string `json:"linkGuestId"`
}

type LoginRequestBody struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func setupRouter(cfg config.Config, staticDir, templatesGlob string) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Aset statis (CSS, JS, gambar) disajikan apa adanya.
	router.Static("/static", staticDir)

	// Template HTML utama
	router.LoadHTMLGlob(templatesGlob)

	// Halaman antarmuka SPA
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "app.html", nil)
	})
	router.GET("/lobby", func(c *gin.Context) {
		c.HTML(http.StatusOK, "app.html", nil)
	})
	router.GET("/play", func(c *gin.Context) {
		c.HTML(http.StatusOK, "app.html", nil)
	})

	// Health check sederhana
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"app":    "drawing-chaos",
		})
	})

	// ================= AUTH API =================

	// POST /api/auth/register
	router.POST("/api/auth/register", func(c *gin.Context) {
		if !auth.GlobalAuthLimiter.Allow(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Terlalu banyak percobaan autentikasi. Silakan tunggu sebentar."})
			return
		}

		var req RegisterRequestBody
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Format data pendaftaran tidak valid"})
			return
		}

		database := db.GetDB()
		user, token, err := auth.Register(c.Request.Context(), database, auth.RegisterParams{
			Username:    req.Username,
			Email:       req.Email,
			Password:    req.Password,
			Avatar:      req.Avatar,
			LinkGuestID: req.LinkGuestID,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		isSecure := cfg.AppEnv == "production"
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("dc_session", token, 30*24*3600, "/", "", isSecure, true)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"user":    user,
			"message": "Pendaftaran akun permanen berhasil!",
		})
	})

	// POST /api/auth/login
	router.POST("/api/auth/login", func(c *gin.Context) {
		if !auth.GlobalAuthLimiter.Allow(c.ClientIP()) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Terlalu banyak percobaan login. Silakan coba 1 menit lagi."})
			return
		}

		var req LoginRequestBody
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username/email dan password wajib diisi"})
			return
		}

		database := db.GetDB()
		user, token, err := auth.Login(c.Request.Context(), database, req.Identifier, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		isSecure := cfg.AppEnv == "production"
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("dc_session", token, 30*24*3600, "/", "", isSecure, true)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"user":    user,
			"message": "Login berhasil! Selamat datang kembali.",
		})
	})

	// POST /api/auth/logout
	router.POST("/api/auth/logout", func(c *gin.Context) {
		token, _ := c.Cookie("dc_session")
		if token != "" {
			database := db.GetDB()
			_ = auth.RevokeSession(c.Request.Context(), database, token)
		}

		isSecure := cfg.AppEnv == "production"
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("dc_session", "", -1, "/", "", isSecure, true)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Sesi telah berakhir. Logout berhasil.",
		})
	})

	// GET /api/auth/me
	router.GET("/api/auth/me", func(c *gin.Context) {
		token, err := c.Cookie("dc_session")
		if err != nil || token == "" {
			c.JSON(http.StatusOK, gin.H{
				"authenticated": false,
				"isGuest":       true,
			})
			return
		}

		database := db.GetDB()
		user, err := auth.GetUserBySession(c.Request.Context(), database, token)
		if err != nil || user == nil {
			c.JSON(http.StatusOK, gin.H{
				"authenticated": false,
				"isGuest":       true,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"authenticated": true,
			"isGuest":       user.IsGuest,
			"user":          user,
		})
	})

	// ================= ROOM & GAME API =================

	manager := game.GetManager()

	router.POST("/api/rooms", func(c *gin.Context) {
		var req CreateRoomRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			req = CreateRoomRequest{Rounds: 5, Duration: 80, MaxPlayers: 6}
		}
		room := manager.CreateRoom(req.Rounds, req.Duration, req.MaxPlayers)
		c.JSON(http.StatusOK, gin.H{
			"code":       room.Code,
			"rounds":     room.TotalRounds,
			"duration":   room.RoundDuration,
			"maxPlayers": room.MaxPlayers,
		})
	})

	router.GET("/api/rooms/:code", func(c *gin.Context) {
		code := strings.ToUpper(c.Param("code"))
		room, exists := manager.GetRoom(code)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room tidak ditemukan"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":        room.Code,
			"state":       room.State,
			"playerCount": len(room.Players),
			"maxPlayers":  room.MaxPlayers,
		})
	})

	// Endpoint Profil & Statistik Lengkap Pemain
	router.GET("/api/players/:id/profile", func(c *gin.Context) {
		id := c.Param("id")
		database := db.GetDB()
		if database != nil {
			profile, err := database.GetPlayerProfile(c.Request.Context(), id)
			if err == nil && profile != nil {
				c.JSON(http.StatusOK, profile)
				return
			}
		}

		// Fallback jika belum tersimpan di DB
		c.JSON(http.StatusOK, gin.H{
			"player": gin.H{
				"id":          id,
				"displayName": "Pemain",
				"avatar":      "cat-orange",
				"isGuest":     true,
			},
			"stats": gin.H{
				"playerId":           id,
				"rating":             1000,
				"rankTier":           "Bronze",
				"coins":              100,
				"matchesPlayed":      0,
				"matchesWon":         0,
				"matchesLost":        0,
				"totalScore":         0,
				"correctGuesses":     0,
				"guessAttempts":      0,
				"guessAccuracy":      0.0,
				"drawingsMade":       0,
				"successfulDrawings": 0,
				"winRate":            0.0,
			},
			"recentMatches": []gin.H{},
		})
	})

	// Endpoint Leaderboard Global (Rating MMR, Total Wins, Total Score)
	router.GET("/api/leaderboard", func(c *gin.Context) {
		sortBy := c.DefaultQuery("sort", "rating")
		database := db.GetDB()
		if database != nil {
			list, err := database.GetLeaderboard(c.Request.Context(), sortBy, 20)
			if err == nil {
				c.JSON(http.StatusOK, gin.H{"leaderboard": list})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"leaderboard": []gin.H{}})
	})

	// Endpoint Katalog Chaos Items (7 Items)
	router.GET("/api/items", func(c *gin.Context) {
		database := db.GetDB()
		items, err := database.GetCatalogItems(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"items": []gin.H{}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	})

	// WebSocket handler
	router.GET("/ws", func(c *gin.Context) {
		handleWebSocket(c.Writer, c.Request, manager)
	})

	router.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Halaman tidak ditemukan")
	})

	return router
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, manager *game.RoomManager) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Gagal upgrade websocket: %v", err)
		return
	}

	query := r.URL.Query()
	roomCode := strings.ToUpper(strings.TrimSpace(query.Get("room")))
	playerName := strings.TrimSpace(query.Get("name"))
	avatar := strings.TrimSpace(query.Get("avatar"))
	isHost := query.Get("is_host") == "true" || query.Get("is_host") == "1"

	// 1. Verifikasi sesi autentikasi dari HttpOnly cookie dc_session
	var authenticatedUser *auth.UserProfile
	database := db.GetDB()
	if cookie, err := r.Cookie("dc_session"); err == nil && cookie.Value != "" && database != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		u, err := auth.GetUserBySession(ctx, database, cookie.Value)
		cancel()
		if err == nil && u != nil {
			authenticatedUser = u
		}
	}

	playerID := ""
	if authenticatedUser != nil {
		// Kredensial server-authoritative yang aman: abaikan spoofing player_id dari query
		playerID = authenticatedUser.ID
		playerName = authenticatedUser.DisplayName
		if authenticatedUser.Avatar != "" {
			avatar = authenticatedUser.Avatar
		}
	} else {
		// Jika guest, gunakan ID yang dikirim browser atau generate baru
		playerID = strings.TrimSpace(query.Get("player_id"))
		if playerID == "" {
			playerID = fmt.Sprintf("guest_%d", time.Now().UnixNano())
		}
	}

	if playerName == "" {
		playerName = "Pemain"
	}
	if avatar == "" {
		avatar = "cat-orange"
	}

	var room *game.Room
	if roomCode == "" {
		room = manager.CreateRoom(5, 80, 6)
	} else {
		var exists bool
		room, exists = manager.GetRoom(roomCode)
		if !exists {
			if isHost {
				room, _ = manager.GetOrCreateRoom(roomCode)
			} else {
				conn.WriteJSON(game.WSMessage{
					Type:    game.MsgTypeError,
					Payload: gin.H{"message": "Room tidak ditemukan atau sudah ditutup"},
				})
				conn.Close()
				return
			}
		}
	}

	// Baca setting opsional dari query jika dikirim oleh host
	if isHost {
		if rStr := query.Get("rounds"); rStr != "" {
			if rInt, err := strconv.Atoi(rStr); err == nil && rInt >= 3 && rInt <= 10 {
				room.TotalRounds = rInt
			}
		}
		if dStr := query.Get("duration"); dStr != "" {
			if dInt, err := strconv.Atoi(dStr); err == nil && dInt >= 30 && dInt <= 180 {
				room.RoundDuration = dInt
			}
		}
		if mStr := query.Get("max_players"); mStr != "" {
			if mInt, err := strconv.Atoi(mStr); err == nil && mInt >= 2 && mInt <= 12 {
				room.MaxPlayers = mInt
			}
		}
	}

	player := game.NewPlayer(playerID, playerName, avatar, isHost, conn)

	// Pastikan data profil dan statistik pemain terdaftar di database
	if database != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		database.EnsurePlayer(ctx, playerID, playerName, avatar)

		if stats, err := database.GetPlayerStats(ctx, playerID); err == nil && stats != nil {
			player.Coins = stats.Coins
			player.Rating = stats.Rating
			player.RankTier = stats.RankTier
		}
	}

	go player.WritePump()

	if err := room.AddPlayer(player); err != nil {
		player.SendMessage(game.WSMessage{
			Type:    game.MsgTypeError,
			Payload: gin.H{"message": err.Error()},
		})
		player.Close()
		return
	}

	defer func() {
		room.RemovePlayer(player.ID)
		player.Close()
	}()

	// Read pump loop
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] Read error for %s: %v", player.Name, err)
			}
			break
		}

		var wsMsg struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}

		if err := json.Unmarshal(msgBytes, &wsMsg); err != nil {
			continue
		}

		switch wsMsg.Type {
		case game.MsgTypeStartGame:
			if err := room.StartGame(player.ID); err != nil {
				player.SendMessage(game.WSMessage{
					Type:    game.MsgTypeError,
					Payload: gin.H{"message": err.Error()},
				})
			}

		case game.MsgTypeDrawStroke:
			var p game.StrokePayload
			if err := json.Unmarshal(wsMsg.Payload, &p); err == nil {
				room.HandleDrawStroke(player.ID, p.Stroke)
			}

		case game.MsgTypeClearCanvas:
			room.HandleClearCanvas(player.ID)

		case game.MsgTypeUndoStroke:
			room.HandleUndoStroke(player.ID)

		case game.MsgTypeSubmitGuess:
			var p game.GuessPayload
			if err := json.Unmarshal(wsMsg.Payload, &p); err == nil {
				if err := room.HandleSubmitGuess(player.ID, p.Text); err != nil {
					player.SendMessage(game.WSMessage{
						Type:    game.MsgTypeError,
						Payload: gin.H{"message": err.Error()},
					})
				}
			}

		case game.MsgTypeSendChat:
			var p game.GuessPayload
			if err := json.Unmarshal(wsMsg.Payload, &p); err == nil {
				room.HandleSendChat(player.ID, p.Text)
			}

		case game.MsgTypeSendReaction:
			var p game.ReactionPayload
			if err := json.Unmarshal(wsMsg.Payload, &p); err == nil {
				room.HandleSendReaction(player.ID, p.Emoji)
			}

		case game.MsgTypeUseItem:
			var p game.UseItemPayload
			if err := json.Unmarshal(wsMsg.Payload, &p); err == nil {
				if err := room.HandleUseItem(player.ID, p.ItemID); err != nil {
					player.SendMessage(game.WSMessage{
						Type:    game.MsgTypeError,
						Payload: gin.H{"message": err.Error()},
					})
				}
			}

		case game.MsgTypeRestartGame:
			room.RestartGame(player.ID)
		}
	}
}

func main() {
	cfg := config.Load()

	// Inisialisasi database jika konfigurasi DATABASE_URL tersedia
	if cfg.DatabaseURL != "" {
		database, err := db.Connect(cfg.DatabaseURL)
		if err != nil {
			log.Printf("[Database Warning] Gagal konek DB: %v (berjalan dalam mode tanpa persistensi)", err)
		} else {
			defer database.Close()
			if err := database.Migrate(); err != nil {
				log.Printf("[Database Warning] Migrasi gagal: %v", err)
			}
		}
	}

	router := setupRouter(cfg, "./web/static", "./web/templates/*.html")

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("Drawing Chaos berjalan di http://localhost:%s (%s)", cfg.Port, cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server gagal berjalan: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Menghentikan server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("server berhenti paksa: %v", err)
	}
}
