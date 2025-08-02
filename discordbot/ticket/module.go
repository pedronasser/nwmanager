package ticket

import (
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/globals"
	"os"
	"slices"
	"strings"
)

const ModuleName = "ticket"

type TicketConfig struct {
	Enabled          bool   `json:"enabled"`
	TicketCategoryID string `json:"ticket_category_id"`
	CheckInterval    int    `json:"check_interval_seconds"`
}

type TicketModule struct{}

func (s *TicketModule) Name() string {
	return ModuleName
}

func (s *TicketModule) Setup(ctx *common.ModuleContext, config any) (bool, error) {
	cfg := config.(*TicketConfig)
	if !cfg.Enabled {
		return false, nil
	}

	log.Println("Ticket module is enabled, setting up...")

	globalCfg := ctx.Config("globals").(*globals.GlobalsConfig)
	dg := ctx.Session()

	// Add interaction handlers
	dg.AddHandler(HandleTicketAction(ctx, globalCfg.GuildID))

	// Start background monitoring routine
	go memberRoleMonitoringRoutine(ctx)

	return true, nil
}

func (s *TicketModule) DefaultConfig() any {
	var IsModuleEnabledFromEnv = slices.Contains(strings.Split(os.Getenv("MODULES"), ","), ModuleName)

	return &TicketConfig{
		Enabled:          IsModuleEnabledFromEnv,
		TicketCategoryID: os.Getenv("TICKET_CATEGORY_ID"),
		CheckInterval:    300, // 5 minutes default
	}
}

func GetModuleConfig(ctx *common.ModuleContext) *TicketConfig {
	config, _ := ctx.Config(ModuleName).(*TicketConfig)
	return config
}
