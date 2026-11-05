package main

import (
	"database/sql"
	"net/http"
	"strings"
)

func handleWatched(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID := getUserIDFromRequest(r)
	if userID == 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "GET":
		rows, err := DB.Query(`
			SELECT w.id, w.user_id, w.movie_id, w.rating, w.review, w.created_at, m.title, m.year, m.poster_url, m.genres, m.director, m.synopsis
			FROM watched_reviews w
			JOIN movies m ON w.movie_id = m.id
			WHERE w.user_id = ?
			ORDER BY w.created_at DESC`, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		defer rows.Close()

		type WatchedItem struct {
			ID         int      `json:"id"`
			MovieID    int      `json:"movieId"`
			MovieIDKey int      `json:"movie_id"`
			Rating     int      `json:"rating"`
			Review     string   `json:"review"`
			CreatedAt  string   `json:"created_at"`
			Title      string   `json:"title"`
			Year       int      `json:"year"`
			PosterURL  string   `json:"poster_url"`
			Genres     []string `json:"genres"`
			Director   string   `json:"director"`
			Synopsis   string   `json:"synopsis"`
		}

		var items []WatchedItem
		for rows.Next() {
			var it WatchedItem
			var genresStr string
			var createdAt sql.NullTime
			err := rows.Scan(&it.ID, &it.MovieID, &it.MovieIDKey, &it.Rating, &it.Review, &createdAt, &it.Title, &it.Year, &it.PosterURL, &genresStr, &it.Director, &it.Synopsis)
			if err != nil {
				continue
			}
			it.Genres = strings.Split(genresStr, ",")
			if createdAt.Valid {
				it.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
			} else {
				it.CreatedAt = ""
			}
			items = append(items, it)
		}
		writeJSON(w, http.StatusOK, items)

	case "POST":
		var req WatchedAddRequest
		if err := readJSON(r, &req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}

		movieID := req.MovieID
		rating := req.Rating
		if rating < 1 || rating > 5 {
			http.Error(w, `{"error":"rating must be between 1 and 5"}`, http.StatusBadRequest)
			return
		}
		review := req.Review
		if review == "" {
			review = req.Opinion
		}

		var exists int
		DB.QueryRow("SELECT COUNT(*) FROM movies WHERE id=?", movieID).Scan(&exists)
		if exists == 0 {
			http.Error(w, `{"error":"movie not found"}`, http.StatusNotFound)
			return
		}

		_, err := DB.Exec(`
			INSERT INTO watched_reviews (user_id, movie_id, rating, review) VALUES (?, ?, ?, ?)
			ON CONFLICT(user_id, movie_id) DO UPDATE SET rating=excluded.rating, review=excluded.review`,
			userID, movieID, rating, review)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]interface{}{"message": "reviewed", "movieId": movieID, "rating": rating})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
