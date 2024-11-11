package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const RegisterCollection = "register"

// Guild register
type Register struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	InGameName string             `bson:"ingame_name" json:"in_game_name"`
	DiscordID  string             `bson:"discord_id" json:"discord_id"`
	WeekDays   []string           `bson:"week_days" json:"week_days"`
	Hours      []string           `bson:"hours" json:"hours"`
	Weapons    []string           `bson:"weapons" json:"weapons"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	Approved   bool               `bson:"approved" json:"approved"`
}
