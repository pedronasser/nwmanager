package ticket

import (
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/globals"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const ModuleName = "ticket"

type TicketConfig struct {
	Enabled          bool   `json:"enabled"`
	TicketCategoryID string `json:"ticket_category_id"`
	AbsenceChannelID string `json:"absence_channel_id"`
	CheckInterval    int    `json:"check_interval_seconds"`
}

type TicketModule struct{}

func (s *TicketModule) Name() string {
	return ModuleName
}

func (s *TicketModule) Setup(ctx *common.ModuleContext, config any) (bool, error) {
	cfg := config.(*TicketConfig)
	globalCfg := ctx.Config("globals").(*globals.GlobalsConfig)
	dg := ctx.Session()

	if !cfg.Enabled {
		log.Println("Ticket module is disabled, removing commands...")
		// Remove slash commands when module is disabled
		cmds, err := dg.ApplicationCommands(globalCfg.AppID, globalCfg.GuildID)
		if err == nil {
			for _, cmd := range cmds {
				if cmd.Name == "ausencia" || cmd.Name == "sync-ticket-permissions" {
					err = dg.ApplicationCommandDelete(globalCfg.AppID, globalCfg.GuildID, cmd.ID)
					if err != nil {
						log.Printf("Error deleting /%s command: %v", cmd.Name, err)
					} else {
						log.Printf("Removed /%s slash command", cmd.Name)
					}
				}
			}
		}
		return false, nil
	}

	log.Println("Ticket module is enabled, setting up...")

	// Add interaction handlers
	dg.AddHandler(HandleTicketAction(ctx, globalCfg.GuildID))

	// Add Discord event handlers for real-time member monitoring
	dg.AddHandler(HandleGuildMemberUpdate(ctx))
	dg.AddHandler(HandleGuildMemberRemove(ctx))

	// Register slash command for absence notification
	_, err := dg.ApplicationCommandCreate(globalCfg.AppID, globalCfg.GuildID, &discordgo.ApplicationCommand{
		Name:        "ausencia",
		Description: "Avisar ausência para a guild",
		Type:        discordgo.ChatApplicationCommand,
	})
	if err != nil {
		log.Printf("Failed to create /ausencia command: %v", err)
	} else {
		log.Println("Created /ausencia slash command")
	}

	// Register admin command for syncing ticket permissions
	_, err = dg.ApplicationCommandCreate(globalCfg.AppID, globalCfg.GuildID, &discordgo.ApplicationCommand{
		Name:        "sync-ticket-permissions",
		Description: "Sincronizar permissões de todos os tickets com suas categorias (Admin only)",
		Type:        discordgo.ChatApplicationCommand,
	})
	if err != nil {
		log.Printf("Failed to create /sync-ticket-permissions command: %v", err)
	} else {
		log.Println("Created /sync-ticket-permissions slash command")
	}

	// Update existing ticket messages with latest components
	go func() {
		// Wait a bit for the bot to be fully ready
		time.Sleep(2 * time.Second)
		err := updateExistingTicketMessages(ctx)
		if err != nil {
			log.Printf("Error updating existing ticket messages: %v", err)
		}
	}()

	// Start background monitoring routine
	go memberRoleMonitoringRoutine(ctx)

	return true, nil
}

func (s *TicketModule) DefaultConfig() any {
	var IsModuleEnabledFromEnv = slices.Contains(strings.Split(os.Getenv("MODULES"), ","), ModuleName)

	return &TicketConfig{
		Enabled:          IsModuleEnabledFromEnv,
		TicketCategoryID: os.Getenv("TICKET_CATEGORY_ID"),
		AbsenceChannelID: os.Getenv("ABSENCE_CHANNEL_ID"),
		CheckInterval:    1800, // 30 minutes (since real-time events handle most cases)
	}
}

func GetModuleConfig(ctx *common.ModuleContext) *TicketConfig {
	config, _ := ctx.Config(ModuleName).(*TicketConfig)
	return config
}
