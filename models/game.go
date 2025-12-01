package models

import (
	"time"
)

type Game struct {
	ID         string                 `bson:"_id" json:"id"`
	UserID     string                 `bson:"user_id" json:"userId"`
	GameType   string                 `bson:"game_type" json:"gameType"` // aviation, spinwheel, slot, mines, plinko, dice, limbo, hilo, blackjack
	BetAmount  float64                `bson:"bet_amount" json:"betAmount"`
	WinAmount  float64                `bson:"win_amount" json:"winAmount"`
	Multiplier float64                `bson:"multiplier,omitempty" json:"multiplier,omitempty"`
	ResultData map[string]interface{} `bson:"result_data,omitempty" json:"resultData,omitempty"` // Game-specific data
	Settled    bool                   `bson:"settled" json:"settled"`
	CreatedAt  time.Time              `bson:"created_at" json:"createdAt"`
}

type GameStats struct {
	UserID          string    `bson:"user_id" json:"userId"`
	GameType        string    `bson:"game_type" json:"gameType"`
	TotalGames      int64     `bson:"total_games" json:"totalGames"`
	TotalWagered    float64   `bson:"total_wagered" json:"totalWagered"`
	TotalWon        float64   `bson:"total_won" json:"totalWon"`
	BiggestWin      float64   `bson:"biggest_win" json:"biggestWin"`
	LastPlayedAt    time.Time `bson:"last_played_at" json:"lastPlayedAt"`
}
