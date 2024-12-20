package war

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
	fmt.Println("Loading war")
	_ = setupEventsChannel(ctx, dg, db, globals.ACCESS_ROLE_IDS[globals.EVERYONE_ROLE_NAME], globals.ACCESS_ROLE_IDS[globals.MEMBER_ROLE_NAME])

	_, err := dg.ApplicationCommandCreate(*AppID, *GuildID, &discordgo.ApplicationCommand{
		Name:        "guerra",
		Description: "Iniciar um agendamento de guerra",
	})
	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	dg.AddHandler(discordutils.CreateHandler(*GuildID, WAR_CHANNEL_ID, handlers, db))
	dg.AddHandler(ClearChannelMessages(*GuildID, dg, db))
	dg.AddHandler(HandleWarInteractions(*GuildID, dg, db))

	go eventsCheckRoutine(db, dg, *GuildID)
}
