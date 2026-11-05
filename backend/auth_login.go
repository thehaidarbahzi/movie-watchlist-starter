package main

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := readJSON(r, &req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password required"}`, http.StatusBadRequest)
		return
	}

	var user User
	row := DB.QueryRow(`SELECT id, username, email, password FROM users WHERE email = ?`, req.Email)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil || user.ID == 0 {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(user.ID, user.Username, user.Email)
	if err != nil {
		http.Error(w, `{"error":"token generation failed"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{
		Token: token,
		User: UserPartial{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	})
}
