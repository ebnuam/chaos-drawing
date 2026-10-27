package main

import (
	"log"

	"drawingchaos/internal/config"
	"drawingchaos/internal/db"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL belum diset")
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Koneksi gagal: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(); err != nil {
		log.Fatalf("Migrasi gagal: %v", err)
	}

	var currentDB string
	var currentSchema string
	var currentUser string
	var serverVersion string
	_ = database.DB.QueryRow("SELECT current_database(), current_schema(), current_user, version()").Scan(&currentDB, &currentSchema, &currentUser, &serverVersion)
	log.Printf("INFO DATABASE: db=%s, schema=%s, user=%s", currentDB, currentSchema, currentUser)

	rows, err := database.DB.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = $1 ORDER BY table_name", currentSchema)
	if err == nil {
		defer rows.Close()
		var tables []string
		for rows.Next() {
			var t string
			if err := rows.Scan(&t); err == nil {
				tables = append(tables, t)
			}
		}
		log.Printf("Tabel ditemukan (%d): %v", len(tables), tables)

		for _, t := range tables {
			var cnt int
			_ = database.DB.QueryRow("SELECT COUNT(*) FROM " + t).Scan(&cnt)
			log.Printf(" - %s: %d baris", t, cnt)
		}
	}

	log.Println("Migrasi tabel PostgreSQL Neon selesai dengan sukses!")
}
