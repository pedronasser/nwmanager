package events

import (
	"context"
	"fmt"
	"log"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"nwmanager/types"

	"github.com/bwmarrin/discordgo"
)

func Setup(ctx context.Context, dg *discordgo.Session, AppID, GuildID *string, db types.Database) {
	fmt.Println("Loading events")
	_ = setupEventsChannel(ctx, dg, db, globals.ACCESS_ROLE_IDS[globals.EVERYONE_ROLE_NAME], globals.ACCESS_ROLE_IDS[globals.MEMBER_ROLE_NAME])

	_, err := dg.ApplicationCommandCreate(*AppID, *GuildID, &discordgo.ApplicationCommand{
		Name:        "evento",
		Description: "Iniciar um novo evento",
	})
	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	_, err = dg.ApplicationCommandCreate(*AppID, *GuildID, &discordgo.ApplicationCommand{
		Name:        "encerrar",
		Description: "Encerre um evento",
	})
	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	dg.AddHandler(discordutils.CreateHandler(*GuildID, EVENTS_CHANNEL_ID, handlers, db))
	// dg.AddHandler(HandleReactionAdd(*GuildID, dg, db))
	dg.AddHandler(HandleMessages(*GuildID, dg, db))
	dg.AddHandler(HandleEventClose(*GuildID, dg, db))

	go eventsCheckRoutine(db, dg)
}
