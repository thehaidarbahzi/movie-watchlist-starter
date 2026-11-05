package main

import (
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func handleRegister(w http.ResponseWriter, r *http.Request) {
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

	username := req.Username
	if username == "" {
		username = req.Email
		if idx := strings.Index(req.Email, "@"); idx > 0 {
			username = req.Email[:idx]
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	_, err = DB.Exec(`INSERT INTO users (username, email, password) VALUES (?, ?, ?)`, username, req.Email, string(hashedPassword))
	if err != nil {
		http.Error(w, `{"error":"user already exists or db error"}`, http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, http.StatusCreated, map[string]interface{}{"message": "user registered"})
}
