package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const EventsCollection = "events"

// Event type
type EventType string

const (
	EventTypeDungeonNormal EventType = "DungeonNormal"
	EventTypeDungeonM1     EventType = "DungeonM1"
	EventTypeDungeonM2     EventType = "DungeonM2"
	EventTypeDungeonM3     EventType = "DungeonM3"

	EventTypeRaidGorgon EventType = "RaidGorgon"
	EventTypeRaidDevour EventType = "RaidDevour"

	EventTypeOPR   EventType = "OPR"
	EventTypeArena EventType = "Arena"
)

// Event status
type EventStatus string

const (
	EventStatusOpen      EventStatus = "Open"
	EventStatusCompleted EventStatus = "Completed"
)

// Event
type Event struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Type        EventType          `bson:"type" json:"type"`
	Owner       string             `bson:"owner" json:"owner"`
	CreatedAt   *time.Time         `bson:"created_at,omitempty" json:"created_at"`
	ScheduledAt *time.Time         `bson:"scheduled_at,omitempty" json:"scheduled_at"`
	CompletedAt *time.Time         `bson:"completed_at,omitempty" json:"completed_at"`
	PlayerSlots []string           `bson:"player_slots" json:"player_slots"`
	Status      EventStatus        `bson:"status" json:"status"`
	MessageID   string             `bson:"message_id" json:"message_id"`
}
