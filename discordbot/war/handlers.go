package war

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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var handlers = map[string]func(ctx *common.ModuleContext, i *discordgo.InteractionCreate){
	COMMAND_CREATE_WAR:       handleCreateWarCommand,
	COMMAND_CANCEL_WAR:       handleCancelWarCommand,
	MODAL_CREATE_WAR:         handleCreateWarModal,
	BUTTON_PARTICIPATE_YES:   handleParticipateYes,
	BUTTON_PARTICIPATE_NO:    handleParticipateNo,
	BUTTON_PARTICIPATE_MAYBE: handleParticipateMaybe,
	BUTTON_CHANGE_ANSWER:     handleChangeAnswer,
	BUTTON_EDIT_WAR:          handleEditWar,
	SELECT_CANCEL_WAR:        handleCancelWarSelect,
}

// Handle the /criar-guerra command
func handleCreateWarCommand(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

	// Check if user has admin permission
	if !discordutils.HasRole(i.Member, globalConfig.AdminRoleID) {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Você não tem permissão para criar guerras.", 5*time.Second)
		return
	}

	// Send modal for war creation
	err := discordutils.SendModal(ctx.Session(), i, MODAL_CREATE_WAR, "Criar Nova Guerra",
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "fort_name",
					Label:       "Nome do Forte",
					Placeholder: "Ex: Forte do Vento",
					Style:       discordgo.TextInputShort,
					Required:    true,
					MaxLength:   100,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "war_type",
					Label:       "Tipo de Guerra (ataque ou defesa)",
					Placeholder: "Digite: ataque ou defesa",
					Style:       discordgo.TextInputShort,
					Required:    true,
					MaxLength:   10,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "opponent_guild",
					Label:       "Guild Oponente",
					Placeholder: "Ex: Inimigos Unidos",
					Style:       discordgo.TextInputShort,
					Required:    true,
					MaxLength:   100,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "scheduled_date",
					Label:       "Data da Guerra (DD/MM/AAAA)",
					Placeholder: "Ex: 25/08/2025",
					Style:       discordgo.TextInputShort,
					Required:    true,
					MinLength:   10,
					MaxLength:   10,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "scheduled_time",
					Label:       "Horário da Guerra (HH:MM)",
					Placeholder: "Ex: 20:00",
					Style:       discordgo.TextInputShort,
					Required:    true,
					MinLength:   5,
					MaxLength:   5,
				},
			},
		},
	)

	if err != nil {
		log.Printf("Error sending war creation modal: %v", err)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Erro ao abrir formulário de criação.", 5*time.Second)
	}
}

// Handle war creation modal submission
func handleCreateWarModal(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()

	var fortName, opponentGuild, scheduledDate, scheduledTime, warTypeInput string

	// Extract data from modal components
	for _, component := range data.Components {
		if actionRow, ok := component.(*discordgo.ActionsRow); ok {
			for _, comp := range actionRow.Components {
				if textInput, ok := comp.(*discordgo.TextInput); ok {
					switch textInput.CustomID {
					case "fort_name":
						fortName = strings.TrimSpace(textInput.Value)
					case "war_type":
						warTypeInput = strings.TrimSpace(strings.ToLower(textInput.Value))
					case "opponent_guild":
						opponentGuild = strings.TrimSpace(textInput.Value)
					case "scheduled_date":
						scheduledDate = strings.TrimSpace(textInput.Value)
					case "scheduled_time":
						scheduledTime = strings.TrimSpace(textInput.Value)
					}
				}
			}
		}
	}

	// Validate and convert war type
	var warType types.WarType
	switch warTypeInput {
	case "ataque", "attack":
		warType = types.WarTypeAttack
	case "defesa", "defense":
		warType = types.WarTypeDefense
	default:
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Tipo de guerra inválido. Use 'ataque' ou 'defesa'.", 5*time.Second)
		return
	}

	// Validate inputs
	if fortName == "" || warTypeInput == "" || opponentGuild == "" || scheduledDate == "" || scheduledTime == "" {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Todos os campos são obrigatórios.", 5*time.Second)
		return
	}

	// Parse date and time
	scheduledAt, err := parseDateTime(scheduledDate, scheduledTime)
	if err != nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, fmt.Sprintf("❌ Erro ao processar data/hora: %v", err), 5*time.Second)
		return
	}

	// Check if scheduled time is in the future
	if scheduledAt.Before(time.Now()) {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ A data e hora da guerra deve ser no futuro.", 5*time.Second)
		return
	}

	// Create war object
	now := time.Now()
	war := &types.War{
		ID:             primitive.NewObjectID(),
		FortName:       fortName,
		Type:           warType,
		OpponentGuild:  opponentGuild,
		ScheduledAt:    &scheduledAt,
		CreatedAt:      &now,
		CreatedBy:      i.Member.User.ID,
		Participations: make(map[string]types.WarParticipation),
		PlayerMessages: make(map[string]string),
		Status:         types.WarStatusActive,
	}

	// Save to database
	err = types.InsertWar(ctx.Context, ctx.DB(), war)
	if err != nil {
		log.Printf("Error inserting war: %v", err)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Erro ao salvar guerra no banco de dados.", 5*time.Second)
		return
	}

	// Reply to interaction first
	discordutils.ReplyEphemeralMessage(ctx.Session(), i, "✅ Guerra criada com sucesso!", 2*time.Second)

	// Post war message to channel and send DMs to players
	err = publishWar(ctx, war)
	if err != nil {
		log.Printf("Error publishing war: %v", err)
	}
}

// Handle the /cancelar-guerra command
func handleCancelWarCommand(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

	// Check if user has admin permission
	if !discordutils.HasRole(i.Member, globalConfig.AdminRoleID) {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Você não tem permissão para cancelar guerras.", 5*time.Second)
		return
	}

	// Get all active wars
	activeWars, err := types.GetActiveWars(ctx.Context, ctx.DB())
	if err != nil {
		log.Printf("Error getting active wars: %v", err)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Erro ao buscar guerras ativas.", 5*time.Second)
		return
	}

	if len(activeWars) == 0 {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Não há guerras ativas para cancelar.", 5*time.Second)
		return
	}

	// If there's only one active war, cancel it directly
	if len(activeWars) == 1 {
		war := activeWars[0]

		// Archive the war (this will also clean up messages)
		err = CancelWar(ctx, war)
		if err != nil {
			log.Printf("Error canceling war %s: %v", war.ID.Hex(), err)
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Erro ao cancelar a guerra.", 5*time.Second)
			return
		}

		// Reply with success message
		discordutils.ReplyEphemeralMessage(ctx.Session(), i,
			fmt.Sprintf("✅ Guerra **%s** (vs %s) foi cancelada com sucesso!\n\nTodas as mensagens relacionadas foram removidas.",
				war.FortName, war.OpponentGuild),
			0)

		log.Printf("War %s (%s vs %s) was canceled by admin %s",
			war.ID.Hex(), war.FortName, war.OpponentGuild, i.Member.User.Username)
		return
	}

	// Multiple active wars - show selection menu
	var options []discordgo.SelectMenuOption
	for _, war := range activeWars {
		warTypeEmoji := EMOJI_ATTACK
		warTypeText := "Ataque"
		if war.Type == types.WarTypeDefense {
			warTypeEmoji = EMOJI_DEFENSE
			warTypeText = "Defesa"
		}

		description := fmt.Sprintf("%s vs %s - <t:%d:F>", warTypeText, war.OpponentGuild, war.ScheduledAt.Unix())

		// Truncate description if too long (Discord limit is 100 chars)
		if len(description) > 100 {
			description = description[:97] + "..."
		}

		options = append(options, discordgo.SelectMenuOption{
			Label:       fmt.Sprintf("%s %s", warTypeEmoji, war.FortName),
			Value:       war.ID.Hex(),
			Description: description,
		})
	}

	// Send response with select menu
	err = ctx.Session().InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🔍 Existem **%d** guerras ativas. Selecione qual guerra deseja cancelar:", len(activeWars)),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    SELECT_CANCEL_WAR,
							Placeholder: "Selecione a guerra para cancelar...",
							Options:     options,
						},
					},
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		log.Printf("Error sending war selection menu: %v", err)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Erro ao exibir lista de guerras.", 5*time.Second)
	}
}

// Handle participation responses
func handleParticipateYes(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	handleParticipation(ctx, i, types.WarParticipationYes)
}

func handleParticipateNo(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	handleParticipation(ctx, i, types.WarParticipationNo)
}

func handleParticipateMaybe(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	handleParticipation(ctx, i, types.WarParticipationMaybe)
}

// Generic participation handler
func handleParticipation(ctx *common.ModuleContext, i *discordgo.InteractionCreate, participation types.WarParticipation) {
	log.Printf("User %s (%s) selected participation: %s", i.Member.User.Username, i.Member.User.ID, participation)

	// Extract war ID from custom ID (format: war_participate:yes:WAR_ID)
	parts := strings.Split(i.MessageComponentData().CustomID, ":")
	if len(parts) < 3 {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Erro interno: ID inválido.", 5*time.Second)
		return
	}

	warID := parts[2]

	// Get war from database
	war, err := types.GetWarByID(ctx.Context, ctx.DB(), warID)
	if err != nil || war == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Guerra não encontrada.", 5*time.Second)
		return
	}

	// Check if war is still active
	if war.Status != types.WarStatusActive {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Esta guerra já foi finalizada.", 5*time.Second)
		return
	}

	playerID := i.Member.User.ID
	oldParticipation := war.Participations[playerID]

	// Atomically update participation in database
	err = types.UpdateWarParticipation(ctx.Context, ctx.DB(), war.ID, playerID, participation)
	if err != nil {
		log.Printf("Error updating war participation: %v", err)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Erro ao salvar resposta.", 5*time.Second)
		return
	}

	// Update local war object for subsequent operations
	if war.Participations == nil {
		war.Participations = make(map[string]types.WarParticipation)
	}
	war.Participations[playerID] = participation

	// Get participation text
	participationText := getParticipationText(participation)

	// Reply to user
	if oldParticipation == "" {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i,
			fmt.Sprintf("✅ Resposta registrada: **%s**", participationText), 5*time.Second)
	} else {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i,
			fmt.Sprintf("✅ Resposta alterada para: **%s**", participationText), 5*time.Second)
	}

	// Update player's private message
	err = updatePlayerMessage(ctx, war, playerID, participation)
	if err != nil {
		log.Printf("Error updating player message: %v", err)
	}
}

// Handle changing war participation answer
func handleChangeAnswer(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	log.Printf("User %s (%s) clicked change answer button", i.Member.User.Username, i.Member.User.ID)

	// Extract war ID from custom ID (format: war_change_answer:WAR_ID)
	parts := strings.Split(i.MessageComponentData().CustomID, ":")
	if len(parts) < 2 {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Erro interno: ID inválido.", 5*time.Second)
		return
	}

	warID := parts[1]

	// Get war from database
	war, err := types.GetWarByID(ctx.Context, ctx.DB(), warID)
	if err != nil || war == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Guerra não encontrada.", 5*time.Second)
		return
	}

	// Check if war is still active
	if war.Status != types.WarStatusActive {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Esta guerra já foi finalizada.", 5*time.Second)
		return
	}

	playerID := i.Member.User.ID

	// Get player info
	player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), playerID)
	if err != nil || player == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Jogador não encontrado.", 5*time.Second)
		return
	}

	// Check if this is the player's ticket channel
	if player.TicketChannel != i.ChannelID {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Você só pode alterar sua resposta no seu próprio ticket.", 5*time.Second)
		return
	}

	// Create embed with current status
	embed := createPlayerWarEmbed(war)

	// Add current participation status if exists
	if currentParticipation, exists := war.Participations[playerID]; exists {
		participationText := getParticipationText(currentParticipation)
		participationEmoji := getParticipationEmoji(currentParticipation)

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Resposta Atual",
			Value:  fmt.Sprintf("%s **%s**", participationEmoji, participationText),
			Inline: false,
		})
	}

	// Create participation buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: fmt.Sprintf("war_participate:yes:%s", war.ID.Hex()),
					Label:    "Sim",
					Style:    discordgo.SuccessButton,
					Emoji: &discordgo.ComponentEmoji{
						Name: EMOJI_YES,
					},
				},
				discordgo.Button{
					CustomID: fmt.Sprintf("war_participate:no:%s", war.ID.Hex()),
					Label:    "Não",
					Style:    discordgo.DangerButton,
					Emoji: &discordgo.ComponentEmoji{
						Name: EMOJI_NO,
					},
				},
				discordgo.Button{
					CustomID: fmt.Sprintf("war_participate:maybe:%s", war.ID.Hex()),
					Label:    "Talvez",
					Style:    discordgo.SecondaryButton,
					Emoji: &discordgo.ComponentEmoji{
						Name: EMOJI_MAYBE,
					},
				},
			},
		},
	}

	// Update the message with the participation options
	edit := &discordgo.MessageEdit{
		Channel:    i.ChannelID,
		ID:         i.Message.ID,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}

	_, err = ctx.Session().ChannelMessageEditComplex(edit)
	if err != nil {
		log.Printf("Error updating message with participation options: %v", err)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Erro ao exibir opções de participação.", 5*time.Second)
		return
	}

	// Acknowledge the interaction
	err = ctx.Session().InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("Error acknowledging interaction: %v", err)
	}
}

// Handle war editing (admin only)
func handleEditWar(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

	// Check if user has admin permission
	if !discordutils.HasRole(i.Member, globalConfig.AdminRoleID) {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Você não tem permissão para editar guerras.", 5*time.Second)
		return
	}

	// TODO: Implement war editing functionality
	discordutils.ReplyEphemeralMessage(ctx.Session(), i, "🚧 Funcionalidade de edição em desenvolvimento.", 5*time.Second)
}

// Handle war selection for cancellation
func handleCancelWarSelect(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

	// Check if user has admin permission
	if !discordutils.HasRole(i.Member, globalConfig.AdminRoleID) {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Você não tem permissão para cancelar guerras.", 5*time.Second)
		return
	}

	// Get selected war ID from the select menu
	data := i.MessageComponentData()
	if len(data.Values) == 0 {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Nenhuma guerra selecionada.", 5*time.Second)
		return
	}

	warID := data.Values[0]

	// Get war from database
	war, err := types.GetWarByID(ctx.Context, ctx.DB(), warID)
	if err != nil || war == nil {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Guerra não encontrada.", 5*time.Second)
		return
	}

	// Check if war is still active
	if war.Status != types.WarStatusActive {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Esta guerra já foi finalizada.", 5*time.Second)
		return
	}

	// Cancel the war
	err = CancelWar(ctx, war)
	if err != nil {
		log.Printf("Error canceling war %s: %v", war.ID.Hex(), err)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "❌ Erro ao cancelar a guerra.", 5*time.Second)
		return
	}

	// Reply with success message
	discordutils.ReplyEphemeralMessage(ctx.Session(), i,
		fmt.Sprintf("✅ Guerra **%s** (vs %s) foi cancelada com sucesso!\n\nTodas as mensagens relacionadas foram removidas.",
			war.FortName, war.OpponentGuild),
		0)

	log.Printf("War %s (%s vs %s) was canceled by admin %s",
		war.ID.Hex(), war.FortName, war.OpponentGuild, i.Member.User.Username)
}

// HandleWarAction creates the main interaction handler
func HandleWarAction(ctx *common.ModuleContext, guildID string) func(*discordgo.Session, *discordgo.InteractionCreate) {
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
			// Handle participation buttons with war ID
			if strings.HasPrefix(data.CustomID, "war_participate:") {
				parts := strings.Split(data.CustomID, ":")
				if len(parts) >= 2 {
					handlerKey = "war_participate:" + parts[1]
				}
			} else if strings.HasPrefix(data.CustomID, "war_change_answer:") {
				handlerKey = BUTTON_CHANGE_ANSWER
			} else if strings.Contains(data.CustomID, ":") {
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
