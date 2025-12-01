package models

import (
	"time"
)

type User struct {
	UID              string    `bson:"uid" json:"uid"`
	Email            string    `bson:"email,omitempty" json:"email,omitempty"`
	Name             string    `bson:"name" json:"name"`
	Phone            string    `bson:"phone,omitempty" json:"phone,omitempty"`
	Role             string    `bson:"role" json:"role"` // user, admin
	ProfilePic       string    `bson:"profile_pic,omitempty" json:"profilePic,omitempty"`
	TotalGamesPlayed int64     `bson:"total_games_played" json:"totalGamesPlayed"`
	TotalWagered     float64   `bson:"total_wagered" json:"totalWagered"`
	TotalWon         float64   `bson:"total_won" json:"totalWon"`
	CreatedAt        time.Time `bson:"created_at" json:"createdAt"`
	LastSeenAt       time.Time `bson:"last_seen_at" json:"lastSeenAt"`
}
