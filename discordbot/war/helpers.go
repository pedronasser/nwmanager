package war

import (
	"fmt"
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
)

// parseDateTime parses date and time strings into a time.Time
func parseDateTime(dateStr, timeStr string) (time.Time, error) {
	// Parse date (DD/MM/YYYY)
	dateParts := strings.Split(dateStr, "/")
	if len(dateParts) != 3 {
		return time.Time{}, fmt.Errorf("formato de data inválido. Use DD/MM/AAAA")
	}

	day, err := strconv.Atoi(dateParts[0])
	if err != nil || day < 1 || day > 31 {
		return time.Time{}, fmt.Errorf("dia inválido")
	}

	month, err := strconv.Atoi(dateParts[1])
	if err != nil || month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("mês inválido")
	}

	year, err := strconv.Atoi(dateParts[2])
	if err != nil || year < 2025 {
		return time.Time{}, fmt.Errorf("ano inválido")
	}

	// Parse time (HH:MM)
	timeParts := strings.Split(timeStr, ":")
	if len(timeParts) != 2 {
		return time.Time{}, fmt.Errorf("formato de hora inválido. Use HH:MM")
	}

	hour, err := strconv.Atoi(timeParts[0])
	if err != nil || hour < 0 || hour > 23 {
		return time.Time{}, fmt.Errorf("hora inválida")
	}

	minute, err := strconv.Atoi(timeParts[1])
	if err != nil || minute < 0 || minute > 59 {
		return time.Time{}, fmt.Errorf("minuto inválido")
	}

	// Construct time
	return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.Local), nil
}

// getParticipationText returns the text representation of a participation
func getParticipationText(participation types.WarParticipation) string {
	switch participation {
	case types.WarParticipationYes:
		return "Sim"
	case types.WarParticipationNo:
		return "Não"
	case types.WarParticipationMaybe:
		return "Talvez"
	case types.WarParticipationBench:
		return "Banco"
	default:
		return "Desconhecido"
	}
}

// getParticipationEmoji returns the emoji for a participation type
func getParticipationEmoji(participation types.WarParticipation) string {
	switch participation {
	case types.WarParticipationYes:
		return EMOJI_YES
	case types.WarParticipationNo:
		return EMOJI_NO
	case types.WarParticipationMaybe:
		return EMOJI_MAYBE
	case types.WarParticipationBench:
		return EMOJI_BENCH
	default:
		return "❓"
	}
}

// publishWar posts the war message to the channel and sends DMs to players
func publishWar(ctx *common.ModuleContext, war *types.War) error {
	cfg := GetModuleConfig(ctx)

	// Post to war channel
	if cfg.WarChannelID != "" {
		messageID, err := postWarToChannel(ctx, war, cfg.WarChannelID)
		if err != nil {
			return fmt.Errorf("failed to post war to channel: %v", err)
		}

		// Update war with channel message ID
		war.ChannelMessageID = messageID
		err = types.UpdateWar(ctx.Context, ctx.DB(), war)
		if err != nil {
			log.Printf("Error updating war with channel message ID: %v", err)
		}
	}

	// Send DMs to all active players
	err := sendWarDMsToPlayers(ctx, war)
	if err != nil {
		log.Printf("Error sending war DMs: %v", err)
	}

	return nil
}

// postWarToChannel posts the war embed to the specified channel
func postWarToChannel(ctx *common.ModuleContext, war *types.War, channelID string) (string, error) {
	embed := createWarEmbedWithMemberList(ctx, war)

	message := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		// Components: []discordgo.MessageComponent{
		// 	discordgo.ActionsRow{
		// 		Components: []discordgo.MessageComponent{
		// 			discordgo.Button{
		// 				CustomID: BUTTON_EDIT_WAR + ":" + war.ID.Hex(),
		// 				Label:    "Editar Guerra",
		// 				Style:    discordgo.SecondaryButton,
		// 				Emoji: &discordgo.ComponentEmoji{
		// 					Name: "✏️",
		// 				},
		// 			},
		// 		},
		// 	},
		// },
	}

	msg, err := ctx.Session().ChannelMessageSendComplex(channelID, message)
	if err != nil {
		return "", err
	}

	return msg.ID, nil
}

// createWarEmbedWithMemberList creates the main war embed with members listed by war class
func createWarEmbedWithMemberList(ctx *common.ModuleContext, war *types.War) *discordgo.MessageEmbed {
	confirmedCount := 0
	noCount := 0

	for _, participation := range war.Participations {
		switch participation {
		case types.WarParticipationYes, types.WarParticipationBench:
			confirmedCount++
		case types.WarParticipationNo:
			noCount++
		}
	}

	// Determine color based on war type
	var color int
	if war.Type == types.WarTypeAttack {
		color = 0xFF4444 // Red for attack
	} else {
		color = 0x4444FF // Blue for defense
	}

	// Determine war type text and emoji
	warTypeText := ""
	if war.Type == types.WarTypeAttack {
		warTypeText = EMOJI_ATTACK + " **ATAQUE**"
	} else {
		warTypeText = EMOJI_DEFENSE + " **DEFESA**"
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("🏰 Guerra: %s", war.FortName),
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Tipo",
				Value:  warTypeText,
				Inline: true,
			},
			{
				Name:   "Oponente",
				Value:  fmt.Sprintf("**%s**", war.OpponentGuild),
				Inline: true,
			},
			{
				Name:   "Data/Hora",
				Value:  fmt.Sprintf("<t:%d:F>", war.ScheduledAt.Unix()),
				Inline: true,
			},
			{
				Name:   "Participação",
				Value:  fmt.Sprintf("**%d/50** confirmados", confirmedCount),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use seu ticket para responder sobre sua participação",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Add member list fields grouped by war class
	memberFields := createMemberListFields(ctx, war)
	embed.Fields = append(embed.Fields, memberFields...)

	return embed
}

// createMemberListFields creates embed fields with members grouped by war class
func createMemberListFields(ctx *common.ModuleContext, war *types.War) []*discordgo.MessageEmbedField {
	// Get players who answered "Sim" or "Talvez"
	confirmedPlayers := make(map[string][]string) // war_class -> []IGN
	maybeePlayers := make(map[string][]string)    // war_class -> []IGN

	for playerID, participation := range war.Participations {
		if participation == types.WarParticipationYes || participation == types.WarParticipationMaybe {
			player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), playerID)
			if err != nil || player == nil || player.ArchivedAt != nil {
				continue
			}

			warClass := player.WarClass
			if warClass == "" {
				warClass = "Sem Classe"
			}

			switch participation {
			case types.WarParticipationYes:
				confirmedPlayers[warClass] = append(confirmedPlayers[warClass], player.IGN)
			case types.WarParticipationMaybe:
				maybeePlayers[warClass] = append(maybeePlayers[warClass], player.IGN)
			}
		}
	}

	var fields []*discordgo.MessageEmbedField

	// Add confirmed players ("Sim") grouped by war class
	if len(confirmedPlayers) > 0 {
		for warClass, players := range confirmedPlayers {
			if len(players) > 0 {
				value := fmt.Sprintf("%s %s", EMOJI_YES, strings.Join(players, ", "))
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   warClass,
					Value:  value,
					Inline: false,
				})
			}
		}
	}

	// Add maybe players ("Talvez") grouped by war class
	if len(maybeePlayers) > 0 {
		for warClass, players := range maybeePlayers {
			if len(players) > 0 {
				value := fmt.Sprintf("%s %s", EMOJI_MAYBE, strings.Join(players, ", "))
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   warClass + " (Talvez)",
					Value:  value,
					Inline: false,
				})
			}
		}
	}

	// If no confirmed or maybe players, add a placeholder
	if len(fields) == 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Participantes Confirmados",
			Value:  "Nenhum jogador confirmou participação ainda.",
			Inline: false,
		})
	}

	return fields
}

// sendWarDMsToPlayers sends DM to all active players via their ticket channels
func sendWarDMsToPlayers(ctx *common.ModuleContext, war *types.War) error {
	// Get all active players - using a more flexible query
	// First try to get all non-archived players
	cursor, err := ctx.DB().Collection(globals.DB_PREFIX+types.PlayerCollection).Find(ctx.Context, bson.M{
		"archived_at":    bson.M{"$eq": nil},
		"ticket_channel": bson.M{"$ne": "", "$exists": true},
	})
	if err != nil {
		return fmt.Errorf("failed to query players: %v", err)
	}
	defer cursor.Close(ctx.Context)

	var players []types.Player
	err = cursor.All(ctx.Context, &players)
	if err != nil {
		return fmt.Errorf("failed to decode players: %v", err)
	}

	log.Printf("Found %d active players with ticket channels", len(players))

	sentCount := 0
	skippedCount := 0

	for _, player := range players {
		if player.TicketChannel == "" {
			log.Printf("Skipping player %s (IGN: %s) - no ticket channel", player.DiscordID, player.IGN)
			skippedCount++
			continue // Skip players without ticket channels
		}

		log.Printf("Sending war DM to player %s (IGN: %s) in channel %s", player.DiscordID, player.IGN, player.TicketChannel)

		err := sendWarDMToPlayer(ctx, war, &player)
		if err != nil {
			log.Printf("Error sending war DM to player %s (IGN: %s): %v", player.DiscordID, player.IGN, err)
		} else {
			log.Printf("Successfully sent war DM to player %s (IGN: %s)", player.DiscordID, player.IGN)
			sentCount++
		}
	}

	log.Printf("War DM summary: %d sent, %d skipped (no ticket channel)", sentCount, skippedCount)

	// Update war in database with player message IDs
	if sentCount > 0 {
		log.Printf("Updating war in database with %d player message IDs", len(war.PlayerMessages))
		err = types.UpdateWar(ctx.Context, ctx.DB(), war)
		if err != nil {
			log.Printf("Error updating war with player message IDs: %v", err)
		}
	}

	return nil
}

// sendWarDMToPlayer sends a war participation DM to a specific player
func sendWarDMToPlayer(ctx *common.ModuleContext, war *types.War, player *types.Player) error {
	log.Printf("Creating embed for player %s", player.IGN)
	embed := createPlayerWarEmbed(war)

	log.Printf("Creating components for war %s", war.ID.Hex())
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
				discordgo.Button{
					CustomID: fmt.Sprintf("war_participate:bench:%s", war.ID.Hex()),
					Label:    "Banco",
					Style:    discordgo.PrimaryButton,
					Emoji: &discordgo.ComponentEmoji{
						Name: EMOJI_BENCH,
					},
				},
			},
		},
	}

	message := &discordgo.MessageSend{
		Content:    fmt.Sprintf("📢 <@%s>, uma guerra foi declarada!", player.DiscordID),
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	}

	log.Printf("Sending message to channel %s for player %s", player.TicketChannel, player.IGN)
	msg, err := ctx.Session().ChannelMessageSendComplex(player.TicketChannel, message)
	if err != nil {
		log.Printf("Failed to send message to channel %s: %v", player.TicketChannel, err)
		return err
	}

	log.Printf("Message sent successfully to channel %s, message ID: %s", player.TicketChannel, msg.ID)

	// Store message ID for later updates
	if war.PlayerMessages == nil {
		war.PlayerMessages = make(map[string]string)
	}
	war.PlayerMessages[player.DiscordID] = msg.ID

	log.Printf("Stored message ID %s for player %s", msg.ID, player.DiscordID)

	return nil
} // createPlayerWarEmbed creates the war embed for player DMs
func createPlayerWarEmbed(war *types.War) *discordgo.MessageEmbed {
	// Determine war type text and emoji
	warTypeText := ""
	var color int
	if war.Type == types.WarTypeAttack {
		warTypeText = EMOJI_ATTACK + " Ataque"
		color = 0xFF4444
	} else {
		warTypeText = EMOJI_DEFENSE + " Defesa"
		color = 0x4444FF
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("⚔️ Guerra: %s", war.FortName),
		Color:       color,
		Description: "Você irá participar desta guerra?",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Tipo",
				Value:  warTypeText,
				Inline: true,
			},
			{
				Name:   "Oponente",
				Value:  war.OpponentGuild,
				Inline: true,
			},
			{
				Name:   "Data/Hora",
				Value:  fmt.Sprintf("<t:%d:F>", war.ScheduledAt.Unix()),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Clique nos botões abaixo para responder",
		},
	}

	return embed
}

// updateWarMessage updates the war message in the channel with new participation counts
func updateWarMessage(ctx *common.ModuleContext, war *types.War) error {
	if war.ChannelMessageID == "" {
		return nil // No message to update
	}

	cfg := GetModuleConfig(ctx)
	if cfg.WarChannelID == "" {
		return nil
	}

	embed := createWarEmbedWithMemberList(ctx, war)

	edit := &discordgo.MessageEdit{
		Channel: cfg.WarChannelID,
		ID:      war.ChannelMessageID,
		Embeds:  &[]*discordgo.MessageEmbed{embed},
	}

	_, err := ctx.Session().ChannelMessageEditComplex(edit)
	return err
}

// updatePlayerMessage updates a player's private message with their current participation
func updatePlayerMessage(ctx *common.ModuleContext, war *types.War, playerID string, participation types.WarParticipation) error {
	// Get player info
	player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), playerID)
	if err != nil || player == nil {
		return fmt.Errorf("player not found")
	}

	messageID, exists := war.PlayerMessages[playerID]
	if !exists || player.TicketChannel == "" {
		return nil // No message to update
	}

	// Create updated embed with participation status
	embed := createPlayerWarEmbed(war)

	// Add participation status field
	participationText := getParticipationText(participation)
	participationEmoji := getParticipationEmoji(participation)

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Sua Resposta",
		Value:  fmt.Sprintf("%s **%s**", participationEmoji, participationText),
		Inline: false,
	})

	// Same components as before
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
				discordgo.Button{
					CustomID: fmt.Sprintf("war_participate:bench:%s", war.ID.Hex()),
					Label:    "Banco",
					Style:    discordgo.PrimaryButton,
					Emoji: &discordgo.ComponentEmoji{
						Name: EMOJI_BENCH,
					},
				},
			},
		},
	}

	edit := &discordgo.MessageEdit{
		Channel:    player.TicketChannel,
		ID:         messageID,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}

	_, err = ctx.Session().ChannelMessageEditComplex(edit)
	return err
}

// CancelWar cancels an active war and cleans up its messages
func CancelWar(ctx *common.ModuleContext, war *types.War) error {
	// Delete channel message
	if war.ChannelMessageID != "" {
		cfg := GetModuleConfig(ctx)
		if cfg.WarChannelID != "" {
			err := ctx.Session().ChannelMessageDelete(cfg.WarChannelID, war.ChannelMessageID)
			if err != nil {
				log.Printf("Error deleting war channel message: %v", err)
			}
		}
	}

	// Delete player messages
	for playerID, messageID := range war.PlayerMessages {
		// Get player to find their ticket channel
		player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), playerID)
		if err != nil || player == nil || player.TicketChannel == "" {
			continue
		}

		err = ctx.Session().ChannelMessageDelete(player.TicketChannel, messageID)
		if err != nil {
			log.Printf("Error deleting player message for %s: %v", player.IGN, err)
		}
	}

	// Archive the war in database
	err := types.ArchiveWar(ctx.Context, ctx.DB(), war.ID)
	if err != nil {
		return err
	}

	return nil
}
