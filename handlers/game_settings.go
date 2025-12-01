package handlers

import (
	"betting-app-backend-go/models"
	"betting-app-backend-go/services"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type GameSettingsHandler struct {
	service *services.GameSettingsService
}

func NewGameSettingsHandler(service *services.GameSettingsService) *GameSettingsHandler {
	return &GameSettingsHandler{service: service}
}

// GetGameSettings retrieves settings for a specific game
func (h *GameSettingsHandler) GetGameSettings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameType := vars["gameType"]

	settings, err := h.service.GetGameSettings(r.Context(), gameType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// GetAllGameSettings retrieves settings for all games
func (h *GameSettingsHandler) GetAllGameSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetAllGameSettings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateGameSettings updates settings for a specific game (Admin only)
func (h *GameSettingsHandler) UpdateGameSettings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameType := vars["gameType"]

	// Get admin UID from context (set by auth middleware)
	adminUID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var settings models.GameSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateGameSettings(r.Context(), gameType, &settings, adminUID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Game settings updated successfully"})
}
