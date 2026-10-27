// Drawing Chaos Online — Main Application Controller
// Multiplayer Real-time WebSocket + Mode Gambar lalu Tebak

(function (window, document) {
  'use strict';

  // Persistent Player ID (Guest Token stabil di browser)
  function getOrCreatePlayerId() {
    let id = localStorage.getItem('dc_player_id');
    if (!id) {
      id = 'p_' + Math.random().toString(36).substring(2, 10) + Date.now().toString(36);
      localStorage.setItem('dc_player_id', id);
    }
    return id;
  }

  // State aplikasi
  const state = {
    screen: 'home',
    player: {
      id: getOrCreatePlayerId(),
      name: localStorage.getItem('dc_player_name') || '',
      avatar: localStorage.getItem('dc_player_avatar') || 'cat-orange',
      isHost: false,
      score: 0,
      coins: 100,
      rating: 1000,
      rankTier: 'Bronze',
      hasGuessed: false
    },
    room: {
      code: '',
      isHost: false,
      rounds: 5,
      duration: 80,
      maxPlayers: 6,
      currentRound: 0,
      totalRounds: 5,
      timeRemaining: 0,
      drawerId: '',
      drawerName: '',
      secretWord: '',
      hint: '',
      category: '',
      players: [],
      status: 'waiting'
    },
    ws: null,
    board: null,
    isDrawer: false
  };

  // Helper untuk toast notifikasi
  function showToast(message, icon = '✦') {
    const region = document.getElementById('toast-region');
    if (!region) return;

    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.innerHTML = `<span aria-hidden="true">${icon}</span> <span>${message}</span>`;
    region.appendChild(toast);

    setTimeout(() => {
      toast.style.opacity = '0';
      toast.style.transform = 'translateY(10px)';
      toast.style.transition = 'all 0.2s ease';
      setTimeout(() => toast.remove(), 250);
    }, 2800);
  }

  // Helper floating emoji reaksi
  function spawnFloatingEmoji(emoji) {
    const layer = document.getElementById('reaction-layer');
    if (!layer) return;

    const el = document.createElement('div');
    el.className = 'floating-emoji';
    el.textContent = emoji;

    // Posisi acak horizontal di bagian bawah layar
    const xPos = 20 + Math.random() * 60;
    el.style.left = `${xPos}%`;
    el.style.bottom = '80px';

    layer.appendChild(el);
    setTimeout(() => el.remove(), 2300);
  }

  // Sinkronisasi status profil (Koin & Rank Tier) ke antarmuka
  function updatePlayerStatsUI() {
    const rankEl = document.getElementById('topbar-rank');
    const coinsEl = document.getElementById('topbar-coins');
    const userLabel = document.getElementById('topbar-username');
    if (rankEl) {
      rankEl.textContent = state.player.rankTier || 'Bronze';
      rankEl.className = `badge-rank ${state.player.rankTier || 'Bronze'}`;
    }
    if (coinsEl) {
      coinsEl.textContent = state.player.coins;
    }
    if (userLabel) {
      userLabel.textContent = state.player.name || 'Tamu';
    }
  }

  async function checkAuthSession() {
    try {
      const resp = await fetch('/api/auth/me');
      if (resp.ok) {
        const data = await resp.json();
        const unauthBox = document.getElementById('topbar-auth-actions');
        const userLabel = document.getElementById('topbar-username');

        if (data.authenticated && data.user) {
          state.player.id = data.user.id;
          state.player.name = data.user.displayName || data.user.username;
          state.player.avatar = data.user.avatar || 'cat-orange';
          state.player.isGuest = false;
          state.auth = data.user;

          if (unauthBox) unauthBox.style.display = 'none';
          if (userLabel) userLabel.textContent = state.player.name;

          const nameInput = document.getElementById('player-name');
          if (nameInput && !nameInput.value) nameInput.value = state.player.name;

          if (data.user.stats) {
            state.player.coins = data.user.stats.coins;
            state.player.rating = data.user.stats.rating;
            state.player.rankTier = data.user.stats.rankTier;
          }
          updatePlayerStatsUI();
          return;
        }
      }
    } catch (_) {}

    // Fallback mode guest
    state.player.isGuest = true;
    state.auth = null;
    const unauthBox = document.getElementById('topbar-auth-actions');
    const userLabel = document.getElementById('topbar-username');
    if (unauthBox) unauthBox.style.display = 'flex';
    if (userLabel) userLabel.textContent = state.player.name || 'Tamu';
    fetchPlayerProfile();
  }

  async function fetchPlayerProfile() {
    try {
      const resp = await fetch(`/api/players/${state.player.id}/profile`);
      if (resp.ok) {
        const data = await resp.json();
        const s = data.stats || data;
        if (typeof s.coins === 'number') state.player.coins = s.coins;
        if (typeof s.rating === 'number') state.player.rating = s.rating;
        if (s.rankTier) state.player.rankTier = s.rankTier;
        updatePlayerStatsUI();
      }
    } catch (_) {
      // Abaikan jika offline
    }
  }

  // ================= TEMA & SUARA =================

  function initTheme() {
    const btnTheme = document.getElementById('btn-theme');
    const savedTheme = localStorage.getItem('dc_theme');
    const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
    const initialTheme = savedTheme || (prefersDark ? 'dark' : 'light');

    document.documentElement.setAttribute('data-theme', initialTheme);

    if (btnTheme) {
      btnTheme.addEventListener('click', () => {
        const current = document.documentElement.getAttribute('data-theme');
        const next = current === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', next);
        localStorage.setItem('dc_theme', next);
        if (window.SoundManager) window.SoundManager.playClick();
      });
    }
  }

  function initSoundDialog() {
    const btnSound = document.getElementById('btn-sound');
    const dialog = document.getElementById('sound-dialog');
    if (!btnSound || !dialog) return;

    const musicInput = document.getElementById('music-volume');
    const musicOutput = document.getElementById('music-volume-value');
    const sfxInput = document.getElementById('sfx-volume');
    const sfxOutput = document.getElementById('sfx-volume-value');
    const muteInput = document.getElementById('toggle-mute');
    const effectsInput = document.getElementById('toggle-effects');

    const sm = window.SoundManager;
    if (sm) {
      if (musicInput) {
        musicInput.value = Math.round(sm.musicVolume * 100);
        if (musicOutput) musicOutput.value = `${musicInput.value}%`;
      }
      if (sfxInput) {
        sfxInput.value = Math.round(sm.sfxVolume * 100);
        if (sfxOutput) sfxOutput.value = `${sfxInput.value}%`;
      }
      if (muteInput) muteInput.checked = sm.isMuted;
      if (effectsInput) effectsInput.checked = sm.effectsReduced;
      if (sm.isMuted) document.body.classList.add('sound-muted');
    }

    btnSound.addEventListener('click', () => {
      if (sm) sm.ensureContext();
      dialog.showModal();
    });

    if (musicInput) {
      musicInput.addEventListener('input', () => {
        const val = parseInt(musicInput.value, 10);
        if (musicOutput) musicOutput.value = `${val}%`;
        if (sm) sm.setMusicVolume(val / 100);
      });
    }

    if (sfxInput) {
      sfxInput.addEventListener('input', () => {
        const val = parseInt(sfxInput.value, 10);
        if (sfxOutput) sfxOutput.value = `${val}%`;
        if (sm) sm.setSfxVolume(val / 100);
      });
      sfxInput.addEventListener('change', () => {
        if (sm) sm.playPop();
      });
    }

    if (muteInput) {
      muteInput.addEventListener('change', () => {
        const muted = muteInput.checked;
        if (sm) sm.setMuted(muted);
        document.body.classList.toggle('sound-muted', muted);
      });
    }

    if (effectsInput) {
      effectsInput.addEventListener('change', () => {
        const reduced = effectsInput.checked;
        if (sm) sm.setEffectsReduced(reduced);
        document.body.classList.toggle('effects-reduced', reduced);
      });
    }
  }

  // ================= ROUTING & TRANSISI LAYAR =================

  function showScreen(screenId, updateHistory = true) {
    const screens = {
      home: document.getElementById('screen-home'),
      lobby: document.getElementById('screen-lobby'),
      play: document.getElementById('screen-play')
    };

    Object.entries(screens).forEach(([key, el]) => {
      if (el) {
        el.hidden = key !== screenId;
      }
    });

    state.screen = screenId;

    if (updateHistory) {
      const path = screenId === 'home' ? '/' : `/${screenId}`;
      if (window.location.pathname !== path) {
        window.history.pushState({ screen: screenId }, '', path);
      }
    }

    if (screenId === 'lobby') {
      renderLobby();
      if (window.SoundManager) window.SoundManager.startMusic('lobby');
    } else if (screenId === 'play') {
      if (window.SoundManager) window.SoundManager.startMusic('play');
    } else {
      if (window.SoundManager) window.SoundManager.stopMusic();
    }
  }

  function initRouting() {
    window.addEventListener('popstate', () => {
      const path = window.location.pathname.replace(/^\//, '') || 'home';
      if (['home', 'lobby', 'play'].includes(path)) {
        showScreen(path, false);
      }
    });

    document.querySelectorAll('[data-nav="home"]').forEach(el => {
      el.addEventListener('click', (e) => {
        e.preventDefault();
        disconnectWebSocket();
        showScreen('home');
      });
    });

    // Jalankan dari home secara default
    showScreen('home', false);
  }

  // ================= WEBSOCKET CLIENT =================

  function connectWebSocket(roomCode, isHost = false) {
    disconnectWebSocket();

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const params = new URLSearchParams({
      room: roomCode,
      player_id: state.player.id,
      name: state.player.name,
      avatar: state.player.avatar,
      is_host: isHost ? '1' : '0',
      rounds: state.room.rounds.toString(),
      duration: state.room.duration.toString(),
      max_players: state.room.maxPlayers.toString()
    });

    const url = `${protocol}//${host}/ws?${params.toString()}`;
    const ws = new WebSocket(url);

    ws.onopen = () => {
      logWS('Terhubung ke server game.');
      state.ws = ws;
    };

    ws.onmessage = (event) => {
      try {
        const lines = event.data.split('\n');
        for (const line of lines) {
          if (!line.trim()) continue;
          const msg = JSON.parse(line);
          handleServerMessage(msg);
        }
      } catch (err) {
        console.error('Error parse pesan WS:', err, event.data);
      }
    };

    ws.onclose = (e) => {
      logWS('Koneksi WebSocket terputus.');
      state.ws = null;
      if (state.screen !== 'home') {
        showToast('Koneksi ke room terputus.', '⚠️');
      }
    };

    ws.onerror = (err) => {
      console.error('WebSocket error:', err);
    };
  }

  function disconnectWebSocket() {
    if (state.ws) {
      state.ws.close();
      state.ws = null;
    }
  }

  function sendWS(type, payload = {}) {
    if (state.ws && state.ws.readyState === WebSocket.OPEN) {
      state.ws.send(JSON.stringify({ type, payload }));
    }
  }

  function logWS(msg) {
    console.log(`[DrawingChaos] ${msg}`);
  }

  // ================= DISPATCH PESAN SERVER =================

  function handleServerMessage(msg) {
    switch (msg.type) {
      case 'room_state':
        onRoomState(msg.payload);
        break;

      case 'stroke_drawn':
        if (!state.isDrawer && state.board && msg.payload && msg.payload.stroke) {
          state.board.applyRemoteStroke(msg.payload.stroke);
        }
        break;

      case 'canvas_cleared':
        if (state.board) {
          state.board.applyRemoteClear();
        }
        break;

      case 'stroke_undone':
        if (state.board) {
          state.board.applyRemoteUndo();
        }
        break;

      case 'chat_message':
        onChatMessage(msg.payload);
        break;

      case 'timer_tick':
        onTimerTick(msg.payload);
        break;

      case 'round_ended':
        onRoundEnded(msg.payload);
        break;

      case 'game_ended':
        onGameEnded(msg.payload);
        break;

      case 'reaction_broadcast':
        if (msg.payload && msg.payload.emoji) {
          spawnFloatingEmoji(msg.payload.emoji);
          if (window.SoundManager) window.SoundManager.playPop();
        }
        break;

      case 'item_activated':
        onItemActivated(msg.payload);
        break;

      case 'error':
        if (msg.payload && msg.payload.message) {
          showToast(msg.payload.message, '⚠️');
        }
        break;
    }
  }

  // ================= EVENT HANDLERS SERVER =================

  function onRoomState(payload) {
    if (!payload) return;

    const roundChanged = payload.currentRound !== state.room.currentRound;
    const stateChanged = payload.state !== state.room.status;

    // Bersihkan kanvas otomatis saat pergantian ronde atau pergantian state permainan
    if (roundChanged || stateChanged) {
      if (state.board) {
        state.board.clear(false);
      }
    }

    state.room.code = payload.roomCode;
    state.room.status = payload.state;
    state.room.players = payload.players || [];
    state.room.currentRound = payload.currentRound || 0;
    state.room.totalRounds = payload.totalRounds || 5;
    state.room.roundDuration = payload.roundDuration || 80;
    state.room.timeRemaining = payload.timeRemaining || 0;
    state.room.drawerId = payload.drawerId || '';
    state.room.drawerName = payload.drawerName || '';
    state.room.secretWord = payload.secretWord || '';
    state.room.hint = payload.hint || '';
    state.room.category = payload.category || '';

    // Cari player kita sendiri dalam daftar pemain server
    const me = state.room.players.find(p => p.name === state.player.name);
    if (me) {
      state.player.id = me.id;
      state.player.isHost = me.isHost;
      state.player.hasGuessed = me.hasGuessedCorrectly;
    }

    state.isDrawer = (state.player.id && state.player.id === state.room.drawerId) ||
                     (state.player.name && state.player.name === state.room.drawerName);

    // Update antarmuka berdasarkan state
    if (payload.state === 'waiting') {
      closeAllModals();
      showScreen('lobby');
      renderLobby();
    } else if (payload.state === 'drawing') {
      closeAllModals();
      showScreen('play');
      renderPlayState(payload);
    } else if (payload.state === 'round_result') {
      // Ditangani oleh onRoundEnded
    } else if (payload.state === 'game_over') {
      // Ditangani oleh onGameEnded
    }
  }

  function renderPlayState(payload) {
    const promptText = document.getElementById('prompt-text');
    const toolbar = document.querySelector('.toolbar');
    const guessInput = document.getElementById('input-guess');
    const guessSubmit = document.getElementById('btn-guess-submit');

    // Sinkronkan kanvas jika baru masuk dan ada goresan yang belum dirender
    if (state.board && payload.strokes && payload.strokes.length > 0 && state.board.strokes.length === 0) {
      state.board.importStrokes(payload.strokes);
    }

    if (state.isDrawer) {
      // PENGGAMBAR
      if (promptText) {
        promptText.innerHTML = `Kamu Menggambar: <strong style="color:var(--primary);">${state.room.secretWord}</strong> (${state.room.category})`;
      }
      if (toolbar) toolbar.style.display = 'flex';
      if (state.board) state.board.setReadOnly(false);

      if (guessInput) {
        guessInput.disabled = true;
        guessInput.placeholder = 'Kamu sedang menggambar! Buat mereka menebak!';
      }
      if (guessSubmit) guessSubmit.disabled = true;
    } else {
      // PENEBAK
      if (promptText) {
        const catStr = state.room.category ? ` [${state.room.category}]` : '';
        promptText.innerHTML = `Tebak Kata: <span style="letter-spacing:0.18em; font-weight:800; color:var(--primary);">${state.room.hint}</span>${catStr}`;
      }
      if (toolbar) toolbar.style.display = 'none';
      if (state.board) state.board.setReadOnly(true);

      if (state.player.hasGuessed) {
        if (guessInput) {
          guessInput.disabled = true;
          guessInput.placeholder = 'Tebakanmu sudah benar! Tunggu yang lain...';
        }
        if (guessSubmit) guessSubmit.disabled = true;
      } else {
        if (guessInput) {
          guessInput.disabled = false;
          guessInput.placeholder = 'Ketik tebakanmu...';
        }
        if (guessSubmit) guessSubmit.disabled = false;
      }
    }

    // Render petunjuk (clue) bantuan jika ada
    const clueBanner = document.getElementById('clue-banner');
    const clueContent = document.getElementById('clue-content');
    if (clueBanner && clueContent) {
      if (payload.clue) {
        clueContent.textContent = payload.clue;
        clueBanner.hidden = false;
      } else {
        clueBanner.hidden = true;
      }
    }

    renderPlayPlayers();
  }

  function renderPlayPlayers() {
    const playPlayers = document.getElementById('play-players');
    if (!playPlayers) return;

    playPlayers.innerHTML = '';
    state.room.players.forEach((p) => {
      const li = document.createElement('li');
      let statusClass = '';
      if (p.isDrawing) statusClass = 'is-drawing';
      else if (p.hasGuessedCorrectly) statusClass = 'has-guessed';

      li.className = `rail-player-item ${statusClass}`;
      const avatarHtml = window.AvatarRegistry ? window.AvatarRegistry.render(p.avatar) : '';
      const meTag = p.name === state.player.name ? ' (Kamu)' : '';
      const drawBadge = p.isDrawing ? ' ✏️' : '';
      const checkBadge = p.hasGuessedCorrectly ? ' ✅' : '';

      li.innerHTML = `
        <div class="rail-avatar">${avatarHtml}</div>
        <span class="rail-name" title="${p.name}">${p.name}${meTag}${drawBadge}${checkBadge}</span>
        <span class="rail-score">${p.score}</span>
      `;
      playPlayers.appendChild(li);
    });
  }

  function onChatMessage(payload) {
    const feed = document.getElementById('chat-feed');
    if (!feed || !payload) return;

    const msg = document.createElement('div');
    if (payload.isSystem) {
      msg.className = `chat-msg system ${payload.isCorrect ? 'correct' : ''}`;
      msg.textContent = payload.text;
    } else {
      msg.className = 'chat-msg';
      msg.innerHTML = `<span class="author">${escapeHtml(payload.playerName)}:</span> <span>${escapeHtml(payload.text)}</span>`;
    }

    feed.appendChild(msg);
    feed.scrollTop = feed.scrollHeight;

    if (payload.isCorrect) {
      if (window.SoundManager) window.SoundManager.playCorrect();
    } else if (!payload.isSystem) {
      if (window.SoundManager) window.SoundManager.playPop();
    }
  }

  function onTimerTick(payload) {
    const timerValue = document.getElementById('timer-value');
    const timerCard = document.getElementById('timer');
    if (!timerValue || !payload) return;

    const sec = payload.timeRemaining;
    const m = Math.floor(sec / 60);
    const s = sec % 60;
    timerValue.textContent = `${m}:${s < 10 ? '0' : ''}${s}`;

    if (sec <= 10 && sec > 0) {
      if (timerCard) timerCard.classList.add('danger');
      if (window.SoundManager) window.SoundManager.playTick();
    } else {
      if (timerCard) timerCard.classList.remove('danger');
    }
  }

  function onItemActivated(payload) {
    if (!payload) return;

    if (payload.itemId === 'oil_splash') {
      const overlay = document.getElementById('oil-splat-overlay');
      if (overlay) {
        overlay.hidden = false;
        if (window.SoundManager) window.SoundManager.playPop();
        setTimeout(() => { overlay.hidden = true; }, (payload.durationSeconds || 5) * 1000);
      }
    } else if (payload.itemId === 'ink_bomb') {
      const overlay = document.getElementById('ink-splat-overlay');
      if (overlay) {
        overlay.hidden = false;
        if (window.SoundManager) window.SoundManager.playPop();
        setTimeout(() => { overlay.hidden = true; }, (payload.durationSeconds || 6) * 1000);
      }
    } else if (payload.itemId === 'fog_screen') {
      const overlay = document.getElementById('fog-overlay');
      if (overlay) {
        overlay.hidden = false;
        if (window.SoundManager) window.SoundManager.playPop();
        setTimeout(() => { overlay.hidden = true; }, (payload.durationSeconds || 7) * 1000);
      }
    } else if (payload.itemId === 'annoying_fly') {
      const overlay = document.getElementById('fly-overlay');
      if (overlay) {
        overlay.hidden = false;
        if (window.SoundManager) window.SoundManager.playPop();
        setTimeout(() => { overlay.hidden = true; }, (payload.durationSeconds || 8) * 1000);
      }
    } else if (payload.itemId === 'canvas_earthquake') {
      const wrap = document.getElementById('canvas-wrap');
      if (wrap) {
        wrap.classList.add('canvas-earthquake');
        if (window.SoundManager) window.SoundManager.playPop();
        setTimeout(() => { wrap.classList.remove('canvas-earthquake'); }, (payload.durationSeconds || 5) * 1000);
      }
    } else if (payload.itemId === 'effect_cleaner') {
      // Bersihkan seketika seluruh efek chaos
      const overlays = document.querySelectorAll('.chaos-overlay');
      overlays.forEach(el => el.hidden = true);
      const wrap = document.getElementById('canvas-wrap');
      if (wrap) wrap.classList.remove('canvas-earthquake');
      if (window.SoundManager) window.SoundManager.playCorrect();
    } else if (payload.itemId === 'turbo_time') {
      const timerCard = document.getElementById('timer');
      if (timerCard) {
        timerCard.style.boxShadow = '0 0 20px #eab308';
        setTimeout(() => { timerCard.style.boxShadow = ''; }, 3000);
      }
      if (window.SoundManager) window.SoundManager.playStart();
    }

    showToast(`${payload.userPlayerName} mengaktifkan ${payload.itemName}!`, '💥');
    fetchPlayerProfile();
  }

  function onRoundEnded(payload) {
    if (window.SoundManager) window.SoundManager.playStart();

    const roundDialog = document.getElementById('round-dialog');
    const wordEl = document.getElementById('round-reveal-word');
    const drawerEl = document.getElementById('round-reveal-drawer');
    const scoresList = document.getElementById('round-scores-list');

    if (wordEl) wordEl.textContent = payload.secretWord || '---';
    if (drawerEl) drawerEl.textContent = `Digambar oleh ${payload.drawerName || 'Pemain'}`;

    if (scoresList && payload.scores) {
      scoresList.innerHTML = '';
      payload.scores.forEach(p => {
        const li = document.createElement('li');
        li.className = 'leaderboard-item';
        const avHtml = window.AvatarRegistry ? window.AvatarRegistry.render(p.avatar) : '';
        li.innerHTML = `
          <div class="leaderboard-avatar">${avHtml}</div>
          <span class="leaderboard-name">${escapeHtml(p.name)}</span>
          <span class="leaderboard-score">${p.score} poin</span>
        `;
        scoresList.appendChild(li);
      });
    }

    if (roundDialog && typeof roundDialog.showModal === 'function') {
      roundDialog.showModal();
    }
  }

  function onGameEnded(payload) {
    closeAllModals();
    if (window.SoundManager) window.SoundManager.playCorrect();

    const gameoverDialog = document.getElementById('gameover-dialog');
    const leaderboardList = document.getElementById('gameover-leaderboard');
    const btnPlayAgain = document.getElementById('btn-play-again');

    if (leaderboardList && payload.leaderboard) {
      leaderboardList.innerHTML = '';
      payload.leaderboard.forEach((p, idx) => {
        const li = document.createElement('li');
        li.className = `leaderboard-item rank-${idx + 1}`;
        const medal = idx === 0 ? '🥇' : idx === 1 ? '🥈' : idx === 2 ? '🥉' : `#${idx + 1}`;
        const avHtml = window.AvatarRegistry ? window.AvatarRegistry.render(p.avatar) : '';

        li.innerHTML = `
          <span class="leaderboard-rank">${medal}</span>
          <div class="leaderboard-avatar">${avHtml}</div>
          <span class="leaderboard-name">${escapeHtml(p.name)}</span>
          <span class="leaderboard-score">${p.score}</span>
        `;
        leaderboardList.appendChild(li);
      });
    }

    if (btnPlayAgain) {
      // Hanya host yang bisa merestart game ke lobby
      if (state.player.isHost) {
        btnPlayAgain.style.display = 'block';
        btnPlayAgain.textContent = 'Main Lagi (Kembali ke Lobby)';
      } else {
        btnPlayAgain.style.display = 'block';
        btnPlayAgain.textContent = 'Kembali ke Beranda';
      }
    }

    if (gameoverDialog && typeof gameoverDialog.showModal === 'function') {
      gameoverDialog.showModal();
    }

    // Refresh statistik & koin terbaru dari database
    setTimeout(fetchPlayerProfile, 1000);
  }

  function closeAllModals() {
    ['round-dialog', 'gameover-dialog', 'confirm-dialog'].forEach(id => {
      const d = document.getElementById(id);
      if (d && typeof d.close === 'function') {
        try { d.close(); } catch (_) {}
      }
    });
  }

  function escapeHtml(str) {
    if (!str) return '';
    return str.replace(/[&<>"']/g, (m) => ({
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#039;'
    }[m]));
  }

  // ================= BERANDA / FORM HOME =================

  function initHome() {
    const nameInput = document.getElementById('input-name');
    const avatarGrid = document.getElementById('avatar-grid');
    const errorName = document.getElementById('error-name');
    const errorCode = document.getElementById('error-code');
    const codeInput = document.getElementById('input-code');
    const btnJoin = document.getElementById('btn-join');
    const formHome = document.getElementById('form-home');

    if (nameInput) {
      nameInput.value = state.player.name;
      nameInput.addEventListener('input', () => {
        if (errorName) errorName.textContent = '';
        state.player.name = nameInput.value.trim();
        localStorage.setItem('dc_player_name', state.player.name);
      });
    }

    if (avatarGrid && window.AvatarRegistry) {
      window.AvatarRegistry.renderGrid(avatarGrid, state.player.avatar, (selectedId) => {
        state.player.avatar = selectedId;
        localStorage.setItem('dc_player_avatar', selectedId);
        if (window.SoundManager) window.SoundManager.playPop();
      });
    }

    function validateName() {
      const name = nameInput ? nameInput.value.trim() : '';
      if (!name) {
        if (errorName) errorName.textContent = 'Masukkan nama panggilanmu terlebih dahulu.';
        if (nameInput) nameInput.focus();
        return null;
      }
      if (name.length < 2) {
        if (errorName) errorName.textContent = 'Nama minimal 2 huruf, ya!';
        if (nameInput) nameInput.focus();
        return null;
      }
      return name;
    }

    // Tombol Buat Room (Host)
    if (formHome) {
      formHome.addEventListener('submit', async (e) => {
        e.preventDefault();
        const validName = validateName();
        if (!validName) return;

        state.player.name = validName;
        state.player.isHost = true;

        try {
          // Buat room via API backend
          const resp = await fetch('/api/rooms', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              rounds: state.room.rounds,
              duration: state.room.duration,
              maxPlayers: state.room.maxPlayers
            })
          });

          if (!resp.ok) throw new Error('Gagal membuat room');
          const data = await resp.json();

          state.room.code = data.code;
          state.room.isHost = true;

          // Sambungkan ke WebSocket
          connectWebSocket(data.code, true);

          if (window.SoundManager) window.SoundManager.playJoin();
          showScreen('lobby');
          showToast(`Room ${data.code} berhasil dibuat!`, '🎉');
        } catch (err) {
          showToast('Gagal membuat room: ' + err.message, '⚠️');
        }
      });
    }

    // Tombol Gabung Room
    if (btnJoin && codeInput) {
      btnJoin.addEventListener('click', async () => {
        const validName = validateName();
        if (!validName) return;

        const rawCode = codeInput.value.trim().toUpperCase();
        if (!rawCode || rawCode.length < 4) {
          if (errorCode) errorCode.textContent = 'Masukkan minimal 4 huruf/angka kode room.';
          codeInput.focus();
          return;
        }

        try {
          // Cek ketersediaan room di server
          const resp = await fetch(`/api/rooms/${rawCode}`);
          if (!resp.ok) {
            if (errorCode) errorCode.textContent = 'Room tidak ditemukan. Periksa kembali kodenya.';
            return;
          }
          const info = await resp.json();
          if (info.playerCount >= info.maxPlayers) {
            if (errorCode) errorCode.textContent = 'Room sudah penuh!';
            return;
          }

          state.player.name = validName;
          state.player.isHost = false;
          state.room.code = rawCode;
          state.room.isHost = false;

          // Sambungkan ke WebSocket
          connectWebSocket(rawCode, false);

          if (window.SoundManager) window.SoundManager.playJoin();
          showScreen('lobby');
          showToast(`Bergabung ke room ${rawCode}!`, '👋');
        } catch (err) {
          if (errorCode) errorCode.textContent = 'Koneksi ke server gagal.';
        }
      });
    }
  }

  // ================= LOBBY =================

  function renderLobby() {
    const codeText = document.getElementById('lobby-code-text');
    const playerCount = document.getElementById('lobby-player-count');
    const maxCount = document.getElementById('lobby-max');
    const playerGrid = document.getElementById('lobby-players');
    const meAvatar = document.getElementById('lobby-me-avatar');
    const meName = document.getElementById('lobby-me-name');
    const btnStart = document.getElementById('btn-start');
    const hostSettings = document.querySelector('.lobby-settings');

    if (codeText) codeText.textContent = state.room.code || '------';
    if (maxCount) maxCount.textContent = state.room.maxPlayers;
    if (playerCount) playerCount.textContent = state.room.players.length;

    if (meAvatar && window.AvatarRegistry) {
      meAvatar.innerHTML = window.AvatarRegistry.render(state.player.avatar);
    }
    if (meName) {
      meName.textContent = state.player.name || 'Pemain';
    }

    // Tombol Mulai Permainan
    if (btnStart) {
      if (state.player.isHost) {
        btnStart.disabled = false;
        btnStart.textContent = 'Mulai Permainan 🎨';
      } else {
        btnStart.disabled = true;
        btnStart.textContent = 'Menunggu Host Memulai...';
      }
    }

    // Sembunyikan pengaturan jika bukan host
    if (hostSettings) {
      hostSettings.style.opacity = state.player.isHost ? '1' : '0.6';
      hostSettings.style.pointerEvents = state.player.isHost ? 'auto' : 'none';
    }

    // Render daftar pemain
    if (playerGrid) {
      playerGrid.innerHTML = '';

      state.room.players.forEach((p) => {
        const isMe = p.name === state.player.name;
        const card = document.createElement('li');
        card.className = `player-card ${isMe ? 'is-me' : ''}`;

        const avatarHtml = window.AvatarRegistry ? window.AvatarRegistry.render(p.avatar) : '';
        const hostBadge = p.isHost ? '<span class="player-badge badge-host">Host</span>' : '';
        const readyBadge = '<span class="player-badge badge-ready">Siap</span>';

        card.innerHTML = `
          ${avatarHtml}
          <span class="player-name">${escapeHtml(p.name)} ${isMe ? '(Kamu)' : ''}</span>
          ${hostBadge}
          ${readyBadge}
        `;
        playerGrid.appendChild(card);
      });

      // Render slot kosong
      const emptySlots = Math.max(0, state.room.maxPlayers - state.room.players.length);
      for (let i = 0; i < emptySlots; i++) {
        const empty = document.createElement('li');
        empty.className = 'slot-empty';
        empty.textContent = 'Menunggu...';
        playerGrid.appendChild(empty);
      }
    }
  }

  function initLobby() {
    const copyBtn = document.getElementById('lobby-code');
    const btnLeave = document.getElementById('btn-leave');
    const btnStart = document.getElementById('btn-start');

    if (copyBtn) {
      copyBtn.addEventListener('click', async () => {
        const code = state.room.code;
        if (!code) return;
        try {
          await navigator.clipboard.writeText(code);
          showToast(`Kode ${code} disalin ke clipboard!`, '📋');
        } catch (_) {
          showToast(`Kode: ${code}`, '📋');
        }
        if (window.SoundManager) window.SoundManager.playClick();
      });
    }

    if (btnLeave) {
      btnLeave.addEventListener('click', () => {
        disconnectWebSocket();
        if (window.SoundManager) window.SoundManager.playLeave();
        showScreen('home');
      });
    }

    if (btnStart) {
      btnStart.addEventListener('click', () => {
        if (!state.player.isHost) return;
        if (window.SoundManager) window.SoundManager.playStart();
        // Kirim permintaan mulai game ke server
        sendWS('start_game');
      });
    }

    // Segmented controls untuk pengaturan host
    document.querySelectorAll('.segmented[data-setting]').forEach((seg) => {
      const settingKey = seg.getAttribute('data-setting');
      seg.querySelectorAll('button').forEach((btn) => {
        btn.addEventListener('click', () => {
          if (!state.player.isHost) return;

          seg.querySelectorAll('button').forEach(b => b.classList.remove('active'));
          btn.classList.add('active');

          const val = parseInt(btn.getAttribute('data-value'), 10);
          state.room[settingKey] = val;

          if (settingKey === 'maxPlayers') {
            const maxEl = document.getElementById('lobby-max');
            if (maxEl) maxEl.textContent = val;
            renderLobby();
          }

          if (window.SoundManager) window.SoundManager.playClick();
        });
      });
    });
  }

  // ================= PERMAINAN (PLAY) & TEBAKAN =================

  function initCanvasAndPlay() {
    const canvasEl = document.getElementById('board');
    if (!canvasEl || !window.DrawingBoard) return;

    // Inisialisasi engine kanvas
    state.board = new window.DrawingBoard(canvasEl, {
      width: 1200,
      height: 800,
      onStrokeComplete: (stroke) => {
        // Kirim goresan lokal ke server WebSocket
        if (state.isDrawer) {
          sendWS('draw_stroke', { stroke });
        }
      },
      onClear: () => {
        if (state.isDrawer) {
          sendWS('clear_canvas');
        }
      },
      onUndo: () => {
        if (state.isDrawer) {
          sendWS('undo_stroke');
        }
      }
    });

    window.DrawingBoard.setupToolbar(state.board, {
      colorsEl: document.getElementById('tool-colors'),
      sizesEl: document.getElementById('tool-sizes'),
      eraserBtn: document.getElementById('btn-eraser'),
      undoBtn: document.getElementById('btn-undo'),
      redoBtn: document.getElementById('btn-redo'),
      clearBtn: document.getElementById('btn-clear'),
      saveBtn: document.getElementById('btn-save'),
      confirmDialog: document.getElementById('confirm-dialog')
    });

    // Form submit tebakan
    const guessForm = document.getElementById('form-guess');
    const guessInput = document.getElementById('input-guess');

    if (guessForm && guessInput) {
      guessForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const text = guessInput.value.trim();
        if (!text) return;

        if (state.isDrawer) return;

        sendWS('submit_guess', { text });
        guessInput.value = '';
      });
    }

    // Quick chat buttons
    document.querySelectorAll('.chip-btn[data-chat]').forEach(btn => {
      btn.addEventListener('click', () => {
        const text = btn.getAttribute('data-chat');
        if (text) {
          sendWS('send_chat', { text });
        }
      });
    });

    // Reaction emoji buttons
    document.querySelectorAll('.reaction-btn[data-emoji]').forEach(btn => {
      btn.addEventListener('click', () => {
        const emoji = btn.getAttribute('data-emoji');
        if (emoji) {
          sendWS('send_reaction', { emoji });
        }
      });
    });

    // Tombol keluar saat permainan aktif
    const btnLeavePlay = document.getElementById('btn-leave-play');
    if (btnLeavePlay) {
      btnLeavePlay.addEventListener('click', () => {
        disconnectWebSocket();
        if (window.SoundManager) window.SoundManager.playLeave();
        showScreen('home');
      });
    }

    // Tombol Gunakan Chaos Items (7 Items)
    document.querySelectorAll('.item-btn[data-item]').forEach(btn => {
      btn.addEventListener('click', () => {
        const itemId = btn.getAttribute('data-item');
        if (!itemId) return;

        const costs = {
          oil_splash: 25,
          ink_bomb: 35,
          fog_screen: 30,
          annoying_fly: 40,
          canvas_earthquake: 45,
          effect_cleaner: 20,
          turbo_time: 50
        };

        const cost = costs[itemId] || 25;
        if (state.player.coins < cost) {
          showToast(`Koin tidak mencukupi (memerlukan ${cost} 🪙, saldo kamu: ${state.player.coins})`, '⚠️');
          return;
        }

        if (state.isDrawer && itemId !== 'effect_cleaner' && itemId !== 'turbo_time') {
          showToast('Sebagai penggambar, kamu hanya bisa memakai Pembersih atau Turbo!', '⚠️');
          return;
        }

        sendWS('use_item', { itemId });
      });
    });

    // Tombol Main Lagi di modal game over
    const btnPlayAgain = document.getElementById('btn-play-again');
    if (btnPlayAgain) {
      btnPlayAgain.addEventListener('click', () => {
        closeAllModals();
        if (state.player.isHost) {
          sendWS('restart_game');
        } else {
          disconnectWebSocket();
          showScreen('home');
        }
      });
    }
  }

  // ================= MODAL LEADERBOARD GLOBAL =================

  async function loadGlobalLeaderboard(sortBy = 'rating') {
    const listEl = document.getElementById('global-lb-list');
    if (!listEl) return;

    listEl.innerHTML = '<li class="leaderboard-item" style="justify-content:center;">Memuat data peringkat...</li>';

    try {
      const resp = await fetch(`/api/leaderboard?sort=${sortBy}`);
      if (!resp.ok) throw new Error('Gagal mengambil data peringkat');
      const data = await resp.json();

      listEl.innerHTML = '';
      if (!data.leaderboard || data.leaderboard.length === 0) {
        listEl.innerHTML = '<li class="leaderboard-item" style="justify-content:center;">Belum ada pertandingan terekam di database.</li>';
        return;
      }

      data.leaderboard.forEach(item => {
        const li = document.createElement('li');
        li.className = `leaderboard-item rank-${item.rankPosition}`;
        const medal = item.rankPosition === 1 ? '🥇' : item.rankPosition === 2 ? '🥈' : item.rankPosition === 3 ? '🥉' : `#${item.rankPosition}`;
        const avHtml = window.AvatarRegistry ? window.AvatarRegistry.render(item.avatar) : '';
        const rankBadge = `<span class="badge-rank ${item.rankTier}">${item.rankTier}</span>`;
        let metric = `${item.rating} MMR`;
        if (sortBy === 'wins') {
          metric = `${item.matchesWon} Menang`;
        } else if (sortBy === 'score') {
          metric = `${item.totalScore} Skor`;
        }

        li.innerHTML = `
          <span class="leaderboard-rank">${medal}</span>
          <div class="leaderboard-avatar">${avHtml}</div>
          <span class="leaderboard-name">${escapeHtml(item.displayName)}</span>
          ${rankBadge}
          <span class="leaderboard-score">${metric}</span>
        `;
        listEl.appendChild(li);
      });
    } catch (err) {
      listEl.innerHTML = `<li class="leaderboard-item" style="justify-content:center; color:var(--danger);">${err.message}</li>`;
    }
  }

  function initLeaderboard() {
    const btnOpen = document.getElementById('btn-open-leaderboard');
    const dialog = document.getElementById('leaderboard-dialog');
    const btnClose = document.getElementById('btn-close-lb');
    const tabRating = document.getElementById('tab-lb-rating');
    const tabWins = document.getElementById('tab-lb-wins');
    const tabScore = document.getElementById('tab-lb-score');

    if (!dialog) return;

    if (btnOpen) {
      btnOpen.addEventListener('click', () => {
        dialog.showModal();
        loadGlobalLeaderboard('rating');
        if (window.SoundManager) window.SoundManager.playClick();
      });
    }

    if (btnClose) {
      btnClose.addEventListener('click', () => {
        dialog.close();
      });
    }

    function selectTab(activeTab, sortBy) {
      [tabRating, tabWins, tabScore].forEach(t => t && t.classList.remove('active'));
      if (activeTab) activeTab.classList.add('active');
      loadGlobalLeaderboard(sortBy);
      if (window.SoundManager) window.SoundManager.playClick();
    }

    if (tabRating) tabRating.addEventListener('click', () => selectTab(tabRating, 'rating'));
    if (tabWins) tabWins.addEventListener('click', () => selectTab(tabWins, 'wins'));
    if (tabScore) tabScore.addEventListener('click', () => selectTab(tabScore, 'score'));
  }

  // ================= MODAL AUTENTIKASI (REGISTER & LOGIN) =================

  function initAuthModals() {
    const btnOpenLogin = document.getElementById('btn-open-login');
    const btnOpenRegister = document.getElementById('btn-open-register');
    const loginDialog = document.getElementById('login-dialog');
    const registerDialog = document.getElementById('register-dialog');
    const btnCloseLogin = document.getElementById('btn-close-login');
    const btnCloseRegister = document.getElementById('btn-close-register');
    const btnSwitchToReg = document.getElementById('btn-switch-to-register');
    const btnSwitchToLog = document.getElementById('btn-switch-to-login');

    const formLogin = document.getElementById('form-login');
    const formRegister = document.getElementById('form-register');
    const loginError = document.getElementById('login-error');
    const regError = document.getElementById('reg-error');

    let selectedRegAvatar = 'cat-orange';

    function renderRegAvatarPicker() {
      const picker = document.getElementById('reg-avatar-picker');
      if (!picker || !window.AvatarRegistry) return;
      if (typeof window.AvatarRegistry.renderGrid === 'function') {
        window.AvatarRegistry.renderGrid(picker, selectedRegAvatar, (id) => {
          selectedRegAvatar = id;
        });
      }
    }

    if (btnOpenLogin && loginDialog) {
      btnOpenLogin.addEventListener('click', () => {
        if (loginError) loginError.hidden = true;
        loginDialog.showModal();
        if (window.SoundManager) window.SoundManager.playClick();
      });
    }

    if (btnOpenRegister && registerDialog) {
      btnOpenRegister.addEventListener('click', () => {
        if (regError) regError.hidden = true;
        try {
          renderRegAvatarPicker();
        } catch (err) {
          console.error(err);
        }
        registerDialog.showModal();
        if (window.SoundManager) window.SoundManager.playClick();
      });
    }

    if (btnCloseLogin && loginDialog) {
      btnCloseLogin.addEventListener('click', () => loginDialog.close());
    }
    if (btnCloseRegister && registerDialog) {
      btnCloseRegister.addEventListener('click', () => registerDialog.close());
    }

    if (btnSwitchToReg && loginDialog && registerDialog) {
      btnSwitchToReg.addEventListener('click', () => {
        loginDialog.close();
        try {
          renderRegAvatarPicker();
        } catch (err) {
          console.error(err);
        }
        if (regError) regError.hidden = true;
        registerDialog.showModal();
      });
    }

    if (btnSwitchToLog && loginDialog && registerDialog) {
      btnSwitchToLog.addEventListener('click', () => {
        registerDialog.close();
        if (loginError) loginError.hidden = true;
        loginDialog.showModal();
      });
    }

    // Submit Login
    if (formLogin) {
      formLogin.addEventListener('submit', async (e) => {
        e.preventDefault();
        const identifier = document.getElementById('login-identifier').value.trim();
        const password = document.getElementById('login-password').value;
        const btnSubmit = document.getElementById('btn-submit-login');

        if (!identifier || !password) return;
        btnSubmit.disabled = true;
        if (loginError) loginError.hidden = true;

        try {
          const resp = await fetch('/api/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ identifier, password })
          });
          const data = await resp.json();
          if (!resp.ok) {
            throw new Error(data.error || 'Login gagal');
          }

          loginDialog.close();
          formLogin.reset();
          showToast(`Selamat datang, ${data.user.displayName}!`, '🎉');
          if (window.SoundManager) window.SoundManager.playStart();
          await checkAuthSession();
        } catch (err) {
          if (loginError) {
            loginError.textContent = err.message;
            loginError.hidden = false;
          }
        } finally {
          btnSubmit.disabled = false;
        }
      });
    }

    // Submit Register
    if (formRegister) {
      formRegister.addEventListener('submit', async (e) => {
        e.preventDefault();
        const username = document.getElementById('reg-username').value.trim();
        const email = document.getElementById('reg-email').value.trim();
        const password = document.getElementById('reg-password').value;
        const passwordConfirm = document.getElementById('reg-password-confirm').value;
        const linkGuest = document.getElementById('reg-link-guest-chk').checked;
        const btnSubmit = document.getElementById('btn-submit-register');

        if (password !== passwordConfirm) {
          if (regError) {
            regError.textContent = 'Konfirmasi password tidak cocok dengan password';
            regError.hidden = false;
          }
          return;
        }

        btnSubmit.disabled = true;
        if (regError) regError.hidden = true;

        try {
          const bodyPayload = {
            username,
            email,
            password,
            avatar: selectedRegAvatar,
            linkGuestId: linkGuest ? state.player.id : ''
          };

          const resp = await fetch('/api/auth/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(bodyPayload)
          });
          const data = await resp.json();
          if (!resp.ok) {
            throw new Error(data.error || 'Pendaftaran gagal');
          }

          registerDialog.close();
          formRegister.reset();
          showToast(`Akun permanen ${data.user.username} berhasil dibuat!`, '🚀');
          if (window.SoundManager) window.SoundManager.playStart();
          await checkAuthSession();
        } catch (err) {
          if (regError) {
            regError.textContent = err.message;
            regError.hidden = false;
          }
        } finally {
          btnSubmit.disabled = false;
        }
      });
    }
  }

  // ================= MODAL PROFIL & STATISTIK =================

  function initProfileDialog() {
    const statPill = document.getElementById('topbar-stats');
    const profDialog = document.getElementById('profile-dialog');
    const btnCloseProf = document.getElementById('btn-close-profile');
    const btnCloseProfBtm = document.getElementById('btn-close-profile-bottom');
    const btnLogout = document.getElementById('btn-logout-user');

    if (!profDialog) return;

    if (statPill) {
      statPill.addEventListener('click', async () => {
        profDialog.showModal();
        if (window.SoundManager) window.SoundManager.playClick();
        await loadProfileData();
      });
    }

    if (btnCloseProf) btnCloseProf.addEventListener('click', () => profDialog.close());
    if (btnCloseProfBtm) btnCloseProfBtm.addEventListener('click', () => profDialog.close());

    if (btnLogout) {
      btnLogout.addEventListener('click', async () => {
        try {
          await fetch('/api/auth/logout', { method: 'POST' });
          profDialog.close();
          showToast('Kamu telah berhasil logout', '👋');
          if (window.SoundManager) window.SoundManager.playLeave();
          await checkAuthSession();
        } catch (_) {
          profDialog.close();
        }
      });
    }
  }

  async function loadProfileData() {
    try {
      const resp = await fetch(`/api/players/${state.player.id}/profile`);
      if (!resp.ok) return;
      const data = await resp.json();

      const p = data.player || {};
      const s = data.stats || {};
      const matches = data.recentMatches || [];

      // Avatar & Header
      const avWrap = document.getElementById('prof-avatar-wrap');
      if (avWrap && window.AvatarRegistry) {
        avWrap.innerHTML = window.AvatarRegistry.render(p.avatar || state.player.avatar);
      }
      const profName = document.getElementById('prof-name');
      const profEmail = document.getElementById('prof-email');
      const profRank = document.getElementById('prof-rank');
      const profRating = document.getElementById('prof-rating');
      const profCoins = document.getElementById('prof-coins');
      const btnLogout = document.getElementById('btn-logout-user');

      if (profName) profName.textContent = p.displayName || state.player.name || 'Pemain';
      if (profEmail) profEmail.textContent = p.email || (p.isGuest ? 'Akun Tamu (Guest)' : 'Akun Permanen');
      if (profRank) {
        profRank.textContent = s.rankTier || 'Bronze';
        profRank.className = `badge-rank ${s.rankTier || 'Bronze'}`;
      }
      if (profRating) profRating.textContent = s.rating || 1000;
      if (profCoins) profCoins.textContent = s.coins || 100;
      if (btnLogout) btnLogout.style.display = p.isGuest ? 'none' : 'inline-flex';

      // Stats Grid
      const statMatches = document.getElementById('stat-matches');
      const statWins = document.getElementById('stat-wins');
      const statLosses = document.getElementById('stat-losses');
      const statWinrate = document.getElementById('stat-winrate');
      const statCorrect = document.getElementById('stat-correct');
      const statAttempts = document.getElementById('stat-attempts');
      const statAccuracy = document.getElementById('stat-accuracy');
      const statDrawings = document.getElementById('stat-drawings');

      if (statMatches) statMatches.textContent = s.matchesPlayed || 0;
      if (statWins) statWins.textContent = s.matchesWon || 0;
      if (statLosses) statLosses.textContent = s.matchesLost || 0;
      if (statWinrate) statWinrate.textContent = `${(s.winRate || 0).toFixed(1)}%`;
      if (statCorrect) statCorrect.textContent = s.correctGuesses || 0;
      if (statAttempts) statAttempts.textContent = s.guessAttempts || 0;
      if (statAccuracy) statAccuracy.textContent = `${(s.guessAccuracy || 0).toFixed(1)}%`;
      if (statDrawings) statDrawings.textContent = s.drawingsMade || 0;

      // Recent Matches Table
      const tbody = document.getElementById('prof-matches-tbody');
      if (tbody) {
        tbody.innerHTML = '';
        if (matches.length === 0) {
          tbody.innerHTML = '<tr><td colspan="6" class="text-center hint">Belum ada riwayat pertandingan selesai</td></tr>';
        } else {
          matches.forEach(m => {
            const tr = document.createElement('tr');
            const medal = m.rankPosition === 1 ? '🥇 #1' : m.rankPosition === 2 ? '🥈 #2' : m.rankPosition === 3 ? '🥉 #3' : `#${m.rankPosition}`;
            const deltaSign = m.ratingChange > 0 ? `+${m.ratingChange}` : `${m.ratingChange}`;
            const deltaClass = m.ratingChange > 0 ? 'text-success' : m.ratingChange < 0 ? 'text-danger' : '';
            const dateStr = m.createdAt ? new Date(m.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '-';

            tr.innerHTML = `
              <td><strong>${m.roomCode || '-'}</strong></td>
              <td>${medal}</td>
              <td>${m.finalScore}</td>
              <td>+${m.coinsEarned} 🪙</td>
              <td class="${deltaClass}">${deltaSign}</td>
              <td class="hint">${dateStr}</td>
            `;
            tbody.appendChild(tr);
          });
        }
      }
    } catch (_) {}
  }

  // ================= INISIALISASI =================

  document.addEventListener('DOMContentLoaded', () => {
    initTheme();
    initSoundDialog();
    initHome();
    initLobby();
    initCanvasAndPlay();
    initLeaderboard();
    initAuthModals();
    initProfileDialog();
    initRouting();
    checkAuthSession();
  });

})(window, document);
