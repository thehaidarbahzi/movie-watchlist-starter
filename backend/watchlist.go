package main

import (
	"fmt"
	"net/http"
	"strings"
)

func handleWatchlist(w http.ResponseWriter, r *http.Request) {
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
			SELECT w.id, w.movie_id, w.user_id, m.title, m.year, m.poster_url, m.rating, m.director, m.genres
			FROM watchlists w
			JOIN movies m ON w.movie_id = m.id
			WHERE w.user_id = ?`, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		defer rows.Close()

		var results []map[string]interface{}
		for rows.Next() {
			var wid, mid, uid int
			var title string
			var year int
			var poster string
			var rating float64
			var director, genres string
			rows.Scan(&wid, &mid, &uid, &title, &year, &poster, &rating, &director, &genres)
			results = append(results, map[string]interface{}{
				"movie_id":   mid,
				"movieId":    mid,
				"id":         mid,
				"title":      title,
				"year":       year,
				"poster_url": poster,
				"rating":     rating,
				"director":   director,
				"genres":     strings.Split(genres, ","),
				"user_id":    uid,
			})
		}
		writeJSON(w, http.StatusOK, results)

	case "POST":
		var req WatchlistAddRequest
		if err := readJSON(r, &req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		movieID := req.MovieID
		if movieID == 0 {
			var alt struct {
				MovieId int `json:"movie_id"`
			}
			readJSON(r, &alt)
			if alt.MovieId > 0 {
				movieID = alt.MovieId
			}
		}
		if movieID <= 0 {
			http.Error(w, `{"error":"movieId required"}`, http.StatusBadRequest)
			return
		}

		_, err := DB.Exec(`INSERT OR IGNORE INTO watchlists (user_id, movie_id) VALUES (?, ?)`, userID, movieID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]interface{}{"message": "added", "movieId": movieID})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func handleWatchlistDelete(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "DELETE" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromRequest(r)
	if userID == 0 {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	movieID := 0
	if len(parts) >= 4 {
		fmt.Sscanf(parts[3], "%d", &movieID)
	}
	if movieID <= 0 {
		http.Error(w, `{"error":"invalid movie id"}`, http.StatusBadRequest)
		return
	}

	_, err := DB.Exec(`DELETE FROM watchlists WHERE user_id = ? AND movie_id = ?`, userID, movieID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	writeJSON(w, http.StatusOK, map[string]interface{}{"message": "removed"})
}
