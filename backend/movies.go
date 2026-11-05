package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func handleMovies(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	query := `SELECT id, title, year, director, genres, rating, duration, synopsis, poster_url, cast_members FROM movies WHERE 1=1`
	var params []interface{}

	searchParam := r.URL.Query().Get("search")
	genreParam := r.URL.Query().Get("genre")

	if searchParam != "" {
		query += ` AND (title LIKE ? OR director LIKE ?)`
		likeVal := "%" + searchParam + "%"
		params = append(params, likeVal, likeVal)
	}

	if genreParam != "" && genreParam != "ALL" {
		query += ` AND genres LIKE ?`
		params = append(params, "%"+genreParam+"%")
	}

	rows, err := DB.Query(query, params...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var m Movie
		var castStr string
		err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Director, &m.Genres, &m.Rating, &m.Duration, &m.Synopsis, &m.PosterURL, &castStr)
		if err != nil {
			continue
		}
		m.CastMembers = castStr
		m.Cast = strings.Split(castStr, ",")
		m.parseGenres()
		for i := range m.GenreList {
			m.GenreList[i] = strings.TrimSpace(m.GenreList[i])
		}
		movies = append(movies, m)
	}

	if r.URL.Query().Get("shuffle") == "true" || r.URL.Query().Get("random") == "true" && len(movies) > 0 {
		shuffleMovies(movies)
	}

	writeJSON(w, http.StatusOK, movies)
}

func shuffleMovies(movies []Movie) {
	n := len(movies)
	if n <= 1 {
		return
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := n - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		movies[i], movies[j] = movies[j], movies[i]
	}
}

func handleMovieDetail(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, `{"error":"missing movie id"}`, http.StatusBadRequest)
		return
	}

	var id int
	if _, err := fmt.Sscanf(parts[3], "%d", &id); err != nil {
		http.Error(w, `{"error":"invalid movie id"}`, http.StatusBadRequest)
		return
	}
	if id <= 0 {
		http.Error(w, `{"error":"invalid movie id"}`, http.StatusBadRequest)
		return
	}

	var m Movie
	var castStr string
	err := DB.QueryRow(`SELECT id, title, year, director, genres, rating, duration, synopsis, poster_url, cast_members FROM movies WHERE id = ?`, id).Scan(
		&m.ID, &m.Title, &m.Year, &m.Director, &m.Genres, &m.Rating, &m.Duration, &m.Synopsis, &m.PosterURL, &castStr)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"movie not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	m.CastMembers = castStr
	m.Cast = strings.Split(castStr, ",")
	m.parseGenres()
	for i := range m.GenreList {
		m.GenreList[i] = strings.TrimSpace(m.GenreList[i])
	}

	response := struct {
		Movie
		Genres []string `json:"genres"`
	}{
		Movie:  m,
		Genres: m.GenreList,
	}
	writeJSON(w, http.StatusOK, response)
}
