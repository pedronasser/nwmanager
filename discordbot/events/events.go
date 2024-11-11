package events

import (
	"context"
	"fmt"
	"log"
	. "nwmanager/discordbot/helpers"
	. "nwmanager/helpers"
	"nwmanager/types"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Setup(ctx context.Context, dg *discordgo.Session, AppID, GuildID *string, db types.Database) {
	guild, err := dg.State.Guild(*GuildID)
	if err != nil {
		log.Fatalf("Cannot get guild: %v", err)
	}

	everyoneRole := GetRoleByName(guild, "@everyone")
	if everyoneRole == nil {
		log.Fatalf("Cannot get @everyone role")
	}
	memberRole := GetRoleByName(guild, MEMBER_ROLE_NAME)
	if memberRole == nil {
		log.Fatalf("Cannot get %s role", MEMBER_ROLE_NAME)
	}

	_ = setupEventsChannel(ctx, dg, db, everyoneRole, memberRole)

	_, err = dg.ApplicationCommandCreate(*AppID, *GuildID, &discordgo.ApplicationCommand{
		Name:        "evento",
		Description: "Iniciar um novo evento",
	})
	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	_, err = dg.ApplicationCommandCreate(*AppID, *GuildID, &discordgo.ApplicationCommand{
		Name:        "encerrar",
		Description: "Encerre um evento",
	})
	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	dg.AddHandler(CreateHandler(handlers, db))
	dg.AddHandler(HandleReactionAdd(dg, db))
	dg.AddHandler(HandleMessages(dg, db))

	go func() {
		// Cleanup completed events
		for {
			time.Sleep(EVENT_CLEANUP_INTERVAL)
			fmt.Println("Cleaning up events...")
			ctx := context.Background()
			res, err := db.Collection(types.EventsCollection).Find(ctx, bson.M{})
			if err != nil {
				log.Fatalf("Cannot get events: %v", err)
			}
			for res.Next(ctx) {
				var event types.Event
				err := res.Decode(&event)
				if err != nil {
					log.Fatalf("Cannot decode event: %v", err)
				}

				if event.Status == types.EventStatusCompleted && event.CompletedAt.Add(EVENT_COMPLETE_EXPIRE_DURATION).Before(time.Now()) {
					removeEvent(ctx, db, dg, &event)
					fmt.Println("Event removed", event.ID)
					continue
				}

				if event.Status == types.EventStatusOpen && event.CreatedAt.Add(EVENT_MAX_DURATION).Before(time.Now()) {
					removeEvent(ctx, db, dg, &event)
					fmt.Println("Event removed", event.ID)
					continue
				}
			}
		}
	}()
}

var handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database){
	"/evento": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		ctx := context.Background()
		if ownerHasEvent(ctx, db, i.Member.User) {
			SendInteractiveMessage(s, i, "evento:already", "Você já possui um evento em andamento.")
			return
		}

		SendInteractiveMessage(s, i, "evento:create", "Qual tipo de evento gostaria de iniciar?",
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "evento:create",
						MenuType:    discordgo.StringSelectMenu,
						Placeholder: "Clique aqui para selecionar o tipo de evento",
						MinValues:   Some(1),
						MaxValues:   *Some(1),
						Options:     EVENT_TYPE_OPTIONS,
					},
				},
			},
		)
	},

	"/encerrar": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		ctx := context.Background()
		res := db.Collection(types.EventsCollection).FindOne(ctx, bson.M{"owner": i.Member.User.ID, "status": types.EventStatusOpen})
		if res.Err() != nil {
			SendInteractiveMessage(s, i, "evento:already", "Você não possui um evento em andamento.")
			return
		}

		var event types.Event
		err := res.Decode(&event)
		if err != nil {
			log.Fatalf("Cannot decode event: %v", err)
		}

		removeEvent(ctx, db, s, &event)

		SendInteractiveMessage(s, i, "evento:close", "**EVENTO ENCERRADO.**")
	},

	"msg:evento:create": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		data := i.MessageComponentData()
		tipo := data.Values[0]
		createEvent(s, i, db, tipo)
	},
}

func HandleMessages(s *discordgo.Session, db types.Database) func(s *discordgo.Session, i *discordgo.MessageCreate) {
	return func(s *discordgo.Session, i *discordgo.MessageCreate) {
		if i.Author.ID == s.State.User.ID {
			return
		}

		if i.ChannelID == EVENTS_CHANNEL_ID {
			_ = s.ChannelMessageDelete(i.ChannelID, i.ID)
		}
	}
}

func HandleReactionAdd(s *discordgo.Session, db types.Database) func(s *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	return func(s *discordgo.Session, i *discordgo.MessageReactionAdd) {
		if i.ChannelID != EVENTS_CHANNEL_ID || i.UserID == s.State.User.ID {
			return
		}

		_ = s.MessageReactionRemove(i.ChannelID, i.MessageID, i.Emoji.Name, i.UserID)

		ctx := context.Background()
		res := db.Collection(types.EventsCollection).FindOne(ctx, bson.M{"message_id": i.MessageID})
		if res.Err() != nil {
			return
		}

		var event types.Event
		err := res.Decode(&event)
		if err != nil {
			log.Fatalf("Cannot decode event: %v", err)
		}

		isPlayerInEvent := isUserAlreadyInEvent(&event, i.UserID)
		if i.Emoji.Name == "❌" && isPlayerInEvent {
			removePlayerFromEvent(i.Member.User, db, &event)
			updateEventMessage(s, &event)
			return
		}

		if isPlayerInEvent {
			return
		}

		role := resolveEventSlotFromEmoji(i.Emoji.Name)
		if role == EventSlotAny {
			return
		}

		if event.Status != types.EventStatusOpen {
			return
		}

		freeSlots := getEventFreeSlotsByRole(&event, role)
		if len(freeSlots) == 0 {
			return
		}

		slot := freeSlots[0]
		event.PlayerSlots[slot] = i.UserID

		remainingSlots := getEventFreeSlots(&event)
		if len(remainingSlots) == 0 {
			event.Status = types.EventStatusCompleted
			event.CompletedAt = Some(time.Now())
		}

		_, err = db.Collection(types.EventsCollection).UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"player_slots": event.PlayerSlots}})
		if err != nil {
			return
		}

		updateEventMessage(s, &event)
	}
}

func removePlayerFromEvent(u *discordgo.User, db types.Database, event *types.Event) {
	for i, slot := range event.PlayerSlots {
		if slot == u.ID {
			event.PlayerSlots[i] = ""
		}
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

	_, err = s.ChannelMessageEdit(events_channel.ID, message.ID, buildEventMessage(event))
	if err != nil {
		log.Fatalf("Cannot edit message: %v", err)
	}
}

func createEvent(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database, tipo string) {
	event := types.Event{
		ID:          primitive.NewObjectID(),
		Type:        types.EventType(tipo),
		Owner:       i.Interaction.Member.User.ID,
		Status:      types.EventStatusOpen,
		CreatedAt:   Some(time.Now()),
		PlayerSlots: []string{},
	}

	for i := 0; i < getEventSlotCount(event.Type); i++ {
		event.PlayerSlots = append(event.PlayerSlots, "")
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

	SendInteractiveMessage(s, i, "evento:confirm", fmt.Sprintf("**EVENTO CRIADO.**\n\nPara encerrar o evento envie **/encerrar**.\nVeja mais informações em <#%s>.", EVENTS_CHANNEL_ID))
}

func setupEventsChannel(
	ctx context.Context,
	dg *discordgo.Session,
	db types.Database,
	everyoneRole *discordgo.Role,
	memberRole *discordgo.Role,
) *discordgo.Channel {

	events_channel, err := dg.ChannelEdit(EVENTS_CHANNEL_ID, &discordgo.ChannelEdit{
		Name:     EVENTS_CHANNEL_NAME,
		Locked:   Some(true),
		Position: Some(1),
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:    everyoneRole.ID,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: 0,
				Deny:  discordgo.PermissionReadMessageHistory | discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionAddReactions,
			},
			{
				ID:    memberRole.ID,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: discordgo.PermissionReadMessageHistory | discordgo.PermissionAddReactions | discordgo.PermissionViewChannel | discordgo.PermissionUseSlashCommands | discordgo.PermissionSendMessages,
				Deny:  0,
			},
		},
	})
	if err != nil {
		log.Fatalf("Cannot edit welcome channel: %v", err)
	}

	msgs, err := dg.ChannelMessages(events_channel.ID, 100, "", "", "")
	if err != nil {
		log.Fatalf("Cannot get channel messages: %v", err)
	}
	for _, msg := range msgs {
		err = dg.ChannelMessageDelete(events_channel.ID, msg.ID)
		if err != nil {
			log.Fatalf("Cannot delete channel message: %v", err)
		}
	}

	_, err = dg.ChannelMessageSend(events_channel.ID, EVENTS_CHANNEL_INIT_MESSAGE)
	if err != nil {
		log.Fatalf("Cannot send setup message: %v", err)
	}

	// Query all events
	c, err := db.Collection(types.EventsCollection).Find(ctx, bson.M{
		"status": types.EventStatusOpen,
	})
	if err != nil {
		log.Fatalf("Cannot query events: %v", err)
	}

	var events []types.Event
	if err = c.All(ctx, &events); err != nil {
		log.Fatalf("Cannot decode events: %v", err)
	}

	for _, event := range events {
		message := createEventMessage(dg, events_channel, &event)

		_, err = db.Collection(types.EventsCollection).UpdateOne(ctx, bson.M{
			"_id": event.ID,
		}, bson.M{
			"$set": bson.M{
				"message_id": message.ID,
			},
		})
		if err != nil {
			log.Fatalf("Cannot update event message id: %v", err)
		}
	}

	return events_channel
}

func buildEventMessage(event *types.Event) string {
	msg := fmt.Sprintf("**%s**\nOrganizador: <@%s>\n\n**Vagas:**\n", getEventName(event.Type), event.Owner)
	for i, player := range event.PlayerSlots {
		role := getEventRoleNameByPosition(event.Type, i)
		playerName := "_[ABERTO]_"
		if player != "" {
			playerName = fmt.Sprintf("<@%s>", player)
		}
		msg += fmt.Sprintf("%s・%s\n", role, playerName)
		if (i+1)%5 == 0 {
			msg += "\n"
		}
	}
	msg += "・Reaja com 🛡️ para participar como **TANK**.\n"
	msg += "・Reaja com 🌿 para participar como **HEALER**.\n"
	msg += "・Reaja com ⚔️ para participar como **DPS**.\n"
	msg += "・Reaja com ❌ para sair do evento.\n"

	return msg
}

func createEventMessage(dg *discordgo.Session, events_channel *discordgo.Channel, event *types.Event) *discordgo.Message {
	message, err := dg.ChannelMessageSend(events_channel.ID, buildEventMessage(event))
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

func ownerHasEvent(ctx context.Context, db types.Database, owner *discordgo.User) bool {
	res := db.Collection(types.EventsCollection).FindOne(ctx, bson.M{"owner": owner.ID})
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
