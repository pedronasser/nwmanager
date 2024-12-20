package types

import (
	"context"
	"fmt"
	"nwmanager/helpers"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const PlayerCollection = "players"

type Player struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	DiscordID     string             `json:"discord_id" bson:"discord_id"`
	IGN           string             `json:"ign" bson:"ign"`
	WarClass      string             `json:"war_class" bson:"war_class"`
	TicketChannel string             `json:"ticket_channel" bson:"ticket_channel"`
	RegisteredAt  *time.Time         `json:"registered_at" bson:"registered_at"`
	ArchivedAt    *time.Time         `json:"archived_at" bson:"archived_at"`
}

func GetPlayerByDiscordID(ctx context.Context, db Database, discordID string) (*Player, error) {
	q := db.Collection(PlayerCollection).FindOne(ctx, bson.M{"discord_id": discordID})
	if q.Err() != nil {
		if q.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("Cannot find player: %v", q.Err())
	}

	var player Player
	err := q.Decode(&player)
	if err != nil {
		return nil, fmt.Errorf("Cannot decode player: %v", err)
	}

	return &player, nil
}

func InsertPlayer(ctx context.Context, db Database, player *Player) error {
	_, err := db.Collection(PlayerCollection).InsertOne(ctx, player)
	if err != nil {
		return fmt.Errorf("Cannot insert player: %v", err)
	}

	return nil
}

func DeletePlayer(ctx context.Context, db Database, player *Player) error {
	_, err := db.Collection(PlayerCollection).DeleteOne(ctx, bson.M{"_id": player.ID})
	if err != nil {
		return fmt.Errorf("Cannot delete player: %v", err)
	}

	return nil
}

func ArchivePlayer(ctx context.Context, db Database, player *Player) error {
	_, err := db.Collection(PlayerCollection).UpdateOne(ctx, bson.M{"_id": player.ID}, bson.M{"$set": bson.M{"archived_at": helpers.GetCurrentTimeAsUTC()}})
	if err != nil {
		return fmt.Errorf("Cannot delete player: %v", err)
	}

	return nil
}

func UpdatePlayer(ctx context.Context, db Database, player *Player) error {
	_, err := db.Collection(PlayerCollection).UpdateOne(ctx, bson.M{"_id": player.ID}, bson.M{"$set": player})
	if err != nil {
		return fmt.Errorf("Cannot update player: %v", err)
	}

	return nil
}

func GetPlayers(ctx context.Context, db Database) ([]Player, error) {
	cursor, err := db.Collection(PlayerCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("Cannot get players: %v", err)
	}
	defer cursor.Close(ctx)

	var players []Player
	err = cursor.All(ctx, &players)
	if err != nil {
		return nil, fmt.Errorf("Cannot decode players: %v", err)
	}

	return players, nil
}

func GetArchivedPlayers(ctx context.Context, db Database) ([]Player, error) {
	cursor, err := db.Collection(PlayerCollection).Find(ctx, bson.M{"archived_at": bson.M{"$ne": nil}})
	if err != nil {
		return nil, fmt.Errorf("Cannot get archived players: %v", err)
	}
	defer cursor.Close(ctx)

	var players []Player
	err = cursor.All(ctx, &players)
	if err != nil {
		return nil, fmt.Errorf("Cannot decode archived players: %v", err)
	}

	return players, nil
}
