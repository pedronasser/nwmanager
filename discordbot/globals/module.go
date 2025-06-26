package globals

import (
	"nwmanager/discordbot/common"
	"os"

	"github.com/bwmarrin/discordgo"
)

const ModuleName = "globals"

type GlobalsConfig struct {
	AppID   string
	GuildID string

	AdminRoleID      string
	DBPrefix         string
	EveryoneRoleName string
}

type GlobalsModule struct {
}

func (s *GlobalsModule) Name() string {
	return ModuleName
}

func (s *GlobalsModule) Setup(ctx *common.ModuleContext, config any) (bool, error) {
	dg := ctx.Session()

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessageReactions | discordgo.IntentGuildVoiceStates | discordgo.IntentGuilds

	cfg := config.(*GlobalsConfig)

	// dg.State.TrackChannels = true
	// dg.State.TrackRoles = true
	// guild, err := dg.Guild(GuildID)
	// if err != nil {
	// 	log.Fatalf("Cannot get guild: %v", err)
	// }
	// err = dg.State.GuildAdd(guild)
	// if err != nil {
	// 	log.Fatalf("Cannot add guild to state: %v", err)
	// }

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

	DB_PREFIX = cfg.DBPrefix
	ADMIN_ROLE_ID = cfg.AdminRoleID

	return true, nil
}

func (s *GlobalsModule) DefaultConfig() any {
	var AppID = os.Getenv("DISCORD_APP_ID")
	var GuildID = os.Getenv("DISCORD_GUILD_ID")

	return &GlobalsConfig{
		AppID:   AppID,
		GuildID: GuildID,

		AdminRoleID: ADMIN_ROLE_ID,
		DBPrefix:    DB_PREFIX,
	}
}
