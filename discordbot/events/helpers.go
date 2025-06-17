package events

import (
	"context"
	"errors"
	"fmt"
	"log"
	"nwmanager/database"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"slices"
	"time"

	. "nwmanager/helpers"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func buildEventMessage(event *types.Event) *discordgo.MessageEmbed {
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

	slotsCount := getEventSlotCount(event.Type)
	if slotsCount != -1 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Vagas",
			Value:  fmt.Sprintf("%d/%d", getEventFreeSlotsCount(event), slotsCount),
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
			role := getEventRoleNameByPosition(event.Type, i)
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
		Title:       getEventTitle(event),
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

func addPlayerToEvent(userId string, db database.Database, event *types.Event, slotType EventSlotRole) error {
	if EventSlots[event.Type] != "" {
		freeSlots := getEventFreeSlotsByRole(event, slotType)
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

	ctx := context.Background()
	_, err := db.Collection(globals.DB_PREFIX+types.EventsCollection).UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"player_slots": event.PlayerSlots}})
	if err != nil {
		return errors.New("Não foi possível adicionar jogador ao evento.")
	}

	return nil
}

func removePlayerFromEvent(userId string, db database.Database, event *types.Event) {
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
		event.PlayerSlots = append(event.PlayerSlots[:foundIndex], event.PlayerSlots[foundIndex+1:]...)
	}

	ctx := context.Background()
	_, err := db.Collection(globals.DB_PREFIX+types.EventsCollection).UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"player_slots": event.PlayerSlots}})
	if err != nil {
		return
	}
}

func updateEventPlayerRole(u *discordgo.User, db database.Database, event *types.Event, slotType EventSlotRole) error {
	if EventSlots[event.Type] == "" {
		return errors.New("Este evento não possui slots de função.")
	}

	freeSlots := getEventFreeSlotsByRole(event, slotType)
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

	ctx := context.Background()
	_, err := db.Collection(globals.DB_PREFIX+types.EventsCollection).UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"player_slots": event.PlayerSlots}})
	if err != nil {
		return errors.New("Não foi possível atualizar a função do jogador.")
	}

	return nil
}

func updateEventMessage(s *discordgo.Session, event *types.Event) {
	events_channel, err := s.Channel(event.ChannelID)
	if err != nil {
		log.Fatalf("Cannot get events channel: %v", err)
	}

	message, err := s.ChannelMessage(events_channel.ID, event.MessageID)
	if err != nil {
		log.Fatalf("Cannot get message: %v", err)
	}

	_, err = s.ChannelMessageEditEmbed(events_channel.ID, message.ID, buildEventMessage(event))
	if err != nil {
		log.Fatalf("Cannot edit message: %v", err)
	}
}

func createEvent(s *discordgo.Session, i *discordgo.InteractionCreate, db database.Database, tipo types.EventType, channel_id, title, description string, scheduledAt *time.Time, isInviteOnly bool) {
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
		for i := 0; i < getEventSlotCount(event.Type); i++ {
			event.PlayerSlots = append(event.PlayerSlots, "")
		}
	} else {
		event.PlayerSlots = []string{i.Interaction.Member.User.ID}
	}

	events_channel, err := s.Channel(event.ChannelID)
	if err != nil {
		log.Fatalf("Cannot get events channel: %v", err)
	}

	message := createEventMessage(s, events_channel, &event)
	event.MessageID = message.ID

	ctx := context.Background()
	_, err = db.Collection(globals.DB_PREFIX+types.EventsCollection).InsertOne(ctx, event)
	if err != nil {
		log.Fatalf("Cannot insert event: %v", err)
	}
}

func createEventMessage(dg *discordgo.Session, events_channel *discordgo.Channel, event *types.Event) *discordgo.Message {

	components := []discordgo.MessageComponent{}

	joinActionsRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{},
	}

	btnLength := 0

	for _, slot := range getEventSlotTypes(event) {
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

	message, err := dg.ChannelMessageSendComplex(events_channel.ID,
		&discordgo.MessageSend{
			Embed:      buildEventMessage(event),
			Components: components,
		},
	)
	if err != nil {
		log.Fatalf("Cannot send event message: %v", err)
	}

	// for _, slot := range getEventSlotTypes(event) {
	// 	err = dg.MessageReactionAdd(events_channel.ID, message.ID, EventSlotRoleEmoji[slot])
	// 	if err != nil {
	// 		log.Fatalf("Cannot add reaction to message: %v", err)
	// 	}
	// }

	// err = dg.MessageReactionAdd(events_channel.ID, message.ID, "❌")
	// if err != nil {
	// 	log.Fatalf("Cannot add reaction to message: %v", err)
	// }

	return message
}

func isUserAlreadyInEvent(event *types.Event, userID string) bool {
	return slices.Contains(event.PlayerSlots, userID)
}

func removeEvent(ctx context.Context, db database.Database, s *discordgo.Session, event *types.Event) {
	_, err := db.Collection(globals.DB_PREFIX+types.EventsCollection).DeleteOne(ctx, bson.M{"_id": event.ID})
	if err != nil {
		log.Fatalf("Cannot delete event: %v", err)
	}

	_ = s.ChannelMessageDelete(event.ChannelID, event.MessageID)
}

func closeEvent(ctx context.Context, db database.Database, s *discordgo.Session, event *types.Event) {
	_, err := db.Collection(globals.DB_PREFIX+types.EventsCollection).UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"status": types.EventStatusClosed, "closed_at": time.Now()}})
	if err != nil {
		log.Fatalf("Cannot update event: %v", err)
	}

	_ = s.ChannelMessageDelete(event.ChannelID, event.MessageID)
}

func ownerHasEvent(ctx context.Context, db database.Database, owner *discordgo.User) bool {
	res := db.Collection(globals.DB_PREFIX+types.EventsCollection).FindOne(ctx, bson.M{"owner": owner.ID, "status": types.EventStatusOpen})
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

func getEventTitle(event *types.Event) string {
	title := fmt.Sprintf("%s - %s", getEventTypeName(event.Type), event.Title)
	if event.ScheduledAt != nil {
		title += fmt.Sprintf(" (%s)", (*event.ScheduledAt).Format("02/01/2006 às 15:04"))
	}

	return title
}

func sendJoinRequest(
	ctx context.Context,
	db database.Database,
	s *discordgo.Session,
	event *types.Event,
	user *discordgo.User,
	slotType EventSlotRole,
) error {
	channel, err := s.UserChannelCreate(event.Owner)
	if err != nil {
		return err
	}

	var content string
	if slotType == EventSlotAny {
		content = fmt.Sprintf("O jogador <@%s> solicitou participar do evento **%s**. Você pode aprovar ou rejeitar a solicitação.", user.ID, getEventTitle(event))
	} else {
		content = fmt.Sprintf("O jogador <@%s> solicitou participar do evento **%s** como **%s**. Você pode aprovar ou rejeitar a solicitação.", user.ID, getEventTitle(event), EventSlotRoleName[slotType])
	}

	_, err = s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
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

func canCreateEvent(member *discordgo.Member) bool {
	if EVENTS_REQUIRE_ADMIN && !discordutils.IsMemberAdmin(member) {
		return false
	}

	return true
}
