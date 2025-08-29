package ticket

import (
	"context"
	"fmt"
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Helper functions

func isTicketOwner(ctx *common.ModuleContext, channelID, userID string) bool {
	// Get ticket from database to check ownership
	ticket, err := getTicketByChannelID(ctx, channelID)
	if err != nil || ticket == nil {
		return false
	}
	return ticket.DiscordID == userID
}

func createClassSelectOptions(classEmojis map[string]string) []discordgo.SelectMenuOption {
	var options []discordgo.SelectMenuOption

	for class, emoji := range classEmojis {
		options = append(options, discordgo.SelectMenuOption{
			Label: globals.PVP_CLASS_NAMES[globals.PVPClassType(class)],
			Value: class,
			Emoji: &discordgo.ComponentEmoji{Name: emoji},
		})
	}

	return options
}

func getPlayerByTicketChannel(ctx *common.ModuleContext, channelID string) (*types.Player, error) {
	// Get ticket first
	ticket, err := getTicketByChannelID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}
	if ticket == nil {
		return nil, fmt.Errorf("ticket not found")
	}

	// Get player by discord ID
	player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), ticket.DiscordID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	return player, nil
}

func updatePlayerClass(ctx *common.ModuleContext, player *types.Player, selectedClass string) error {
	// Update player's war class
	player.WarClass = selectedClass

	// Update in database
	err := types.UpdatePlayer(context.Background(), ctx.DB(), player)
	if err != nil {
		return fmt.Errorf("failed to update player class: %w", err)
	}

	return nil
}

func getCurrentBuildInfo(ctx *common.ModuleContext, player *types.Player) string {
	log.Printf("getCurrentBuildInfo called for player %s, BuildImages count: %d", player.IGN, len(player.BuildImages))

	// Basic player info
	classDisplay := player.WarClass
	if classDisplay == "" {
		classDisplay = "Não definida"
	}
	buildInfo := fmt.Sprintf("**Classe:** %s\n", classDisplay)

	// Check for last update time
	if player.Stats != nil && player.Stats.LastEquipUpdate != nil {
		buildInfo += fmt.Sprintf("**Última atualização:** %s\n", player.Stats.LastEquipUpdate.Format("02/01/2006 15:04"))
	}

	// Check for build images (BuildImages is directly on Player struct)
	if len(player.BuildImages) > 0 {
		log.Printf("Found %d build images for player %s", len(player.BuildImages), player.IGN)
		buildInfo += "\n**Screenshots da Build:**\n"
		for i, imageURL := range player.BuildImages {
			buildInfo += fmt.Sprintf("[Imagem %d](%s)\n", i+1, imageURL)
		}
	} else {
		log.Printf("No build images found for player %s", player.IGN)
		buildInfo += "\n**Nenhuma screenshot enviada ainda**"
	}

	return buildInfo
}

func setupBuildThread(ctx *common.ModuleContext, threadID, playerIGN string) {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📸 Build - %s", playerIGN),
		Description: "**Envie as screenshots da sua build aqui!**\n\nPode enviar várias imagens mostrando:\n• Equipamentos\n• Atributos\n• Skills\n• Armas\n\n*Suas imagens serão automaticamente salvas quando você confirmar a build.*\n\nQuando terminar, clique no botão abaixo para confirmar.",
		Color:       0x0099ff,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "✅ Confirmar Build",
					Style:    discordgo.SuccessButton,
					CustomID: "ticket:submit_build",
				},
				discordgo.Button{
					Label:    "❌ Fechar Thread",
					Style:    discordgo.SecondaryButton,
					CustomID: "ticket:close_thread",
				},
			},
		},
	}

	_, err := ctx.Session().ChannelMessageSendComplex(threadID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		log.Printf("Error setting up build thread: %v", err)
	}
}

func setupQuestionThread(ctx *common.ModuleContext, threadID, playerIGN string) {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("❓ Dúvidas - %s", playerIGN),
		Description: "**Envie suas dúvidas aqui!**\n\nPode perguntar sobre:\n• Builds e equipamentos\n• Estratégias de PvP\n• Eventos da guild\n• Qualquer outra coisa relacionada ao jogo\n\nQuando sua dúvida for resolvida, clique no botão para fechar a thread.",
		Color:       0xff9900,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "✅ Dúvida Resolvida",
					Style:    discordgo.SuccessButton,
					CustomID: "ticket:close_thread",
				},
			},
		},
	}

	_, err := ctx.Session().ChannelMessageSendComplex(threadID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		log.Printf("Error setting up question thread: %v", err)
	}
}

func createTicketMessageComponents() []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		// discordgo.ActionsRow{
		// 	Components: []discordgo.MessageComponent{
		// 		discordgo.Button{
		// 			Label:    "Enviar Build",
		// 			Style:    discordgo.PrimaryButton,
		// 			CustomID: "ticket:send_build",
		// 			Emoji: &discordgo.ComponentEmoji{
		// 				Name: "📸",
		// 			},
		// 		},
		// 		discordgo.Button{
		// 			Label:    "Enviar Dúvida",
		// 			Style:    discordgo.SecondaryButton,
		// 			CustomID: "ticket:send_question",
		// 			Emoji: &discordgo.ComponentEmoji{
		// 				Name: "❓",
		// 			},
		// 		},
		// 	},
		// },
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Trocar Classe Guerra",
					Style:    discordgo.SecondaryButton,
					CustomID: "ticket:change_class",
					Emoji: &discordgo.ComponentEmoji{
						Name: "⚔️",
					},
				},
				discordgo.Button{
					Label:    "Avisar Ausência",
					Style:    discordgo.SecondaryButton,
					CustomID: "ticket:notify_absence",
					Emoji: &discordgo.ComponentEmoji{
						Name: "📅",
					},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Alterar Status",
					Style:    discordgo.PrimaryButton,
					CustomID: "ticket:change_build_status",
					Emoji: &discordgo.ComponentEmoji{
						Name: "🔧",
					},
				},
			},
		},
	}
}

func setupTicketMessage(ctx *common.ModuleContext, channel *discordgo.Channel, player *types.Player) error {
	// Get build status emoji for title
	statusEmoji := globals.BUILD_STATUS_EMOJIS[player.BuildStatus]
	if statusEmoji == "" {
		statusEmoji = globals.BUILD_STATUS_EMOJIS[globals.BUILD_MISSING] // Default fallback
	}

	// Build embed fields with player information
	var embedFields []*discordgo.MessageEmbedField

	// War experience field
	warExperienceValue := "Não"
	if player.HasWarExperience {
		warExperienceValue = "Sim"
	}
	embedFields = append(embedFields, &discordgo.MessageEmbedField{
		Name:   "⚔️ Experiência em Guerras?",
		Value:  warExperienceValue,
		Inline: true,
	})

	// Previous guild name field
	guildValue := "Nenhuma"
	if player.HasWarExperience && player.PreviousGuildName != "" {
		guildValue = player.PreviousGuildName
	} else if !player.HasWarExperience {
		guildValue = "N/A" // User has no war experience, so previous guilds don't apply
	}
	embedFields = append(embedFields, &discordgo.MessageEmbedField{
		Name:   "🏛️ Guild(s) Anterior(es):",
		Value:  guildValue,
		Inline: true,
	})

	// Get embed color based on build status
	embedColor := globals.BUILD_STATUS_COLORS[player.BuildStatus]
	if embedColor == 0 {
		embedColor = globals.BUILD_STATUS_COLORS[globals.BUILD_MISSING] // Default fallback
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s 🎫 Ticket - %s", statusEmoji, player.IGN),
		Description: fmt.Sprintf("Bem-vindo(a) ao seu ticket pessoal, **%s**!", player.IGN),
		Color:       embedColor,
		Fields:      embedFields,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use os botões abaixo para interagir com o ticket",
		},
	}

	components := createTicketMessageComponents()

	message, err := ctx.Session().ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		return fmt.Errorf("failed to send ticket message: %w", err)
	}

	// Pin the ticket message
	err = ctx.Session().ChannelMessagePin(channel.ID, message.ID)
	if err != nil {
		log.Printf("Error pinning ticket message for player %s: %v", player.IGN, err)
		// Don't return error as the ticket message was sent successfully
	} else {
		log.Printf("Successfully pinned ticket message for player %s in channel %s", player.IGN, channel.ID)
	}

	// Store ticket in database
	ticket := &Ticket{
		DiscordID:     player.DiscordID,
		ChannelID:     channel.ID,
		MessageID:     message.ID,
		PlayerIGN:     player.IGN,
		PlayerClass:   player.WarClass,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
		IsActive:      true,
	}

	err = insertTicket(ctx, ticket)
	if err != nil {
		log.Printf("Error storing ticket in database: %v", err)
		// Don't return error as the ticket channel was created successfully
	}

	// Update player.ticket_channel with the channel ID
	player.TicketChannel = channel.ID
	err = types.UpdatePlayer(ctx.Context, ctx.DB(), player)
	if err != nil {
		log.Printf("Error updating player ticket_channel for player %s: %v", player.IGN, err)
		// Don't return error as the ticket channel was created successfully
	} else {
		log.Printf("Successfully updated player %s with ticket channel %s", player.IGN, channel.ID)
	}

	return nil
}

func updateExistingTicketMessages(ctx *common.ModuleContext) error {
	log.Println("Updating existing ticket messages with latest components...")

	// Get all active tickets
	activeTickets, err := getAllActiveTickets(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active tickets: %w", err)
	}

	components := createTicketMessageComponents()
	updatedCount := 0
	errorCount := 0

	for _, ticket := range activeTickets {
		// Try to get the original message
		message, err := ctx.Session().ChannelMessage(ticket.ChannelID, ticket.MessageID)
		if err != nil {
			log.Printf("Error fetching message %s in channel %s: %v", ticket.MessageID, ticket.ChannelID, err)
			errorCount++
			continue
		}

		// Preserve the original embed but update components
		var embeds []*discordgo.MessageEmbed
		if len(message.Embeds) > 0 {
			embeds = message.Embeds
		}

		// Update the message with new components
		_, err = ctx.Session().ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel:    ticket.ChannelID,
			ID:         ticket.MessageID,
			Embeds:     &embeds,
			Components: &components,
		})
		if err != nil {
			log.Printf("Error updating message %s in channel %s: %v", ticket.MessageID, ticket.ChannelID, err)
			errorCount++
			continue
		}

		updatedCount++
		log.Printf("Updated ticket message for player %s (channel: %s)", ticket.PlayerIGN, ticket.ChannelID)
	}

	log.Printf("Ticket message update completed. Updated: %d, Errors: %d", updatedCount, errorCount)
	return nil
}

func syncAllTicketPermissions(ctx *common.ModuleContext) error {
	log.Println("Syncing permissions for all active tickets with their categories...")

	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)
	ticketConfig := GetModuleConfig(ctx)

	// Get all active tickets
	activeTickets, err := getAllActiveTickets(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active tickets: %w", err)
	}

	if len(activeTickets) == 0 {
		log.Println("No active tickets found to sync")
		return nil
	}

	syncedCount := 0
	errorCount := 0
	movedCount := 0

	for i, ticket := range activeTickets {
		log.Printf("Processing ticket %d/%d: %s (player: %s)", i+1, len(activeTickets), ticket.ChannelID, ticket.PlayerIGN)

		moved, err := syncTicketPermissions(ctx, &ticket, globalConfig, ticketConfig)
		if err != nil {
			log.Printf("Error syncing permissions for ticket %s (player: %s): %v", ticket.ChannelID, ticket.PlayerIGN, err)
			errorCount++
			continue
		}

		syncedCount++
		if moved {
			movedCount++
		}
		log.Printf("Successfully synced permissions for ticket %s (player: %s)", ticket.ChannelID, ticket.PlayerIGN)
	}

	log.Printf("Permission sync completed. Total: %d, Synced: %d, Moved: %d, Errors: %d", len(activeTickets), syncedCount, movedCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("completed with %d errors out of %d tickets", errorCount, len(activeTickets))
	}

	return nil
}

func syncTicketPermissions(ctx *common.ModuleContext, ticket *Ticket, globalConfig *globals.GlobalsConfig, ticketConfig *TicketConfig) (bool, error) {
	// Get current channel information
	channel, err := ctx.Session().Channel(ticket.ChannelID)
	if err != nil {
		return false, fmt.Errorf("failed to get channel information: %w", err)
	}

	// Get player information to determine the correct category
	player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), ticket.DiscordID)
	if err != nil {
		return false, fmt.Errorf("failed to get player information: %w", err)
	}
	if player == nil {
		return false, fmt.Errorf("player not found for discord ID: %s", ticket.DiscordID)
	}

	// Determine the correct category for this ticket
	var targetCategoryID string
	if player.WarClass != "" {
		// Player has a class defined, check if there's a class-specific category
		if classCategory, exists := globalConfig.ClassCategoryIDs[player.WarClass]; exists {
			targetCategoryID = classCategory
		} else {
			// Fallback to main ticket category if class category doesn't exist
			targetCategoryID = ticketConfig.TicketCategoryID
		}
	} else {
		// No class defined, use main ticket category
		targetCategoryID = ticketConfig.TicketCategoryID
	}

	// If no target category is configured, skip this ticket
	if targetCategoryID == "" {
		log.Printf("No target category configured for ticket %s, skipping permission sync", ticket.ChannelID)
		return false, nil
	}

	// Get the target category to copy its permissions
	category, err := ctx.Session().Channel(targetCategoryID)
	if err != nil {
		return false, fmt.Errorf("failed to get category %s: %w", targetCategoryID, err)
	}

	// Create the permission overwrites based on category permissions
	permissionOverwrites := make([]*discordgo.PermissionOverwrite, len(category.PermissionOverwrites))
	copy(permissionOverwrites, category.PermissionOverwrites)

	// Ensure the ticket owner has the proper permissions
	hasOwnerPermission := false
	for _, overwrite := range permissionOverwrites {
		if overwrite.ID == ticket.DiscordID && overwrite.Type == discordgo.PermissionOverwriteTypeMember {
			// Update existing owner permission
			overwrite.Allow = discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory
			hasOwnerPermission = true
			break
		}
	}

	// If owner permission doesn't exist, add it
	if !hasOwnerPermission {
		permissionOverwrites = append(permissionOverwrites, &discordgo.PermissionOverwrite{
			ID:    ticket.DiscordID,
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
		})
	}

	// Update the channel with synced permissions and correct category
	channelEdit := &discordgo.ChannelEdit{
		PermissionOverwrites: permissionOverwrites,
	}

	moved := false
	// If the channel is not in the correct category, move it
	if channel.ParentID != targetCategoryID {
		channelEdit.ParentID = targetCategoryID
		moved = true
		log.Printf("Moving ticket %s from category %s to %s", ticket.ChannelID, channel.ParentID, targetCategoryID)
	}

	_, err = ctx.Session().ChannelEdit(ticket.ChannelID, channelEdit)
	if err != nil {
		return false, fmt.Errorf("failed to sync permissions: %w", err)
	}

	return moved, nil
}
