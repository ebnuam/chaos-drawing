package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"drawingchaos/internal/config"
)

func TestServerRoutes(t *testing.T) {
	cfg := config.Config{
		Port:   "8080",
		AppEnv: "test",
	}

	router := setupRouter(cfg, "../../web/static", "../../web/templates/*.html")

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		containsBody   string
	}{
		{
			name:           "Healthz endpoint",
			method:         "GET",
			url:            "/healthz",
			expectedStatus: http.StatusOK,
			containsBody:   `"app":"drawing-chaos"`,
		},
		{
			name:           "Home page",
			method:         "GET",
			url:            "/",
			expectedStatus: http.StatusOK,
			containsBody:   "Drawing Chaos",
		},
		{
			name:           "Lobby page",
			method:         "GET",
			url:            "/lobby",
			expectedStatus: http.StatusOK,
			containsBody:   "Kode room",
		},
		{
			name:           "Play page",
			method:         "GET",
			url:            "/play",
			expectedStatus: http.StatusOK,
			containsBody:   "board",
		},
		{
			name:           "Static CSS app.css",
			method:         "GET",
			url:            "/static/css/app.css",
			expectedStatus: http.StatusOK,
			containsBody:   "Drawing Chaos Online",
		},
		{
			name:           "Static JS app.js",
			method:         "GET",
			url:            "/static/js/app.js",
			expectedStatus: http.StatusOK,
			containsBody:   "SoundManager",
		},
		{
			name:           "Static JS canvas.js",
			method:         "GET",
			url:            "/static/js/canvas.js",
			expectedStatus: http.StatusOK,
			containsBody:   "DrawingBoard",
		},
		{
			name:           "Static JS avatars.js",
			method:         "GET",
			url:            "/static/js/avatars.js",
			expectedStatus: http.StatusOK,
			containsBody:   "AvatarRegistry",
		},
		{
			name:           "Static JS audio.js",
			method:         "GET",
			url:            "/static/js/audio.js",
			expectedStatus: http.StatusOK,
			containsBody:   "SoundManager",
		},
		{
			name:           "Static favicon.svg",
			method:         "GET",
			url:            "/static/img/favicon.svg",
			expectedStatus: http.StatusOK,
			containsBody:   "<svg",
		},
		{
			name:           "Create room via API",
			method:         "POST",
			url:            "/api/rooms",
			expectedStatus: http.StatusOK,
			containsBody:   `"code":`,
		},
		{
			name:           "Get player profile API",
			method:         "GET",
			url:            "/api/players/test_player_1/profile",
			expectedStatus: http.StatusOK,
			containsBody:   `"playerId"`,
		},
		{
			name:           "Get leaderboard API",
			method:         "GET",
			url:            "/api/leaderboard",
			expectedStatus: http.StatusOK,
			containsBody:   `"leaderboard"`,
		},
		{
			name:           "Get catalog items API",
			method:         "GET",
			url:            "/api/items",
			expectedStatus: http.StatusOK,
			containsBody:   `"items"`,
		},
		{
			name:           "404 Not Found",
			method:         "GET",
			url:            "/not-a-real-page",
			expectedStatus: http.StatusNotFound,
			containsBody:   "Halaman tidak ditemukan",
		},
		{
			name:           "Auth me unauthenticated",
			method:         "GET",
			url:            "/api/auth/me",
			expectedStatus: http.StatusOK,
			containsBody:   `"authenticated":false`,
		},
		{
			name:           "Auth logout",
			method:         "POST",
			url:            "/api/auth/logout",
			expectedStatus: http.StatusOK,
			containsBody:   `"success":true`,
		},
		{
			name:           "Player profile",
			method:         "GET",
			url:            "/api/players/p_test_123/profile",
			expectedStatus: http.StatusOK,
			containsBody:   `"playerId":"p_test_123"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.url, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			if tc.containsBody != "" && !strings.Contains(w.Body.String(), tc.containsBody) {
				t.Errorf("expected body to contain '%s', got '%s'", tc.containsBody, w.Body.String())
			}
		})
	}
}
