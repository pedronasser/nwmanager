package main

import (
	"fmt"
	"log"
	"nwmanager/database"
	"nwmanager/discordbot/events"
	"nwmanager/discordbot/globals"
	"nwmanager/discordbot/management"
	"nwmanager/discordbot/signup"
	"nwmanager/discordbot/war"
	"nwmanager/discordbot/web"
	"os"
	"os/signal"
	"slices"
	"strings"
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

	ctx := context.Background()
	os.MkdirAll("static", os.ModePerm)
	web.Setup(ctx)

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

	db, err := database.NewMongoDB(ctx, os.Getenv("MONGO_URI"))
	if err != nil {
		log.Fatalf("failed to create database: %v", err)
	}

	globals.Setup(ctx, dg, &AppID, &GuildID)
	if shouldLoadModule("events") {
		events.Setup(ctx, dg, &AppID, &GuildID, db)
	}
	if shouldLoadModule("war") {
		war.Setup(ctx, dg, &AppID, &GuildID, db)
	}
	if shouldLoadModule("management") {
		management.Setup(ctx, dg, &AppID, &GuildID, db)
	}
	if shouldLoadModule("signup") {
		signup.Setup(ctx, dg, &AppID, &GuildID, db)
	}

	// Just like the ping pong example, we only care about receiving message
	// events in this example.
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessageReactions | discordgo.IntentGuildVoiceStates | discordgo.IntentGuilds

	// dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
	// 	for _, guild := range r.Guilds {
	// 		// fmt.Println("Connected to guild:", guild.Name, "with ID:", guild.ID)
	// 		guild, err := dg.Guild(guild.ID) // Ensure the guild is cached
	// 		if err != nil {
	// 			log.Printf("Error fetching guild %s: %v", guild.ID, err)
	// 			continue
	// 		}
	// 		fmt.Printf("Guild Name: %s, ID: %s, Member Count: %d\n", guild.Name, guild.ID, guild.MemberCount)
	// 	}
	// })

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

func shouldLoadModule(m string) bool {
	modulesEnv := os.Getenv("MODULES")
	if modulesEnv == "" {
		return true
	}

	if modulesEnv == "all" {
		return true
	}

	var modules []string = strings.Split(modulesEnv, ",")

	return slices.Contains(modules, m)
}
