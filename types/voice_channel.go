package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const VoiceChannelCollection = "voice_channel"

type VoiceChannel struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	ChannelID string             `bson:"channel_id" json:"channel_id"`
	OwnerID   string             `bson:"owner_id" json:"owner_id"`
	Title     string             `bson:"title" json:"title"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
}
