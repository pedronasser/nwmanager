package register

import (
	"context"
	"fmt"
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"nwmanager/helpers"
	"nwmanager/types"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var handlers = map[string]func(ctx *common.ModuleContext, i *discordgo.InteractionCreate){
	"btn:start_registration":   startRegistration,
	"select:pvp_classes":       handleClassSelection,
	"select:times":             handleTimeSelection,
	"select:weekdays":          handleWeekdaySelection,
	"btn:war_yes":              handleWarExperienceYes,
	"btn:war_no":               handleWarExperienceNo,
	"modal:guild_name":         handleGuildNameModal,
	"btn:approve_registration": approveRegistration,
	"btn:reject_registration":  rejectRegistration,
}

func setupWelcomeChannel(ctx *common.ModuleContext, channelID string) error {
	cfg := GetModuleConfig(ctx)
	dg := ctx.Session()

	// Check if welcome message already exists
	messages, err := dg.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		log.Printf("Error fetching channel messages: %v", err)
		// Continue with setup if we can't check existing messages
	} else {
		// Look for existing registration button message
		// Get bot user ID safely
		var botUserID string
		if dg.State != nil && dg.State.User != nil {
			botUserID = dg.State.User.ID
		}

		for _, msg := range messages {
			// Check if message is from bot (if we have bot ID) and has components
			if (botUserID == "" || msg.Author.ID == botUserID) && len(msg.Components) > 0 {
				// Check if this message has a registration button
				for _, component := range msg.Components {
					if actionRow, ok := component.(*discordgo.ActionsRow); ok {
						for _, comp := range actionRow.Components {
							if button, ok := comp.(*discordgo.Button); ok && button.CustomID == "btn:start_registration" {
								log.Printf("Welcome message already exists in channel %s", channelID)
								return nil // Message already exists, don't send another
							}
						}
					}
				}
			}
		}
	}

	// Send welcome message with registration button
	embed := &discordgo.MessageEmbed{
		Title:       "🎮 Registrar na Guild",
		Description: cfg.WelcomeMessage,
		Color:       0x00ff00,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Iniciar Recrutamento",
					Style:    discordgo.PrimaryButton,
					CustomID: "btn:start_registration",
					Emoji: &discordgo.ComponentEmoji{
						Name: "📝",
					},
				},
			},
		},
	}

	_, err = dg.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})

	return err
}

// cleanupExistingRegistrationChannels finds and deletes any existing channels with the same registration name
func cleanupExistingRegistrationChannels(dg *discordgo.Session, guildID, channelName, categoryID string) error {
	// Get all channels in the guild
	channels, err := dg.GuildChannels(guildID)
	if err != nil {
		return fmt.Errorf("error fetching guild channels: %v", err)
	}

	// Look for channels with the same name
	for _, channel := range channels {
		// Only check text channels
		if channel.Type != discordgo.ChannelTypeGuildText {
			continue
		}

		// If a category is specified, only check channels in that category
		if categoryID != "" && channel.ParentID != categoryID {
			continue
		}

		// If channel name matches, delete it
		if channel.Name == channelName {
			log.Printf("Found existing registration channel '%s' (ID: %s), deleting...", channel.Name, channel.ID)

			// Clean up any registration state that might reference this channel
			cleanupRegistrationStateByChannelID(channel.ID)

			_, err := dg.ChannelDelete(channel.ID)
			if err != nil {
				log.Printf("Error deleting existing registration channel %s: %v", channel.ID, err)
				// Continue trying to delete other channels even if one fails
			} else {
				log.Printf("Successfully deleted existing registration channel '%s'", channel.Name)
			}
		}
	}

	return nil
}

// cleanupRegistrationStateByChannelID removes any registration state that references the given channel ID
func cleanupRegistrationStateByChannelID(channelID string) {
	for userID, state := range RegisterData {
		if state.TopicID == channelID {
			log.Printf("Cleaning up orphaned registration state for user %s (channel: %s)", userID, channelID)
			delete(RegisterData, userID)
		}
	}
}

func handleWelcomeChannelMessage(ctx *common.ModuleContext, s *discordgo.Session, m *discordgo.MessageCreate) {
	cfg := GetModuleConfig(ctx)

	// Only handle messages in the welcome channel
	if cfg.WelcomeChannelID == "" || m.ChannelID != cfg.WelcomeChannelID {
		return
	}

	// Don't delete bot messages
	if m.Author.Bot {
		return
	}

	// Delete the message
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		log.Printf("Error deleting message in welcome channel: %v", err)
	} else {
		log.Printf("Deleted message from user %s in welcome channel", m.Author.Username)
	}
}

func startRegistration(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	cfg := GetModuleConfig(ctx)
	dg := ctx.Session()

	// Check if user already has a registration in progress
	if _, exists := RegisterData[i.Member.User.ID]; exists {
		discordutils.ReplyEphemeralMessage(dg, i, "❌ Você já possui um registro em andamento!", 5*time.Second)
		return
	}

	// Check if user already has the member role (is already fully registered)
	globalCfg, _ := ctx.Config("globals").(*globals.GlobalsConfig)
	// if globalCfg.MemberRoleID != "" && discordutils.HasRole(i.Member, globalCfg.MemberRoleID) {
	// 	discordutils.ReplyEphemeralMessage(dg, i, "✅ Você já está registrado na guild!", 5*time.Second)
	// 	return
	// }

	// Create private text channel for registration
	channelName := fmt.Sprintf("registro-%s", strings.ToLower(i.Member.User.Username))

	// Check for existing registration channels with the same name and delete them
	err := cleanupExistingRegistrationChannels(dg, i.GuildID, channelName, cfg.RegistrationCategoryID)
	if err != nil {
		log.Printf("Error cleaning up existing registration channels: %v", err)
		// Continue with creation despite cleanup errors
	}

	// Determine parent category
	var parentID string
	if cfg.RegistrationCategoryID != "" {
		parentID = cfg.RegistrationCategoryID
	}

	// Create channel permissions - only the user and bot can see it
	permissionOverwrites := []*discordgo.PermissionOverwrite{
		{
			ID:   i.GuildID, // @everyone role
			Type: discordgo.PermissionOverwriteTypeRole,
			Deny: discordgo.PermissionViewChannel,
		},
		{
			ID:    i.Member.User.ID, // The registering user
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
		},
	}

	// Add bot permissions if we can get the bot user ID
	if dg.State != nil && dg.State.User != nil {
		permissionOverwrites = append(permissionOverwrites, &discordgo.PermissionOverwrite{
			ID:    dg.State.User.ID, // Bot user
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory | discordgo.PermissionManageMessages,
		})
	}

	// Add admin role permissions if configured
	if globalCfg.AdminRoleID != "" {
		permissionOverwrites = append(permissionOverwrites, &discordgo.PermissionOverwrite{
			ID:    globalCfg.AdminRoleID, // Admin role
			Type:  discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
		})
	}

	channel, err := dg.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
		Name:                 channelName,
		Type:                 discordgo.ChannelTypeGuildText,
		ParentID:             parentID,
		PermissionOverwrites: permissionOverwrites,
	})
	if err != nil {
		log.Printf("Error creating registration channel: %v", err)
		discordutils.ReplyEphemeralMessage(dg, i, "❌ Erro ao criar canal de registro.", 5*time.Second)
		return
	}

	// Initialize registration state
	RegisterData[i.Member.User.ID] = &RegistrationState{
		DiscordID: i.Member.User.ID,
		TopicID:   channel.ID,
		StepIndex: 0, // Start with first step (0-based index)
	}

	// Start the first step using the new step processor
	processor := GetStepProcessor()
	err = processor.ProcessStep(ctx, RegisterData[i.Member.User.ID], 0)
	if err != nil {
		log.Printf("Error processing first step: %v", err)
		discordutils.ReplyEphemeralMessage(dg, i, "❌ Erro ao iniciar registro.", 5*time.Second)
		return
	}

	// Reply to original interaction
	discordutils.ReplyEphemeralMessage(dg, i, fmt.Sprintf("✅ Canal de registro criado: <#%s>", channel.ID), 10*time.Second)
}

func handleClassSelection(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	state, exists := RegisterData[i.Member.User.ID]
	if !exists {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Registro não encontrado.", 5*time.Second)
		return
	}

	// Find the current step by handler type instead of hardcoded index
	processor := GetStepProcessor()
	currentStep := processor.GetStepByIndex(state.StepIndex)
	if currentStep == nil || currentStep.Handler == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Passo inválido.", 5*time.Second)
		return
	}

	// Reply to interaction first
	discordutils.ReplyEphemeralMessage(ctx.Session(), i, "✅ Classes selecionadas com sucesso!", 1*time.Second)

	// Handle using the new step system
	err := processor.HandleStepResponse(ctx, state, state.StepIndex, i)
	if err != nil {
		log.Printf("Error handling PVP classes step: %v", err)
	}
}

func handleTimeSelection(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	state, exists := RegisterData[i.Member.User.ID]
	if !exists {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Registro não encontrado.", 5*time.Second)
		return
	}

	// Find the current step by handler type instead of hardcoded index
	processor := GetStepProcessor()
	currentStep := processor.GetStepByIndex(state.StepIndex)
	if currentStep == nil || currentStep.Handler == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Passo inválido.", 5*time.Second)
		return
	}

	// Reply to interaction first
	go discordutils.ReplyEphemeralMessage(ctx.Session(), i, "✅ Horários selecionados com sucesso!", 1*time.Second)

	// Handle using the new step system
	err := processor.HandleStepResponse(ctx, state, state.StepIndex, i)
	if err != nil {
		log.Printf("Error handling times step: %v", err)
	}
}

func handleWeekdaySelection(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	state, exists := RegisterData[i.Member.User.ID]
	if !exists {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Registro não encontrado.", 5*time.Second)
		return
	}

	// Find the current step by handler type instead of hardcoded index
	processor := GetStepProcessor()
	currentStep := processor.GetStepByIndex(state.StepIndex)
	if currentStep == nil || currentStep.Handler == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Passo inválido.", 5*time.Second)
		return
	}

	// Reply to interaction first
	discordutils.ReplyEphemeralMessage(ctx.Session(), i, "✅ Dias selecionados com sucesso!", 1*time.Second)

	// Handle using the new step system
	err := processor.HandleStepResponse(ctx, state, state.StepIndex, i)
	if err != nil {
		log.Printf("Error handling weekdays step: %v", err)
	}
}

func handleWarExperienceYes(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	state, exists := RegisterData[i.Member.User.ID]
	if !exists {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Registro não encontrado.", 5*time.Second)
		return
	}

	// Find the current step by handler type instead of hardcoded index
	processor := GetStepProcessor()
	currentStep := processor.GetStepByIndex(state.StepIndex)
	if currentStep == nil || currentStep.Handler == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Passo inválido.", 5*time.Second)
		return
	}

	// Handle using the new step system
	err := processor.HandleStepResponse(ctx, state, state.StepIndex, i)
	if err != nil {
		log.Printf("Error handling war experience step: %v", err)
	}
}

func handleWarExperienceNo(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	state, exists := RegisterData[i.Member.User.ID]
	if !exists {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Registro não encontrado.", 5*time.Second)
		return
	}

	// Find the current step by handler type instead of hardcoded index
	processor := GetStepProcessor()
	currentStep := processor.GetStepByIndex(state.StepIndex)
	if currentStep == nil || currentStep.Handler == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Passo inválido.", 5*time.Second)
		return
	}

	// Reply to interaction first
	discordutils.ReplyEphemeralMessage(ctx.Session(), i, "✅ Resposta registrada!", 1*time.Second)

	// Handle using the new step system
	err := processor.HandleStepResponse(ctx, state, state.StepIndex, i)
	if err != nil {
		log.Printf("Error handling war experience step: %v", err)
	}
}

func handleGuildNameModal(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	state, exists := RegisterData[i.Member.User.ID]
	if !exists {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Registro não encontrado.", 5*time.Second)
		return
	}

	// Extract guild name from modal
	data := i.ModalSubmitData()
	var guildName string
	for _, component := range data.Components {
		if actionRow, ok := component.(*discordgo.ActionsRow); ok {
			for _, comp := range actionRow.Components {
				if textInput, ok := comp.(*discordgo.TextInput); ok && textInput.CustomID == "guild_name_input" {
					guildName = strings.TrimSpace(textInput.Value)
					break
				}
			}
		}
	}

	if guildName == "" {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Nome da guild não pode estar vazio.", 5*time.Second)
		return
	}

	// Store guild name and complete registration
	state.PreviousGuildName = guildName

	// Reply to interaction first
	discordutils.ReplyEphemeralMessage(ctx.Session(), i, "✅ Informações registradas com sucesso!", 1*time.Second)

	// Complete registration
	completeRegistration(ctx, state, i)
}

func completeRegistration(ctx *common.ModuleContext, state *RegistrationState, i *discordgo.InteractionCreate) {
	dg := ctx.Session()

	// Create registration record in database
	registration := &types.Register{
		ID:         primitive.NewObjectID(),
		DiscordID:  state.DiscordID,
		InGameName: state.IGN,
		WeekDays:   state.Weekdays,   // Can be nil/empty if step was skipped
		Hours:      state.Times,      // Can be nil/empty if step was skipped
		PVPClasses: state.PVPClasses, // Using weapons field for PVP classes
		CreatedAt:  time.Now(),
	}

	// Store in database (implement this based on your database layer)
	err := insertRegistration(ctx, registration)
	if err != nil {
		log.Printf("Error storing registration: %v", err)
		discordutils.ReplyEphemeralMessage(dg, i, "❌ Erro ao salvar registro. Tente novamente.", 5*time.Second)
		return
	}

	// Reply to interaction first
	discordutils.ReplyEphemeralMessage(dg, i, "✅ Dias selecionados com sucesso!", 1*time.Second)

	// Send completion message with admin approval buttons
	sendCompletionMessage(ctx, state, registration)

	// Clean up registration state
	delete(RegisterData, state.DiscordID)
}

func sendCompletionMessage(ctx *common.ModuleContext, state *RegistrationState, registration *types.Register) {
	dg := ctx.Session()
	globalCfg, _ := ctx.Config("globals").(*globals.GlobalsConfig)

	// Create readable lists for the summary
	var classNames []string
	for _, class := range state.PVPClasses {
		if name, exists := globals.PVP_CLASS_NAMES[class]; exists {
			if emoji, emojiExists := globalCfg.ClassEmojiIDs[string(class)]; emojiExists {
				classNames = append(classNames, fmt.Sprintf("%s %s", emoji, name))
			} else {
				classNames = append(classNames, name)
			}
		}
	}

	var timeNames []string
	if len(state.Times) > 0 {
		for _, timeKey := range state.Times {
			if name, exists := TIMES[timeKey]; exists {
				timeNames = append(timeNames, name)
			}
		}
	}

	var weekdayNames []string
	if len(state.Weekdays) > 0 {
		for _, weekdayKey := range state.Weekdays {
			if name, exists := WEEKDAYS[weekdayKey]; exists {
				weekdayNames = append(weekdayNames, name)
			}
		}
	}

	// Check if this is a re-registration
	existingPlayer, _ := types.GetPlayerByDiscordID(context.Background(), ctx.DB(), state.DiscordID)
	isReRegistration := existingPlayer != nil

	description := "Seu registro foi enviado com sucesso! Nossa equipe irá revisar suas informações."
	if isReRegistration {
		description = "Seu novo registro foi enviado com sucesso! Este registro substituirá seus dados anteriores após aprovação."
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📋 Registro Completo - Aguardando Aprovação",
		Description: description,
		Color:       0xffaa00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🎮 Nome no Jogo",
				Value:  state.IGN,
				Inline: true,
			},
			{
				Name:   "⚔️ Classes de PvP",
				Value:  strings.Join(classNames, ", "),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Aguarde a aprovação de um administrador",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Only add times field if times were collected
	if len(timeNames) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⏰ Horários",
			Value:  strings.Join(timeNames, ", "),
			Inline: false,
		})
	}

	// Only add weekdays field if weekdays were collected
	if len(weekdayNames) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📅 Dias da Semana",
			Value:  strings.Join(weekdayNames, ", "),
			Inline: false,
		})
	}

	// Add war experience information
	warExperienceValue := "Não participou de guerras"
	if state.HasWarExperience {
		if state.PreviousGuildName != "" {
			warExperienceValue = fmt.Sprintf("Sim - Guild: %s", state.PreviousGuildName)
		} else {
			warExperienceValue = "Sim"
		}
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "⚔️ Experiência em Guerras",
		Value:  warExperienceValue,
		Inline: false,
	})

	if isReRegistration {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🔄 Re-registro",
			Value:  "Este registro substituirá os dados anteriores",
			Inline: false,
		})
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "✅ Aprovar",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("btn:approve_registration:%s", registration.ID.Hex()),
				},
				discordgo.Button{
					Label:    "❌ Rejeitar",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("btn:reject_registration:%s", registration.ID.Hex()),
				},
			},
		},
	}

	_, err := dg.ChannelMessageSendComplex(state.TopicID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		log.Printf("Error sending completion message: %v", err)
	}
}

func approveRegistration(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	globalCfg := ctx.Config("globals").(*globals.GlobalsConfig)

	// Extract registration ID from custom ID
	parts := strings.Split(i.MessageComponentData().CustomID, ":")
	if len(parts) != 3 {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ ID de registro inválido.", 5*time.Second)
		return
	}

	registrationID := parts[2]

	// Check if user has admin permissions
	if !hasAdminPermission(i.Member, globalCfg.AdminRoleID) {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Você não tem permissão para aprovar registros.", 5*time.Second)
		return
	}

	// Get registration from database and create player
	err := processApproval(ctx, registrationID, i.Member.User.ID, i.GuildID)
	if err != nil {
		log.Printf("Error processing approval: %v", err)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Erro ao processar aprovação.", 5*time.Second)
		return
	}

	// Update the message to show approval
	embed := &discordgo.MessageEmbed{
		Title:       "✅ Registro Aprovado",
		Description: fmt.Sprintf("Registro aprovado por <@%s>", i.Member.User.ID),
		Color:       0x00ff00,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	_, err = ctx.Session().ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    i.ChannelID,
		ID:         i.Message.ID,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &[]discordgo.MessageComponent{}, // Remove buttons
	})
	if err != nil {
		log.Printf("Error updating approval message: %v", err)
	}

	responseErr := ctx.Session().InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ Registro aprovado com sucesso!",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if responseErr != nil {
		log.Printf("Error responding to approval: %v", responseErr)
	}

	// Delete the registration channel after a delay
	go func() {
		time.Sleep(10 * time.Second)
		_, err := ctx.Session().ChannelDelete(i.ChannelID)
		if err != nil {
			log.Printf("Error deleting registration channel: %v", err)
		} else {
			log.Printf("Successfully deleted registration channel %s", i.ChannelID)
		}
	}()
}

func rejectRegistration(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	globalCfg := ctx.Config("globals").(*globals.GlobalsConfig)

	// Extract registration ID from custom ID
	parts := strings.Split(i.MessageComponentData().CustomID, ":")
	if len(parts) != 3 {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ ID de registro inválido.", 5*time.Second)
		return
	}

	registrationID := parts[2]

	// Check if user has admin permissions
	if !hasAdminPermission(i.Member, globalCfg.AdminRoleID) {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Você não tem permissão para rejeitar registros.", 5*time.Second)
		return
	}

	// Mark registration as rejected
	err := processRejection(ctx, registrationID, i.Member.User.ID)
	if err != nil {
		log.Printf("Error processing rejection: %v", err)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Erro ao processar rejeição.", 5*time.Second)
		return
	}

	// Update the message to show rejection
	embed := &discordgo.MessageEmbed{
		Title:       "❌ Seu registro não foi aprovado pela equipe",
		Description: fmt.Sprintf("Registro rejeitado por <@%s>", i.Member.User.ID),
		Color:       0xff0000,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	_, err = ctx.Session().ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    i.ChannelID,
		ID:         i.Message.ID,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &[]discordgo.MessageComponent{}, // Remove buttons
	})
	if err != nil {
		log.Printf("Error updating rejection message: %v", err)
	}

	responseErr := ctx.Session().InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "❌ Registro rejeitado.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if responseErr != nil {
		log.Printf("Error responding to rejection: %v", responseErr)
	}

	// Delete the registration channel after a delay
	go func() {
		time.Sleep(10 * time.Second)
		_, err := ctx.Session().ChannelDelete(i.ChannelID)
		if err != nil {
			log.Printf("Error deleting registration channel: %v", err)
		} else {
			log.Printf("Successfully deleted registration channel %s", i.ChannelID)
		}
	}()
}

// Helper functions
func GetModuleConfig(ctx *common.ModuleContext) *RegisterConfig {
	config, _ := ctx.Config(ModuleName).(*RegisterConfig)
	return config
}

func hasAdminPermission(member *discordgo.Member, adminRoleID string) bool {
	return discordutils.HasRole(member, adminRoleID)
}

// Remove the duplicate hasRole function since we're now using discordutils.HasRole

// These functions would need to be implemented based on your database layer
func insertRegistration(ctx *common.ModuleContext, registration *types.Register) error {
	return types.InsertRegister(context.Background(), ctx.DB(), registration)
}

func processApproval(ctx *common.ModuleContext, registrationID, approverID, guildID string) error {
	// Get registration from database
	registration, err := types.GetRegisterByID(context.Background(), ctx.DB(), registrationID)
	if err != nil {
		return fmt.Errorf("error getting registration: %v", err)
	}
	if registration == nil {
		return fmt.Errorf("registration not found")
	}

	// Check if player already exists
	existingPlayer, err := types.GetPlayerByDiscordID(context.Background(), ctx.DB(), registration.DiscordID)
	if err != nil {
		log.Printf("Error checking existing player: %v", err)
		// Continue with creation if we can't check
	}

	// Create new player record
	now := helpers.GetCurrentTimeAsUTC()

	// Ensure empty slices are not nil for database consistency
	availableTimes := registration.Hours
	if availableTimes == nil {
		availableTimes = []string{}
	}

	availableWeekdays := registration.WeekDays
	if availableWeekdays == nil {
		availableWeekdays = []string{}
	}

	player := &types.Player{
		ID:                primitive.NewObjectID(),
		DiscordID:         registration.DiscordID,
		IGN:               registration.InGameName,
		PVPClasses:        registration.PVPClasses,
		AvailableTimes:    availableTimes,
		AvailableWeekdays: availableWeekdays,
		WarClass:          string(registration.PVPClasses[0]), // Assuming first class is the war class
		RegisteredAt:      &now,
		Stats:             &types.PlayerStats{},
	}

	// If player exists, delete the old record first
	if existingPlayer != nil {
		log.Printf("Existing player found for Discord ID %s, removing old record", registration.DiscordID)
		err = types.DeletePlayer(context.Background(), ctx.DB(), existingPlayer)
		if err != nil {
			log.Printf("Error deleting existing player: %v", err)
			// Continue anyway - we'll create the new record
		}
	}

	// Insert new player record
	err = types.InsertPlayer(context.Background(), ctx.DB(), player)
	if err != nil {
		return fmt.Errorf("error creating player: %v", err)
	}

	// Assign member role
	globalCfg, _ := ctx.Config("globals").(*globals.GlobalsConfig)
	if globalCfg.MemberRoleID != "" && guildID != "" {
		dg := ctx.Session()
		err = dg.GuildMemberRoleAdd(guildID, registration.DiscordID, globalCfg.MemberRoleID)
		if err != nil {
			log.Printf("Error assigning member role to user %s: %v", registration.DiscordID, err)
		} else {
			log.Printf("Successfully assigned member role to user %s", registration.DiscordID)
		}
	}

	// Change member nickname to their IGN
	if guildID != "" {
		dg := ctx.Session()
		err = dg.GuildMemberNickname(guildID, registration.DiscordID, registration.InGameName)
		if err != nil {
			log.Printf("Error changing nickname for user %s to %s: %v", registration.DiscordID, registration.InGameName, err)
		} else {
			log.Printf("Successfully changed nickname for user %s to %s", registration.DiscordID, registration.InGameName)
		}
	}

	// Mark registration as approved
	return types.ApproveRegister(context.Background(), ctx.DB(), registrationID, approverID)
}

func processRejection(ctx *common.ModuleContext, registrationID, rejecterID string) error {
	return types.RejectRegister(context.Background(), ctx.DB(), registrationID, rejecterID)
}

// HandleRegistrationAction creates the main interaction handler
func HandleRegistrationAction(ctx *common.ModuleContext, guildID string) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.GuildID != guildID {
			return
		}

		var handlerKey string
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			handlerKey = "/" + i.ApplicationCommandData().Name
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
		}

		if handler, exists := handlers[handlerKey]; exists {
			handler(ctx, i)
		}
	}
}

// HandleWelcomeChannelMessages creates a message handler for the welcome channel
func HandleWelcomeChannelMessages(ctx *common.ModuleContext) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		handleWelcomeChannelMessage(ctx, s, m)
	}
}
