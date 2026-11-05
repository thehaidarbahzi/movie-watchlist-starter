# Deploy FilmLog (satu service)

Frontend (`backend/web/`) sudah di-*embed* ke binary Go lewat `//go:embed`.
Artinya **cukup satu deployment** untuk UI + REST API, satu URL, tanpa CORS.

---

## 1. Jalankan lokal

```bash
cd backend
go run .
```

Buka `http://localhost:9994`. UI dan `/api/*` dilayani oleh server yang sama.

Butuh `gcc` karena `mattn/go-sqlite3` memakai CGO.

---

## 2. Environment variables

| Variable | Wajib | Keterangan |
|---|---|---|
| `GEMINI_API_KEY` | ya (untuk fitur AI) | Isi lewat dashboard hosting, jangan commit `.env` |
| `PORT` | tidak | Diisi otomatis oleh sebagian besar PaaS; default `9994` |
| `DB_PATH` | tidak | Default `./filmlog.db`. Di Docker: `/app/data/filmlog.db` |
| `GEMINI_MODEL` | tidak | Override model Gemini |

---

## 3. Deploy dengan Docker

Dockerfile ada di root repo.

```bash
docker build -t filmlog .
docker run -p 9994:9994 \
  -e GEMINI_API_KEY=xxx \
  -v filmlog-data:/app/data \
  filmlog
```

Buka `http://localhost:9994`.

> Volume `/app/data` penting supaya database SQLite tidak hilang saat container
> di-redeploy. Tanpa volume, akun/watchlist/review akan reset.

---

## 4. Deploy ke PaaS

### Render

1. **New → Web Service** → hubungkan repo GitHub.
2. Runtime: **Docker**. Dockerfile Path: `Dockerfile` (root).
3. Tambahkan env `GEMINI_API_KEY`.
4. (Opsional) Tambahkan **Persistent Disk** dengan mount path `/app/data`
   agar SQLite awet. Di free tier tidak tersedia — data akan reset saat restart.
5. Deploy, lalu buka URL `https://<nama-service>.onrender.com`.

### Railway

1. **New Project → Deploy from GitHub Repo**.
2. Railway mendeteksi `Dockerfile` otomatis.
3. Set env `GEMINI_API_KEY`.
4. Tambahkan **Volume** dengan mount path `/app/data`.
5. Generate domain dari tab **Settings → Networking**.

### Fly.io

```bash
fly launch --no-deploy          # pilih Dockerfile yang ada
fly volumes create filmlog_data --size 1
fly secrets set GEMINI_API_KEY=xxx
```

Tambahkan di `fly.toml`:

```toml
[http_service]
  internal_port = 9994
  force_https = true
  auto_stop_machines = "stop"
  auto_start_machines = true
  min_machines_running = 0

[mounts]
  source = "filmlog_data"
  destination = "/app/data"
```

> Fly **tidak** mengisi `PORT` otomatis. Jadi `internal_port` harus sama dengan
> port container (`9994`, sesuai `ENV PORT` di Dockerfile).

Lalu:

```bash
fly deploy
```

---

## 4b. Soal port di hosting

Kode backend membaca `PORT` lebih dulu, baru fallback ke `9994`:

```go
port := os.Getenv("PORT")
if port == "" {
    port = "9994"
}
```

Artinya:

- **Render / Railway** → mereka menyuntik `PORT` sendiri (misal `10000` / `8080`).
  Jangan set `PORT` manual; biarkan platform yang menentukan. Dockerfile `ENV PORT=9994`
  otomatis kalah oleh env dari platform.
- **Fly.io** → tidak menyuntik `PORT`; samakan `internal_port` di `fly.toml` dengan `9994`.
- **VPS + Nginx** → Go tetap listen di `127.0.0.1:9994`, lalu Nginx reverse proxy:
  ```nginx
  location / {
      proxy_pass http://127.0.0.1:9994;
      proxy_set_header Host $host;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
  }
  ```
  Buka port 80/443 di firewall, **bukan** 9994.
- **Docker lokal** → `docker run -p 9994:9994 ...`

> URL publik di hosting tetap tanpa `:9994` (misal `https://filmlog.onrender.com`).
> `:9994` hanya port internal container.

---

## 5. Catatan SQLite

SQLite cocok untuk demo/tugas, tapi **filesystem container bersifat sementara**.
Kalau butuh data permanen tanpa volume, migrasi ke Postgres (Supabase/Neon/Railway)
lebih aman.

Kalau pindah ke Postgres, ganti driver `github.com/mattn/go-sqlite3` dan sesuaikan
query `INSERT OR IGNORE` / placeholder `?` (Postgres memakai `$1`, `$2`, ...).

---

## 6. Checklist sebelum deploy

- [ ] `GEMINI_API_KEY` sudah diset di dashboard hosting
- [ ] `.env` tidak ikut ter-commit (sudah ada di `.gitignore`)
- [ ] Volume `/app/data` terpasang (kalau butuh data permanen)
- [ ] Buka `/` → UI muncul, lalu cek API Inspector di bawah
- [ ] Coba register/login dan tambah watchlist
