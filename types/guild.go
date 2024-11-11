package types

import (
	"fmt"
	"strings"
	"time"
	"nwmanager/helpers"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const GuildMemberCollection = "guild_member"

type GuildRank int

const (
	GuildRankLeader GuildRank = iota
	GuildRankMember
)

func (r GuildRank) String() string {
	return [...]string{"Guild Leader", "Guild Member"}[r]
}

type GuildMember struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	Name       string             `json:"name" bson:"name"`
	Rank       GuildRank          `json:"rank" bson:"rank"`
	Reputation int                `json:"reputation" bson:"reputation"`
	LastActive time.Time          `json:"last_active" bson:"last_active"`
}

func ParseRank(rank string) GuildRank {
	switch rank {
	case "Guild Leader":
		return GuildRankLeader
	case "Guild Member":
		return GuildRankMember
	default:
		return GuildRankMember
	}
}

func ParseLastActive(lastActive string) time.Time {
	if !strings.HasSuffix(lastActive, " ago") {
		return time.Now()
	}

	t, err := helpers.ParseDuration(strings.TrimRight(lastActive, " ago"))
	if err != nil {
		t = 0
	}

	fmt.Println(t)

	return time.Now().Add(-t)
}
