package services

import (
	"context"
	"fmt"
	"time"

	"betting-app-backend-go/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GameService struct {
	db            *mongo.Database
	walletService *WalletService
}

func NewGameService(db *mongo.Database, walletService *WalletService) *GameService {
	return &GameService{
		db:            db,
		walletService: walletService,
	}
}

// RecordGame records a game play with bet deduction and win credit
func (s *GameService) RecordGame(ctx context.Context, game *models.Game) error {
	if game.BetAmount <= 0 {
		return fmt.Errorf("bet amount must be positive")
	}
	
	// Validate game type
	validGameTypes := map[string]bool{
		"aviation": true, "spinwheel": true, "slot": true, "mines": true,
		"plinko": true, "dice": true, "limbo": true, "hilo": true, "blackjack": true,
	}
	
	if !validGameTypes[game.GameType] {
		return fmt.Errorf("invalid game type: %s", game.GameType)
	}
	
	gamesCol := s.db.Collection("games")
	usersCol := s.db.Collection("users")
	
	// Start transaction
	session, err := s.db.Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)
	
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Deduct bet amount from wallet
		err := s.walletService.DeductBalance(
			sessCtx,
			game.UserID,
			game.BetAmount,
			fmt.Sprintf("%s game bet", game.GameType),
			"game_loss",
		)
		if err != nil {
			return nil, fmt.Errorf("failed to deduct bet: %w", err)
		}
		
		// If there's a win, credit the wallet
		if game.WinAmount > 0 {
			err = s.walletService.CreditBalance(
				sessCtx,
				game.UserID,
				game.WinAmount,
				fmt.Sprintf("%s game win", game.GameType),
				"game_win",
			)
			if err != nil {
				return nil, fmt.Errorf("failed to credit winnings: %w", err)
			}
		}
		
		// Record game in database
		game.ID = fmt.Sprintf("game_%d", time.Now().UnixNano())
		game.Settled = true
		game.CreatedAt = time.Now()
		
		_, err = gamesCol.InsertOne(sessCtx, game)
		if err != nil {
			return nil, fmt.Errorf("failed to insert game: %w", err)
		}
		
		// Update user statistics
		netProfit := game.WinAmount - game.BetAmount
		_, err = usersCol.UpdateOne(
			sessCtx,
			bson.M{"uid": game.UserID},
			bson.M{
				"$inc": bson.M{
					"total_games_played": 1,
					"total_wagered":      game.BetAmount,
					"total_won":          netProfit,
				},
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update user stats: %w", err)
		}
		
		return nil, nil
	})
	
	return err
}

// GetGameHistory retrieves game history for a user
func (s *GameService) GetGameHistory(ctx context.Context, userID string, gameType string, limit int64) ([]models.Game, error) {
	collection := s.db.Collection("games")
	
	filter := bson.M{"user_id": userID}
	if gameType != "" && gameType != "all" {
		filter["game_type"] = gameType
	}
	
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit)
	
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get game history: %w", err)
	}
	defer cursor.Close(ctx)
	
	var games []models.Game
	if err = cursor.All(ctx, &games); err != nil {
		return nil, fmt.Errorf("failed to decode games: %w", err)
	}
	
	return games, nil
}

// GetGameStats retrieves aggregated statistics for a user's games
func (s *GameService) GetGameStats(ctx context.Context, userID string, gameType string) (map[string]interface{}, error) {
	collection := s.db.Collection("games")
	
	filter := bson.M{"user_id": userID}
	if gameType != "" && gameType != "all" {
		filter["game_type"] = gameType
	}
	
	// Aggregation pipeline
	pipeline := []bson.M{
		{"$match": filter},
		{
			"$group": bson.M{
				"_id": "$game_type",
				"total_games": bson.M{"$sum": 1},
				"total_wagered": bson.M{"$sum": "$bet_amount"},
				"total_won": bson.M{"$sum": "$win_amount"},
				"biggest_win": bson.M{"$max": "$win_amount"},
				"last_played": bson.M{"$max": "$created_at"},
			},
		},
	}
	
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate game stats: %w", err)
	}
	defer cursor.Close(ctx)
	
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}
	
	// Calculate overall stats
	stats := map[string]interface{}{
		"total_games":   int64(0),
		"total_wagered": 0.0,
		"total_won":     0.0,
		"biggest_win":   0.0,
		"by_game_type":  results,
	}
	
	for _, result := range results {
		if totalGames, ok := result["total_games"].(int64); ok {
			stats["total_games"] = stats["total_games"].(int64) + totalGames
		}
		if totalWagered, ok := result["total_wagered"].(float64); ok {
			stats["total_wagered"] = stats["total_wagered"].(float64) + totalWagered
		}
		if totalWon, ok := result["total_won"].(float64); ok {
			stats["total_won"] = stats["total_won"].(float64) + totalWon
		}
		if biggestWin, ok := result["biggest_win"].(float64); ok {
			if biggestWin > stats["biggest_win"].(float64) {
				stats["biggest_win"] = biggestWin
			}
		}
	}
	
	return stats, nil
}

// ValidateGameResult performs basic validation on game results
func (s *GameService) ValidateGameResult(game *models.Game) error {
	// Basic validation rules
	if game.BetAmount <= 0 {
		return fmt.Errorf("bet amount must be positive")
	}
	
	if game.WinAmount < 0 {
		return fmt.Errorf("win amount cannot be negative")
	}
	
	// Game-specific validation
	switch game.GameType {
	case "spinwheel":
		// Max multiplier is typically 100x
		if game.Multiplier > 100 {
			return fmt.Errorf("invalid multiplier for spinwheel")
		}
		
	case "slot":
		// Max payout is typically 1000x
		if game.WinAmount > game.BetAmount*1000 {
			return fmt.Errorf("win amount exceeds maximum payout")
		}
		
	case "dice":
		// Win amount should match odds
		if game.WinAmount > 0 && game.Multiplier > 0 {
			expectedWin := game.BetAmount * game.Multiplier
			if game.WinAmount != expectedWin {
				return fmt.Errorf("win amount doesn't match multiplier")
			}
		}
		
	case "aviation":
		// Multiplier should be >= 1.0
		if game.WinAmount > 0 && game.Multiplier < 1.0 {
			return fmt.Errorf("invalid aviation multiplier")
		}
	}
	
	return nil
}

// GetRecentBets retrieves recent bets across all users (for dashboard display)
func (s *GameService) GetRecentBets(ctx context.Context, limit int64) ([]models.Game, error) {
	collection := s.db.Collection("games")
	
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit)
	
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent bets: %w", err)
	}
	defer cursor.Close(ctx)
	
	var games []models.Game
	if err = cursor.All(ctx, &games); err != nil {
		return nil, fmt.Errorf("failed to decode games: %w", err)
	}
	
	return games, nil
}
