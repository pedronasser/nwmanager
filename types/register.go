package types

import (
	"context"
	"fmt"
	"nwmanager/database"
	"nwmanager/discordbot/globals"
	"nwmanager/helpers"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	ApprovedBy string             `bson:"approved_by" json:"approved_by"`
	ApprovedAt *time.Time         `bson:"approved_at" json:"approved_at"`
	Rejected   bool               `bson:"rejected" json:"rejected"`
	RejectedBy string             `bson:"rejected_by" json:"rejected_by"`
	RejectedAt *time.Time         `bson:"rejected_at" json:"rejected_at"`
}

func InsertRegister(ctx context.Context, db database.Database, register *Register) error {
	_, err := db.Collection(globals.DB_PREFIX+RegisterCollection).InsertOne(ctx, register)
	if err != nil {
		return fmt.Errorf("Cannot insert register: %v", err)
	}
	return nil
}

func GetRegisterByID(ctx context.Context, db database.Database, id string) (*Register, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("Invalid register ID: %v", err)
	}

	q := db.Collection(globals.DB_PREFIX+RegisterCollection).FindOne(ctx, bson.M{"_id": objectID})
	if q.Err() != nil {
		if q.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("Cannot find register: %v", q.Err())
	}

	var register Register
	err = q.Decode(&register)
	if err != nil {
		return nil, fmt.Errorf("Cannot decode register: %v", err)
	}

	return &register, nil
}

func GetPendingRegisters(ctx context.Context, db database.Database) ([]*Register, error) {
	cursor, err := db.Collection(globals.DB_PREFIX+RegisterCollection).Find(ctx, bson.M{
		"approved": false,
		"rejected": false,
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot find pending registers: %v", err)
	}
	defer cursor.Close(ctx)

	var registers []*Register
	err = cursor.All(ctx, &registers)
	if err != nil {
		return nil, fmt.Errorf("Cannot decode registers: %v", err)
	}

	return registers, nil
}

func ApproveRegister(ctx context.Context, db database.Database, registerID, approverID string) error {
	objectID, err := primitive.ObjectIDFromHex(registerID)
	if err != nil {
		return fmt.Errorf("Invalid register ID: %v", err)
	}

	now := helpers.GetCurrentTimeAsUTC()
	_, err = db.Collection(globals.DB_PREFIX+RegisterCollection).UpdateOne(ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{
			"approved":    true,
			"approved_by": approverID,
			"approved_at": now,
		}})
	if err != nil {
		return fmt.Errorf("Cannot approve register: %v", err)
	}

	return nil
}

func RejectRegister(ctx context.Context, db database.Database, registerID, rejecterID string) error {
	objectID, err := primitive.ObjectIDFromHex(registerID)
	if err != nil {
		return fmt.Errorf("Invalid register ID: %v", err)
	}

	now := helpers.GetCurrentTimeAsUTC()
	_, err = db.Collection(globals.DB_PREFIX+RegisterCollection).UpdateOne(ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{
			"rejected":    true,
			"rejected_by": rejecterID,
			"rejected_at": now,
		}})
	if err != nil {
		return fmt.Errorf("Cannot reject register: %v", err)
	}

	return nil
}
