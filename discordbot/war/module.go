package war

import (
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"os"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const ModuleName = "war"

type WarConfig struct {
	Enabled      bool   `json:"enabled"`
	WarChannelID string `json:"war_channel_id"`
}

type WarModule struct{}

func (w *WarModule) Name() string {
	return ModuleName
}

func (w *WarModule) Setup(ctx *common.ModuleContext, config any) (bool, error) {
	var cfg = config.(*WarConfig)
	if !cfg.Enabled {
		return false, nil
	}

	log.Println("War module is enabled, setting up...")

	global, _ := ctx.Config("globals").(*globals.GlobalsConfig)
	dg := ctx.Session()

	// Register slash command
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "criar-guerra",
			Description: "Cria uma nova guerra para a guild",
		},
		{
			Name:        "cancelar-guerra",
			Description: "Cancela a guerra ativa atual",
		},
	}

	for _, command := range commands {
		_, err := dg.ApplicationCommandCreate(global.AppID, global.GuildID, command)
		if err != nil {
			log.Printf("Cannot create slash command %s: %v", command.Name, err)
		}
	}

	// Add handlers for war channel if configured
	if cfg.WarChannelID != "" {
		channel, err := dg.Channel(cfg.WarChannelID)
		if err != nil {
			log.Printf("Could not retrieve war channel: %v", err)
		} else {
			dg.AddHandler(discordutils.CreateChannelHandler(ctx, channel, handlers))
		}
	}

	// Add interaction handlers
	dg.AddHandler(HandleWarAction(ctx, global.GuildID))

	// Start cleanup routine
	go warCleanupRoutine(ctx)

	return true, nil
}

func (w *WarModule) DefaultConfig() any {
	var IsModuleEnabledFromEnv = slices.Contains(strings.Split(os.Getenv("MODULES"), ","), ModuleName)

	return &WarConfig{
		Enabled:      IsModuleEnabledFromEnv,
		WarChannelID: os.Getenv("WAR_CHANNEL_ID"),
	}
}

func GetModuleConfig(ctx *common.ModuleContext) *WarConfig {
	if module, ok := ctx.Config(ModuleName).(*WarConfig); ok {
		return module
	}
	return nil
}
