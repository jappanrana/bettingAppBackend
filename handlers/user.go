package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"betting-app-backend-go/middleware"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserHandler struct {
	db *mongo.Database
}

func NewUserHandler(db *mongo.Database) *UserHandler {
	return &UserHandler{db: db}
}

// GetProfile handles GET /api/user/profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	log.Printf("[User] Getting profile for user: %s\n", userID)
	
	collection := h.db.Collection("users")
	
	var user bson.M
	err := collection.FindOne(context.Background(), bson.M{"uid": userID}).Decode(&user)
	if err != nil {
		log.Printf("[User] ❌ Failed to get profile: %v\n", err)
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	
	log.Println("[User] ✅ Profile retrieved")
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateProfile handles PUT /api/user/profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	var updates struct {
		Name       string `json:"name,omitempty"`
		ProfilePic string `json:"profilePic,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("[User] ❌ Invalid request body: %v\n", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	log.Printf("[User] Updating profile for user: %s\n", userID)
	
	collection := h.db.Collection("users")
	
	updateFields := bson.M{}
	if updates.Name != "" {
		updateFields["name"] = updates.Name
	}
	if updates.ProfilePic != "" {
		updateFields["profile_pic"] = updates.ProfilePic
	}
	
	if len(updateFields) == 0 {
		http.Error(w, "no fields to update", http.StatusBadRequest)
		return
	}
	
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"uid": userID},
		bson.M{"$set": updateFields},
	)
	
	if err != nil {
		log.Printf("[User] ❌ Failed to update profile: %v\n", err)
		http.Error(w, "failed to update profile", http.StatusInternalServerError)
		return
	}
	
	log.Println("[User] ✅ Profile updated")
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Profile updated successfully",
	})
}

// GetStats handles GET /api/user/stats
func (h *UserHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	log.Printf("[User] Getting stats for user: %s\n", userID)
	
	collection := h.db.Collection("users")
	
	var user bson.M
	err := collection.FindOne(context.Background(), bson.M{"uid": userID}).Decode(&user)
	if err != nil {
		log.Printf("[User] ❌ Failed to get stats: %v\n", err)
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	
	stats := map[string]interface{}{
		"totalGamesPlayed": user["total_games_played"],
		"totalWagered":     user["total_wagered"],
		"totalWon":         user["total_won"],
	}
	
	log.Println("[User] ✅ Stats retrieved")
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
