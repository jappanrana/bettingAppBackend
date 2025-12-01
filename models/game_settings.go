package models

import (
	"time"
)

// GameSettings stores configurable parameters for each game
type GameSettings struct {
	GameType    string                 `bson:"game_type" json:"game_type"`
	DisplayName string                 `bson:"display_name" json:"display_name"`
	Enabled     bool                   `bson:"enabled" json:"enabled"`
	MinBet      float64                `bson:"min_bet" json:"min_bet"`
	MaxBet      float64                `bson:"max_bet" json:"max_bet"`
	HouseEdge   float64                `bson:"house_edge" json:"house_edge"` // Percentage (e.g., 2.5 = 2.5%)
	Config      map[string]interface{} `bson:"config" json:"config"`         // Game-specific configuration
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
	UpdatedBy   string                 `bson:"updated_by" json:"updated_by"` // Admin UID
}

// Game-specific configuration structures

// SpinWheelConfig holds SpinWheel game configuration
type SpinWheelConfig struct {
	Multipliers []float64 `json:"multipliers"` // e.g., [1.2, 1.5, 2.0, 0, 1.8, 0.5, 3.0, 0]
}

// AviationConfig holds Aviation game configuration
type AviationConfig struct {
	MinMultiplier float64 `json:"min_multiplier"` // e.g., 1.01
	MaxMultiplier float64 `json:"max_multiplier"` // e.g., 50.0
	CrashChances  struct {
		Low    float64 `json:"low"`    // % chance for 1.01-3x (e.g., 50)
		Medium float64 `json:"medium"` // % chance for 3-5x (e.g., 30)
		High   float64 `json:"high"`   // % chance for 5-10x (e.g., 15)
		VeryHigh float64 `json:"very_high"` // % chance for 10-50x (e.g., 5)
	} `json:"crash_chances"`
}

// SlotConfig holds Slot game configuration
type SlotConfig struct {
	Symbols     []string           `json:"symbols"` // e.g., ["🍒", "🍋", "🍊", "🍇", "💎", "⭐", "7️⃣"]
	Multipliers map[string]float64 `json:"multipliers"` // Symbol -> Multiplier mapping
}

// MinesConfig holds Mines game configuration
type MinesConfig struct {
	GridSize      int     `json:"grid_size"`       // e.g., 5
	MinMines      int     `json:"min_mines"`       // e.g., 3
	MaxMines      int     `json:"max_mines"`       // e.g., 10
	DefaultMines  int     `json:"default_mines"`   // e.g., 5
	MultiplierBase float64 `json:"multiplier_base"` // Base for calculating multiplier
}

// PlinkoConfig holds Plinko game configuration
type PlinkoConfig struct {
	Rows int `json:"rows"` // Number of rows
	MultipliersLow []float64 `json:"multipliers_low"` // Low risk multipliers
	MultipliersMedium []float64 `json:"multipliers_medium"` // Medium risk
	MultipliersHigh []float64 `json:"multipliers_high"` // High risk
}

// DiceConfig holds Dice game configuration
type DiceConfig struct {
	MinTarget float64 `json:"min_target"` // e.g., 1.00
	MaxTarget float64 `json:"max_target"` // e.g., 99.99
	HouseEdge float64 `json:"house_edge"` // e.g., 1.0 (1%)
}

// LimboConfig holds Limbo game configuration
type LimboConfig struct {
	MinMultiplier float64 `json:"min_multiplier"` // e.g., 1.01
	MaxMultiplier float64 `json:"max_multiplier"` // e.g., 1000.0
	HouseEdge    float64 `json:"house_edge"`     // e.g., 1.0 (1%)
}

// HiLoConfig holds HiLo game configuration
type HiLoConfig struct {
	MultiplierPerWin float64 `json:"multiplier_per_win"` // e.g., 1.5 (each correct guess multiplies by this)
	MaxStreak        int     `json:"max_streak"`          // Maximum winning streak allowed
}

// BlackjackConfig holds Blackjack game configuration
type BlackjackConfig struct {
	NumDecks          int     `json:"num_decks"`           // e.g., 6
	BlackjackPayout   float64 `json:"blackjack_payout"`    // e.g., 2.5 (3:2 payout = bet * 2.5)
	WinPayout         float64 `json:"win_payout"`          // e.g., 2.0 (1:1 payout = bet * 2)
	DealerHitsSoft17  bool    `json:"dealer_hits_soft_17"` // Whether dealer hits on soft 17
	AllowDoubleDown   bool    `json:"allow_double_down"`
	AllowSplit        bool    `json:"allow_split"`
}
