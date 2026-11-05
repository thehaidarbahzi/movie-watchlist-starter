package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

type Movie struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Year        int      `json:"year"`
	Director    string   `json:"director"`
	Genres      string   `json:"genres"`
	Rating      float64  `json:"rating"`
	Duration    string   `json:"duration"`
	Synopsis    string   `json:"synopsis"`
	PosterURL   string   `json:"poster_url"`
	CastMembers string   `json:"cast_members"`
	Cast        []string `json:"cast"`
	GenreList   []string `json:"-"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type WatchlistItem struct {
	ID      int `json:"id"`
	UserID  int `json:"user_id"`
	MovieID int `json:"movie_id"`
}

type WatchedReview struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	MovieID   int       `json:"movie_id"`
	Rating    int       `json:"rating"`
	Review    string    `json:"review"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserPartial `json:"user"`
}

type UserPartial struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type RecommendRequest struct {
	Prompt      string `json:"prompt"`
	Query       string `json:"query"`
	Description string `json:"description"`
}

type RecommendResponse struct {
	MovieID   int    `json:"movie_id"`
	Reasoning string `json:"reasoning"`
}

type WatchlistAddRequest struct {
	MovieID int `json:"movieId"`
}

type WatchedAddRequest struct {
	MovieID int    `json:"movieId"`
	Rating  int    `json:"rating"`
	Review  string `json:"review"`
	Opinion string `json:"opinion"`
}

func (m *Movie) parseGenres() {
	m.GenreList = strings.Split(m.Genres, ",")
}

func readJSON(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.Unmarshal(body, v)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
