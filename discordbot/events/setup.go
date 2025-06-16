package events

import (
	"context"
	"fmt"
	"log"
	"nwmanager/database"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func Setup(ctx context.Context, dg *discordgo.Session, AppID, GuildID *string, db database.Database) {
	fmt.Println("Loading events")

	_ = godotenv.Load()
	EVENTS_CHANNEL_ID = os.Getenv("EVENTS_CHANNEL_ID")
	if EVENTS_CHANNEL_NAME != "" {
		channels, err := dg.GuildChannels(*GuildID)
		if err != nil {
			log.Fatalf("Cannot get guild channels: %v", err)
		}
		for _, channel := range channels {
			if channel.Name == EVENTS_CHANNEL_NAME {
				EVENTS_CHANNEL_ID = channel.ID
				break
			}
		}
	}

	if EVENTS_CHANNEL_ID == "" {
		fmt.Println("EVENTS_CHANNEL_ID is not set")
		os.Exit(1)
	}

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
