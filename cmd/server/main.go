package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/block52/poker-equity/pkg/equity"
)

// EquityRequest is the JSON request body for equity calculations
type EquityRequest struct {
	Hands       [][]string `json:"hands"`
	Board       []string   `json:"board"`
	Dead        []string   `json:"dead"`
	Simulations int        `json:"simulations"`
}

// EquityResponse is the JSON response body
type EquityResponse struct {
	Results     []equity.EquityResult `json:"results"`
	Simulations int                   `json:"simulations"`
	Stage       string                `json:"stage"`
	DurationMs  float64               `json:"duration_ms"`
	HandsPerSec float64               `json:"hands_per_sec"`
	BoardCards  []string              `json:"board_cards"`
}

// ErrorResponse is the JSON error response body
type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /api/v1/equity", handleEquity)
	mux.HandleFunc("POST /api/v1/evaluate", handleEvaluate)

	handler := withCORS(withLogging(mux))

	log.Printf("Poker equity server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"service": "poker-equity",
	})
}

func handleEquity(w http.ResponseWriter, r *http.Request) {
	var req EquityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if len(req.Hands) < 2 {
		writeError(w, http.StatusBadRequest, "must provide at least 2 hands")
		return
	}

	simulations := req.Simulations
	if simulations <= 0 {
		simulations = 10000
	}
	if simulations > 100000 {
		simulations = 100000
	}

	calc := equity.NewCalculator(
		equity.WithSimulations(simulations),
		equity.WithWorkers(4),
	)

	result, err := calc.CalculateEquity(req.Hands, req.Board, req.Dead)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := EquityResponse{
		Results:     result.Results,
		Simulations: result.Simulations,
		Stage:       result.Stage.String(),
		DurationMs:  float64(result.Duration.Microseconds()) / 1000.0,
		HandsPerSec: result.HandsPerSec,
		BoardCards:  result.BoardCards,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// EvaluateRequest is the JSON request body for hand evaluation
type EvaluateRequest struct {
	Cards []string `json:"cards"`
}

// EvaluateResponse is the JSON response body for hand evaluation
type EvaluateResponse struct {
	Rank     string `json:"rank"`
	Category int    `json:"category"`
	Score    uint32 `json:"score"`
}

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	var req EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if len(req.Cards) < 5 || len(req.Cards) > 7 {
		writeError(w, http.StatusBadRequest, "must provide 5-7 cards")
		return
	}

	result, err := equity.EvaluateHandFastFromMnemonics(req.Cards)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	rankNames := []string{
		"High Card", "One Pair", "Two Pair", "Three of a Kind",
		"Straight", "Flush", "Full House", "Four of a Kind", "Straight Flush",
	}

	rankName := "Unknown"
	if int(result.Category) < len(rankNames) {
		rankName = rankNames[result.Category]
	}

	resp := EvaluateResponse{
		Rank:     rankName,
		Category: int(result.Category),
		Score:    result.Score,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

func init() {
	// Verify build
	_ = fmt.Sprintf("poker-equity server")
}
