package globals

import (
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/helpers"
	"os"

	"github.com/bwmarrin/discordgo"
)

const ModuleName = "globals"

type GlobalsConfig struct {
	// Env variables
	AppID   string
	GuildID string

	// DB configurable values
	AdminRoleID    string `json:"admin_role_id"`
	MemberRoleID   string `json:"member_role_id"`
	CompleteRoleID string `json:"complete_role_id"`

	ClassRoleIDs     map[string]string `json:"class_role_ids"`
	ClassEmojiIDs    map[string]string `json:"class_emoji_ids"`
	ClassCategoryIDs map[string]string `json:"class_category_ids"`
}

type GlobalsModule struct {
}

func (s *GlobalsModule) Name() string {
	return ModuleName
}

func (s *GlobalsModule) Setup(ctx *common.ModuleContext, config any) (bool, error) {
	dg := ctx.Session()
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessageReactions | discordgo.IntentGuildVoiceStates | discordgo.IntentGuilds

	cfg := GetModuleConfig(ctx)

	if cfg.AdminRoleID == "" {
		panic("AdminRoleID is not set")
	}

	// dg.State.TrackChannels = true
	// dg.State.TrackRoles = true
	guild, err := dg.Guild(cfg.GuildID)
	if err != nil {
		log.Fatalf("Cannot get guild: %v", err)
	}
	err = dg.State.GuildAdd(guild)
	if err != nil {
		log.Fatalf("Cannot add guild to state: %v", err)
	}

	DB_PREFIX = ctx.GuildName() + "_"
	ADMIN_ROLE_ID = cfg.AdminRoleID

	return true, nil
}

func (s *GlobalsModule) DefaultConfig() any {
	var AppID = os.Getenv("DISCORD_APP_ID")
	var GuildID = os.Getenv("DISCORD_GUILD_ID")
	ADMIN_ROLE_ID = helpers.LoadOrDefault("ADMIN_ROLE_ID", "")
	MEMBER_ROLE_ID := helpers.LoadOrDefault("MEMBER_ROLE_ID", "")
	COMPLETE_ROLE_ID := helpers.LoadOrDefault("COMPLETE_ROLE_ID", "")

	return &GlobalsConfig{
		AppID:   AppID,
		GuildID: GuildID,

		AdminRoleID:    ADMIN_ROLE_ID,
		MemberRoleID:   MEMBER_ROLE_ID,
		CompleteRoleID: COMPLETE_ROLE_ID,

		ClassRoleIDs:     make(map[string]string),
		ClassEmojiIDs:    make(map[string]string),
		ClassCategoryIDs: make(map[string]string),
	}
}

func GetModuleConfig(ctx *common.ModuleContext) *GlobalsConfig {
	if module, ok := ctx.Config(ModuleName).(*GlobalsConfig); ok {
		return module
	}
	return nil
}
