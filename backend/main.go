package main

import (
	"net/http"
	"os"
	"strings"
)

func main() {
	loadDotEnv(".env")
	loadDotEnv("../.env")

	if err := initDB(); err != nil {
		panic(err)
	}

	if getGeminiKey() != "" {
		println("Gemini API key: detected")
	} else {
		println("Gemini API key: NOT SET — AI recommendations will return 503 until GEMINI_API_KEY is configured")
	}

	http.HandleFunc("/api/auth/register", handleRegister)
	http.HandleFunc("/api/auth/login", handleLogin)

	http.HandleFunc("/api/movies/recommend", handleRecommend)
	http.HandleFunc("/api/movies/", handleMovieDetail)
	http.HandleFunc("/api/movies", handleMovies)

	http.HandleFunc("/api/watchlist", authMiddleware(handleWatchlist))
	http.HandleFunc("/api/watchlist/", authMiddleware(handleWatchlistDelete))

	http.HandleFunc("/api/watched", authMiddleware(handleWatched))

	server := &http.Server{
		Addr:    ":8000",
		Handler: http.DefaultServeMux,
	}

	println("Server starting on :8000")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}
