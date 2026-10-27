# Drawing Chaos Online

## Project Brief untuk OpenCode

> **Tujuan:** Membangun game party multiplayer berbasis browser,
> terinspirasi dari permainan gambar-dan-tebak serta gambar berantai.
> Game harus memiliki visual menarik, animasi, audio, dan interaksi
> real-time antarpemain.
>
> **Catatan orisinalitas:** Gunakan nama, identitas visual, ilustrasi,
> ikon, dan aset buatan sendiri. Jangan menyalin logo, aset, atau materi
> berhak cipta dari game lain.

------------------------------------------------------------------------

## 1. Instruksi Utama untuk OpenCode

Baca dokumen ini sampai selesai sebelum mengubah kode apa pun.

1.  Periksa seluruh isi folder proyek dan pahami struktur serta
    teknologi yang sudah tersedia.
2.  Jika folder berisi proyek lama, identifikasi bagian yang bisa
    digunakan kembali dan bagian yang tidak relevan. Jangan menghapus
    atau menimpa pekerjaan lama tanpa alasan yang jelas.
3.  Jelaskan secara singkat kondisi proyek saat ini, struktur folder,
    dan rencana implementasi.
4.  Buat rencana kerja bertahap berdasarkan bagian **Tahapan
    Pengembangan** di bawah.
5.  Kerjakan satu tahap pada satu waktu. Setelah setiap tahap, jalankan
    pemeriksaan atau tes yang relevan, perbaiki error, dan laporkan apa
    yang berhasil serta apa yang belum.
6.  Jangan mengklaim fitur berfungsi jika belum dijalankan dan diuji.
7.  Jangan membuat semua fitur sekaligus dalam satu perubahan besar.
8.  Jika ada keputusan teknis yang belum jelas, pilih solusi sederhana
    yang sesuai dengan proyek dan jelaskan asumsi tersebut. Tanyakan
    hanya jika keputusan itu benar-benar menghalangi pekerjaan.
9.  Berikan instruksi menjalankan proyek dengan jelas, termasuk perintah
    terminal yang diperlukan.
10. Utamakan kode yang mudah dipahami dan dirawat oleh pemula.

**Mulai dengan audit dan rencana. Setelah itu, implementasikan fondasi
minimum yang dapat dijalankan.**

------------------------------------------------------------------------

## 2. Konsep Produk

**Nama kerja:** Drawing Chaos\
**Jenis:** Party game multiplayer online berbasis browser\
**Target:** Bermain bersama teman melalui room privat dan kode undangan\
**Platform awal:** Browser desktop dan mobile\
**Bahasa antarmuka awal:** Bahasa Indonesia\
**Gaya visual:** Ceria, ekspresif, modern, responsif, dan penuh karakter

Pemain dapat masuk ke room, menggambar berdasarkan prompt, berinteraksi
dengan pemain lain, lalu melihat hasil permainan dalam tampilan akhir
yang lucu dan dramatis.

Game ini perlu terasa hidup, tetapi animasi dan suara tidak boleh
mengganggu fungsi utama atau membuat perangkat lambat.

### Mode permainan yang direncanakan

Buat struktur yang memungkinkan dua mode:

1.  **Gambar lalu tebak:** seorang pemain menggambar prompt, pemain lain
    menebak.
2.  **Gambar berantai:** prompt atau deskripsi diteruskan melalui
    rangkaian pemain, bergantian antara menggambar dan mendeskripsikan.
    Semua hasil dibuka pada akhir ronde.

Untuk versi awal, implementasikan satu mode terlebih dahulu dengan alur
paling stabil. Rekomendasi awal: **gambar lalu tebak**, kemudian
tambahkan mode gambar berantai setelah fondasi multiplayer teruji.
Jangan menganggap pilihan mode atau jumlah pemain sudah final; buat
konfigurasi yang mudah diubah.

------------------------------------------------------------------------

## 3. Teknologi yang Diutamakan

Gunakan teknologi berikut jika cocok dengan kondisi repositori:

-   **Frontend:** HTML, CSS, JavaScript.
-   **Menggambar:** Canvas API.
-   **Backend:** Go dengan Gin.
-   **Real-time:** WebSocket.
-   **Penyimpanan:** mulai tanpa database jika belum diperlukan; gunakan
    PostgreSQL ketika akun, statistik, atau riwayat permanen dibutuhkan.
-   **Pengujian:** unit test untuk aturan game dan pengujian
    manual/otomatis untuk alur multiplayer.
-   **Deployment:** pilih pendekatan yang sesuai setelah aplikasi lokal
    stabil.

Jangan menambahkan framework besar, layanan berbayar, atau layanan
eksternal tanpa kebutuhan nyata. Jangan mengasumsikan kredensial atau
layanan hosting sudah tersedia.

### Prinsip arsitektur

-   Server menjadi sumber kebenaran untuk room, timer, giliran, skor,
    dan status pertandingan.
-   Browser menangani tampilan, input gambar, animasi, dan audio.
-   Validasi aturan penting dilakukan di server.
-   Data yang dikirim kepada pemain hanya boleh mencakup informasi yang
    memang boleh mereka ketahui. Contohnya, prompt rahasia tidak boleh
    dikirim kepada pemain yang tidak berhak melihatnya.
-   Pisahkan logika permainan dari kode UI agar mudah diuji.
-   Kelola koneksi terputus, pemain keluar, dan pemain yang mencoba
    bergabung di tengah ronde.
-   Jangan mengirim ulang seluruh kanvas pada setiap gerakan jika data
    goresan yang lebih kecil sudah cukup.
-   Berikan batas ukuran pesan dan validasi input untuk mengurangi
    penyalahgunaan.

------------------------------------------------------------------------

## 4. Fitur Produk

### 4.1 Halaman utama

-   Nama dan identitas visual orisinal.
-   Tombol **Buat Room** dan **Gabung Room**.
-   Kolom nama pemain.
-   Pemilihan avatar sederhana.
-   Pengaturan suara dan musik.
-   Transisi halaman yang halus.
-   Tampilan responsif untuk desktop dan mobile.

### 4.2 Room dan lobby

-   Host dapat membuat room dengan kode unik.
-   Pemain dapat bergabung menggunakan kode.
-   Tampilkan avatar, nama, dan status siap.
-   Tampilkan jumlah pemain.
-   Host dapat memulai permainan ketika syarat minimum terpenuhi.
-   Host dapat mengatur jumlah ronde, durasi, dan mode jika sudah
    didukung.
-   Tampilkan notifikasi ketika pemain bergabung atau keluar.
-   Tangani kode room yang tidak valid dan room yang penuh.
-   Room harus memiliki batas pemain yang dapat dikonfigurasi.

### 4.3 Kanvas menggambar

Implementasikan komponen kanvas dengan:

-   Pensil dan pilihan ukuran goresan.
-   Palet warna.
-   Penghapus.
-   Undo dan redo.
-   Hapus seluruh kanvas dengan konfirmasi.
-   Tombol kirim gambar.
-   Dukungan mouse dan layar sentuh.
-   Penyesuaian ukuran kanvas untuk berbagai layar.
-   Indikator waktu yang tersisa.
-   Pemulihan kondisi gambar jika koneksi sementara terputus, sejauh
    memungkinkan.

Gunakan koordinat yang dinormalisasi atau pendekatan lain yang membuat
gambar tetap proporsional pada ukuran kanvas berbeda. Jangan menganggap
ukuran layar semua pemain sama.

### 4.4 Tebakan dan skor

-   Pemain dapat mengirim tebakan melalui antarmuka yang jelas.
-   Validasi jawaban dilakukan di server.
-   Aturan skor harus konsisten dan didokumentasikan.
-   Hindari menampilkan prompt rahasia kepada penebak sebelum waktunya.
-   Beri umpan balik ketika tebakan benar, salah, atau waktu habis.
-   Cegah satu pemain mendapat poin berulang kali untuk jawaban yang
    sama.
-   Tampilkan skor sementara tanpa mengganggu kanvas.

### 4.5 Interaksi antarpemain

-   Reaksi emoji yang dapat dikirim selama ronde.
-   Emoji tampil sementara di dekat avatar atau pada area reaksi.
-   Quick chat dengan pesan yang sudah disediakan, misalnya "Aku tahu!",
    "Apa itu?", dan "Waktunya habis!".
-   Notifikasi pemain masuk/keluar.
-   Indikator siapa yang sedang menggambar.
-   Batasi frekuensi pesan dan reaksi agar tidak menjadi spam.
-   Sediakan cara mematikan atau mengecilkan efek visual yang terlalu
    ramai.

### 4.6 Animasi

Gunakan CSS Animation dan JavaScript sesuai kebutuhan. Canvas dapat
digunakan untuk efek yang memang memerlukan penggambaran dinamis.

Animasi yang direncanakan:

-   Transisi antarhalaman.
-   Avatar masuk ke lobby.
-   Tombol dan kartu prompt dengan efek hover/tekan.
-   Transisi saat ronde dimulai.
-   Timer yang berubah visual saat mendekati nol.
-   Emoji yang melayang.
-   Efek jawaban benar.
-   Pemutaran ulang goresan jika fitur replay sudah diimplementasikan.
-   Pembukaan hasil secara berurutan.
-   Perayaan pemenang dan confetti.

Jangan mengorbankan keterbacaan atau performa demi animasi. Hormati
preferensi `prefers-reduced-motion` dan sediakan cara mengurangi efek.

### 4.7 Audio

Sediakan tiga kategori audio:

1.  **Musik latar:** lobby, gameplay, dan hasil pertandingan.
2.  **Efek UI:** klik, pemain bergabung, ronde dimulai, dan timer
    selesai.
3.  **Efek pertandingan:** jawaban benar, pembukaan hasil, dan
    kemenangan.

Persyaratan:

-   Kontrol volume musik dan efek suara secara terpisah.
-   Tombol mute yang mudah ditemukan.
-   Jangan mengandalkan audio untuk menyampaikan informasi penting.
-   Tangani pembatasan autoplay browser dengan meminta interaksi awal
    pengguna.
-   Gunakan aset audio orisinal atau aset yang lisensinya jelas untuk
    penggunaan yang direncanakan.
-   Jangan membuat audio terus diputar ketika pengguna telah mematikan
    suara.

### 4.8 Hasil pertandingan

-   Tampilkan prompt awal dan hasil gambar.
-   Untuk mode gambar berantai, tampilkan rangkaian deskripsi dan gambar
    sesuai urutan.
-   Berikan kesempatan kepada pemain untuk mengirim reaksi.
-   Tampilkan skor akhir dan pemenang.
-   Tombol **Main Lagi** dan kembali ke lobby.
-   Jangan kehilangan hasil ronde sebelum seluruh pemain selesai
    melihatnya.

------------------------------------------------------------------------

## 5. Arah Desain UI/UX

Gunakan gaya visual game party modern: ceria, ekspresif, rapi, dan
memiliki karakter.

### Prinsip desain

-   Gunakan tipografi yang mudah dibaca.
-   Buat tombol utama jelas dan mudah disentuh.
-   Gunakan warna dengan kontras memadai.
-   Berikan umpan balik visual pada interaksi.
-   Buat layout berbeda yang sesuai untuk layar lebar dan layar sempit.
-   Pastikan kanvas tetap menjadi fokus saat menggambar.
-   Hindari memenuhi layar dengan emoji, pop-up, dan animasi sekaligus.
-   Sediakan tema terang dan gelap jika tidak memperlambat tahap awal.
-   Buat avatar, ikon, ilustrasi, dan identitas visual orisinal.
-   Gunakan SVG atau CSS untuk ikon dan elemen sederhana jika lebih
    efisien daripada gambar raster.

Sebelum membuat banyak halaman, buat satu sistem desain kecil: warna,
tipografi, tombol, panel, avatar, modal, notifikasi, dan komponen kartu
prompt.

------------------------------------------------------------------------

## 6. Model Status Permainan

Rancang status pertandingan secara eksplisit. Contoh status:

-   `waiting`: room menunggu pemain.
-   `ready_check`: pemain mengonfirmasi kesiapan, jika diperlukan.
-   `round_start`: ronde dimulai.
-   `drawing`: pemain menggambar.
-   `guessing`: pemain menebak.
-   `round_result`: hasil ronde ditampilkan.
-   `game_result`: hasil akhir ditampilkan.
-   `finished`: permainan selesai.

Sesuaikan status dengan mode yang benar-benar diimplementasikan. Jangan
menggunakan status ini secara mekanis jika model yang lebih sederhana
lebih tepat.

Server harus memvalidasi setiap perpindahan status dan menolak aksi yang
tidak sesuai dengan fase permainan.

------------------------------------------------------------------------

## 7. Keamanan dan Keandalan

-   Validasi semua aksi permainan di server.
-   Buat kode room yang tidak mudah ditebak dan tangani benturan kode.
-   Batasi jumlah pesan WebSocket, ukuran payload, dan frekuensi aksi.
-   Sanitasi teks yang ditampilkan.
-   Jangan pernah mengirim jawaban rahasia atau prompt tersembunyi
    kepada pemain yang tidak berhak.
-   Tangani koneksi terputus dan reconnect secara terencana.
-   Jangan mempercayai skor atau timer yang dikirim browser.
-   Jangan menyimpan rahasia, token, atau kredensial di repository.
-   Tambahkan logging yang berguna tanpa mencatat data sensitif yang
    tidak diperlukan.
-   Uji kondisi host keluar, pemain keluar, room penuh, dan aksi ganda.

------------------------------------------------------------------------

## 8. Tahapan Pengembangan

### Tahap 0 --- Audit dan rencana

-   Periksa repository dan file yang sudah ada.
-   Identifikasi bahasa, dependency, entry point, cara menjalankan, dan
    tes yang tersedia.
-   Laporkan risiko dan komponen yang dapat digunakan kembali.
-   Buat rencana implementasi kecil dengan urutan yang jelas.
-   Jangan melakukan refactor besar sebelum memahami proyek.

**Kriteria selesai:** struktur dan rencana dijelaskan; perintah
menjalankan proyek diketahui atau hambatannya didokumentasikan.

### Tahap 1 --- Fondasi UI

-   Buat halaman utama, lobby, panel pemain, dan layout gameplay.
-   Buat sistem desain dasar dan responsivitas.
-   Buat kanvas lokal yang dapat digunakan.
-   Tambahkan kontrol suara dan animasi dasar tanpa menghubungkan semua
    fitur.

**Kriteria selesai:** UI berjalan di browser, layout menyesuaikan ukuran
layar, dan alat gambar dasar bekerja secara lokal.

### Tahap 2 --- Aturan game lokal

-   Implementasikan state machine permainan.
-   Implementasikan timer, prompt, fase menggambar, fase menebak, dan
    skor.
-   Buat data prompt awal yang orisinal.
-   Uji kondisi jawaban benar, waktu habis, dan pergantian ronde.

**Kriteria selesai:** satu permainan lokal dapat diselesaikan dari awal
sampai hasil akhir tanpa kesalahan alur utama.

### Tahap 3 --- Multiplayer real-time

-   Buat server Go dan endpoint WebSocket.
-   Implementasikan pembuatan dan penggabungan room.
-   Sinkronkan status pemain, fase game, timer, dan skor.
-   Sinkronkan data goresan dengan format pesan yang efisien.
-   Validasi semua aksi di server.
-   Tangani pemain keluar dan koneksi terputus.

**Kriteria selesai:** beberapa browser dapat bermain dalam room yang
sama dan menerima state yang konsisten.

### Tahap 4 --- Interaksi dan audio

-   Tambahkan emoji dan quick chat.
-   Tambahkan notifikasi pemain.
-   Tambahkan musik dan efek suara.
-   Tambahkan kontrol volume, mute, dan pembatasan spam.
-   Uji autoplay dan perilaku audio di browser.

**Kriteria selesai:** interaksi tersinkron, audio dapat dikendalikan,
dan fitur tidak mengganggu gameplay.

### Tahap 5 --- Animasi dan hasil pertandingan

-   Tambahkan transisi dan animasi penting.
-   Tambahkan layar hasil ronde.
-   Tambahkan replay goresan jika arsitektur data mendukungnya.
-   Tambahkan hasil akhir, skor, dan perayaan pemenang.
-   Optimalkan animasi untuk perangkat dengan kemampuan terbatas.

**Kriteria selesai:** alur dari lobby sampai hasil pertandingan terasa
utuh dan animasi tetap responsif.

### Tahap 6 --- Mode tambahan dan pemolesan

-   Tambahkan mode gambar berantai.
-   Tambahkan pengaturan host lebih lengkap.
-   Tambahkan reconnect yang lebih baik.
-   Tambahkan statistik atau riwayat hanya jika dibutuhkan.
-   Periksa aksesibilitas, performa, dan keamanan.

**Kriteria selesai:** fitur tambahan diuji tanpa merusak mode dasar.

### Tahap 7 --- Persiapan publikasi

-   Tulis README dan panduan instalasi.
-   Siapkan konfigurasi environment tanpa membocorkan rahasia.
-   Tambahkan panduan deployment.
-   Uji room dengan beberapa pemain dan koneksi yang tidak stabil.
-   Dokumentasikan batasan yang masih ada.

**Kriteria selesai:** proyek dapat dijalankan orang lain berdasarkan
dokumentasi dan siap diuji pada lingkungan hosting yang dipilih.

------------------------------------------------------------------------

## 9. Strategi Pengujian Minimum

Uji skenario berikut sebelum menyatakan multiplayer siap:

1.  Dua atau lebih pemain berhasil masuk ke room yang sama.
2.  Pemain tidak dapat mengambil giliran atau aksi yang bukan haknya.
3.  Semua pemain melihat perubahan state yang konsisten.
4.  Jawaban benar hanya menghasilkan skor sesuai aturan.
5.  Timer berakhir secara konsisten di server.
6.  Pemain yang reconnect tidak merusak ronde.
7.  Host keluar ditangani dengan jelas.
8.  Room penuh menolak pemain tambahan.
9.  Input dan pesan berlebihan dibatasi.
10. Tampilan tetap dapat digunakan pada layar HP.
11. Game tetap dapat dimainkan ketika musik dan efek suara dimatikan.
12. Hasil pertandingan dapat ditampilkan tanpa kehilangan data ronde.

------------------------------------------------------------------------

## 10. Aturan Kerja untuk OpenCode

-   Jangan hanya membuat mockup. Implementasikan fitur yang berfungsi.
-   Jangan mengarang hasil tes; jalankan tes dan laporkan hasil
    sebenarnya.
-   Jangan menyelesaikan semua tahap sekaligus.
-   Jangan mengubah banyak file yang tidak berkaitan dengan tahap aktif.
-   Setelah perubahan, ringkas file yang dibuat/diubah dan alasannya.
-   Jika ada error, telusuri penyebabnya dan perbaiki sebelum lanjut.
-   Jika butuh dependency baru, jelaskan fungsinya.
-   Hindari pola kode terlalu rumit untuk kebutuhan saat ini.
-   Gunakan nama variabel dan fungsi yang jelas.
-   Tambahkan komentar untuk logika yang memang sulit dipahami, bukan
    untuk setiap baris.
-   Pertahankan perubahan yang sudah berfungsi.
-   Jika ada beberapa opsi teknis, pilih opsi paling sederhana yang
    memenuhi kebutuhan saat ini.
-   Pada akhir setiap tahap, berikan:
    1.  Fitur yang selesai.
    2.  File yang berubah.
    3.  Cara menjalankan dan menguji.
    4.  Hasil tes yang benar-benar dilakukan.
    5.  Masalah yang masih tersisa.
    6.  Tahap berikutnya yang disarankan.

------------------------------------------------------------------------

## 11. Hasil Akhir yang Diharapkan

Sebuah game party multiplayer online berbasis browser yang:

-   Dapat dimainkan bersama teman melalui kode room.
-   Memiliki kanvas gambar yang nyaman digunakan.
-   Menyinkronkan permainan secara real-time.
-   Mendukung reaksi dan quick chat.
-   Memiliki animasi serta audio yang dapat dikendalikan.
-   Memiliki hasil pertandingan yang menarik.
-   Responsif untuk desktop dan mobile.
-   Menggunakan identitas visual orisinal.
-   Memiliki kode yang terstruktur, dapat diuji, dan dapat dikembangkan.
-   Memiliki README dan petunjuk menjalankan proyek.

**Mulai sekarang:** audit folder proyek, jelaskan temuannya, buat
rencana implementasi, lalu kerjakan Tahap 1 setelah rencana jelas.
Jangan langsung mengimplementasikan seluruh dokumen sekaligus.
