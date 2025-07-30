package admin_commands

import (
	"fmt"
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Admin Commands Module
//
// This module provides administrative commands for Discord server management.
// Currently includes:
// - remover-cargo: Remove a specific role from all members in the server
//
// To add a new admin command:
// 1. Define a constant for the command name (e.g., const CmdNewCommand = "new-command")
// 2. Create a handler function with signature: func(ctx *common.ModuleContext, s *discordgo.Session, i *discordgo.InteractionCreate)
// 3. Add an entry to the adminCommands map with the command definition and handler
// 4. The module will automatically handle setup and cleanup of the command

const ModuleName = "admin_commands"

const CmdRemoveRole = "remover-cargo"

type AdminCommandsConfig struct {
	Enabled bool `json:"enabled"`
}

type AdminCommandsModule struct{}

type CommandHandler func(ctx *common.ModuleContext, s *discordgo.Session, i *discordgo.InteractionCreate)

// Map of command names to their definitions and handlers
var adminCommands = map[string]struct {
	Definition *discordgo.ApplicationCommand
	Handler    CommandHandler
}{
	CmdRemoveRole: {
		Definition: &discordgo.ApplicationCommand{
			Name:        CmdRemoveRole,
			Description: "Remove um cargo específico de todos os membros do servidor",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role-id",
					Description: "ID do cargo a ser removido",
					Required:    true,
				},
			},
		},
		Handler: handleRoleRemoval,
	},
}

func (s *AdminCommandsModule) Name() string {
	return ModuleName
}

func (s *AdminCommandsModule) Setup(ctx *common.ModuleContext, config any) (bool, error) {
	var cfg = config.(*AdminCommandsConfig)

	global, _ := ctx.Config("globals").(*globals.GlobalsConfig)
	dg := ctx.Session()

	if !cfg.Enabled {
		log.Println("Admin Commands module is disabled, removing commands...")
		return s.removeAllCommands(dg, global)
	}

	log.Println("Admin Commands module is enabled, setting up...")

	// Create all admin commands
	err := s.createAllCommands(dg, global)
	if err != nil {
		return false, err
	}

	// Add the main handler for all admin commands
	dg.AddHandler(s.createCommandHandler(ctx))

	return true, nil
}

func (s *AdminCommandsModule) DefaultConfig() any {
	var IsModuleEnabledFromEnv = slices.Contains(strings.Split(os.Getenv("MODULES"), ","), ModuleName)
	return &AdminCommandsConfig{
		Enabled: IsModuleEnabledFromEnv,
	}
}

func GetModuleConfig(ctx *common.ModuleContext) *AdminCommandsConfig {
	if module, ok := ctx.Config(ModuleName).(*AdminCommandsConfig); ok {
		return module
	}
	return nil
}

// Helper function to remove all admin commands when module is disabled
func (s *AdminCommandsModule) removeAllCommands(dg *discordgo.Session, global *globals.GlobalsConfig) (bool, error) {
	// Get existing commands
	cmds, err := dg.ApplicationCommands(global.AppID, global.GuildID)
	if err != nil {
		log.Printf("Error fetching slash commands: %v", err)
		return false, err
	}

	// Remove each admin command
	for cmdName := range adminCommands {
		var cmd *discordgo.ApplicationCommand
		for _, c := range cmds {
			if c.Name == cmdName {
				cmd = c
				break
			}
		}

		if cmd == nil {
			continue
		}

		err = dg.ApplicationCommandDelete(global.AppID, global.GuildID, cmd.ID)
		if err != nil {
			log.Printf("Error deleting slash command %s: %v", cmdName, err)
		} else {
			log.Printf("Admin Commands module is disabled, slash command %s removed.", cmdName)
		}
	}

	return false, nil
}

// Helper function to create all admin commands
func (s *AdminCommandsModule) createAllCommands(dg *discordgo.Session, global *globals.GlobalsConfig) error {
	for cmdName, cmdInfo := range adminCommands {
		_, err := dg.ApplicationCommandCreate(global.AppID, global.GuildID, cmdInfo.Definition)
		if err != nil {
			log.Printf("Cannot create slash command %s: %v", cmdName, err)
			return err
		}
		log.Printf("Created admin command: %s", cmdName)
	}
	return nil
}

// Helper function to create the main command handler
func (s *AdminCommandsModule) createCommandHandler(ctx *common.ModuleContext) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		data := i.ApplicationCommandData()

		// Check if this is one of our admin commands
		if cmdInfo, exists := adminCommands[data.Name]; exists {
			cmdInfo.Handler(ctx, s, i)
		}
	}
}

func handleRoleRemoval(ctx *common.ModuleContext, s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if user has admin permissions
	if !globals.IsMemberAdmin(ctx, i.Member) {
		discordutils.ReplyEphemeralMessage(s, i, "❌ Você não possui permissão para usar este comando.", 5*time.Second)
		return
	}

	// Get the role ID from the command options
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		discordutils.ReplyEphemeralMessage(s, i, "❌ ID do cargo não fornecido.", 5*time.Second)
		return
	}

	roleID := options[0].StringValue()
	if roleID == "" {
		discordutils.ReplyEphemeralMessage(s, i, "❌ ID do cargo não pode estar vazio.", 5*time.Second)
		return
	}

	// Verify the role exists
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		discordutils.ReplyEphemeralMessage(s, i, "❌ Erro ao acessar informações do servidor.", 5*time.Second)
		return
	}

	var targetRole *discordgo.Role
	for _, role := range guild.Roles {
		if role.ID == roleID {
			targetRole = role
			break
		}
	}

	if targetRole == nil {
		discordutils.ReplyEphemeralMessage(s, i, "❌ Cargo não encontrado no servidor.", 5*time.Second)
		return
	}

	// Send initial response
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🔄 Iniciando remoção do cargo **%s** de todos os membros...", targetRole.Name),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Error responding to interaction: %v", err)
		return
	}

	// Start the role removal process in a goroutine to avoid timeout
	go removeRoleFromAllMembers(ctx, s, i, roleID, targetRole.Name)
}

func removeRoleFromAllMembers(ctx *common.ModuleContext, s *discordgo.Session, i *discordgo.InteractionCreate, roleID, roleName string) {
	// Get all members with the specified role
	members, err := discordutils.GetGuildMembers(s, i.GuildID, roleID)
	if err != nil {
		updateMessage(s, i, fmt.Sprintf("❌ Erro ao buscar membros: %v", err))
		return
	}

	if len(members) == 0 {
		updateMessage(s, i, fmt.Sprintf("ℹ️ Nenhum membro encontrado com o cargo **%s**.", roleName))
		return
	}

	updateMessage(s, i, fmt.Sprintf("🔄 Encontrados %d membros com o cargo **%s**. Iniciando remoção...", len(members), roleName))

	successCount := 0
	errorCount := 0

	for idx, member := range members {
		err := s.GuildMemberRoleRemove(i.GuildID, member.User.ID, roleID)
		if err != nil {
			log.Printf("Error removing role from member %s (%s): %v", member.User.Username, member.User.ID, err)
			errorCount++
		} else {
			successCount++
			log.Printf("Successfully removed role from member %s (%s)", member.User.Username, member.User.ID)
		}

		// Update progress every 10 members or on the last member
		if (idx+1)%10 == 0 || idx == len(members)-1 {
			progress := fmt.Sprintf("🔄 Progresso: %d/%d membros processados (%d sucessos, %d erros)",
				idx+1, len(members), successCount, errorCount)
			updateMessage(s, i, progress)
		}

		// Small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	// Final message
	finalMessage := fmt.Sprintf("✅ **Processo concluído!**\n\n"+
		"📊 **Estatísticas:**\n"+
		"• Total de membros processados: %d\n"+
		"• Remoções bem-sucedidas: %d\n"+
		"• Erros: %d\n"+
		"• Cargo removido: **%s**",
		len(members), successCount, errorCount, roleName)

	updateMessage(s, i, finalMessage)
}

func updateMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	if err != nil {
		log.Printf("Error updating message: %v", err)
	}
}
