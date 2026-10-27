// Package game menyediakan aturan permainan, manajemen room, dan bank prompt.
package game

import (
	"math/rand"
	"strings"
	"time"
	"unicode"
)

// Prompt merepresentasikan sebuah kata/frasa rahasia yang harus digambar.
type Prompt struct {
	Word     string `json:"word"`
	Category string `json:"category"`
	Hint     string `json:"hint,omitempty"`
}

var defaultPrompts = []Prompt{
	// Hewan
	{Word: "Kucing", Category: "Hewan"},
	{Word: "Gajah", Category: "Hewan"},
	{Word: "Kelinci", Category: "Hewan"},
	{Word: "Jerapah", Category: "Hewan"},
	{Word: "Kura-Kura", Category: "Hewan"},
	{Word: "Singa", Category: "Hewan"},
	{Word: "Panda", Category: "Hewan"},
	{Word: "Katak", Category: "Hewan"},
	{Word: "Burung Hantu", Category: "Hewan"},
	{Word: "Bebek", Category: "Hewan"},
	{Word: "Kupu-Kupu", Category: "Hewan"},
	{Word: "Ikan Paus", Category: "Hewan"},
	{Word: "Dinosaurus", Category: "Hewan"},
	{Word: "Lumba-Lumba", Category: "Hewan"},
	{Word: "Ular", Category: "Hewan"},
	{Word: "Seto Rawon", Category: "Orang"},

	// Makanan
	{Word: "Nasi Goreng", Category: "Makanan"},
	{Word: "Bakso", Category: "Makanan"},
	{Word: "Pizza", Category: "Makanan"},
	{Word: "Es Krim", Category: "Makanan"},
	{Word: "Sate Ayam", Category: "Makanan"},
	{Word: "Burger", Category: "Makanan"},
	{Word: "Martabak", Category: "Makanan"},
	{Word: "Semangka", Category: "Makanan"},
	{Word: "Pisang", Category: "Makanan"},
	{Word: "Donat", Category: "Makanan"},
	{Word: "Roti Bakar", Category: "Makanan"},
	{Word: "Telur Mata Sapi", Category: "Makanan"},

	// Benda & Kendaraan
	{Word: "Helikopter", Category: "Kendaraan"},
	{Word: "Kapal Selam", Category: "Kendaraan"},
	{Word: "Sepeda", Category: "Kendaraan"},
	{Word: "Roket", Category: "Kendaraan"},
	{Word: "Gitar", Category: "Benda"},
	{Word: "Payung", Category: "Benda"},
	{Word: "Kacamata", Category: "Benda"},
	{Word: "Jam Dinding", Category: "Benda"},
	{Word: "Kamera", Category: "Benda"},
	{Word: "Lampu Tidur", Category: "Benda"},
	{Word: "Sepatu Roda", Category: "Benda"},
	{Word: "Mahkota", Category: "Benda"},

	// Alam & Karakter
	{Word: "Gunung Berapi", Category: "Alam"},
	{Word: "Pelangi", Category: "Alam"},
	{Word: "Pohon Kelapa", Category: "Alam"},
	{Word: "Air Terjun", Category: "Alam"},
	{Word: "Astronot", Category: "Profesi"},
	{Word: "Penyihir", Category: "Karakter"},
	{Word: "Bajak Laut", Category: "Karakter"},
	{Word: "Superhero", Category: "Karakter"},
	{Word: "Detektif", Category: "Profesi"},
	{Word: "Robot", Category: "Karakter"},
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// GetRandomPrompt memilih prompt rahasia secara acak yang belum dipakai pada sesi saat ini.
func GetRandomPrompt(used map[string]bool) Prompt {
	var candidates []Prompt
	for _, p := range defaultPrompts {
		if !used[strings.ToLower(p.Word)] {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		// Jika semua sudah terpakai, gunakan kembali seluruh daftar
		candidates = defaultPrompts
	}
	idx := rng.Intn(len(candidates))
	return candidates[idx]
}

// GenerateHint membuat petunjuk huruf berupa garis bawah, mempertahankan spasi dan tanda minus.
// Contoh: "Burung Hantu" -> "_ _ _ _ _ _   _ _ _ _ _"
func GenerateHint(word string, revealFraction float64) string {
	runes := []rune(word)
	result := make([]string, len(runes))

	// Kumpulkan indeks huruf alfabet yang bisa di-reveal
	var letterIndices []int
	for i, r := range runes {
		if unicode.IsLetter(r) {
			letterIndices = append(letterIndices, i)
		}
	}

	// Tentukan berapa banyak huruf yang dibuka (jika revealFraction > 0)
	toReveal := int(float64(len(letterIndices)) * revealFraction)
	revealedSet := make(map[int]bool)
	if toReveal > 0 && len(letterIndices) > 0 {
		// Buka huruf pertama atau acak
		revealedSet[letterIndices[0]] = true
		for len(revealedSet) < toReveal {
			pick := letterIndices[rng.Intn(len(letterIndices))]
			revealedSet[pick] = true
		}
	}

	for i, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			// Spasi atau tanda baca ditampilkan apa adanya
			if unicode.IsSpace(r) {
				result[i] = "  "
			} else {
				result[i] = string(r)
			}
		} else if revealedSet[i] {
			result[i] = strings.ToUpper(string(r))
		} else {
			result[i] = "_"
		}
	}

	return strings.Join(result, " ")
}

// NormalizeGuess membersihkan input tebakan dari spasi berlebih, tanda baca, dan case sensitivity.
func NormalizeGuess(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	// Normalisasi multiple spasi menjadi single space
	fields := strings.Fields(b.String())
	return strings.Join(fields, " ")
}

// IsMatchGuess memeriksa apakah tebakan pemain cocok dengan kata rahasia.
func IsMatchGuess(guess, secret string) bool {
	return NormalizeGuess(guess) == NormalizeGuess(secret)
}
