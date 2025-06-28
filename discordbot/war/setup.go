package war

import (
	"context"
	"nwmanager/database"

	"github.com/bwmarrin/discordgo"
)

func Setup(ctx context.Context, dg *discordgo.Session, AppID, GuildID *string, db database.Database) {
	// fmt.Println("Loading war")

	// _ = godotenv.Load()
	// WAR_CHANNEL_ID = os.Getenv("WAR_CHANNEL_ID")
	// if WAR_CHANNEL_ID == "" {
	// 	fmt.Println("WAR_CHANNEL_ID is not set")
	// 	os.Exit(1)
	// }

	// _ = setupEventsChannel(ctx, dg, db, globals.ACCESS_ROLE_IDS[globals.EVERYONE_ROLE_NAME], globals.ACCESS_ROLE_IDS[globals.MEMBER_ROLE_NAME])

	// _, err := dg.ApplicationCommandCreate(*AppID, *GuildID, &discordgo.ApplicationCommand{
	// 	Name:        "guerra",
	// 	Description: "Iniciar um agendamento de guerra",
	// })
	// if err != nil {
	// 	log.Fatalf("Cannot create slash command: %v", err)
	// }

	// dg.AddHandler(discordutils.CreateChannelHandler(*GuildID, WAR_CHANNEL_ID, handlers, db))
	// dg.AddHandler(ClearChannelMessages(*GuildID, dg, db))
	// dg.AddHandler(HandleWarInteractions(*GuildID, dg, db))

	// go eventsCheckRoutine(db, dg, *GuildID)
}
