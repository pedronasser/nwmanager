package events

import (
	"errors"
	"fmt"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"slices"
	"time"

	. "nwmanager/helpers"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func buildEventMessage(config *EventsConfig, event *types.Event) *discordgo.MessageEmbed {
	desc := ""

	if event.Description != "" {
		desc += "\n```"
		desc += fmt.Sprintf("\n%s\n", event.Description)
		desc += "```\n"
	}

	fields := []*discordgo.MessageEmbedField{}

	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   "Organizador",
		Value:  fmt.Sprintf("<@%s>", event.Owner),
		Inline: true,
	})

	slotsCount := getEventSlotCount(config, event.Type)
	if slotsCount != -1 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Vagas",
			Value:  fmt.Sprintf("%d/%d", getEventFreeSlotsCount(config, event), slotsCount),
			Inline: true,
		})
	}

	fields = append(fields, &discordgo.MessageEmbedField{
		Name:   globals.SEPARATOR,
		Value:  "",
		Inline: false,
	})

	partyField := &discordgo.MessageEmbedField{
		Name:   "PT 1",
		Value:  "",
		Inline: true,
	}
	for i, player := range event.PlayerSlots {
		if i != 0 && i%5 == 0 {
			fields = append(fields, partyField)
			if i%10 == 0 {
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   "-",
					Value:  "",
					Inline: false,
				})
			}
			ptIndex := i/5 + 1
			partyField = &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("PT %d", ptIndex),
				Value:  "",
				Inline: true,
			}
		}

		if EventSlots[event.Type] != "" {
			role := getEventRoleNameByPosition(config, event.Type, i)
			playerName := "_[ABERTO]_"
			if player != "" {
				playerName = fmt.Sprintf("<@%s>", player)
			}
			partyField.Value += fmt.Sprintf("%s・%s\n", role, playerName)
		} else {
			partyField.Value += fmt.Sprintf("・<@%s>\n", player)
		}
	}
	fields = append(fields, partyField)

	if event.IsInviteOnly {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  "OBS:",
			Value: "_Este evento é fechado e requer aprovação do organizador para participar._",
		})
	}

	// footer := ""
	// for _, slot := range getEventSlotTypes(event) {
	// 	footer += fmt.Sprintf("・Reaja com %s para participar como %s.\n", EventSlotRoleEmoji[slot], EventSlotRoleName[slot])
	// }

	embed := &discordgo.MessageEmbed{
		Title:       getEventTitle(config, event),
		Description: desc,
		// Footer: &discordgo.MessageEmbedFooter{
		// 	Text: footer,
		// },
		Color: 0x00ff00,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://dqzvgunkova5o.cloudfront.net/statics/2024-10-31/images/NWA_logo.png",
		},
		Fields: fields,
	}

	return embed
}

func addPlayerToEvent(ctx *common.ModuleContext, userId string, event *types.Event, slotType EventSlotRole) error {
	config := GetModuleConfig(ctx)
	if EventSlots[event.Type] != "" {
		freeSlots := getEventFreeSlotsByRole(config, event, slotType)
		if len(freeSlots) == 0 {
			return errors.New("Não há vagas disponíveis para a função escolhida.")
		}

		slot := freeSlots[0]
		event.PlayerSlots[slot] = userId

		remainingSlots := getEventFreeSlots(event)
		if len(remainingSlots) == 0 {
			event.Status = types.EventStatusCompleted
			event.CompletedAt = Some(time.Now())
		}
	} else {
		event.PlayerSlots = append(event.PlayerSlots, userId)
	}

	_, err := ctx.DB().Collection(globals.DB_PREFIX+types.EventsCollection).UpdateOne(ctx.Context, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"player_slots": event.PlayerSlots}})
	if err != nil {
		return errors.New("Não foi possível adicionar jogador ao evento.")
	}

	return nil
}

func removePlayerFromEvent(ctx *common.ModuleContext, userId string, event *types.Event) {
	foundIndex := -1
	for i, slot := range event.PlayerSlots {
		if slot == userId {
			foundIndex = i
		}
	}

	if foundIndex == -1 {
		return
	}

	if EventSlots[event.Type] != "" {
		event.PlayerSlots[foundIndex] = ""
	} else {
		event.PlayerSlots = slices.Delete(event.PlayerSlots, foundIndex, foundIndex+1)
	}

	_, err := ctx.DB().Collection(globals.DB_PREFIX+types.EventsCollection).UpdateOne(ctx.Context, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"player_slots": event.PlayerSlots}})
	if err != nil {
		return
	}
}

func updateEventPlayerRole(ctx *common.ModuleContext, u *discordgo.User, event *types.Event, slotType EventSlotRole) error {
	config := GetModuleConfig(ctx)

	if EventSlots[event.Type] == "" {
		return errors.New("Este evento não possui slots de função.")
	}

	freeSlots := getEventFreeSlotsByRole(config, event, slotType)
	if len(freeSlots) == 0 {
		return errors.New("Não há vagas disponíveis para a função escolhida.")
	}

	foundIndex := -1
	for i, slot := range event.PlayerSlots {
		if slot == u.ID {
			foundIndex = i
		}
	}

	if foundIndex == -1 {
		return errors.New("Você não está inscrito neste evento.")
	}

	event.PlayerSlots[foundIndex] = ""
	event.PlayerSlots[freeSlots[0]] = u.ID

	_, err := ctx.DB().Collection(globals.DB_PREFIX+types.EventsCollection).UpdateOne(ctx.Context, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"player_slots": event.PlayerSlots}})
	if err != nil {
		return errors.New("Não foi possível atualizar a função do jogador.")
	}

	return nil
}

func updateEventMessage(ctx *common.ModuleContext, event *types.Event) error {
	config := GetModuleConfig(ctx)
	events_channel, err := ctx.Session().Channel(event.ChannelID)
	if err != nil {
		return fmt.Errorf("Cannot get events channel: %v", err)
	}

	message, err := ctx.Session().ChannelMessage(events_channel.ID, event.MessageID)
	if err != nil {
		return fmt.Errorf("Cannot get event message: %v", err)
	}

	_, err = ctx.Session().ChannelMessageEditEmbed(events_channel.ID, message.ID, buildEventMessage(config, event))
	if err != nil {
		return fmt.Errorf("Cannot edit event message: %v", err)
	}

	return nil
}

func createEvent(ctx *common.ModuleContext, i *discordgo.InteractionCreate, tipo types.EventType, channel_id, title, description string, scheduledAt *time.Time, isInviteOnly bool) error {
	config := GetModuleConfig(ctx)
	event := types.Event{
		ID:           primitive.NewObjectID(),
		Title:        title,
		Description:  description,
		Type:         tipo,
		Owner:        i.Interaction.Member.User.ID,
		Status:       types.EventStatusOpen,
		CreatedAt:    Some(time.Now()),
		PlayerSlots:  []string{},
		ScheduledAt:  scheduledAt,
		IsInviteOnly: isInviteOnly,
		ChannelID:    channel_id,
	}

	if EventSlots[event.Type] != "" {
		for range getEventSlotCount(config, event.Type) {
			event.PlayerSlots = append(event.PlayerSlots, "")
		}
	} else {
		event.PlayerSlots = []string{i.Interaction.Member.User.ID}
	}

	events_channel, err := ctx.Session().Channel(event.ChannelID)
	if err != nil {
		return fmt.Errorf("Cannot get events channel: %v", err)
	}

	message, err := createEventMessage(ctx, events_channel, &event)
	if err != nil {
		return fmt.Errorf("Cannot create event message: %v", err)
	}

	event.MessageID = message.ID

	_, err = ctx.DB().Collection(globals.DB_PREFIX+types.EventsCollection).InsertOne(ctx.Context, event)
	if err != nil {
		return fmt.Errorf("Cannot insert event into database: %v", err)
	}

	return nil
}

func createEventMessage(ctx *common.ModuleContext, events_channel *discordgo.Channel, event *types.Event) (*discordgo.Message, error) {
	config := GetModuleConfig(ctx)
	components := []discordgo.MessageComponent{}

	joinActionsRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{},
	}

	btnLength := 0

	for _, slot := range getEventSlotTypes(config, event) {
		if btnLength == 4 {
			components = append(components, joinActionsRow)
			joinActionsRow = discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{},
			}
			btnLength = 0
		}
		joinActionsRow.Components = append(joinActionsRow.Components, discordgo.Button{
			Label:    fmt.Sprintf("Entrar %s", EventSlotRoleName[slot]),
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("join_%s_%s", event.ID.Hex(), string(slot)),
			Emoji:    &discordgo.ComponentEmoji{Name: EventSlotRoleEmoji[slot]},
		})
		btnLength++
	}

	if EventSlots[event.Type] == "" {
		joinActionsRow.Components = append(joinActionsRow.Components, discordgo.Button{
			Label:    "Entrar",
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("join_%s_A", event.ID.Hex()),
			Emoji:    &discordgo.ComponentEmoji{Name: "🎮"},
		})
	}

	joinActionsRow.Components = append(joinActionsRow.Components, discordgo.Button{
		Label:    "Sair",
		Style:    discordgo.SecondaryButton,
		CustomID: fmt.Sprintf("leave_%s", event.ID.Hex()),
		Emoji:    &discordgo.ComponentEmoji{Name: "❌"},
	})

	components = append(components, joinActionsRow)
	components = append(components,
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Editar",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("edit_%s", event.ID.Hex()),
				},
				// discordgo.Button{
				// 	Label:    "Remover Participante",
				// 	Style:    discordgo.SecondaryButton,
				// 	CustomID: fmt.Sprintf("uninvite_%s", event.ID.Hex()),
				// },
				discordgo.Button{
					Label:    "Encerrar",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("close_event_%s", event.ID.Hex()),
				},
			},
		},
	)

	eventMessage := buildEventMessage(config, event)
	message, err := ctx.Session().ChannelMessageSendComplex(events_channel.ID,
		&discordgo.MessageSend{
			Embed:      eventMessage,
			Components: components,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Cannot send event message: %v", err)
	}

	if config.CreateThread {
		thread_channel, err := ctx.Session().MessageThreadStartComplex(events_channel.ID, message.ID, &discordgo.ThreadStart{
			Name: eventMessage.Title,
			Type: discordgo.ChannelTypeGuildPublicThread,
		})
		if err != nil {
			return nil, fmt.Errorf("Cannot create thread for event: %v", err)
		}

		_, err = ctx.Session().ChannelMessageSend(thread_channel.ID, fmt.Sprintf("Este é o canal de discussão do evento **%s**. Aqui você pode conversar com os participantes e tirar dúvidas.", eventMessage.Title))
		if err != nil {
			return nil, fmt.Errorf("Cannot send thread message: %v", err)
		}
	}

	return message, nil
}

func isUserAlreadyInEvent(event *types.Event, userID string) bool {
	return slices.Contains(event.PlayerSlots, userID)
}

func removeEventMessage(ctx *common.ModuleContext, event *types.Event) error {
	msg, err := ctx.Session().ChannelMessage(event.ChannelID, event.MessageID)
	if err != nil {
		return fmt.Errorf("Cannot get event message: %v", err)
	}

	if msg.Thread != nil {
		_, err = ctx.Session().ChannelDelete(msg.Thread.ID)
		if err != nil {
			return fmt.Errorf("Cannot delete event thread: %v", err)
		}
	}

	err = ctx.Session().ChannelMessageDelete(event.ChannelID, event.MessageID)
	if err != nil {
		return fmt.Errorf("Cannot delete event message: %v", err)
	}

	return nil
}

func closeEvent(ctx *common.ModuleContext, event *types.Event) error {
	_, err := ctx.DB().Collection(globals.DB_PREFIX+types.EventsCollection).UpdateOne(ctx.Context, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"status": types.EventStatusClosed, "closed_at": time.Now()}})
	if err != nil {
		return fmt.Errorf("Cannot close event: %v", err)
	}

	err = removeEventMessage(ctx, event)
	if err != nil {
		return fmt.Errorf("Cannot remove event: %v", err)
	}

	return nil
}

func ownerHasEvent(ctx *common.ModuleContext, owner *discordgo.User) bool {
	res := ctx.DB().Collection(globals.DB_PREFIX+types.EventsCollection).FindOne(ctx.Context, bson.M{"owner": owner.ID, "status": types.EventStatusOpen})
	if res.Err() != nil {
		return false
	}

	var event types.Event
	err := res.Decode(&event)
	if err != nil {
		return false
	}

	return true
}

func getEventTitle(config *EventsConfig, event *types.Event) string {
	title := fmt.Sprintf("%s - %s", getEventTypeName(config, event.Type), event.Title)
	if event.ScheduledAt != nil {
		title += fmt.Sprintf(" (%s)", (*event.ScheduledAt).Format("02/01/2006 às 15:04"))
	}

	return title
}

func sendJoinRequest(
	ctx *common.ModuleContext,
	event *types.Event,
	user *discordgo.User,
	slotType EventSlotRole,
) error {
	config := GetModuleConfig(ctx)
	channel, err := ctx.Session().UserChannelCreate(event.Owner)
	if err != nil {
		return err
	}

	var content string
	if slotType == EventSlotAny {
		content = fmt.Sprintf("O jogador <@%s> solicitou participar do evento **%s**. Você pode aprovar ou rejeitar a solicitação.", user.ID, getEventTitle(config, event))
	} else {
		content = fmt.Sprintf("O jogador <@%s> solicitou participar do evento **%s** como **%s**. Você pode aprovar ou rejeitar a solicitação.", user.ID, getEventTitle(config, event), EventSlotRoleName[slotType])
	}

	_, err = ctx.Session().ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Content: content,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Aceitar",
						Style:    discordgo.SuccessButton,
						CustomID: fmt.Sprintf("approve_%s_%s_%s", event.ID.Hex(), user.ID, string(slotType)),
					},
					discordgo.Button{
						Label:    "Rejeitar",
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("reject_%s_%s", event.ID.Hex(), user.ID),
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func canCreateEvent(ctx *common.ModuleContext, member *discordgo.Member) bool {
	if EVENTS_REQUIRE_ADMIN && !globals.IsMemberAdmin(ctx, member) {
		return false
	}

	return true
}
