# FilmLog — Frontend Companion

Aplikasi frontend sederhana bertema **Letterboxd / IMDb** yang dirancang sebagai antarmuka pengujian bagi peserta penugasan Backend GDGoC.

Frontend ini dibuat **Zero-Build & Standalone** (menggunakan Tailwind CSS CDN dan Vanilla Modern JS). Peserta **tidak perlu** menginstall `npm` dependencies apa pun untuk menjalankannya.

---

## 🚀 Cara Menjalankan Frontend

Frontend ini sudah **di-embed ke backend Go**, jadi cukup jalankan backend-nya saja (satu service untuk UI + API):

```bash
cd backend
go run .
```

Buka browser di: `http://localhost:9994`

> File di folder ini (`backend/web/`) ikut ter-*compile* ke dalam binary lewat `//go:embed`.
> Kalau kamu mengubah `index.html` atau `app.js`, cukup restart `go run .`.

### Menjalankan frontend terpisah (opsional, untuk development)

Kalau mau pakai live-reload dari static server, kamu tetap bisa:

```bash
cd backend/web
npx serve .
```

Lalu buka Settings (`⚙️`) di pojok kanan atas dan arahkan **API Base URL** ke `http://localhost:9994/api`.

---

## ⚙️ Mengarahkan ke Backend Anda
Karena UI dan API disajikan dari origin yang sama, default-nya frontend memanggil `/api` (relatif), sehingga otomatis bekerja di local maupun setelah deploy.

Kalau backend kamu berjalan terpisah (misal `:3000`, `:5000`, atau URL Render lain):
1. Buka web frontend di browser.
2. Klik tombol indikator endpoint di pojok kanan atas (`⚙️ Settings`).
3. Ubah **API Base URL** ke port/URL backend Anda, lalu klik **Save & Reconnect**.

> 💡 **Penting (CORS):** Kalau frontend dan backend beda origin, pastikan backend mengaktifkan middleware **CORS** (`Access-Control-Allow-Origin: *` atau izinkan origin frontend Anda).

---

## 📋 Fitur yang Tersedia di Frontend

1. **🎬 Discover Movies (`GET /api/movies`)**:
   - Menampilkan poster, rating, sutradara, genre.
   - Filter pencarian judul dan filter pill genre.
   - Tombol **Shuffle / Randomize** untuk mengacak tampilan film.
2. **🔍 Detail Film (`GET /api/movies/:id`)**:
   - Modal informasi lengkap (sinopsis, durasi, sutradara, pemeran/cast).
3. **🔐 Autentikasi (`POST /api/auth/register` & `POST /api/auth/login`)**:
   - Login & Register modal.
   - Token JWT disimpan di `localStorage` dan otomatis dikirimkan di header `Authorization: Bearer <token>` pada setiap protected request.
4. **📌 Watchlist (`GET /api/watchlist`, `POST /api/watchlist`, `DELETE /api/watchlist/:id`)**:
   - Menambah dan menghapus film dari daftar tontonan pengguna yang sedang login.
5. **⭐ Watched & Reviews (`GET /api/watched`, `POST /api/watched`)**:
   - Mencatat film yang telah ditonton dengan rating bintang (1-5) dan catatan opini/review.
6. **🤖 AI Movie Recommendation (`POST /api/movies/recommend`)**:
   - Meminta rekomendasi film berdasarkan deskripsi/mood pengguna.
   - Backend harus memanfaatkan Gemini API untuk memilihkan film **yang tersedia di database lokal** beserta alasan logisnya.
7. **📡 API Inspector (Drawer di bagian bawah)**:
   - Mencatat setiap HTTP method, URL, status code (200, 401, 500, dll.), dan durasi request secara *real-time* untuk mempermudah debugging backend.
