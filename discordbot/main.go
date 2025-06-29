package main

import (
	"fmt"
	"log"
	"nwmanager/database"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/events"
	"nwmanager/discordbot/globals"
	"nwmanager/discordbot/register"
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
	GuildName := os.Getenv("GUILD_NAME")
	Token = os.Getenv("DISCORD_BOT_TOKEN")

	ctx := context.Background()
	// os.MkdirAll("static", os.ModePerm)
	// web.Setup(ctx)

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	dg.StateEnabled = true

	db, err := database.NewMongoDB(ctx, os.Getenv("MONGO_URI"))
	if err != nil {
		log.Fatalf("failed to create database: %v", err)
	}

	ml := common.NewModuleManager(GuildName, db, dg)
	ml.RegisterModule(globals.ModuleName, &globals.GlobalsModule{})
	ml.RegisterModule(register.ModuleName, &register.RegisterModule{})
	ml.RegisterModule(events.ModuleName, &events.EventsModule{})

	ml.Run(ctx)

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
