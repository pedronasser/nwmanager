package register

import (
	"context"
	"fmt"
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/discordutils"
	"nwmanager/helpers"
	"nwmanager/types"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Registration steps
const (
	STEP_START    = 0
	STEP_IGN      = 1
	STEP_CLASSES  = 2
	STEP_TIMES    = 3
	STEP_WEEKDAYS = 4
	STEP_COMPLETE = 5
)

var handlers = map[string]func(ctx *common.ModuleContext, i *discordgo.InteractionCreate){
	"btn:start_registration":   startRegistration,
	"select:pvp_classes":       handleClassSelection,
	"select:times":             handleTimeSelection,
	"select:weekdays":          handleWeekdaySelection,
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
	if cfg.MemberRoleID != "" && hasRole(i.Member, cfg.MemberRoleID) {
		discordutils.ReplyEphemeralMessage(dg, i, "✅ Você já está registrado na guild!", 5*time.Second)
		return
	}

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
	if cfg.AdminRoleID != "" {
		permissionOverwrites = append(permissionOverwrites, &discordgo.PermissionOverwrite{
			ID:    cfg.AdminRoleID, // Admin role
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
		Step:      STEP_IGN,
	}

	// Send first step message
	askForIGN(ctx, channel.ID, i.Member.User.ID)

	// Reply to original interaction
	discordutils.ReplyEphemeralMessage(dg, i, fmt.Sprintf("✅ Canal de registro criado: <#%s>", channel.ID), 10*time.Second)
}

func askForIGN(ctx *common.ModuleContext, channelID, userID string) {
	dg := ctx.Session()

	embed := &discordgo.MessageEmbed{
		Title:       "📝 Registro - Passo 1/4",
		Description: "**Qual é o seu nome no jogo (IGN)?**\n\nPor favor, digite seu nome exatamente como aparece no New World.",
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Digite sua resposta na próxima mensagem",
		},
	}

	_, err := dg.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Printf("Error sending IGN question: %v", err)
	}
}

func handleClassSelection(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	state, exists := RegisterData[i.Member.User.ID]
	if !exists || state.Step != STEP_CLASSES {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Registro não encontrado ou passo inválido.", 5*time.Second)
		return
	}

	// Store selected classes
	state.PVPClasses = i.MessageComponentData().Values
	state.Step = STEP_TIMES

	// Reply to interaction first
	discordutils.ReplyEphemeralMessage(ctx.Session(), i, "✅ Classes selecionadas com sucesso!", 1*time.Second)

	// Then ask for available times
	askForTimes(ctx, state.TopicID, i.Member.User.ID)
}

func askForTimes(ctx *common.ModuleContext, channelID, userID string) {
	dg := ctx.Session()

	// Create time options from constants using ordered array
	var timeOptions []discordgo.SelectMenuOption
	for _, timeKey := range TIME_OPTIONS {
		timeName := TIMES[timeKey]
		timeOptions = append(timeOptions, discordgo.SelectMenuOption{
			Label: timeName,
			Value: timeKey,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⏰ Registro - Passo 3/4",
		Description: "**Quais horários você costuma jogar?**\n\nSelecione todos os horários em que você geralmente está disponível.",
		Color:       0x0099ff,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "select:times",
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: "Selecione seus horários disponíveis",
					MinValues:   &[]int{1}[0],
					MaxValues:   len(timeOptions),
					Options:     timeOptions,
				},
			},
		},
	}

	_, err := dg.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		log.Printf("Error sending times question: %v", err)
	}
}

func handleTimeSelection(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	state, exists := RegisterData[i.Member.User.ID]
	if !exists || state.Step != STEP_TIMES {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Registro não encontrado ou passo inválido.", 5*time.Second)
		return
	}

	// Store selected times
	state.Times = i.MessageComponentData().Values
	state.Step = STEP_WEEKDAYS

	// Reply to interaction first
	go discordutils.ReplyEphemeralMessage(ctx.Session(), i, "✅ Horários selecionados com sucesso!", 1*time.Second)

	// Then ask for weekdays
	askForWeekdays(ctx, state.TopicID, i.Member.User.ID)
}

func askForWeekdays(ctx *common.ModuleContext, channelID, userID string) {
	dg := ctx.Session()

	// Create weekday options from constants using ordered array
	var weekdayOptions []discordgo.SelectMenuOption
	for _, weekdayKey := range WEEKDAY_OPTIONS {
		weekdayName := WEEKDAYS[weekdayKey]
		weekdayOptions = append(weekdayOptions, discordgo.SelectMenuOption{
			Label: weekdayName,
			Value: weekdayKey,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       "📅 Registro - Passo 4/4",
		Description: "**Quais dias da semana você costuma jogar?**\n\nSelecione todos os dias em que você geralmente joga.",
		Color:       0x0099ff,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "select:weekdays",
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: "Selecione os dias da semana",
					MinValues:   &[]int{1}[0],
					MaxValues:   len(weekdayOptions),
					Options:     weekdayOptions,
				},
			},
		},
	}

	_, err := dg.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		log.Printf("Error sending weekdays question: %v", err)
	}
}

func handleWeekdaySelection(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	state, exists := RegisterData[i.Member.User.ID]
	if !exists || state.Step != STEP_WEEKDAYS {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Registro não encontrado ou passo inválido.", 5*time.Second)
		return
	}

	// Store selected weekdays
	state.Weekdays = i.MessageComponentData().Values
	state.Step = STEP_COMPLETE

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
		WeekDays:   state.Weekdays,
		Hours:      state.Times,
		Weapons:    state.PVPClasses, // Using weapons field for PVP classes
		CreatedAt:  time.Now(),
		Approved:   false,
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

	// Create readable lists for the summary
	var classNames []string
	for _, class := range state.PVPClasses {
		if name, exists := PVP_CLASSES[class]; exists {
			if emoji, emojiExists := PVP_CLASSES_EMOJI[class]; emojiExists {
				classNames = append(classNames, fmt.Sprintf("%s %s", emoji, name))
			} else {
				classNames = append(classNames, name)
			}
		}
	}

	var timeNames []string
	for _, timeKey := range state.Times {
		if name, exists := TIMES[timeKey]; exists {
			timeNames = append(timeNames, name)
		}
	}

	var weekdayNames []string
	for _, weekdayKey := range state.Weekdays {
		if name, exists := WEEKDAYS[weekdayKey]; exists {
			weekdayNames = append(weekdayNames, name)
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
			{
				Name:   "⏰ Horários",
				Value:  strings.Join(timeNames, ", "),
				Inline: false,
			},
			{
				Name:   "📅 Dias da Semana",
				Value:  strings.Join(weekdayNames, ", "),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Aguarde a aprovação de um administrador",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

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
	cfg := GetModuleConfig(ctx)

	// Extract registration ID from custom ID
	parts := strings.Split(i.MessageComponentData().CustomID, ":")
	if len(parts) != 3 {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ ID de registro inválido.", 5*time.Second)
		return
	}

	registrationID := parts[2]

	// Check if user has admin permissions
	if !hasAdminPermission(i.Member, cfg.AdminRoleID) {
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
	cfg := GetModuleConfig(ctx)

	// Extract registration ID from custom ID
	parts := strings.Split(i.MessageComponentData().CustomID, ":")
	if len(parts) != 3 {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ ID de registro inválido.", 5*time.Second)
		return
	}

	registrationID := parts[2]

	// Check if user has admin permissions
	if !hasAdminPermission(i.Member, cfg.AdminRoleID) {
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
	if adminRoleID == "" {
		return false
	}

	for _, roleID := range member.Roles {
		if roleID == adminRoleID {
			return true
		}
	}
	return false
}

func hasRole(member *discordgo.Member, roleID string) bool {
	if roleID == "" {
		return false
	}

	for _, memberRoleID := range member.Roles {
		if memberRoleID == roleID {
			return true
		}
	}
	return false
}

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
	player := &types.Player{
		ID:                primitive.NewObjectID(),
		DiscordID:         registration.DiscordID,
		IGN:               registration.InGameName,
		WarClass:          registration.Weapons, // Keep existing field for compatibility
		PVPClasses:        registration.Weapons,
		AvailableTimes:    registration.Hours,
		AvailableWeekdays: registration.WeekDays,
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
	cfg := GetModuleConfig(ctx)
	if cfg.MemberRoleID != "" && guildID != "" {
		dg := ctx.Session()
		err = dg.GuildMemberRoleAdd(guildID, registration.DiscordID, cfg.MemberRoleID)
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
