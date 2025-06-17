package events

import (
	"context"
	"fmt"
	"log"
	"nwmanager/database"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func Setup(ctx context.Context, dg *discordgo.Session, AppID, GuildID *string, db database.Database) {
	fmt.Println("Loading events")

	_ = godotenv.Load()
	EVENTS_CHANNEL_IDS = strings.Split(os.Getenv("EVENTS_CHANNEL_IDS"), ",")
	if len(EVENTS_CHANNEL_IDS) == 0 {
		fmt.Println("EVENTS_CHANNEL_IDS is not set")
		os.Exit(1)
	}

	if os.Getenv("EVENTS_REQUIRE_ADMIN") != "" {
		EVENTS_REQUIRE_ADMIN = true
	}

	for _, channel_id := range EVENTS_CHANNEL_IDS {
		_ = setupEventsChannel(ctx, dg, db, globals.ACCESS_ROLE_IDS[globals.EVERYONE_ROLE_NAME], globals.ACCESS_ROLE_IDS[globals.MEMBER_ROLE_NAME], channel_id)
		dg.AddHandler(discordutils.CreateHandler(*GuildID, channel_id, handlers, db))
	}

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

	// dg.AddHandler(HandleReactionAdd(*GuildID, dg, db))
	dg.AddHandler(HandleMessages(*GuildID, dg, db))
	dg.AddHandler(HandleEventClose(*GuildID, dg, db))

	go eventsCheckRoutine(db, dg)
}
