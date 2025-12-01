package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"betting-app-backend-go/middleware"
	"betting-app-backend-go/models"
	"betting-app-backend-go/services"
)

type GameHandler struct {
	service *services.GameService
}

func NewGameHandler(service *services.GameService) *GameHandler {
	return &GameHandler{service: service}
}

// PlayGame handles POST /api/game/play
func (h *GameHandler) PlayGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	var game models.Game
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		log.Printf("[Game] ❌ Invalid request body: %v\n", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	game.UserID = userID
	
	log.Printf("[Game] User %s playing %s - bet: %.2f\n", userID, game.GameType, game.BetAmount)
	
	// Validate game result
	if err := h.service.ValidateGameResult(&game); err != nil {
		log.Printf("[Game] ❌ Invalid game result: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Record game (handles wallet transactions)
	err := h.service.RecordGame(context.Background(), &game)
	if err != nil {
		log.Printf("[Game] ❌ Failed to record game: %v\n", err)
		if err.Error() == "insufficient balance" {
			http.Error(w, "insufficient balance", http.StatusPaymentRequired)
		} else {
			http.Error(w, "failed to record game", http.StatusInternalServerError)
		}
		return
	}
	
	log.Printf("[Game] ✅ Game recorded: %s - win: %.2f\n", game.ID, game.WinAmount)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(game)
}

// GetGameHistory handles GET /api/game/history
func (h *GameHandler) GetGameHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	gameType := r.URL.Query().Get("game_type")
	limitStr := r.URL.Query().Get("limit")
	limit := int64(50) // default
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			limit = l
		}
	}
	
	log.Printf("[Game] Getting history for user: %s (type: %s, limit: %d)\n", userID, gameType, limit)
	
	games, err := h.service.GetGameHistory(context.Background(), userID, gameType, limit)
	if err != nil {
		log.Printf("[Game] ❌ Failed to get game history: %v\n", err)
		http.Error(w, "failed to get game history", http.StatusInternalServerError)
		return
	}
	
	log.Printf("[Game] ✅ Retrieved %d games\n", len(games))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

// GetGameStats handles GET /api/game/stats
func (h *GameHandler) GetGameStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	gameType := r.URL.Query().Get("game_type")
	
	log.Printf("[Game] Getting stats for user: %s (type: %s)\n", userID, gameType)
	
	stats, err := h.service.GetGameStats(context.Background(), userID, gameType)
	if err != nil {
		log.Printf("[Game] ❌ Failed to get game stats: %v\n", err)
		http.Error(w, "failed to get game stats", http.StatusInternalServerError)
		return
	}
	
	log.Printf("[Game] ✅ Stats retrieved\n")
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetRecentBets handles GET /api/game/recent-bets
func (h *GameHandler) GetRecentBets(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := int64(20) // default
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			limit = l
		}
	}
	
	log.Printf("[Game] Getting recent bets (limit: %d)\n", limit)
	
	games, err := h.service.GetRecentBets(context.Background(), limit)
	if err != nil {
		log.Printf("[Game] ❌ Failed to get recent bets: %v\n", err)
		http.Error(w, "failed to get recent bets", http.StatusInternalServerError)
		return
	}
	
	log.Printf("[Game] ✅ Retrieved %d recent bets\n", len(games))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}
