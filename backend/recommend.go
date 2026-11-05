package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var defaultGeminiModels = []string{
	"gemini-flash-lite-latest",
	"gemini-3.5-flash-lite",
	"gemini-flash-latest",
	"gemini-3.8-flash",
	"gemini-3.6-flash",
	"gemini-3.5-flash",
}

func handleRecommend(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req RecommendRequest
	if err := readJSON(r, &req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(req.Query)
	}
	if prompt == "" {
		prompt = strings.TrimSpace(req.Description)
	}
	if prompt == "" {
		http.Error(w, `{"error":"prompt is required"}`, http.StatusBadRequest)
		return
	}

	movies, err := loadAllMovies()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if len(movies) == 0 {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "movie database is empty"})
		return
	}

	apiKey := getGeminiKey()
	if apiKey == "" {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"error":  "GEMINI_API_KEY belum dikonfigurasi",
			"detail": "Set environment variable GEMINI_API_KEY, atau buat file backend/.env berisi GEMINI_API_KEY=... lalu restart server.",
			"hint":   "Dapatkan API key gratis di https://aistudio.google.com/apikey",
		})
		return
	}

	geminiPrompt := buildGeminiPrompt(movies, prompt)

	result, model, err := callGemini(apiKey, geminiPrompt, movies)
	if err != nil {
		if os.Getenv("ALLOW_AI_FALLBACK") == "true" {
			best := pickBestByKeyword(prompt, movies)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"movie_id":  best.ID,
				"reasoning": fmt.Sprintf("Fallback (AI gagal: %v). Berdasarkan kata kunci, saya sarankan %s.", err, best.Title),
				"source":    "fallback",
			})
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]interface{}{
			"error":  "Gemini API request gagal",
			"detail": err.Error(),
			"hint":   "Cek GEMINI_API_KEY yang valid dan model yang tersedia (set GEMINI_MODEL bila perlu).",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"movie_id":  result.MovieID,
		"reasoning": result.Reasoning,
		"source":    "gemini",
		"model":     model,
	})
}

func loadAllMovies() ([]Movie, error) {
	rows, err := DB.Query("SELECT id, title, year, director, genres, rating, duration, synopsis FROM movies ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Year, &m.Director, &m.Genres, &m.Rating, &m.Duration, &m.Synopsis); err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, rows.Err()
}

func buildGeminiPrompt(movies []Movie, userPrompt string) string {
	var sb strings.Builder
	sb.WriteString("You are a movie recommendation assistant for a curated film database.\n")
	sb.WriteString("You MUST recommend ONLY from the exact list below. Never invent a movie that is not listed.\n\n")
	sb.WriteString("Available movies:\n")
	for _, m := range movies {
		sb.WriteString(fmt.Sprintf("- ID %d: %s (%d), directed by %s. Genres: %s. Synopsis: %s\n",
			m.ID, m.Title, m.Year, m.Director, m.Genres, m.Synopsis))
	}
	sb.WriteString("\nUser request: ")
	sb.WriteString(userPrompt)
	sb.WriteString("\n\nPick the SINGLE best matching movie from the list above. ")
	sb.WriteString("Return only JSON with keys: movie_id (integer, must be an ID from the list) and reasoning ")
	sb.WriteString("(string, 1-3 sentences explaining why it matches the user's mood or description).")
	return sb.String()
}

func callGemini(apiKey, prompt string, movies []Movie) (RecommendResponse, string, error) {
	models := geminiModelsToTry()

	var lastErr error
	for _, model := range models {
		result, err := callGeminiWithModel(apiKey, model, prompt, movies)
		if err == nil {
			return result, model, nil
		}
		lastErr = err

		if !isRetryableError(err) {
			return RecommendResponse{}, model, err
		}
	}
	return RecommendResponse{}, "", fmt.Errorf("semua model Gemini gagal dicoba: %w", lastErr)
}

func geminiModelsToTry() []string {
	if custom := strings.TrimSpace(os.Getenv("GEMINI_MODEL")); custom != "" {
		return []string{custom}
	}
	return defaultGeminiModels
}

type geminiError struct {
	Status int
	Body   string
}

func (e *geminiError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Status, e.Body)
}

func isRetryableError(err error) bool {
	ge, ok := err.(*geminiError)
	if !ok {
		return true
	}
	lower := strings.ToLower(ge.Body)
	switch ge.Status {
	case http.StatusNotFound,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return strings.Contains(lower, "unavailable") ||
		strings.Contains(lower, "high demand") ||
		strings.Contains(lower, "not found") ||
		strings.Contains(lower, "not supported") ||
		strings.Contains(lower, "overloaded")
}

func callGeminiWithModel(apiKey, model, prompt string, movies []Movie) (RecommendResponse, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", model)

	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":      0.6,
			"responseMimeType": "application/json",
			"responseSchema": map[string]interface{}{
				"type": "OBJECT",
				"properties": map[string]interface{}{
					"movie_id":  map[string]interface{}{"type": "INTEGER"},
					"reasoning": map[string]interface{}{"type": "STRING"},
				},
				"required": []string{"movie_id", "reasoning"},
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return RecommendResponse{}, err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return RecommendResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return RecommendResponse{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return RecommendResponse{}, &geminiError{Status: resp.StatusCode, Body: strings.TrimSpace(string(respBody))}
	}

	var geminiResp map[string]interface{}
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return RecommendResponse{}, fmt.Errorf("respons Gemini bukan JSON valid: %w", err)
	}

	text := extractGeminiText(geminiResp)
	if text == "" {
		return RecommendResponse{}, fmt.Errorf("Gemini tidak mengembalikan teks jawaban: %s", strings.TrimSpace(string(respBody)))
	}

	text = stripCodeFence(text)

	var result RecommendResponse
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return RecommendResponse{}, fmt.Errorf("gagal parse JSON dari Gemini (%q): %w", text, err)
	}

	valid := false
	var chosen Movie
	for _, m := range movies {
		if m.ID == result.MovieID {
			valid = true
			chosen = m
			break
		}
	}
	if !valid {
		return RecommendResponse{}, fmt.Errorf("Gemini memilih movie_id %d yang tidak ada di database", result.MovieID)
	}
	if strings.TrimSpace(result.Reasoning) == "" {
		result.Reasoning = fmt.Sprintf("%s cocok dengan permintaanmu berdasarkan genre dan sinopsisnya.", chosen.Title)
	}

	return result, nil
}

func extractGeminiText(resp map[string]interface{}) string {
	candidates, ok := resp["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return ""
	}
	first, ok := candidates[0].(map[string]interface{})
	if !ok {
		return ""
	}
	content, ok := first["content"].(map[string]interface{})
	if !ok {
		return ""
	}
	parts, ok := content["parts"].([]interface{})
	if !ok {
		return ""
	}
	var sb strings.Builder
	for _, p := range parts {
		part, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		if t, ok := part["text"].(string); ok {
			sb.WriteString(t)
		}
	}
	return sb.String()
}

func stripCodeFence(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
	}
	return strings.TrimSpace(text)
}

func getGeminiKey() string {
	return strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
}

func pickBestByKeyword(prompt string, movies []Movie) Movie {
	best := movies[0]
	bestScore := -1
	lower := strings.ToLower(prompt)
	for _, m := range movies {
		score := 0
		for _, g := range strings.Split(m.Genres, ",") {
			if strings.Contains(lower, strings.ToLower(strings.TrimSpace(g))) {
				score += 2
			}
		}
		if strings.Contains(lower, strings.ToLower(m.Title)) || strings.Contains(lower, strings.ToLower(m.Director)) {
			score += 3
		}
		if score > bestScore {
			bestScore = score
			best = m
		}
	}
	return best
}
