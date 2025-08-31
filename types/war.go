package types

import (
	"context"
	"fmt"
	"nwmanager/database"
	"nwmanager/discordbot/globals"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const WarsCollection = "wars"

// War participation types
type WarParticipation string

const (
	WarParticipationYes   WarParticipation = "yes"   // Sim
	WarParticipationNo    WarParticipation = "no"    // Não
	WarParticipationMaybe WarParticipation = "maybe" // Talvez
	WarParticipationBench WarParticipation = "bench" // Banco
)

// War types
type WarType string

const (
	WarTypeAttack  WarType = "attack"  // Ataque
	WarTypeDefense WarType = "defense" // Defesa
)

// War status
type WarStatus string

const (
	WarStatusActive   WarStatus = "active"   // Ativa
	WarStatusArchived WarStatus = "archived" // Arquivada
)

// War
type War struct {
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	FortName      string             `bson:"fort_name" json:"fort_name"`
	Type          WarType            `bson:"war_type" json:"war_type"`
	OpponentGuild string             `bson:"opponent_guild" json:"opponent_guild"`
	Description   string             `bson:"description" json:"description"`
	Territory     string             `bson:"territory" json:"territory"`
	CreatedAt     *time.Time         `bson:"created_at,omitempty" json:"created_at"`
	CreatedBy     string             `bson:"created_by" json:"created_by"`
	ScheduledAt   *time.Time         `bson:"scheduled_at,omitempty" json:"scheduled_at"`
	CompletedAt   *time.Time         `bson:"completed_at,omitempty" json:"completed_at"`
	ArchivedAt    *time.Time         `bson:"archived_at,omitempty" json:"archived_at"`
	NotifiedAt    *time.Time         `bson:"notified_at,omitempty" json:"notified_at"`
	ConfirmedAt   *time.Time         `bson:"confirmed_at,omitempty" json:"confirmed_at"`

	// Participação dos jogadores (PlayerID -> Resposta)
	Participations map[string]WarParticipation `bson:"participations" json:"participations"`

	// IDs das mensagens para cleanup posterior
	ChannelMessageID string            `bson:"channel_message_id" json:"channel_message_id"`
	PlayerMessages   map[string]string `bson:"player_messages" json:"player_messages"` // PlayerID -> MessageID

	Status    WarStatus `bson:"status" json:"status"`
	MessageID string    `bson:"message_id" json:"message_id"` // Legacy field
}

// Database functions for War management
func InsertWar(ctx context.Context, db database.Database, war *War) error {
	_, err := db.Collection(globals.DB_PREFIX+WarsCollection).InsertOne(ctx, war)
	if err != nil {
		return fmt.Errorf("Cannot insert war: %v", err)
	}
	return nil
}

func GetWarByID(ctx context.Context, db database.Database, id string) (*War, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("Invalid war ID: %v", err)
	}

	var war War
	err = db.Collection(globals.DB_PREFIX+WarsCollection).FindOne(ctx, bson.M{"_id": oid}).Decode(&war)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("Cannot find war: %v", err)
	}

	return &war, nil
}

func UpdateWar(ctx context.Context, db database.Database, war *War) error {
	_, err := db.Collection(globals.DB_PREFIX+WarsCollection).UpdateOne(ctx, bson.M{"_id": war.ID}, bson.M{"$set": war})
	if err != nil {
		return fmt.Errorf("Cannot update war: %v", err)
	}
	return nil
}

// UpdateWarParticipation atomically updates a single player's participation in a war
func UpdateWarParticipation(ctx context.Context, db database.Database, warID primitive.ObjectID, playerID string, participation WarParticipation) error {
	update := bson.M{
		"$set": bson.M{
			"participations." + playerID: participation,
		},
	}
	
	_, err := db.Collection(globals.DB_PREFIX+WarsCollection).UpdateOne(ctx, bson.M{"_id": warID}, update)
	if err != nil {
		return fmt.Errorf("cannot update war participation: %v", err)
	}
	return nil
}

func GetActiveWars(ctx context.Context, db database.Database) ([]*War, error) {
	cursor, err := db.Collection(globals.DB_PREFIX+WarsCollection).Find(ctx, bson.M{
		"status": WarStatusActive,
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot get active wars: %v", err)
	}
	defer cursor.Close(ctx)

	var wars []*War
	err = cursor.All(ctx, &wars)
	if err != nil {
		return nil, fmt.Errorf("Cannot decode wars: %v", err)
	}

	return wars, nil
}

func ArchiveWar(ctx context.Context, db database.Database, warID primitive.ObjectID) error {
	now := time.Now()
	_, err := db.Collection(globals.DB_PREFIX+WarsCollection).UpdateOne(
		ctx,
		bson.M{"_id": warID},
		bson.M{
			"$set": bson.M{
				"status":      WarStatusArchived,
				"archived_at": now,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("Cannot archive war: %v", err)
	}
	return nil
}
