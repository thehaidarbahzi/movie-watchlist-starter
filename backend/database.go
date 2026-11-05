package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func initDB() error {
	var err error
	DB, err = sql.Open("sqlite3", "./filmlog.db")
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	sqlStatements := []string{
		`CREATE TABLE IF NOT EXISTS movies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			year INTEGER NOT NULL,
			director TEXT NOT NULL,
			genres TEXT NOT NULL,
			rating REAL NOT NULL,
			duration TEXT NOT NULL,
			synopsis TEXT NOT NULL,
			poster_url TEXT NOT NULL,
			cast_members TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS watchlists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			movie_id INTEGER NOT NULL,
			UNIQUE(user_id, movie_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS watched_reviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			movie_id INTEGER NOT NULL,
			rating INTEGER NOT NULL CHECK(rating >= 1 AND rating <= 5),
			review TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, movie_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
		);`,
	}

	for _, stmt := range sqlStatements {
		_, err := DB.Exec(stmt)
		if err != nil {
			return fmt.Errorf("table creation error: %w", err)
		}
	}

	var count int
	DB.QueryRow("SELECT COUNT(*) FROM movies").Scan(&count)
	if count == 0 {
		seedMovies()
	}

	return nil
}

func seedMovies() {
	movies := []Movie{
		{Title: "Inception", Year: 2010, Director: "Christopher Nolan", Genres: "Sci-Fi,Action,Thriller", Rating: 8.8, Duration: "148 min", Synopsis: "Seorang pencuri ulung yang mengekstrak rahasia perusahaan melalui teknologi berbagi mimpi ditugaskan untuk melakukan 'inception', menanamkan ide ke dalam alam bawah sadar seorang CEO.", PosterURL: "https://image.tmdb.org/t/p/w500/oYuLEt3zVCKq57qu2F8dT7NIa6f.jpg", CastMembers: "Leonardo DiCaprio, Joseph Gordon-Levitt, Elliot Page, Tom Hardy"},
		{Title: "Parasite", Year: 2019, Director: "Bong Joon-ho", Genres: "Drama,Thriller,Comedy", Rating: 8.5, Duration: "132 min", Synopsis: "Keluarga Kim yang miskin secara perlahan menyusup dan bekerja untuk keluarga Park yang kaya raya, namun rahasia gelap di rumah mewah tersebut mengancam kehidupan mereka semua.", PosterURL: "https://image.tmdb.org/t/p/w500/7IiTTgloJzvGI1TAYymCfbfl3vT.jpg", CastMembers: "Song Kang-ho, Lee Sun-kyun, Cho Yeo-jeong, Choi Woo-shik"},
		{Title: "The Dark Knight", Year: 2008, Director: "Christopher Nolan", Genres: "Action,Crime,Drama", Rating: 9.0, Duration: "152 min", Synopsis: "Ketika ancaman yang dikenal sebagai Joker menebar teror dan kekacauan di Gotham, Batman harus menerima salah satu ujian psikologis dan fisik terbesar untuk melawan ketidakadilan.", PosterURL: "https://image.tmdb.org/t/p/w500/qJ2tW6WMUDux911r6m7haRef0WH.jpg", CastMembers: "Christian Bale, Heath Ledger, Aaron Eckhart, Michael Caine"},
		{Title: "Interstellar", Year: 2014, Director: "Christopher Nolan", Genres: "Sci-Fi,Drama,Adventure", Rating: 8.7, Duration: "169 min", Synopsis: "Ketika masa depan bumi terancam oleh krisis pangan dan iklim, sekelompok penjelajah melintasi lubang cacing di luar angkasa demi mencari planet baru yang layak huni bagi umat manusia.", PosterURL: "https://image.tmdb.org/t/p/w500/gEU2QniE6E77NI6lCU6MxlNBvIx.jpg", CastMembers: "Matthew McConaughey, Anne Hathaway, Jessica Chastain, Michael Caine"},
		{Title: "Pulp Fiction", Year: 1994, Director: "Quentin Tarantino", Genres: "Crime,Drama", Rating: 8.9, Duration: "154 min", Synopsis: "Kisah kehidupan dua pembunuh bayaran mafia, seorang petinju, istri seorang gangster, dan sepasang perampok restoran yang saling bertaut dalam alur non-linear yang penuh ironi dan aksi kriminal.", PosterURL: "https://image.tmdb.org/t/p/w500/d5iIlFn5s0ImszYzBPb8JPIfbXD.jpg", CastMembers: "John Travolta, Samuel L. Jackson, Uma Thurman, Bruce Willis"},
		{Title: "Spirited Away", Year: 2001, Director: "Hayao Miyazaki", Genres: "Animation,Adventure,Fantasy", Rating: 8.6, Duration: "125 min", Synopsis: "Chihiro, seorang gadis berusia 10 tahun, tanpa sengaja tersesat ke dunia roh. Setelah orang tuanya diubah menjadi babi, ia harus bekerja di pemandian air panas para dewa untuk membebaskan mereka.", PosterURL: "https://image.tmdb.org/t/p/w500/39wmItIWsg5sZMyRUHLkWBcuVCM.jpg", CastMembers: "Rumi Hiiragi, Miyu Irino, Mari Natsuki, Takashi Naito"},
		{Title: "Whiplash", Year: 2014, Director: "Damien Chazelle", Genres: "Drama,Music", Rating: 8.5, Duration: "106 min", Synopsis: "Seorang drummer jazz muda yang ambisius terdaftar di konservatorium musik paling bergengsi, di mana gurunya yang kejam dan perfeksionis mendorongnya hingga ke batas kewarasan.", PosterURL: "https://image.tmdb.org/t/p/w500/7fn624j5lj3xTme2SgiLCeuedmO.jpg", CastMembers: "Miles Teller, J.K. Simmons, Paul Reiser, Melissa Benoist"},
		{Title: "The Matrix", Year: 1999, Director: "Lana Wachowski, Lilly Wachowski", Genres: "Sci-Fi,Action", Rating: 8.7, Duration: "136 min", Synopsis: "Seorang peretas komputer bernama Neo menemukan bahwa realitas yang selama ini ia jalani hanyalah simulasi buatan mesin cerdas yang memperbudak manusia, dan ia bergabung dalam pemberontakan.", PosterURL: "https://image.tmdb.org/t/p/w500/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg", CastMembers: "Keanu Reeves, Laurence Fishburne, Carrie-Anne Moss, Hugo Weaving"},
		{Title: "Spider-Man: Into the Spider-Verse", Year: 2018, Director: "Bob Persichetti, Peter Ramsey, Rodney Rothman", Genres: "Animation,Action,Adventure", Rating: 8.4, Duration: "117 min", Synopsis: "Remaja Miles Morales menjadi Spider-Man baru di dimensinya dan harus bekerja sama dengan lima pahlawan bertopeng laba-laba lain dari dimensi berbeda untuk menghentikan ancaman multiverse.", PosterURL: "https://image.tmdb.org/t/p/w500/iiZZdoQBEYBv6id8su7ImL0oCbD.jpg", CastMembers: "Shameik Moore, Jake Johnson, Hailee Steinfeld, Mahershala Ali"},
		{Title: "The Godfather", Year: 1972, Director: "Francis Ford Coppola", Genres: "Crime,Drama", Rating: 9.2, Duration: "175 min", Synopsis: "Kisah kepemimpinan keluarga mafia Corleone di New York pasca Perang Dunia II, di mana sang putra bungsu yang awalnya enggan, Michael Corleone, perlahan terseret menjadi pemimpin baru dinasti kejahatan.", PosterURL: "https://image.tmdb.org/t/p/w500/3bhkrj58Vtu7enYsRolD1fZdja1.jpg", CastMembers: "Marlon Brando, Al Pacino, James Caan, Robert Duvall"},
		{Title: "Fight Club", Year: 1999, Director: "David Fincher", Genres: "Drama,Thriller", Rating: 8.8, Duration: "139 min", Synopsis: "Seorang pekerja kantoran yang menderita insomnia bertemu dengan penjual sabun misterius bernama Tyler Durden. Keduanya mendirikan klub tarung rahasia bawah tanah yang segera berkembang menjadi sesuatu yang berbahaya.", PosterURL: "https://image.tmdb.org/t/p/w500/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg", CastMembers: "Edward Norton, Brad Pitt, Helena Bonham Carter, Meat Loaf"},
		{Title: "Alien", Year: 1979, Director: "Ridley Scott", Genres: "Sci-Fi,Horror", Rating: 8.5, Duration: "117 min", Synopsis: "Kru kapal kargo luar angkasa komersial Nostromo menyelidiki panggilan darurat dari planet tak dikenal, dan tanpa sadar membawa masuk organisme parasit mematikan yang mulai memangsa mereka satu per satu.", PosterURL: "https://image.tmdb.org/t/p/w500/vfrQk5IPloGg1v9Rzbh2Eg3VGyM.jpg", CastMembers: "Sigourney Weaver, Tom Skerritt, John Hurt, Veronica Cartwright"},
		{Title: "WALL-E", Year: 2008, Director: "Andrew Stanton", Genres: "Animation,Sci-Fi,Family", Rating: 8.4, Duration: "98 min", Synopsis: "Di masa depan di mana bumi ditinggalkan karena tertutup sampah, robot pemadat sampah kecil bernama WALL-E menemukan robot pencari kehidupan bernama EVE dan memulai petualangan kosmik yang menentukan nasib manusia.", PosterURL: "https://image.tmdb.org/t/p/w500/hbhFnRzzg6ZDmm8YAmxBnQpQIPh.jpg", CastMembers: "Ben Burtt, Elissa Knight, Jeff Garlin, Fred Willard"},
		{Title: "Memento", Year: 2000, Director: "Christopher Nolan", Genres: "Mystery,Thriller", Rating: 8.4, Duration: "113 min", Synopsis: "Seorang pria yang kehilangan ingatan jangka pendek berusaha melacak pembunuh istrinya dengan bantuan foto polaroid dan tato di tubuhnya, diceritakan dalam dua garis waktu yang bergerak berlawanan.", PosterURL: "https://image.tmdb.org/t/p/w500/yuNs09hvpHVU1cBTCAk9zxsL2oW.jpg", CastMembers: "Guy Pearce, Carrie-Anne Moss, Joe Pantoliano, Mark Boone Junior"},
		{Title: "Joker", Year: 2019, Director: "Todd Phillips", Genres: "Crime,Drama,Thriller", Rating: 8.4, Duration: "122 min", Synopsis: "Di Gotham City yang penuh ketimpangan pada tahun 1981, Arthur Fleck, seorang badut pesta yang kesepian dan terabaikan oleh masyarakat, perlahan mengalami degradasi mental hingga menjadi sosok ikonik yang memicu revolusi brutal.", PosterURL: "https://image.tmdb.org/t/p/w500/udDclJoHjfjb8Ekgsd4FDteOkCU.jpg", CastMembers: "Joaquin Phoenix, Robert De Niro, Zazie Beetz, Frances Conroy"},
		{Title: "Blade Runner 2049", Year: 2017, Director: "Denis Villeneuve", Genres: "Sci-Fi,Drama,Mystery", Rating: 8.0, Duration: "164 min", Synopsis: "Tiga puluh tahun setelah peristiwa film pertama, seorang blade runner baru dari LAPD, Officer K, menemukan rahasia lama yang terkubur yang berpotensi memicu kekacauan besar dalam masyarakat distopia.", PosterURL: "https://image.tmdb.org/t/p/w500/gajva2L0rPYkEWjzgFlBXCAVBE5.jpg", CastMembers: "Ryan Gosling, Harrison Ford, Ana de Armas, Sylvia Hoeks"},
		{Title: "The Shawshank Redemption", Year: 1994, Director: "Frank Darabont", Genres: "Drama", Rating: 9.3, Duration: "142 min", Synopsis: "Seorang bankir sukses dijatuhi hukuman seumur hidup di Penjara Shawshank atas pembunuhan istri dan kekasihnya yang tidak pernah ia lakukan. Selama puluhan tahun, ia menemukan keteguhan dan harapan melalui persahabatan.", PosterURL: "https://image.tmdb.org/t/p/w500/9cqNxx0GxF0bflZmeSMuL5tnGzr.jpg", CastMembers: "Tim Robbins, Morgan Freeman, Bob Gunton, William Sadler"},
		{Title: "Dune", Year: 2021, Director: "Denis Villeneuve", Genres: "Sci-Fi,Adventure", Rating: 8.0, Duration: "155 min", Synopsis: "Paul Atreides, pemuda berbakat yang ditakdirkan untuk takdir yang melampaui pemahamannya, harus melakukan perjalanan ke planet paling berbahaya di alam semesta demi memastikan masa depan keluarga dan bangsanya.", PosterURL: "https://image.tmdb.org/t/p/w500/d5NXSklXo0qyIYkgV94XAgMIckC.jpg", CastMembers: "Timothée Chalamet, Rebecca Ferguson, Oscar Isaac, Zendaya"},
		{Title: "Oppenheimer", Year: 2023, Director: "Christopher Nolan", Genres: "Drama,History,Biography", Rating: 8.9, Duration: "180 min", Synopsis: "Kisah fisikawan teoretis J. Robert Oppenheimer yang memimpin Proyek Manhattan di Los Alamos untuk mengembangkan bom atom pertama di dunia selama Perang Dunia II, dan dilema moral yang menghantuinya setelah itu.", PosterURL: "https://image.tmdb.org/t/p/w500/8Gxv8gSFCU0XGDykEGv7zR1n2ua.jpg", CastMembers: "Cillian Murphy, Emily Blunt, Matt Damon, Robert Downey Jr."},
		{Title: "Goodfellas", Year: 1990, Director: "Martin Scorsese", Genres: "Crime,Drama", Rating: 8.7, Duration: "145 min", Synopsis: "Kisah nyata Henry Hill yang sejak remaja bercita-cita menjadi gangster, naik pangkat melalui hierarki kejahatan terorganisir di New York, hingga akhirnya terjerumus dalam lingkaran kecurigaan dan bahaya narkotika.", PosterURL: "https://image.tmdb.org/t/p/w500/aKuFiU82s5ISJpGZp7YkIr3kCUd.jpg", CastMembers: "Robert De Niro, Ray Liotta, Joe Pesci, Lorraine Bracco"},
	}

	stmt, err := DB.Prepare(`INSERT OR IGNORE INTO movies (id, title, year, director, genres, rating, duration, synopsis, poster_url, cast_members) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return
	}
	defer stmt.Close()

	for i, m := range movies {
		stmt.Exec(i+1, m.Title, m.Year, m.Director, m.Genres, m.Rating, m.Duration, m.Synopsis, m.PosterURL, m.CastMembers)
	}
}
