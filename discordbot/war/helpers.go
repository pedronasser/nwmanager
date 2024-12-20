package war

import (
	"context"
	"fmt"
	"log"
	"nwmanager/discordbot/discordutils"
	"nwmanager/types"
	"strings"
	"time"

	. "nwmanager/helpers"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func buildWarMessage(event *types.War) *discordgo.MessageEmbed {
	desc := ""

	classes := map[string][]string{}
	players := 0
	for player, class := range event.Players {
		if _, ok := classes[string(class)]; !ok {
			classes[string(class)] = []string{}
		}

		classes[string(class)] = append(classes[string(class)], fmt.Sprintf("<@%s>", player))
		players++
	}

	desc += "\n```"
	if event.Description != "" {
		desc += fmt.Sprintf("\n%s\n", event.Description)
	}
	desc += "```\n"

	// TODO: add players list here by class

	// footer := "・Reaja com 🛡️ para participar como TANK.\n"
	// footer += "・Reaja com 🌿 para participar como HEALER.\n"
	// footer += "・Reaja com ⚔️ para participar como DPS.\n"
	// footer += "・Reaja com ❌ para sair do evento.\n"

	scheduled := ""
	if event.ScheduledAt != nil {
		scheduled = fmt.Sprintf(" (%s)", (*event.ScheduledAt).Format("02/01/2006 às 15:04"))
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:  "Jogadores Incritos",
			Value: fmt.Sprintf("%d\n", players),
		},
	}

	for class, players := range classes {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s%s", WarClassEmojis[types.WarClass(class)], WarClassNames[types.WarClass(class)]),
			Value:  fmt.Sprintf("%s", strings.Join(players, "\n")),
			Inline: true,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Guerra em **%s**%s", event.Territory, scheduled),
		Description: desc,
		// Footer: &discordgo.MessageEmbedFooter{
		// 	Text: footer,
		// },
		Color: 0x00ff00,
		// Thumbnail: &discordgo.MessageEmbedThumbnail{
		// 	URL: "https://dqzvgunkova5o.cloudfront.net/statics/2024-10-31/images/NWA_logo.png",
		// },
		Fields: fields,
	}

	return embed
}

func removePlayerFromEvent(u *discordgo.User, db types.Database, war *types.War) {
	delete(war.Players, u.ID)

	ctx := context.Background()
	_, err := db.Collection(types.WarsCollection).UpdateOne(ctx, bson.M{"_id": war.ID}, bson.M{"$set": bson.M{"players": war.Players}})
	if err != nil {
		return
	}
}

func updateEventMessage(s *discordgo.Session, event *types.War) {
	events_channel, err := s.Channel(WAR_CHANNEL_ID)
	if err != nil {
		log.Fatalf("Cannot get events channel: %v", err)
	}

	message, err := s.ChannelMessage(events_channel.ID, event.MessageID)
	if err != nil {
		log.Fatalf("Cannot get message: %v", err)
	}

	_, err = s.ChannelMessageEditEmbed(events_channel.ID, message.ID, buildWarMessage(event))
	if err != nil {
		log.Fatalf("Cannot edit message: %v", err)
	}
}

func createWar(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database, territory, description string, scheduledAt *time.Time) {
	war := types.War{
		ID:          primitive.NewObjectID(),
		Territory:   territory,
		Description: description,
		Status:      types.EventStatusOpen,
		CreatedAt:   Some(time.Now()),
		Players:     map[string]types.WarClass{},
		ScheduledAt: scheduledAt,
	}

	war_channel, err := s.Channel(WAR_CHANNEL_ID)
	if err != nil {
		log.Fatalf("Cannot get events channel: %v", err)
	}

	message := createWarMessage(s, war_channel, &war)
	war.MessageID = message.ID

	ctx := context.Background()
	_, err = db.Collection(types.WarsCollection).InsertOne(ctx, war)
	if err != nil {
		log.Fatalf("Cannot insert event: %v", err)
	}
}

func createWarMessage(dg *discordgo.Session, events_channel *discordgo.Channel, event *types.War) *discordgo.Message {
	message, err := dg.ChannelMessageSendComplex(events_channel.ID,
		&discordgo.MessageSend{
			Embed: buildWarMessage(event),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Inscrever-se",
							Style:    discordgo.SuccessButton,
							CustomID: fmt.Sprintf("join_%s", event.ID.Hex()),
							Emoji: &discordgo.ComponentEmoji{
								Name: "✅",
							},
						},
						discordgo.Button{
							Label:    "Desinscrever",
							Style:    discordgo.PrimaryButton,
							CustomID: fmt.Sprintf("leave_%s", event.ID.Hex()),
							Emoji: &discordgo.ComponentEmoji{
								Name: "❌",
							},
						},
						discordgo.Button{
							Label:    "Editar",
							Style:    discordgo.SecondaryButton,
							CustomID: fmt.Sprintf("edit_%s", event.ID.Hex()),
							Emoji: &discordgo.ComponentEmoji{
								Name: "✏️",
							},
						},
						discordgo.Button{
							Label:    "Encerrar",
							Style:    discordgo.DangerButton,
							CustomID: fmt.Sprintf("close_event_%s", event.ID.Hex()),
							Emoji: &discordgo.ComponentEmoji{
								Name: "🗑️",
							},
						},
					},
				},
			},
		},
	)
	if err != nil {
		log.Fatalf("Cannot send event message: %v", err)
	}

	return message
}

func isUserAlreadyInEvent(event *types.War, userID string) bool {
	if _, ok := event.Players[userID]; ok {
		return true
	}
	return false
}

func removeEvent(ctx context.Context, db types.Database, s *discordgo.Session, event *types.War) {
	_, err := db.Collection(types.WarsCollection).DeleteOne(ctx, bson.M{"_id": event.ID})
	if err != nil {
		log.Fatalf("Cannot delete event: %v", err)
	}

	_ = s.ChannelMessageDelete(WAR_CHANNEL_ID, event.MessageID)
}

func closeWar(ctx context.Context, db types.Database, s *discordgo.Session, event *types.War) {
	_, err := db.Collection(types.WarsCollection).UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"status": types.EventStatusClosed, "closed_at": time.Now()}})
	if err != nil {
		log.Fatalf("Cannot update event: %v", err)
	}

	_ = s.ChannelMessageDelete(WAR_CHANNEL_ID, event.MessageID)
}

func hasActiveWar(ctx context.Context, db types.Database) bool {
	res := db.Collection(types.WarsCollection).FindOne(ctx, bson.M{"status": types.EventStatusOpen})
	if res.Err() != nil {
		return false
	}

	var event types.War
	err := res.Decode(&event)
	if err != nil {
		return false
	}

	return true
}

func getWarClassesOptions() []discordgo.SelectMenuOption {
	options := []discordgo.SelectMenuOption{}
	for id, className := range WarClassNames {
		options = append(options, discordgo.SelectMenuOption{
			Label: className,
			Value: string(id),
			Emoji: &discordgo.ComponentEmoji{
				Name: WarClassEmojis[id],
			},
		})
	}
	return options
}

func getWarData(ctx context.Context, db types.Database, oid primitive.ObjectID) *types.War {
	res := db.Collection(types.WarsCollection).FindOne(ctx, bson.M{
		"_id": oid,
	})
	if res.Err() != nil {
		return nil
	}

	var war types.War
	err := res.Decode(&war)
	if err != nil {
		return nil
	}

	return &war
}

func sendReminderNotification(ctx context.Context, db types.Database, dg *discordgo.Session, war *types.War) {
	now := GetCurrentTimeAsUTC()

	fmt.Println("Sending notifications for war", war)
	go func(dg *discordgo.Session) {
		for player, _ := range war.Players {
			if player != "" {
				discordutils.SendMemberDM(dg, player, fmt.Sprintf("A guerra em **%s** na qual você está inscrito(a) iniciará em **15 minutos**. Verifique o canal de guerras para mais informações.", war.Territory))
			}
		}
	}(dg)

	_, err := db.Collection(types.WarsCollection).UpdateOne(ctx, bson.M{"_id": war.ID}, bson.M{"$set": bson.M{"notified_at": now}})
	if err != nil {
		log.Fatalf("Cannot update war: %v", err)
	}
}

func sendConfirmationRequest(ctx context.Context, db types.Database, dg *discordgo.Session, war *types.War, members []*discordgo.Member) {
	now := GetCurrentTimeAsUTC()

	fmt.Println("Sending confirmation request for war", war)
	go func(dg *discordgo.Session) {
		for _, member := range members {
			if _, exists := war.Players[member.User.ID]; !exists {
				discordutils.SendMemberDM(dg, member.User.ID, fmt.Sprintf("Há uma guerra em **%s** está marcada para iniciar em **6 horas**. Por favor, confirme sua presença no evento no canal de guerra ou notifique sua ausencia no canal de ausências", war.Territory))
			}
		}
	}(dg)

	_, err := db.Collection(types.WarsCollection).UpdateOne(ctx, bson.M{"_id": war.ID}, bson.M{"$set": bson.M{"confirmed_at": now}})
	if err != nil {
		log.Fatalf("Cannot update war: %v", err)
	}
}
