package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const WarsCollection = "wars"

type WarClass string

const (
	WarClassBruiser           = "bruiser"
	WarClassHealer            = "healer-pocket"
	WarClassHealerAOE         = "healer-aoe"
	WarClassTank              = "tank"
	WarClassVoidFlail         = "void-flail"
	WarClassVoidIce           = "void-ice"
	WarClassFlailIce          = "flail-ice"
	WarClassFireIce           = "fire-ice"
	WarClassFireAbyss         = "fire-abyss"
	WarClassFireBlunder       = "fire-blunderbuss"
	WarClassFireRapier        = "fire-finisher"
	WarClassDisruptorScorpion = "disruptor-scorpion"
	WarClassDisruptor         = "disruptor"
	WarClassDisruptorHatchet  = "disruptor-hatchet"
	WarClassDisruptorPoison   = "disruptor-poison"
	WarClassDisruptorGS       = "disruptor-greatsword"
	WarClassAssassinHatchet   = "assassin-hatchet"
	WarClassAssassinGS        = "assassin-greatsword"
	WarClassBow               = "bow"
)

type WarSlot struct {
	Role WarClass `bson:"role" json:"role"`
}

// War
type War struct {
	ID          primitive.ObjectID  `bson:"_id" json:"id"`
	Description string              `bson:"description" json:"description"`
	Territory   string              `bson:"territory" json:"territory"`
	Type        EventType           `bson:"type" json:"type"`
	CreatedAt   *time.Time          `bson:"created_at,omitempty" json:"created_at"`
	ScheduledAt *time.Time          `bson:"scheduled_at,omitempty" json:"scheduled_at"`
	CompletedAt *time.Time          `bson:"completed_at,omitempty" json:"completed_at"`
	ClosedAt    *time.Time          `bson:"closed_at,omitempty" json:"closed_at"`
	NotifiedAt  *time.Time          `bson:"notified_at,omitempty" json:"notified_at"`
	ConfirmedAt *time.Time          `bson:"confirmed_at,omitempty" json:"confirmed_at"`
	Players     map[string]WarClass `bson:"players" json:"players"`
	Status      EventStatus         `bson:"status" json:"status"`
	MessageID   string              `bson:"message_id" json:"message_id"`
}
