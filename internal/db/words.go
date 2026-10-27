package db

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// WordRecord merepresentasikan entitas kata di database.
type WordRecord struct {
	ID         int      `json:"id"`
	Word       string   `json:"word"`
	Category   string   `json:"category"`
	Difficulty string   `json:"difficulty"`
	Clues      []string `json:"clues"`
}

// SeedWordItem struktur bantuan untuk seeding data kata dan petunjuk.
type SeedWordItem struct {
	Word       string
	Category   string
	Difficulty string
	Clues      []string
}

// InitialWordsBank daftar 110 kata Indonesia dengan petunjuk relevan dalam 10 kategori.
var InitialWordsBank = []SeedWordItem{
	// 1. Hewan (11)
	{Word: "Kucing", Category: "Hewan", Difficulty: "easy", Clues: []string{"Suka mengeong dan menangkap tikus", "Memiliki kumis sensitif dan bulu lembut"}},
	{Word: "Anjing", Category: "Hewan", Difficulty: "easy", Clues: []string{"Sahabat setia manusia yang suka menggonggong", "Punya indra penciuman yang sangat tajam"}},
	{Word: "Gajah", Category: "Hewan", Difficulty: "easy", Clues: []string{"Mamalia darat terbesar dengan belalai panjang", "Memiliki sepasang gading dan telinga lebar"}},
	{Word: "Harimau", Category: "Hewan", Difficulty: "medium", Clues: []string{"Kucing besar bergaris loreng oranye hitam", "Dijuluki raja hutan di Asia"}},
	{Word: "Jerapah", Category: "Hewan", Difficulty: "medium", Clues: []string{"Hewan darat tertinggi dengan leher panjang", "Makan dedaunan dari pucuk pohon tinggi"}},
	{Word: "Kura-Kura", Category: "Hewan", Difficulty: "easy", Clues: []string{"Hewan lambat bertempurung keras di punggungnya", "Bisa menyembunyikan kepalanya ke dalam cangkang"}},
	{Word: "Kupu-Kupu", Category: "Hewan", Difficulty: "medium", Clues: []string{"Serangga bersayap indah hasil metamorfosis ulat", "Suka hinggap mengisap nektar bunga"}},
	{Word: "Ikan Hiu", Category: "Hewan", Difficulty: "medium", Clues: []string{"Predator laut bergigi tajam dengan sirip punggung menonjol", "Bisa mendeteksi setetes darah dari jarak jauh"}},
	{Word: "Burung Hantu", Category: "Hewan", Difficulty: "medium", Clues: []string{"Burung nokturnal yang matanya besar dan bisa memutar kepala", "Aktif berburu di malam hari tanpa suara"}},
	{Word: "Kelinci", Category: "Hewan", Difficulty: "easy", Clues: []string{"Hewan bertelinga panjang yang suka wortel", "Bergerak dengan cara melompat-lompat lincah"}},
	{Word: "Lumba-Lumba", Category: "Hewan", Difficulty: "medium", Clues: []string{"Mamalia laut cerdas yang suka melompat di atas air", "Berkomunikasi menggunakan suara ultrasonik"}},

	// 2. Makanan & Minuman (11)
	{Word: "Nasi Goreng", Category: "Makanan & Minuman", Difficulty: "easy", Clues: []string{"Masakan nasi yang digoreng kecap dengan telur dan ayam", "Menu sarapan dan makan malam khas nusantara"}},
	{Word: "Rendang", Category: "Makanan & Minuman", Difficulty: "medium", Clues: []string{"Masakan daging sapi kaya rempah khas Minangkabau", "Dimasak berjam-jam hingga kuah santan mengering dan gelap"}},
	{Word: "Sate Ayam", Category: "Makanan & Minuman", Difficulty: "easy", Clues: []string{"Daging ayam tusuk yang dibakar di atas arang", "Disajikan dengan lumuran bumbu kacang atau kecap"}},
	{Word: "Bakso", Category: "Makanan & Minuman", Difficulty: "easy", Clues: []string{"Bola daging sapi gurih disajikan dengan kuah kaldu panas", "Biasanya dilengkapi mi kuning, bihun, dan tahu"}},
	{Word: "Martabak", Category: "Makanan & Minuman", Difficulty: "medium", Clues: []string{"Kue tebal manis bertabur cokelat keju atau martabak telur gurih", "Camilan malam yang sangat populer di Indonesia"}},
	{Word: "Es Teh Manis", Category: "Makanan & Minuman", Difficulty: "easy", Clues: []string{"Minuman penyegar sejuta umat saat cuaca terik", "Dibuat dari seduhan teh celup, gula pasir, dan es batu"}},
	{Word: "Gado-Gado", Category: "Makanan & Minuman", Difficulty: "medium", Clues: []string{"Salad sayuran rebus khas Indonesia disiram saus kacang", "Sering disajikan dengan kerupuk, lontong, dan telur rebus"}},
	{Word: "Roti Bakar", Category: "Makanan & Minuman", Difficulty: "easy", Clues: []string{"Roti tawar panggang dengan olesan selai atau cokelat", "Sering disantap sebagai kudapan sore atau malam"}},
	{Word: "Soto Ayam", Category: "Makanan & Minuman", Difficulty: "medium", Clues: []string{"Sup ayam berkuah kuning kunyit harum dengan suwiran ayam", "Diberi taburan koya gurih dan perasan jeruk nipis"}},
	{Word: "Pisang Goreng", Category: "Makanan & Minuman", Difficulty: "easy", Clues: []string{"Camilan pisang berbalut tepung renyah yang digoreng", "Sangat nikmat disantap hangat bersama kopi atau teh"}},
	{Word: "Nasi Uduk", Category: "Makanan & Minuman", Difficulty: "medium", Clues: []string{"Nasi gurih dimasak dengan santan, daun pandan, dan serai", "Khas Betawi yang harum dengan lauk orek tempe dan bihun"}},

	// 3. Benda Sehari-hari (11)
	{Word: "Sepeda", Category: "Benda Sehari-hari", Difficulty: "easy", Clues: []string{"Alat transportasi roda dua yang dikayuh pedal", "Memiliki stang kemudi dan rantai penggerak"}},
	{Word: "Jam Dinding", Category: "Benda Sehari-hari", Difficulty: "easy", Clues: []string{"Penunjuk waktu berdetak yang dipasang di tembok", "Memiliki jarum pendek jam, jarum panjang menit, dan detik"}},
	{Word: "Kacamata", Category: "Benda Sehari-hari", Difficulty: "easy", Clues: []string{"Alat bantu penglihatan berbingkai dengan sepasang lensa", "Dikenakan di atas hidung dan bertumpu di telinga"}},
	{Word: "Payung", Category: "Benda Sehari-hari", Difficulty: "easy", Clues: []string{"Alat pelindung dari guyuran hujan atau terik matahari", "Bisa dibuka mengembang dan dilipat menjadi ramping"}},
	{Word: "Kursi", Category: "Benda Sehari-hari", Difficulty: "easy", Clues: []string{"Tempat duduk berkaki empat dengan sandaran punggung", "Sering dipasangkan bersama meja kerja atau makan"}},
	{Word: "Meja", Category: "Benda Sehari-hari", Difficulty: "easy", Clues: []string{"Perabot berkaki empat dengan permukaan bidang datar", "Digunakan untuk meletakkan barang, makan, atau menulis"}},
	{Word: "Sepatu", Category: "Benda Sehari-hari", Difficulty: "easy", Clues: []string{"Alas kaki pelindung bertali atau berperekat", "Dipakai saat bepergian, sekolah, atau berolahraga"}},
	{Word: "Termos", Category: "Benda Sehari-hari", Difficulty: "medium", Clues: []string{"Wadah tertutup untuk menjaga air tetap panas atau dingin", "Memiliki dinding isolasi berlapis ruang hampa"}},
	{Word: "Kipas Angin", Category: "Benda Sehari-hari", Difficulty: "easy", Clues: []string{"Alat berbaling-baling pembuat hembusan angin sejuk", "Berputar dengan motor listrik untuk mengusir rasa gerah"}},
	{Word: "Cermin", Category: "Benda Sehari-hari", Difficulty: "easy", Clues: []string{"Permukaan kaca mengilap pemantul bayangan wajah kita", "Sering digunakan saat berdandan atau merapikan pakaian"}},
	{Word: "Sapu", Category: "Benda Sehari-hari", Difficulty: "easy", Clues: []string{"Alat pembersih lantai dari ijuk atau lidi dengan gagang", "Digunakan untuk mengumpulkan debu dan sampah di lantai"}},

	// 4. Profesi (11)
	{Word: "Dokter", Category: "Profesi", Difficulty: "easy", Clues: []string{"Tenaga medis yang memeriksa penyakit dan meresepkan obat", "Biasa memakai jas putih dan stetoskop di leher"}},
	{Word: "Guru", Category: "Profesi", Difficulty: "easy", Clues: []string{"Pendidik yang mengajar ilmu pengetahuan di sekolah", "Berdiri di depan kelas membagikan wawasan kepada murid"}},
	{Word: "Polisi", Category: "Profesi", Difficulty: "easy", Clues: []string{"Aparat penegak hukum yang menjaga ketertiban masyarakat", "Mengenakan seragam dinas dan mengatur lalu lintas jalan"}},
	{Word: "Koki", Category: "Profesi", Difficulty: "easy", Clues: []string{"Juru masak profesional yang meracik masakan lezat di restoran", "Memakai topi koki putih tinggi dan celemek"}},
	{Word: "Pilot", Category: "Profesi", Difficulty: "easy", Clues: []string{"Pengemudi andal pesawat terbang di kokpit udara", "Menerbangkan penumpang melintasi langit antar negara"}},
	{Word: "Pemadam Kebakaran", Category: "Profesi", Difficulty: "medium", Clues: []string{"Pahlawan berseragam oranye peredam kobaran api bencana", "Mengendarai mobil merah bertangki air dan selang besar"}},
	{Word: "Nelayan", Category: "Profesi", Difficulty: "easy", Clues: []string{"Pencari ikan di laut luas menggunakan jaring dan perahu", "Berangkat petang melaut dan pulang membawa hasil tangkapan pagi hari"}},
	{Word: "Petani", Category: "Profesi", Difficulty: "easy", Clues: []string{"Penggarap sawah dan ladang yang menanam padi serta sayur", "Memakai caping bambu dan mencangkul tanah yang subur"}},
	{Word: "Arsitek", Category: "Profesi", Difficulty: "medium", Clues: []string{"Perancang desain bangunan dan gedung bertingkat modern", "Menggambar denah dan maket sebelum proyek didirikan"}},
	{Word: "Astronot", Category: "Profesi", Difficulty: "medium", Clues: []string{"Penjelajah luar angkasa dengan baju astronot tebal", "Melayang di stasiun luar angkasa dalam gravitasi nol"}},
	{Word: "Hakim", Category: "Profesi", Difficulty: "medium", Clues: []string{"Pemimpin sidang pengadilan pemutus perkara hukum yang adil", "Mengetukkan palu sidang setelah membacakan vonis"}},

	// 5. Tempat (11)
	{Word: "Sekolah", Category: "Tempat", Difficulty: "easy", Clues: []string{"Tempat para siswa belajar menuntut ilmu bersama guru", "Terdapat ruang kelas, papan tulis, dan lapangan upacara"}},
	{Word: "Rumah Sakit", Category: "Tempat", Difficulty: "easy", Clues: []string{"Pusat pelayanan kesehatan untuk pasien rawat inap dan darurat", "Terdapat dokter, perawat, ruang operasi, dan ambulans"}},
	{Word: "Bandara", Category: "Tempat", Difficulty: "medium", Clues: []string{"Tempat lepas landas dan mendaratnya pesawat terbang", "Memiliki landasan pacu panjang dan terminal penumpang"}},
	{Word: "Stasiun Kereta", Category: "Tempat", Difficulty: "medium", Clues: []string{"Tempat pemberhentian rangkaian kereta api antar kota", "Memiliki peron tempat penumpang menunggu dan rel panjang"}},
	{Word: "Pantai", Category: "Tempat", Difficulty: "easy", Clues: []string{"Kawasan pesisir berpasir di tepi deburan ombak laut", "Banyak pohon kelapa dan sering dikunjungi untuk bersantai"}},
	{Word: "Gunung", Category: "Tempat", Difficulty: "easy", Clues: []string{"Dataran tanah tinggi menjulang ke langit dengan puncak sejuk", "Destinasi favorit para pendaki untuk melihat pemandangan awan"}},
	{Word: "Pasar Tradisional", Category: "Tempat", Difficulty: "medium", Clues: []string{"Tempat transaksi tawar-menawar antara penjual dan pembeli", "Menjual sayuran segar, buah-buahan, bumbu, dan ikan"}},
	{Word: "Kebun Binatang", Category: "Tempat", Difficulty: "easy", Clues: []string{"Taman margasatwa tempat memelihara dan melihat berbagai satwa", "Destinasi wisata keluarga untuk mengenal flora dan fauna"}},
	{Word: "Perpustakaan", Category: "Tempat", Difficulty: "medium", Clues: []string{"Gedung hening berisi ribuan koleksi buku yang dipinjamkan", "Tempat favorit bagi yang suka membaca dan mencari referensi"}},
	{Word: "Pelabuhan", Category: "Tempat", Difficulty: "medium", Clues: []string{"Tempat berlabuhnya kapal-kapal laut pengangkut kontainer", "Terdapat dermaga dan kapal feri penyeberangan antarpulau"}},
	{Word: "Museum", Category: "Tempat", Difficulty: "medium", Clues: []string{"Tempat menyimpan dan memamerkan benda peninggalan sejarah purbakala", "Menyajikan karya seni dan artefak masa lampau"}},

	// 6. Transportasi (11)
	{Word: "Pesawat Terbang", Category: "Transportasi", Difficulty: "easy", Clues: []string{"Burung besi bermesin jet yang mengangkut penumpang di langit", "Melesat kencang melintasi awan dengan sayap lebar"}},
	{Word: "Kapal Laut", Category: "Transportasi", Difficulty: "easy", Clues: []string{"Kendaraan besar pengarung samudra di permukaan air laut", "Mengangkut ribuan kontainer barang dan penumpang antarpulau"}},
	{Word: "Kereta Api", Category: "Transportasi", Difficulty: "easy", Clues: []string{"Rangkaian gerbong panjang yang berjalan di atas rel besi", "Ditarik lokomotif kuat dan berbunyi klakson nyaring"}},
	{Word: "Sepeda Motor", Category: "Transportasi", Difficulty: "easy", Clues: []string{"Kendaraan bermotor roda dua yang lincah di jalanan", "Dikendarai menggunakan helm pengaman dan pegangan gas"}},
	{Word: "Bus Sekolah", Category: "Transportasi", Difficulty: "easy", Clues: []string{"Bus berwarna kuning pengantar jemput murid belajar", "Memiliki banyak kursi untuk mengangkut anak-anak sekolah"}},
	{Word: "Helikopter", Category: "Transportasi", Difficulty: "medium", Clues: []string{"Kendaraan udara dengan baling-baling besar berputar di atapnya", "Bisa terbang vertikal tanpa memerlukan landasan pacu"}},
	{Word: "Perahu Dayung", Category: "Transportasi", Difficulty: "easy", Clues: []string{"Perahu kayu kecil yang digerakkan menggunakan dayung kayu", "Sering digunakan di danau tenang atau sungai mengalir"}},
	{Word: "Kapal Selam", Category: "Transportasi", Difficulty: "hard", Clues: []string{"Kapal militer berteknologi canggih yang menyelam di bawah laut", "Menjelajah kedalaman samudra tanpa terlihat di permukaan"}},
	{Word: "Becak", Category: "Transportasi", Difficulty: "easy", Clues: []string{"Kendaraan roda tiga tradisional yang dikayuh pengemudi di belakang", "Menampung satu atau dua penumpang di bagian depan"}},
	{Word: "Delman", Category: "Transportasi", Difficulty: "easy", Clues: []string{"Kereta beroda dua atau empat yang ditarik oleh seekor kuda", "Suara lonceng bergemerincing seiring langkah kaki kuda"}},
	{Word: "Truk Gandeng", Category: "Transportasi", Difficulty: "medium", Clues: []string{"Kendaraan angkutan barang berbobot raksasa dengan bak panjang", "Membawa muatan logistik jarak jauh di jalan tol"}},

	// 7. Alam & Cuaca (11)
	{Word: "Matahari", Category: "Alam & Cuaca", Difficulty: "easy", Clues: []string{"Bintang pusat tata surya pemancar cahaya dan kehangatan bumi", "Terbit di ufuk timur dan terbenam di ufuk barat"}},
	{Word: "Pelangi", Category: "Alam & Cuaca", Difficulty: "easy", Clues: []string{"Lengkungan tujuh warna indah di langit setelah hujan reda", "Terjadi akibat pembiasan cahaya matahari oleh tetesan air"}},
	{Word: "Petir", Category: "Alam & Cuaca", Difficulty: "easy", Clues: []string{"Kilatan kilau listrik terang di awan mendung disusul guruh", "Menghantarkan energi listrik jutaan volt saat badai"}},
	{Word: "Hujan Salju", Category: "Alam & Cuaca", Difficulty: "medium", Clues: []string{"Tetesan air beku berbentuk butiran kristal putih lembut", "Turun di musim dingin dan membuat jalanan berselimut es"}},
	{Word: "Gunung Berapi", Category: "Alam & Cuaca", Difficulty: "medium", Clues: []string{"Gunung dengan kawah aktif yang bisa memuntahkan lava panas", "Menyemburkan abu vulkanik dan bebatuan pijar saat erupsi"}},
	{Word: "Air Terjun", Category: "Alam & Cuaca", Difficulty: "easy", Clues: []string{"Aliran air sungai yang jatuh curam dari tebing tinggi", "Menghasilkan suara gemuruh dan cipratan embun segar"}},
	{Word: "Hutan Tropis", Category: "Alam & Cuaca", Difficulty: "medium", Clues: []string{"Kawasan rimbun pepohonan lebat dengan keanekaragaman hayati", "Berfungsi sebagai paru-paru bumi penghasil oksigen"}},
	{Word: "Bintang Kejora", Category: "Alam & Cuaca", Difficulty: "medium", Clues: []string{"Titik cahaya berkilau di langit malam yang tampak paling terang", "Planet Venus yang tampak indah sebelum fajar atau petang"}},
	{Word: "Danau Vulkanik", Category: "Alam & Cuaca", Difficulty: "medium", Clues: []string{"Genangan air luas di kawah kaldera gunung letusan masa lalu", "Airnya jernih membiru dikelilingi perbukitan hijau"}},
	{Word: "Gurun Pasir", Category: "Alam & Cuaca", Difficulty: "medium", Clues: []string{"Hamparan tanah tandus berpasir luas dengan suhu ekstrem", "Hampir tidak pernah hujan dan ditumbuhi pohon kaktus"}},
	{Word: "Gempa Bumi", Category: "Alam & Cuaca", Difficulty: "hard", Clues: []string{"Guncangan hebat di permukaan tanah akibat pergeseran lempeng", "Bisa merobohkan bangunan dan memicu gelombang tsunami"}},

	// 8. Olahraga (11)
	{Word: "Sepak Bola", Category: "Olahraga", Difficulty: "easy", Clues: []string{"Olahraga beregu menendang bola ke dalam gawang lawan", "Dimainkan oleh sebelas pemain di lapangan rumput hijau"}},
	{Word: "Bulu Tangkis", Category: "Olahraga", Difficulty: "easy", Clues: []string{"Olahraga tepuk kok menggunakan raket melintasi net", "Sangat populer di Indonesia dengan segudang juara dunia"}},
	{Word: "Bola Basket", Category: "Olahraga", Difficulty: "easy", Clues: []string{"Olahraga memasukkan bola besar oranye ke dalam ring gantung", "Pemain saling mendribel dan melempar bola dengan tangan"}},
	{Word: "Renang", Category: "Olahraga", Difficulty: "easy", Clues: []string{"Olahraga meluncur di air kolam dengan berbagai gaya kayuhan", "Ada gaya dada, gaya bebas, gaya punggung, dan gaya kupu-kupu"}},
	{Word: "Catur", Category: "Olahraga", Difficulty: "medium", Clues: []string{"Permainan asah otak di atas papan kotak hitam putih", "Mengatur strategi raja, ratu, benteng, kuda, dan pion"}},
	{Word: "Bola Voli", Category: "Olahraga", Difficulty: "easy", Clues: []string{"Olahraga memukul bola melayang melewati net tanpa menyentuh lantai", "Tim saling smash dan blocking di area lapangan"}},
	{Word: "Lari Maraton", Category: "Olahraga", Difficulty: "medium", Clues: []string{"Lomba lari jarak jauh menempuh lintasan sejauh 42 kilometer", "Menguji ketahanan fisik dan napas para pelari tangguh"}},
	{Word: "Panahan", Category: "Olahraga", Difficulty: "medium", Clues: []string{"Olahraga membidik anak panah ke sasaran lingkaran target", "Memerlukan konsentrasi tinggi dan tarikan busur yang stabil"}},
	{Word: "Tenis Meja", Category: "Olahraga", Difficulty: "easy", Clues: []string{"Permainan pingpong bola kecil menggunakan bet karet di atas meja", "Bola memantul cepat melewati jaring net meja kecil"}},
	{Word: "Angkat Besi", Category: "Olahraga", Difficulty: "medium", Clues: []string{"Olahraga kekuatan mengangkat barbel berat ke atas kepala", "Menguji tenaga otot atlet dalam angkatan snatch dan clean & jerk"}},
	{Word: "Tinju", Category: "Olahraga", Difficulty: "medium", Clues: []string{"Pertarungan dua petarung bersarung tangan di atas ring bertali", "Saling jual beli pukulan jab, hook, dan uppercut"}},

	// 9. Hiburan & Seni (11)
	{Word: "Gitar", Category: "Hiburan & Seni", Difficulty: "easy", Clues: []string{"Alat musik berdawai enam yang dipetik atau digenjreng", "Punya kotak resonansi kayu dan leher berpapan fret"}},
	{Word: "Piano", Category: "Hiburan & Seni", Difficulty: "easy", Clues: []string{"Alat musik akustik dengan deretan tuts hitam dan putih", "Ditekan dengan sepuluh jari menghasilkan alunan melodi indah"}},
	{Word: "Lukisan", Category: "Hiburan & Seni", Difficulty: "easy", Clues: []string{"Karya seni visual pada kanvas yang diwarnai dengan kuas cat", "Mengabadikan keindahan alam atau ekspresi imajinasi seniman"}},
	{Word: "Kamera", Category: "Hiburan & Seni", Difficulty: "easy", Clues: []string{"Alat penangkap gambar dan rekaman momen foto atau video", "Memiliki lensa fokus, sensor digital, dan tombol rana"}},
	{Word: "Mikrofon", Category: "Hiburan & Seni", Difficulty: "easy", Clues: []string{"Alat penangkap getaran suara untuk diperkuat pengeras suara", "Digunakan oleh vokalis penyanyi atau pembawa acara di panggung"}},
	{Word: "Televisi", Category: "Hiburan & Seni", Difficulty: "easy", Clues: []string{"Perangkat elektronik layar kaca penayang siaran audio visual", "Menampilkan berita, film, kartun, dan acara hiburan di rumah"}},
	{Word: "Wayang Kulit", Category: "Hiburan & Seni", Difficulty: "medium", Clues: []string{"Seni pertunjukan boneka bayangan kulit kerbau khas Jawa", "Dimainkan oleh seorang dalang di balik layar kelir kain"}},
	{Word: "Biola", Category: "Hiburan & Seni", Difficulty: "medium", Clues: []string{"Alat musik dawai berdada ramping yang dimainkan dengan digesek", "Menghasilkan suara lengkingan merdu bernada tinggi"}},
	{Word: "Drum", Category: "Hiburan & Seni", Difficulty: "easy", Clues: []string{"Alat musik pukul terdiri dari simbal, snar, tom, dan bas drum", "Dipukul menggunakan sepasang stik kayu penentu ritme lagu"}},
	{Word: "Patung", Category: "Hiburan & Seni", Difficulty: "medium", Clues: []string{"Karya seni tiga dimensi yang dipahat dari batu, kayu, atau logam", "Membentuk figur manusia, pahlawan, atau hewan secara nyata"}},
	{Word: "Tari Topeng", Category: "Hiburan & Seni", Difficulty: "medium", Clues: []string{"Tarian tradisional di mana penari mengenakan topeng karakter", "Gerakan gemulai melambangkan sifat manusia yang beraneka rupa"}},

	// 10. Teknologi (11)
	{Word: "Komputer", Category: "Teknologi", Difficulty: "easy", Clues: []string{"Perangkat elektronik pengolah data digital dengan CPU dan monitor", "Digunakan untuk mengetik, membuat program, dan bermain game"}},
	{Word: "Ponsel Pintar", Category: "Teknologi", Difficulty: "easy", Clues: []string{"Telepon genggam layar sentuh pintar dengan koneksi internet", "Bisa untuk chatting, foto, bermain aplikasi, dan sosial media"}},
	{Word: "Robot", Category: "Teknologi", Difficulty: "easy", Clues: []string{"Mesin otomatis terprogram yang dapat membantu pekerjaan manusia", "Bisa berbentuk lengan mekanik pabrik atau humanoid cerdas"}},
	{Word: "Drone", Category: "Teknologi", Difficulty: "medium", Clues: []string{"Pesawat mini nirawak berkamera yang dikendalikan remote nirkabel", "Terbang tinggi mengambil foto pemandangan dari udara"}},
	{Word: "Roket", Category: "Teknologi", Difficulty: "medium", Clues: []string{"Wahana peluncur bermesin propulsi pendorong ke luar angkasa", "Meluncur tegak melepaskan diri dari gravitasi bumi"}},
	{Word: "Kacamata VR", Category: "Teknologi", Difficulty: "hard", Clues: []string{"Headset realitas virtual yang menampilkan dunia simulasi 3D", "Pengguna merasa seolah-olah masuk ke dalam dunia game"}},
	{Word: "Headphone", Category: "Teknologi", Difficulty: "easy", Clues: []string{"Pengeras suara sepasang bantalan yang dikenakan di telinga", "Menghasilkan suara musik jernih tanpa mengganggu sekitar"}},
	{Word: "Satelit", Category: "Teknologi", Difficulty: "hard", Clues: []string{"Benda buatan manusia yang mengorbit bumi untuk telekomunikasi", "Memancarkan sinyal GPS, cuaca, dan siaran televisi global"}},
	{Word: "Jam Pintar", Category: "Teknologi", Difficulty: "medium", Clues: []string{"Smartwatch di pergelangan tangan pengukur detak jantung dan langkah", "Tersambung ke ponsel untuk menampilkan notifikasi pesan"}},
	{Word: "Mobil Listrik", Category: "Teknologi", Difficulty: "medium", Clues: []string{"Kendaraan ramah lingkungan bertenaga baterai tanpa bahan bakar fosil", "Bisa diisi daya di colokan listrik dan tidak bersuara bising"}},
	{Word: "Printer 3D", Category: "Teknologi", Difficulty: "hard", Clues: []string{"Mesin pencetak yang mencairkan filamen menjadi objek tiga dimensi nyata", "Membentuk benda lapis demi lapis sesuai rancangan komputer"}},
}

// SeedWords memasukkan kata dan petunjuk awal ke tabel PostgreSQL jika belum ada.
func (w *DBWrapper) SeedWords(ctx context.Context) error {
	if w == nil || w.DB == nil {
		return nil
	}

	tx, err := w.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi seed kata: %w", err)
	}
	defer tx.Rollback()

	queryWord := `
	INSERT INTO words (word, category, difficulty, is_active)
	VALUES ($1, $2, $3, TRUE)
	ON CONFLICT (word) DO UPDATE SET
		category = EXCLUDED.category,
		difficulty = EXCLUDED.difficulty
	RETURNING id;
	`
	stmtWord, err := tx.PrepareContext(ctx, queryWord)
	if err != nil {
		return fmt.Errorf("gagal siapkan statement word: %w", err)
	}
	defer stmtWord.Close()

	queryClue := `
	INSERT INTO word_clues (word_id, clue_text, clue_order)
	VALUES ($1, $2, $3)
	ON CONFLICT (word_id, clue_order) DO UPDATE SET
		clue_text = EXCLUDED.clue_text;
	`
	stmtClue, err := tx.PrepareContext(ctx, queryClue)
	if err != nil {
		return fmt.Errorf("gagal siapkan statement clue: %w", err)
	}
	defer stmtClue.Close()

	for _, item := range InitialWordsBank {
		var wordID int
		err := stmtWord.QueryRowContext(ctx, item.Word, item.Category, item.Difficulty).Scan(&wordID)
		if err != nil {
			return fmt.Errorf("gagal seed word %s: %w", item.Word, err)
		}

		for idx, clue := range item.Clues {
			if _, err := stmtClue.ExecContext(ctx, wordID, clue, idx+1); err != nil {
				return fmt.Errorf("gagal seed clue for %s: %w", item.Word, err)
			}
		}
	}

	return tx.Commit()
}

var wordRng = rand.New(rand.NewSource(time.Now().UnixNano()))

// GetRandomWordWithClues mengambil kata acak beserta daftar petunjuknya dari database.
// Jika database offline atau terjadi error, mengambil dari InitialWordsBank.
func (w *DBWrapper) GetRandomWordWithClues(ctx context.Context, usedWords map[string]bool) (*WordRecord, error) {
	if w != nil && w.DB != nil {
		// Ambil satu kata acak yang belum dipakai
		var rows *sql.Rows
		var err error

		if len(usedWords) > 0 {
			var excluded []string
			var args []interface{}
			idx := 1
			for word := range usedWords {
				excluded = append(excluded, fmt.Sprintf("$%d", idx))
				args = append(args, strings.ToLower(word))
				idx++
			}
			query := fmt.Sprintf(`
			SELECT id, word, category, difficulty
			FROM words
			WHERE is_active = TRUE AND LOWER(word) NOT IN (%s)
			ORDER BY RANDOM()
			LIMIT 1;
			`, strings.Join(excluded, ", "))
			rows, err = w.DB.QueryContext(ctx, query, args...)
		}

		if rows == nil || err != nil || !rows.Next() {
			if rows != nil {
				rows.Close()
			}
			// Fallback: ambil sembarang kata aktif jika semua sudah dipakai
			query := `
			SELECT id, word, category, difficulty
			FROM words
			WHERE is_active = TRUE
			ORDER BY RANDOM()
			LIMIT 1;
			`
			rows, err = w.DB.QueryContext(ctx, query)
		}

		if err == nil && rows != nil {
			defer rows.Close()
			if rows.Next() {
				var rec WordRecord
				if err := rows.Scan(&rec.ID, &rec.Word, &rec.Category, &rec.Difficulty); err == nil {
					// Ambil petunjuk
					clueQuery := `
					SELECT clue_text
					FROM word_clues
					WHERE word_id = $1
					ORDER BY clue_order ASC;
					`
					cRows, cErr := w.DB.QueryContext(ctx, clueQuery, rec.ID)
					if cErr == nil {
						defer cRows.Close()
						for cRows.Next() {
							var clue string
							if err := cRows.Scan(&clue); err == nil {
								rec.Clues = append(rec.Clues, clue)
							}
						}
					}
					return &rec, nil
				}
			}
		}
	}

	// Fallback jika database tidak tersedia atau query kosong
	var candidates []SeedWordItem
	for _, it := range InitialWordsBank {
		if !usedWords[strings.ToLower(it.Word)] {
			candidates = append(candidates, it)
		}
	}
	if len(candidates) == 0 {
		candidates = InitialWordsBank
	}
	pick := candidates[wordRng.Intn(len(candidates))]

	return &WordRecord{
		ID:         0,
		Word:       pick.Word,
		Category:   pick.Category,
		Difficulty: pick.Difficulty,
		Clues:      pick.Clues,
	}, nil
}
