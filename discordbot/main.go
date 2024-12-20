package main

import (
	"fmt"
	"log"
	"nwmanager/database"
	"nwmanager/discordbot/events"
	"nwmanager/discordbot/globals"
	"nwmanager/discordbot/management"
	"nwmanager/discordbot/war"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"golang.org/x/net/context"
)

// Variables used for command line parameters
var (
	Token string
)

func init() {
	// .env
	_ = godotenv.Load()
}

func main() {
	Token = os.Getenv("DISCORD_BOT_TOKEN")

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	dg.StateEnabled = true

	var AppID = os.Getenv("DISCORD_APP_ID")
	var GuildID = os.Getenv("DISCORD_GUILD_ID")

	dg.State.TrackChannels = true
	dg.State.TrackRoles = true
	guild, err := dg.Guild(GuildID)
	if err != nil {
		log.Fatalf("Cannot get guild: %v", err)
	}
	err = dg.State.GuildAdd(guild)
	if err != nil {
		log.Fatalf("Cannot add guild to state: %v", err)
	}

	ctx := context.Background()
	db, err := database.NewMongoDB(ctx, os.Getenv("MONGO_URI"))
	if err != nil {
		log.Fatalf("failed to create database: %v", err)
	}

	globals.Setup(ctx, dg, &AppID, &GuildID, db)
	// register.Setup(dg, &AppID, &GuildID, db)
	events.Setup(ctx, dg, &AppID, &GuildID, db)
	war.Setup(ctx, dg, &AppID, &GuildID, db)
	management.Setup(ctx, dg, &AppID, &GuildID, db)

	// Just like the ping pong example, we only care about receiving message
	// events in this example.
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessageReactions

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}
