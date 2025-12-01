package services

import (
	"betting-app-backend-go/models"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GameSettingsService struct {
	collection *mongo.Collection
}

func NewGameSettingsService(db *mongo.Database) *GameSettingsService {
	return &GameSettingsService{
		collection: db.Collection("game_settings"),
	}
}

// GetGameSettings retrieves settings for a specific game
func (s *GameSettingsService) GetGameSettings(ctx context.Context, gameType string) (*models.GameSettings, error) {
	var settings models.GameSettings
	err := s.collection.FindOne(ctx, bson.M{"game_type": gameType}).Decode(&settings)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("game settings not found")
		}
		return nil, err
	}
	return &settings, nil
}

// GetAllGameSettings retrieves settings for all games
func (s *GameSettingsService) GetAllGameSettings(ctx context.Context) ([]models.GameSettings, error) {
	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var settings []models.GameSettings
	if err := cursor.All(ctx, &settings); err != nil {
		return nil, err
	}
	return settings, nil
}

// UpdateGameSettings updates settings for a specific game
func (s *GameSettingsService) UpdateGameSettings(ctx context.Context, gameType string, settings *models.GameSettings, adminUID string) error {
	settings.GameType = gameType
	settings.UpdatedAt = time.Now()
	settings.UpdatedBy = adminUID

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"game_type": gameType}
	update := bson.M{"$set": settings}

	_, err := s.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// InitializeDefaultSettings creates default settings for all games if they don't exist
func (s *GameSettingsService) InitializeDefaultSettings(ctx context.Context) error {
	defaults := []models.GameSettings{
		{
			GameType:    "spinwheel",
			DisplayName: "Spin Wheel",
			Enabled:     true,
			MinBet:      10,
			MaxBet:      10000,
			HouseEdge:   2.5,
			Config: map[string]interface{}{
				"multipliers": []float64{1.2, 1.5, 2.0, 0, 1.8, 0.5, 3.0, 0},
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "system",
		},
		{
			GameType:    "aviation",
			DisplayName: "Aviation",
			Enabled:     true,
			MinBet:      10,
			MaxBet:      10000,
			HouseEdge:   2.0,
			Config: map[string]interface{}{
				"min_multiplier": 1.01,
				"max_multiplier": 50.0,
				"crash_chances": map[string]float64{
					"low":       50.0, // 1.01-3x
					"medium":    30.0, // 3-5x
					"high":      15.0, // 5-10x
					"very_high": 5.0,  // 10-50x
				},
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "system",
		},
		{
			GameType:    "slot",
			DisplayName: "Slot Machine",
			Enabled:     true,
			MinBet:      10,
			MaxBet:      5000,
			HouseEdge:   3.0,
			Config: map[string]interface{}{
				"symbols": []string{"🍒", "🍋", "🍊", "🍇", "💎", "⭐", "7️⃣"},
				"multipliers": map[string]float64{
					"7️⃣": 100,
					"💎": 50,
					"⭐": 20,
					"🍇": 10,
					"🍊": 5,
					"🍋": 3,
					"🍒": 2,
				},
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "system",
		},
		{
			GameType:    "mines",
			DisplayName: "Mines",
			Enabled:     true,
			MinBet:      10,
			MaxBet:      10000,
			HouseEdge:   2.0,
			Config: map[string]interface{}{
				"grid_size":       5,
				"min_mines":       3,
				"max_mines":       10,
				"default_mines":   5,
				"multiplier_base": 0.5,
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "system",
		},
		{
			GameType:    "plinko",
			DisplayName: "Plinko",
			Enabled:     true,
			MinBet:      10,
			MaxBet:      10000,
			HouseEdge:   2.5,
			Config: map[string]interface{}{
				"rows":                12,
				"multipliers_low":     []float64{1.5, 1.3, 1.1, 1.0, 0.5, 1.0, 1.1, 1.3, 1.5},
				"multipliers_medium":  []float64{3.0, 1.6, 1.4, 1.1, 1.0, 0.5, 1.0, 1.1, 1.4, 1.6, 3.0},
				"multipliers_high":    []float64{10.0, 3.0, 1.6, 1.4, 1.1, 1.0, 0.2, 1.0, 1.1, 1.4, 1.6, 3.0, 10.0},
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "system",
		},
		{
			GameType:    "dice",
			DisplayName: "Dice",
			Enabled:     true,
			MinBet:      10,
			MaxBet:      10000,
			HouseEdge:   1.0,
			Config: map[string]interface{}{
				"min_target": 1.00,
				"max_target": 99.99,
				"house_edge": 1.0,
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "system",
		},
		{
			GameType:    "limbo",
			DisplayName: "Limbo",
			Enabled:     true,
			MinBet:      10,
			MaxBet:      10000,
			HouseEdge:   1.0,
			Config: map[string]interface{}{
				"min_multiplier": 1.01,
				"max_multiplier": 1000.0,
				"house_edge":     1.0,
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "system",
		},
		{
			GameType:    "hilo",
			DisplayName: "Hi-Lo",
			Enabled:     true,
			MinBet:      10,
			MaxBet:      10000,
			HouseEdge:   2.0,
			Config: map[string]interface{}{
				"multiplier_per_win": 1.5,
				"max_streak":         10,
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "system",
		},
		{
			GameType:    "blackjack",
			DisplayName: "Blackjack",
			Enabled:     true,
			MinBet:      10,
			MaxBet:      10000,
			HouseEdge:   0.5,
			Config: map[string]interface{}{
				"num_decks":           6,
				"blackjack_payout":    2.5,
				"win_payout":          2.0,
				"dealer_hits_soft_17": true,
				"allow_double_down":   true,
				"allow_split":         true,
			},
			UpdatedAt: time.Now(),
			UpdatedBy: "system",
		},
	}

	for _, defaultSettings := range defaults {
		// Only insert if doesn't exist
		filter := bson.M{"game_type": defaultSettings.GameType}
		count, err := s.collection.CountDocuments(ctx, filter)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err := s.collection.InsertOne(ctx, defaultSettings)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
