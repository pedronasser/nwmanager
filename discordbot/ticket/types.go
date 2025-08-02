package ticket

import (
	"fmt"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/globals"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const TicketCollection = "tickets"

type Ticket struct {
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	DiscordID     string             `bson:"discord_id" json:"discord_id"`
	ChannelID     string             `bson:"channel_id" json:"channel_id"`
	MessageID     string             `bson:"message_id" json:"message_id"`
	PlayerIGN     string             `bson:"player_ign" json:"player_ign"`
	PlayerClass   string             `bson:"player_class" json:"player_class"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	LastUpdatedAt time.Time          `bson:"last_updated_at" json:"last_updated_at"`
	IsActive      bool               `bson:"is_active" json:"is_active"`
}

func insertTicket(ctx *common.ModuleContext, ticket *Ticket) error {
	ticket.ID = primitive.NewObjectID()
	_, err := ctx.DB().Collection(globals.DB_PREFIX+TicketCollection).InsertOne(ctx.Context, ticket)
	if err != nil {
		return fmt.Errorf("cannot insert ticket: %v", err)
	}
	return nil
}

func getTicketByChannelID(ctx *common.ModuleContext, channelID string) (*Ticket, error) {
	q := ctx.DB().Collection(globals.DB_PREFIX+TicketCollection).FindOne(ctx.Context, bson.M{
		"channel_id": channelID,
		"is_active":  true,
	})
	if q.Err() != nil {
		if q.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot find ticket: %v", q.Err())
	}

	var ticket Ticket
	err := q.Decode(&ticket)
	if err != nil {
		return nil, fmt.Errorf("cannot decode ticket: %v", err)
	}

	return &ticket, nil
}

func getTicketByDiscordID(ctx *common.ModuleContext, discordID string) (*Ticket, error) {
	q := ctx.DB().Collection(globals.DB_PREFIX+TicketCollection).FindOne(ctx.Context, bson.M{
		"discord_id": discordID,
		"is_active":  true,
	})
	if q.Err() != nil {
		if q.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot find ticket: %v", q.Err())
	}

	var ticket Ticket
	err := q.Decode(&ticket)
	if err != nil {
		return nil, fmt.Errorf("cannot decode ticket: %v", err)
	}

	return &ticket, nil
}

func getAllActiveTickets(ctx *common.ModuleContext) ([]Ticket, error) {
	cursor, err := ctx.DB().Collection(globals.DB_PREFIX+TicketCollection).Find(ctx.Context, bson.M{
		"is_active": true,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot find tickets: %v", err)
	}
	defer cursor.Close(ctx.Context)

	var tickets []Ticket
	if err = cursor.All(ctx.Context, &tickets); err != nil {
		return nil, fmt.Errorf("cannot decode tickets: %v", err)
	}

	return tickets, nil
}

func deleteTicket(ctx *common.ModuleContext, ticket *Ticket) error {
	_, err := ctx.DB().Collection(globals.DB_PREFIX+TicketCollection).UpdateOne(ctx.Context,
		bson.M{"_id": ticket.ID},
		bson.M{"$set": bson.M{
			"is_active":       false,
			"last_updated_at": time.Now(),
		}})
	if err != nil {
		return fmt.Errorf("cannot deactivate ticket: %v", err)
	}
	return nil
}

func updateTicket(ctx *common.ModuleContext, ticket *Ticket) error {
	ticket.LastUpdatedAt = time.Now()
	_, err := ctx.DB().Collection(globals.DB_PREFIX+TicketCollection).UpdateOne(ctx.Context,
		bson.M{"_id": ticket.ID},
		bson.M{"$set": ticket})
	if err != nil {
		return fmt.Errorf("cannot update ticket: %v", err)
	}
	return nil
}
