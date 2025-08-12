package ticket

import (
	"fmt"
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var handlers = map[string]func(ctx *common.ModuleContext, i *discordgo.InteractionCreate){
	"ticket:send_build":        handleSendBuild,
	"ticket:send_question":     handleSendQuestion,
	"ticket:change_class":      handleChangeClass,
	"ticket:view_build":        handleViewBuild,
	"ticket:close_thread":      handleCloseThread,
	"ticket:submit_build":      handleSubmitBuild,
	"ticket:notify_absence":    handleNotifyAbsence,
	"modal:absence_form":       handleAbsenceModal,
	"select:class_selection":   handleClassSelection,
	"/ausencia":                handleNotifyAbsence, // Slash command uses same handler as button
	"/sync-ticket-permissions": handleSyncTicketPermissions,
}

func handleSendBuild(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	// Get player data from database
	player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), i.Member.User.ID)
	if err != nil || player == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao encontrar dados do jogador.", 5*time.Second)
		return
	}

	// Create thread for build submission
	threadName := fmt.Sprintf("build-%s", player.IGN)
	thread, err := ctx.Session().MessageThreadStartComplex(i.ChannelID, i.Message.ID, &discordgo.ThreadStart{
		Name:                threadName,
		AutoArchiveDuration: 1440, // 24 hours auto-archive timeout
		Type:                discordgo.ChannelTypeGuildPrivateThread,
	})
	if err != nil {
		log.Printf("Error creating build thread: %v", err)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao criar thread para build.", 5*time.Second)
		return
	}

	// Send instructions and confirm button in thread
	setupBuildThread(ctx, thread.ID, player.IGN)
	discordutils.ReplyEphemeralMessage(ctx.Session(), i, fmt.Sprintf("Thread criada: <#%s>", thread.ID), 5*time.Second)
}

func handleNotifyAbsence(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	// Create modal for absence notification
	modal := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "absence_date",
					Label:       "Data da Ausência (DD/MM/AAAA)",
					Style:       discordgo.TextInputShort,
					Placeholder: "Ex: 15/08/2025",
					Required:    true,
					MaxLength:   10,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "absence_reason",
					Label:       "Motivo (Opcional)",
					Style:       discordgo.TextInputParagraph,
					Placeholder: "Descreva brevemente o motivo da ausência...",
					Required:    false,
					MaxLength:   500,
				},
			},
		},
	}

	err := discordutils.SendModal(ctx.Session(), i, "modal:absence_form", "Avisar Ausência", modal...)
	if err != nil {
		log.Printf("Error sending absence modal: %v", err)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao abrir formulário de ausência.", 5*time.Second)
	}
}

func handleAbsenceModal(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	// Get form data
	data := i.ModalSubmitData()
	var absenceDate, absenceReason string

	for _, component := range data.Components {
		if actionRow, ok := component.(*discordgo.ActionsRow); ok {
			for _, comp := range actionRow.Components {
				if textInput, ok := comp.(*discordgo.TextInput); ok {
					switch textInput.CustomID {
					case "absence_date":
						absenceDate = textInput.Value
					case "absence_reason":
						absenceReason = textInput.Value
					}
				}
			}
		}
	}

	// Get player data
	player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), i.Member.User.ID)
	if err != nil || player == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao encontrar dados do jogador.", 5*time.Second)
		return
	}

	// Get config for absence channel
	ticketConfig := GetModuleConfig(ctx)
	if ticketConfig.AbsenceChannelID == "" {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Canal de ausências não configurado.", 5*time.Second)
		return
	}

	// Create absence notification embed
	embed := &discordgo.MessageEmbed{
		Title:     "📅 Notificação de Ausência",
		Color:     0xffa500, // Orange color
		Timestamp: time.Now().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Jogador",
				Value:  fmt.Sprintf("%s (<@%s>)", player.IGN, player.DiscordID),
				Inline: true,
			},
			{
				Name:   "Data da Ausência",
				Value:  absenceDate,
				Inline: true,
			},
		},
	}

	// Add reason field if provided
	if absenceReason != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Motivo",
			Value:  absenceReason,
			Inline: false,
		})
	}

	// Send notification to absence channel
	_, err = ctx.Session().ChannelMessageSendEmbed(ticketConfig.AbsenceChannelID, embed)
	if err != nil {
		log.Printf("Error sending absence notification: %v", err)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao enviar notificação de ausência.", 5*time.Second)
		return
	}

	// Respond to user
	discordutils.ReplyEphemeralMessage(ctx.Session(), i, "✅ Ausência notificada com sucesso!", 0)
}

func handleSendQuestion(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), i.Member.User.ID)
	if err != nil || player == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao encontrar dados do jogador.", 5*time.Second)
		return
	}

	threadName := fmt.Sprintf("duvida-%s", player.IGN)
	thread, err := ctx.Session().MessageThreadStartComplex(i.ChannelID, i.Message.ID, &discordgo.ThreadStart{
		Name:                threadName,
		AutoArchiveDuration: 0,
		Type:                discordgo.ChannelTypeGuildPrivateThread,
	})
	if err != nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao criar thread para dúvida.", 5*time.Second)
		return
	}

	setupQuestionThread(ctx, thread.ID, player.IGN)
	discordutils.ReplyEphemeralMessage(ctx.Session(), i, fmt.Sprintf("Thread criada: <#%s>", thread.ID), 5*time.Second)
}

func handleChangeClass(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

	// Check if user is admin or the ticket owner
	isAdmin := discordutils.HasRole(i.Member, globalConfig.AdminRoleID)
	isOwner := isTicketOwner(ctx, i.ChannelID, i.Member.User.ID)

	if !isAdmin && !isOwner {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você não tem permissão para usar este comando.", 5*time.Second)
		return
	}

	// Create class selection dropdown using global class emojis
	classOptions := createClassSelectOptions(globalConfig.ClassEmojiIDs)

	discordutils.SendInteractiveMessage(ctx.Session(), i, "select:class_selection", "Selecione a nova classe:",
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "select:class_selection",
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: "Escolha uma classe...",
					MinValues:   &[]int{1}[0],
					MaxValues:   1,
					Options:     classOptions,
				},
			},
		},
	)
}

func handleClassSelection(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

	selectedClass := i.MessageComponentData().Values[0]

	// Get player and update their class
	player, err := getPlayerByTicketChannel(ctx, i.ChannelID)
	if err != nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao encontrar dados do jogador.", 5*time.Second)
		return
	}

	// Check if the player is trying to select the same class they already have
	if player.WarClass == selectedClass {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Você já está na classe selecionada!", 5*time.Second)
		return
	}

	// Store the old class before updating the player
	oldClass := player.WarClass

	// Update player class in database
	err = updatePlayerClass(ctx, player, selectedClass)
	if err != nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao atualizar classe do jogador.", 5*time.Second)
		return
	}

	// Get class emoji
	classEmoji := globalConfig.ClassEmojiIDs[selectedClass]

	// FIRST: Respond by updating the original message (removes dropdown, shows success)
	err = ctx.Session().InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    fmt.Sprintf("✅ Classe alterada para %s %s!", classEmoji, selectedClass),
			Components: []discordgo.MessageComponent{}, // Remove the dropdown
		},
	})
	if err != nil {
		log.Printf("Error updating interaction message: %v", err)
		return
	}

	// Update Discord nickname
	newNickname := fmt.Sprintf("%s %s", classEmoji, player.IGN)
	err = ctx.Session().GuildMemberNickname(globalConfig.GuildID, player.DiscordID, newNickname)
	if err != nil {
		log.Printf("Error updating nickname: %v", err)
	}

	// Add PVP class role if it exists in the configuration
	if classRoleID, exists := globalConfig.ClassRoleIDs[selectedClass]; exists {
		// Remove old class roles first (if player had a different class before)
		if oldClass != "" && oldClass != selectedClass {
			if oldRoleID, oldExists := globalConfig.ClassRoleIDs[oldClass]; oldExists {
				err = ctx.Session().GuildMemberRoleRemove(globalConfig.GuildID, player.DiscordID, oldRoleID)
				if err != nil {
					log.Printf("Error removing old class role %s: %v", oldRoleID, err)
				} else {
					log.Printf("Successfully removed old class role %s from player %s", oldRoleID, player.IGN)
				}
			}
		}

		// Add new class role
		err = ctx.Session().GuildMemberRoleAdd(globalConfig.GuildID, player.DiscordID, classRoleID)
		if err != nil {
			log.Printf("Error adding class role %s: %v", classRoleID, err)
		} else {
			log.Printf("Successfully added class role %s to player %s", classRoleID, player.IGN)
		}
	}

	// Move ticket to class-specific category, update name, and sync permissions in a single edit
	channelEdit := &discordgo.ChannelEdit{
		Name: player.IGN,
	}

	if categoryID, exists := globalConfig.ClassCategoryIDs[selectedClass]; exists {
		channelEdit.ParentID = categoryID

		// Get category permissions to sync them in the same edit
		category, err := ctx.Session().Channel(categoryID)
		if err != nil {
			log.Printf("Error getting category %s for permission sync: %v", categoryID, err)
		} else {
			// Include category permissions in the same edit
			channelEdit.PermissionOverwrites = category.PermissionOverwrites
		}
		channelEdit.PermissionOverwrites = append(channelEdit.PermissionOverwrites, &discordgo.PermissionOverwrite{
			ID:    player.DiscordID,
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
		})
	}

	_, err = ctx.Session().ChannelEdit(i.ChannelID, channelEdit)
	if err != nil {
		log.Printf("Error updating channel (name: %s, category: %v): %v", player.IGN, channelEdit.ParentID, err)
	} else if channelEdit.ParentID != "" {
		log.Printf("Successfully moved channel %s to category %s and synced permissions", i.ChannelID, channelEdit.ParentID)
	}

	// Optional: Delete the success message after a delay
	go func() {
		time.Sleep(3 * time.Second)
		err = ctx.Session().InteractionResponseDelete(i.Interaction)
		if err != nil {
			log.Printf("Error deleting interaction response: %v", err)
		}
	}()
}

func handleViewBuild(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), i.Member.User.ID)
	if err != nil || player == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Erro ao encontrar dados do jogador.", 5*time.Second)
		return
	}

	// Get latest build from player's build thread or database
	buildInfo := getCurrentBuildInfo(ctx, player)

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Build Atual - %s", player.IGN),
		Description: buildInfo,
		Color:       0x00ff00,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	// Add first build image as embed image if available
	if len(player.BuildImages) > 0 {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: player.BuildImages[0],
		}
	}

	err = ctx.Session().InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Error responding to view build: %v", err)
	}
}

func handleCloseThread(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	// First, respond to the interaction
	err := ctx.Session().InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "�️ Excluindo thread...",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Error responding to close thread interaction: %v", err)
	}

	// Delete the thread
	_, err = ctx.Session().ChannelDelete(i.ChannelID)
	if err != nil {
		log.Printf("Error deleting thread: %v", err)
		return
	}

	log.Printf("Thread %s deleted successfully", i.ChannelID)
}

func handleSubmitBuild(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	// First, respond to the interaction
	err := ctx.Session().InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ Build enviada com sucesso! Excluindo thread...",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Error responding to submit build interaction: %v", err)
	}

	// Mark build as submitted
	player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), i.Member.User.ID)
	if err != nil || player == nil {
		log.Printf("Error finding player data: %v", err)
		return
	}

	// Collect all images from the thread before deletion
	messages, err := ctx.Session().ChannelMessages(i.ChannelID, 100, "", "", "")
	if err != nil {
		log.Printf("Error fetching thread messages: %v", err)
	} else {
		var allImageURLs []string
		log.Printf("Processing %d messages in thread %s for image collection", len(messages), i.ChannelID)

		// Process messages to collect all image URLs
		for _, message := range messages {
			// Skip bot messages
			if message.Author.Bot {
				continue
			}

			log.Printf("Processing message from %s with %d attachments and %d embeds", message.Author.Username, len(message.Attachments), len(message.Embeds))

			// Check attachments
			for _, attachment := range message.Attachments {
				if isImageAttachment(attachment) {
					log.Printf("Found image attachment: %s", attachment.URL)
					allImageURLs = append(allImageURLs, attachment.URL)
				}
			}

			// Check embeds for images
			for _, embed := range message.Embeds {
				if embed.Image != nil && embed.Image.URL != "" {
					log.Printf("Found embed image: %s", embed.Image.URL)
					allImageURLs = append(allImageURLs, embed.Image.URL)
				}
				if embed.Thumbnail != nil && embed.Thumbnail.URL != "" {
					log.Printf("Found embed thumbnail: %s", embed.Thumbnail.URL)
					allImageURLs = append(allImageURLs, embed.Thumbnail.URL)
				}
			}
		}

		log.Printf("Collected %d total image URLs from thread", len(allImageURLs))

		// Store all collected images and update stats in one operation
		if len(allImageURLs) > 0 {
			log.Printf("Before update - Player %s has %d existing BuildImages", player.IGN, len(player.BuildImages))

			// Update both build images and stats
			player.BuildImages = allImageURLs
			if player.Stats == nil {
				player.Stats = &types.PlayerStats{}
			}
			now := time.Now()
			player.Stats.LastEquipUpdate = &now

			log.Printf("About to update player %s with %d build images", player.IGN, len(player.BuildImages))

			err = types.UpdatePlayer(ctx.Context, ctx.DB(), player)
			if err != nil {
				log.Printf("Error updating player with build images and stats: %v", err)
			} else {
				log.Printf("Successfully updated player %s with %d build images and stats", player.IGN, len(allImageURLs))

				// Verify the update worked by re-fetching the player
				updatedPlayer, verifyErr := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), i.Member.User.ID)
				if verifyErr == nil && updatedPlayer != nil {
					log.Printf("Verification - Player %s now has %d BuildImages in database", updatedPlayer.IGN, len(updatedPlayer.BuildImages))
				} else {
					log.Printf("Error verifying player update: %v", verifyErr)
				}
			}
		} else {
			log.Printf("No images found in thread to store")
			// Still update the stats even if no images
			if player.Stats == nil {
				player.Stats = &types.PlayerStats{}
			}
			now := time.Now()
			player.Stats.LastEquipUpdate = &now

			err = types.UpdatePlayer(ctx.Context, ctx.DB(), player)
			if err != nil {
				log.Printf("Error updating player stats: %v", err)
			}
		}
	}

	// Delete the thread
	_, err = ctx.Session().ChannelDelete(i.ChannelID)
	if err != nil {
		log.Printf("Error deleting thread after build submission: %v", err)
		return
	}

	log.Printf("Thread %s deleted after build submission", i.ChannelID)
}

func handleSyncTicketPermissions(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

	// Check if user is admin
	if !discordutils.HasRole(i.Member, globalConfig.AdminRoleID) {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Você não tem permissão para usar este comando. Apenas administradores podem sincronizar permissões de tickets.", 10*time.Second)
		return
	}

	// Respond immediately to acknowledge the command
	err := ctx.Session().InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🔄 Iniciando sincronização de permissões dos tickets...",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		log.Printf("Error responding to sync command: %v", err)
		return
	}

	// Start the sync process in a goroutine
	go func() {
		log.Printf("Admin %s (%s) initiated ticket permission sync", i.Member.User.Username, i.Member.User.ID)

		// Get ticket count first for progress
		activeTickets, err := getAllActiveTickets(ctx)
		if err != nil {
			followupContent := fmt.Sprintf("❌ Erro ao obter tickets: %v", err)
			ctx.Session().FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: followupContent,
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			return
		}

		if len(activeTickets) == 0 {
			ctx.Session().FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "ℹ️ Nenhum ticket ativo encontrado para sincronizar.",
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			return
		}

		// Send progress update
		ctx.Session().FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("📊 Encontrados %d tickets ativos. Sincronizando...", len(activeTickets)),
			Flags:   discordgo.MessageFlagsEphemeral,
		})

		err = syncAllTicketPermissions(ctx)

		var followupContent string
		if err != nil {
			log.Printf("Error during permission sync: %v", err)
			followupContent = fmt.Sprintf("⚠️ Sincronização concluída com alguns erros: %v", err)
		} else {
			followupContent = fmt.Sprintf("✅ Sincronização concluída com sucesso! %d tickets processados.", len(activeTickets))
		}

		// Send final result
		_, err = ctx.Session().FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: followupContent,
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			log.Printf("Error sending follow-up message: %v", err)
		}
	}()
}

// isImageAttachment checks if an attachment is an image
func isImageAttachment(attachment *discordgo.MessageAttachment) bool {
	contentType := attachment.ContentType
	return strings.HasPrefix(contentType, "image/") ||
		strings.HasSuffix(strings.ToLower(attachment.Filename), ".jpg") ||
		strings.HasSuffix(strings.ToLower(attachment.Filename), ".jpeg") ||
		strings.HasSuffix(strings.ToLower(attachment.Filename), ".png") ||
		strings.HasSuffix(strings.ToLower(attachment.Filename), ".gif") ||
		strings.HasSuffix(strings.ToLower(attachment.Filename), ".webp")
}

// HandleTicketAction creates the main interaction handler
func HandleTicketAction(ctx *common.ModuleContext, guildID string) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.GuildID != guildID {
			return
		}

		var handlerKey string
		switch i.Type {
		case discordgo.InteractionMessageComponent:
			data := i.MessageComponentData()
			if strings.Contains(data.CustomID, ":") {
				parts := strings.Split(data.CustomID, ":")
				if len(parts) >= 2 {
					handlerKey = parts[0] + ":" + parts[1]
				}
			} else {
				handlerKey = data.CustomID
			}
		case discordgo.InteractionModalSubmit:
			handlerKey = i.ModalSubmitData().CustomID
		case discordgo.InteractionApplicationCommand:
			handlerKey = "/" + i.ApplicationCommandData().Name
		}

		if handler, exists := handlers[handlerKey]; exists {
			handler(ctx, i)
		}
	}
}
