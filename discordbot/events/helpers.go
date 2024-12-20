package events

import (
	"context"
	"fmt"
	"log"
	"nwmanager/types"
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

	for i, player := range event.PlayerSlots {
		if EventSlots[event.Type] != "" {
			role := getEventRoleNameByPosition(event.Type, i)
			playerName := "_[ABERTO]_"
			if player != "" {
				playerName = fmt.Sprintf("<@%s>", player)
			}
			desc += fmt.Sprintf("%s・%s\n", role, playerName)
		} else {
			desc += fmt.Sprintf("・<@%s>\n", player)
		}

		if (i+1)%5 == 0 {
			desc += "\n"
		}
	}

	footer := "・Reaja com 🛡️ para participar como TANK.\n"
	footer += "・Reaja com 🌿 para participar como HEALER.\n"
	footer += "・Reaja com ⚔️ para participar como DPS.\n"
	footer += "・Reaja com ❌ para sair do evento.\n"

	scheduled := ""
	if event.ScheduledAt != nil {
		scheduled = fmt.Sprintf(" (%s)", (*event.ScheduledAt).Format("02/01/2006 às 15:04"))
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s - %s%s", getEventName(event.Type), event.Title, scheduled),
		Description: desc,
		Footer: &discordgo.MessageEmbedFooter{
			Text: footer,
		},
		Color: 0x00ff00,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://dqzvgunkova5o.cloudfront.net/statics/2024-10-31/images/NWA_logo.png",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Organizador",
				Value:  fmt.Sprintf("<@%s>", event.Owner),
				Inline: true,
			},
		},
	}

	return embed
}

func removePlayerFromEvent(u *discordgo.User, db types.Database, event *types.Event) {
	foundIndex := -1
	for i, slot := range event.PlayerSlots {
		if slot == u.ID {
			foundIndex = i
		}
	}

	if foundIndex == -1 {
		return
	}

	if EventRoles[event.Type] != "" {
		event.PlayerSlots[foundIndex] = ""
	} else {
		event.PlayerSlots = append(event.PlayerSlots[:foundIndex], event.PlayerSlots[foundIndex+1:]...)
	}

	ctx := context.Background()
	_, err := db.Collection(types.EventsCollection).UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"player_slots": event.PlayerSlots}})
	if err != nil {
		return
	}
}

func updateEventMessage(s *discordgo.Session, event *types.Event) {
	events_channel, err := s.Channel(EVENTS_CHANNEL_ID)
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

func createEvent(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database, tipo types.EventType, title, description string, scheduledAt *time.Time) {
	event := types.Event{
		ID:          primitive.NewObjectID(),
		Title:       title,
		Description: description,
		Type:        tipo,
		Owner:       i.Interaction.Member.User.ID,
		Status:      types.EventStatusOpen,
		CreatedAt:   Some(time.Now()),
		PlayerSlots: []string{},
		ScheduledAt: scheduledAt,
	}

	if EventSlots[event.Type] != "" {
		for i := 0; i < getEventSlotCount(event.Type); i++ {
			event.PlayerSlots = append(event.PlayerSlots, "")
		}
	}

	events_channel, err := s.Channel(EVENTS_CHANNEL_ID)
	if err != nil {
		log.Fatalf("Cannot get events channel: %v", err)
	}

	message := createEventMessage(s, events_channel, &event)
	event.MessageID = message.ID

	ctx := context.Background()
	_, err = db.Collection(types.EventsCollection).InsertOne(ctx, event)
	if err != nil {
		log.Fatalf("Cannot insert event: %v", err)
	}
}

func createEventMessage(dg *discordgo.Session, events_channel *discordgo.Channel, event *types.Event) *discordgo.Message {
	message, err := dg.ChannelMessageSendComplex(events_channel.ID,
		&discordgo.MessageSend{
			Embed: buildEventMessage(event),
			Components: []discordgo.MessageComponent{
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
			},
		},
	)
	if err != nil {
		log.Fatalf("Cannot send event message: %v", err)
	}

	err = dg.MessageReactionAdd(events_channel.ID, message.ID, "🛡️")
	if err != nil {
		log.Fatalf("Cannot add reaction to message: %v", err)
	}

	err = dg.MessageReactionAdd(events_channel.ID, message.ID, "🌿")
	if err != nil {
		log.Fatalf("Cannot add reaction to message: %v", err)
	}

	err = dg.MessageReactionAdd(events_channel.ID, message.ID, "⚔️")
	if err != nil {
		log.Fatalf("Cannot add reaction to message: %v", err)
	}

	err = dg.MessageReactionAdd(events_channel.ID, message.ID, "❌")
	if err != nil {
		log.Fatalf("Cannot add reaction to message: %v", err)
	}

	return message
}

func isUserAlreadyInEvent(event *types.Event, userID string) bool {
	for _, player := range event.PlayerSlots {
		if player == userID {
			return true
		}
	}
	return false
}

func removeEvent(ctx context.Context, db types.Database, s *discordgo.Session, event *types.Event) {
	_, err := db.Collection(types.EventsCollection).DeleteOne(ctx, bson.M{"_id": event.ID})
	if err != nil {
		log.Fatalf("Cannot delete event: %v", err)
	}

	_ = s.ChannelMessageDelete(EVENTS_CHANNEL_ID, event.MessageID)
}

func closeEvent(ctx context.Context, db types.Database, s *discordgo.Session, event *types.Event) {
	_, err := db.Collection(types.EventsCollection).UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"status": types.EventStatusClosed, "closed_at": time.Now()}})
	if err != nil {
		log.Fatalf("Cannot update event: %v", err)
	}

	_ = s.ChannelMessageDelete(EVENTS_CHANNEL_ID, event.MessageID)
}

func ownerHasEvent(ctx context.Context, db types.Database, owner *discordgo.User) bool {
	res := db.Collection(types.EventsCollection).FindOne(ctx, bson.M{"owner": owner.ID, "status": types.EventStatusOpen})
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
