# FilmLog — GDGoC Backend Recruitment Task Starter Pack

Selamat datang di repositori penugasan backend **FilmLog (Movie Watchlist & Review Platform)**. Repositori ini menyediakan:

1. **Initial Seed Database** (`data/movies.json` dan `data/seed.sql`) berisi 20 film kurasi awal dengan atribut lengkap.
2. **Frontend Companion** (`backend/web/`) bertema Letterboxd yang siap pakai untuk menguji endpoint backend Anda secara visual dan interaktif. Frontend ini di-_embed_ ke binary Go, jadi backend dan UI berjalan sebagai satu service.

---

## 🎯 Objektif Penugasan

Rancang dan bangun layanan **RESTful API** untuk aplikasi pelacak film (_movie watchlist_) yang mendukung alur autentikasi, kurasi film, pencatatan tontonan dan ulasan, serta integrasi **Google Gemini LLM** untuk rekomendasi film berbasis data lokal.

Bebas menggunakan bahasa pemrograman dan framework apa pun (Node.js/Express/NestJS, Go/Gin/Fiber, Python/FastAPI/Django, PHP/Laravel, Java/Spring Boot, dll.) serta database relasional (PostgreSQL, MySQL, SQLite) atau NoSQL.

---

## 📁 Struktur Repositori Starter

```text
movie-watchlist-starter/
├── data/
│   ├── movies.json     # Data 20 film awal (format JSON) untuk database seeder
│   └── seed.sql        # Skrip SQL INSERT kompatibel dengan PostgreSQL & MySQL
├── backend/
│   ├── main.go         # Entrypoint: serve REST API + frontend
│   └── web/            # Frontend Letterboxd-style SPA (di-embed ke binary)
│       ├── index.html
│       ├── app.js
│       └── README.md
├── Dockerfile          # Untuk deploy satu container (API + UI)
└── README.md           # Panduan penugasan & spesifikasi API
```
