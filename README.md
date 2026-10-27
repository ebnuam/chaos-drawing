# Drawing Chaos Online 🎨💥

Party game multiplayer online berbasis browser dengan konsep **Gambar lalu Tebak**, dilengkapi sistem **Akun Permanen (Register & Login)**, **Sesi Cookie HttpOnly**, **7 Chaos Items**, **Ekonomi Koin**, **Rank Kompetitif (Bronze → Diamond)**, **Bank Kata Terpusat (110 Kata Indonesia + Petunjuk)**, **Profil Pemain & Riwayat Pertandingan**, serta **Persistensi Database PostgreSQL Neon**.

Proyek dibangun menggunakan:
- **Backend:** Go (Gin + Gorilla WebSocket) & PostgreSQL (Neon Cloud DB via driver `pgx/v5`).
- **Autentikasi & Keamanan:** Bcrypt Hashing, HttpOnly Cookie Sessions, CSRF/Origin Handshake Validation, In-memory Sliding Window Rate Limiting.
- **Frontend:** HTML5 Semantik, CSS3 Neobrutalism Modern, Vanilla JavaScript, Canvas API, dan Web Audio API murni.

---

## 🚀 Fitur Utama

1. **Sistem Autentikasi Akun Permanen (Register, Login & Logout):**
   - **Register:** Mendaftarkan akun permanen dengan Username unik (3-32 karakter), Email unik valid, Password (bcrypt hashing), konfirmasi password, dan avatar pilihan.
   - **Migrasi / Penautan Guest Account:** Pemain yang sebelumnya bermain sebagai tamu (guest) dapat mencentang *"Tautkan progres bermain saat ini ke akun ini"*, sehingga seluruh saldo koin, rating MMR, dan riwayat pertandingan otomatis ditautkan ke akun permanen baru tanpa kehilangan data.
   - **Login:** Masuk menggunakan Username atau Email beserta Password.
   - **Logout:** Menghapus sesi di database dan menghapus cookie HttpOnly klien secara aman.
   - **Sesi Aman HttpOnly:** Token sesi 32-byte acak kriptografis (64 hex characters) disimpan di tabel `sessions` dengan masa berlaku 30 hari, disajikan melalui cookie `HttpOnly; SameSite=Lax`.
   - **Proteksi WebSocket Handshake:** Handshake `/ws` memverifikasi sesi cookie `dc_session` secara server-authoritative, mencegah pemalsuan identitas (*player_id spoofing*) dan memvalidasi `Origin`.
   - **Rate Limiting:** Sliding-window rate limiter per IP membatasi percobaan brute-force pada endpoint autentikasi.

2. **Bank Kata Terpusat & Petunjuk Bertahap (Centralized Word Bank):**
   - Tersimpan di tabel `words` dan `word_clues` PostgreSQL dengan **110 kata bahasa Indonesia** dan petunjuk relevan dalam 10 kategori:
     1. Hewan (Kucing, Gajah, Harimau, Jerapah, dll.)
     2. Makanan & Minuman (Nasi Goreng, Rendang, Sate Ayam, Bakso, dll.)
     3. Benda Sehari-hari (Sepeda, Jam Dinding, Kacamata, Termos, dll.)
     4. Profesi (Dokter, Guru, Polisi, Koki, Pilot, dll.)
     5. Tempat (Sekolah, Rumah Sakit, Bandara, Stasiun, dll.)
     6. Transportasi (Pesawat Terbang, Kapal Laut, Kereta Api, dll.)
     7. Alam & Cuaca (Matahari, Pelangi, Petir, Gunung Berapi, dll.)
     8. Olahraga (Sepak Bola, Bulu Tangkis, Bola Basket, Renang, dll.)
     9. Hiburan & Seni (Gitar, Piano, Lukisan, Kamera, dll.)
     10. Teknologi (Komputer, Ponsel Pintar, Robot, Drone, Roket, dll.)
   - Kata rahasia **hanya dikirim ke penggambar**, penebak hanya menerima pola garis bawah (*masked*).
   - Petunjuk bertahap (*clue*) otomatis dibongkar dan disiarkan di chat pada sisa waktu **50%** (Petunjuk 1) dan **25%** (Petunjuk 2).

3. **7 Chaos Items Lengkap (Interaktif dengan Koin):**
   - **🛢️ Lempar Oli (25 🪙, 5s):** Menyiram kanvas dengan cipratan oli hitam pekat.
   - **💣 Bom Tinta (35 🪙, 6s):** Ledakan tinta ungu gelap mengaburkan pandangan kanvas.
   - **🌫️ Kabut Tebal (30 🪙, 7s):** Menutup kanvas dengan kabut tebal berawan.
   - **🪰 Lalat Pengganggu (40 🪙, 8s):** Lalat animasi mendengung yang terbang acak mengganggu penglihatan kanvas.
   - **🌋 Gempa Kanvas (45 🪙, 5s):** Mengguncang kanvas gambar dengan getaran gempa dahsyat.
   - **✨ Pembersih Efek (20 🪙):** Menghapus seketika seluruh efek gangguan chaos di kanvas pemain!
   - **⚡ Mode Turbo (50 🪙, 10s):** Menambah 10 detik waktu ronde dan memicu bonus skor.
   - Seluruh mutasi koin divalidasi dengan transaksi database `FOR UPDATE` (saldo tidak bisa negatif).

4. **Profil Pemain & Statistik Akurasi:**
   - Modal Profil lengkap (klik pill statistik di topbar) menyajikan:
     - Avatar, Username, Email / Status Akun, Rank Tier, Rating MMR, dan Saldo Koin.
     - Statistik komprehensif: Pertandingan Dimainkan, Menang, Kalah, Win Rate (%), Tebakan Benar, Total Tebakan, **Akurasi Tebakan (%)**, dan Gambar Dibuat.
     - **Tabel 10 Riwayat Pertandingan Terakhir:** Menampilkan kode room, peringkat akhir, skor, perolehan koin, kenaikan/penurunan MMR, dan waktu selesai.

5. **Rank Kompetitif & Papan Peringkat Global:**
   - Peringkat rank:
     - 🥉 **Bronze** (< 1100 MMR)
     - 🥈 **Silver** (1100–1299 MMR)
     - 🥇 **Gold** (1300–1599 MMR)
     - 💎 **Platinum** (1600–1999 MMR)
     - 👑 **Diamond** (≥ 2000 MMR)
   - Modal Papan Peringkat Global (ikon 🏆 di topbar) dengan 3 tab filter:
     - **Rating Kompetitif** (MMR Tertinggi)
     - **Kemenangan Terbanyak** (Total Wins)
     - **Total Skor** (Akumulasi Skor)

6. **Kanvas Bersih Otomatis (*Canvas Auto-Clear*):**
   - Setiap kali terjadi pergantian ronde atau game restart, kanvas dibersihkan secara instan di seluruh layar pemain.

---

## 🛠️ Prasyarat & Menjalankan Proyek

### Prasyarat
- **Go** versi 1.22 atau lebih baru.
- Konfigurasi `DATABASE_URL` di file `.env` (sudah disiapkan untuk PostgreSQL Neon).

### Menjalankan Server Lokal
1. Buka terminal di folder proyek:
   ```bash
   cd c:\Users\USER\OpenCodeGO
   ```

2. Jalankan aplikasi:
   ```bash
   go run ./cmd/server
   ```
   *(Server otomatis memeriksa koneksi Neon PostgreSQL dan mengeksekusi migrasi skema tabel serta seeding data secara aman).*

3. Buka browser:
   - Akses: [http://localhost:8080](http://localhost:8080)

---

## 🧪 Menjalankan Pengujian Otomatis

Jalankan pengujian unit untuk memverifikasi autentikasi, hashing password, sesi cookie, rate limiting, logika room, WebSocket, database, dan rule scoring:
```bash
go test ./... -v
```

---

## 📂 Struktur Proyek

```text
OpenCodeGO/
├── cmd/
│   ├── migrate/main.go            # Skrip runner migrasi skema & seed Neon DB
│   └── server/
│       ├── main.go                # Server utama Gin, Auth API, & WebSocket handler
│       └── main_test.go           # Unit test endpoint HTTP & auth
├── internal/
│   ├── auth/
│   │   ├── auth.go                # Bcrypt, sessions, rate limiter, register/login
│   │   └── auth_test.go           # Unit test autentikasi
│   ├── config/config.go           # Pemuatan environment variable (.env)
│   ├── db/
│   │   ├── db.go                  # Koneksi connection pool PostgreSQL & migrasi
│   │   ├── words.go               # Bank kata terpusat (110 kata + petunjuk)
│   │   ├── player.go              # Profil, statistik, dan riwayat pertandingan
│   │   ├── item.go                # Katalog & transaksi atomik 7 Chaos Items
│   │   └── match.go               # Pencatatan hasil match & leaderboard
│   └── game/
│       ├── manager.go             # Manajemen room singleton
│       ├── room.go                # Engine giliran, timer, scoring, tebakan, & items
│       ├── player.go              # State pemain aktif & WebSocket read/write pump
│       ├── prompt.go              # Masking kata & normalisasi tebakan
│       └── messages.go            # Payload protokol WebSocket
├── web/
│   ├── static/
│   │   ├── css/
│   │   │   ├── app.css            # Desain UI, auth modals, overlays chaos items
│   │   │   └── tokens.css         # Variabel warna & tipografi neobrutalism
│   │   └── js/
│   │       ├── app.js             # Controller SPA, auth flows, WS client, profile
│   │       ├── canvas.js          # Canvas API drawing engine (mouse & touch)
│   │       ├── audio.js           # Web Audio API synthesizers
│   │       └── avatars.js         # SVG avatar vector procedural
│   └── templates/
│       └── app.html               # Single Page Application HTML5 & dialogs
├── .env                           # Konfigurasi environment (DATABASE_URL Neon)
└── go.mod                         # Go module dependencies
```
