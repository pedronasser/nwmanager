package management

import (
	"context"
	"fmt"
	"os"
	"time"

	"nwmanager/database"
	"nwmanager/discordbot/discordutils"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	OPR_PRINTS_CHANNEL_ID = ""
)

func Setup(ctx context.Context, dg *discordgo.Session, AppID, GuildID *string, db database.Database) {
	_ = godotenv.Load()
	OPR_PRINTS_CHANNEL_ID = os.Getenv("OPR_PRINTS_CHANNEL_ID")
	if OPR_PRINTS_CHANNEL_ID == "" {
		fmt.Println("OPR_PRINTS_CHANNEL_ID is not set")
		os.Exit(1)
	}

	fmt.Println("Loading management")

	dg.AddHandler(HandleTicketMessages(ctx, dg, GuildID, db))
	dg.AddHandler(HandleTicketInteractions(ctx, dg, GuildID, db))

	routineExportPlayersCSV(ctx, db)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			members := discordutils.RetrieveAllMembers(dg, *GuildID)
			routineRegisterNewPlayers(ctx, dg, GuildID, db, members)
			routineArchiveUnavailablePlayers(ctx, dg, GuildID, db, members)
			routineUnarchiveReturningPlayers(ctx, dg, GuildID, db, members)
			routineDeleteArchivedPlayers(ctx, dg, db)
			routineExportPlayersCSV(ctx, db)
		}
	}()
}
