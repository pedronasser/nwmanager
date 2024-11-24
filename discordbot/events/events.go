package events

import (
	"context"
	"fmt"
	"log"
	. "nwmanager/discordbot/helpers"
	. "nwmanager/helpers"
	"nwmanager/types"
	"strings"
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

	// for t, roleName := range EventRoles {
	// 	role := GetRoleByName(guild, roleName)
	// 	if role == nil {
	// 		log.Fatalf("Cannot get %s role", roleName)
	// 	}
	// 	EventRoles[t] = role.ID
	// }

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

	dg.AddHandler(CreateHandler(*GuildID, handlers, db))
	dg.AddHandler(HandleReactionAdd(*GuildID, dg, db))
	dg.AddHandler(HandleMessages(*GuildID, dg, db))
	dg.AddHandler(HandleEventClose(*GuildID, dg, db))

	go eventsCleanup(db, dg)
}

func eventsCleanup(db types.Database, dg *discordgo.Session) {
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
				closeEvent(ctx, db, dg, &event)
				fmt.Println("Event closed", event.ID, "due to expiration", event.CompletedAt, time.Now())
				continue
			}

			// now := time.Now()
			// if event.CreatedAt != nil && event.ClosedAt == nil {
			// 	if now.Day() != event.CreatedAt.Day() && now.Hour() > 6 {
			// 		closeEvent(ctx, db, dg, &event)
			// 		fmt.Println("Event closed", event.ID, "due to day change", *event.CreatedAt, now)
			// 		continue
			// 	}

			// 	// if event.Status == types.EventStatusOpen && event.CreatedAt.Add(EVENT_MAX_DURATION).Before(time.Now()) {
			// 	// 	closeEvent(ctx, db, dg, &event)
			// 	// 	fmt.Println("Event closed", event.ID, "due to max duration", *event.CreatedAt, now)
			// 	// 	continue
			// 	// }
			// }
		}
	}
}

var handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database){
	"/evento": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		ctx := context.Background()
		if ownerHasEvent(ctx, db, i.Member.User) {
			ReplyEphemeralMessage(s, i, "Você já possui um evento em andamento.", 5*time.Second)
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

	"msg:create_event": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		ctx := context.Background()
		if ownerHasEvent(ctx, db, i.Member.User) {
			ReplyEphemeralMessage(s, i, "Você já possui um evento em andamento.", 5*time.Second)
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
			ReplyEphemeralMessage(s, i, "Você não possui um evento em andamento.", 5*time.Second)
			return
		}

		var event types.Event
		err := res.Decode(&event)
		if err != nil {
			log.Fatalf("Cannot decode event: %v", err)
		}

		removeEvent(ctx, db, s, &event)
		ReplyEphemeralMessage(s, i, "**EVENTO ENCERRADO.**", 5*time.Second)
	},

	"msg:evento:create": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		data := i.MessageComponentData()
		tipo := types.EventType(data.Values[0])
		EventData[i.Member.User.ID] = &types.Event{
			Owner: i.Member.User.ID,
			Type:  tipo,
		}
		SendModal(s, i, "create", "Criação de evento",
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "title",
						Label:       "Título do evento",
						Placeholder: "Exemplo: Gorgonas às 17h30",
						Style:       discordgo.TextInputShort,
						Required:    true,
						MaxLength:   50,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "description",
						Label:       "Descrição/requisitos do evento",
						Placeholder: "Exemplo: GS 695+",
						Style:       discordgo.TextInputParagraph,
						Required:    false,
					},
				},
			},
		)

		s.InteractionResponseDelete(i.Interaction)
	},

	"modal:create": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		fmt.Println("modal:create", EventData)
		data := i.ModalSubmitData()
		if v, ok := EventData[i.Member.User.ID]; ok {
			title := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			description := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			createEvent(s, i, db, v.Type, title, description)
		}
		delete(EventData, i.Member.User.ID)
	},
}

var EventData = map[string]*types.Event{}

func HandleMessages(GuildID string, s *discordgo.Session, db types.Database) func(s *discordgo.Session, i *discordgo.MessageCreate) {
	return func(s *discordgo.Session, i *discordgo.MessageCreate) {
		if GuildID != i.GuildID {
			return
		}

		if i.Author.ID == s.State.User.ID {
			return
		}

		if i.ChannelID == EVENTS_CHANNEL_ID {
			_ = s.ChannelMessageDelete(i.ChannelID, i.ID)
		}
	}
}

func HandleEventClose(GuildID string, s *discordgo.Session, db types.Database) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if GuildID != i.GuildID {
			return
		}
		if i.Type == discordgo.InteractionMessageComponent {
			fmt.Println("msg", i.MessageComponentData().CustomID)
			if strings.HasPrefix(i.MessageComponentData().CustomID, "close_event_") {
				handleEventClose(s, i, db)
				return
			}

			if strings.HasPrefix(i.MessageComponentData().CustomID, "edit_") {
				id := strings.TrimPrefix(i.MessageComponentData().CustomID, "edit_")
				handleEventEditPrompt(s, i, db, id)
				return
			}
		} else if i.Type == discordgo.InteractionModalSubmit {
			fmt.Println("modal", i.ModalSubmitData().CustomID)
			if strings.HasPrefix(i.ModalSubmitData().CustomID, "edit_") {
				id := strings.TrimPrefix(i.ModalSubmitData().CustomID, "edit_")
				handleEventEdit(s, i, db, id)
				return
			}
		}
	}
}

func handleEventClose(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
	id := strings.TrimPrefix(i.MessageComponentData().CustomID, "close_event_")
	ctx := context.Background()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Cannot convert id to ObjectID: %v", err)
	}
	res := db.Collection(types.EventsCollection).FindOne(ctx, bson.M{"_id": oid})
	if res.Err() != nil {
		log.Fatalf("Cannot find event: %v", res.Err())
		return
	}

	var event types.Event
	err = res.Decode(&event)
	if err != nil {
		log.Fatalf("Cannot decode event: %v", err)
	}

	if event.Owner != i.Member.User.ID {
		ReplyEphemeralMessage(s, i, "Você não é o organizador do evento.", 5*time.Second)
		return
	}

	removeEvent(ctx, db, s, &event)
	ReplyEphemeralMessage(s, i, "**EVENTO ENCERRADO.**", 5*time.Second)
}

func handleEventEditPrompt(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database, id string) {
	ctx := context.Background()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Cannot convert id to ObjectID: %v", err)
	}
	res := db.Collection(types.EventsCollection).FindOne(ctx, bson.M{"_id": oid})
	if res.Err() != nil {
		log.Fatalf("Cannot find event: %v", res.Err())
		return
	}

	var event types.Event
	err = res.Decode(&event)
	if err != nil {
		log.Fatalf("Cannot decode event: %v", err)
	}

	if event.Owner != i.Member.User.ID {
		ReplyEphemeralMessage(s, i, "Você não é o organizador do evento.", 5*time.Second)
		return
	}

	SendModal(s, i, "edit_"+event.ID.Hex(), "Editando evento",
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "title",
					Label:       "Título do evento",
					Placeholder: "Exemplo: Gorgonas às 17h30",
					Style:       discordgo.TextInputShort,
					Required:    true,
					MaxLength:   50,
					Value:       event.Title,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "description",
					Label:       "Descrição/requisitos do evento",
					Placeholder: "Exemplo: GS 695+",
					Style:       discordgo.TextInputParagraph,
					Required:    false,
					Value:       event.Description,
				},
			},
		},
	)
}

func handleEventEdit(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database, id string) {
	ctx := context.Background()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Cannot convert id to ObjectID: %v", err)
	}
	res := db.Collection(types.EventsCollection).FindOne(ctx, bson.M{"_id": oid})
	if res.Err() != nil {
		log.Fatalf("Cannot find event: %v", res.Err())
		return
	}

	var event types.Event
	err = res.Decode(&event)
	if err != nil {
		log.Fatalf("Cannot decode event: %v", err)
	}

	if event.Owner != i.Member.User.ID {
		ReplyEphemeralMessage(s, i, "Você não é o organizador do evento.", 5*time.Second)
		return
	}

	data := i.ModalSubmitData()
	title := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	description := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	event.Title = title
	event.Description = description
	_, err = db.Collection(types.EventsCollection).UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"title": title, "description": description}},
	)
	if err != nil {
		log.Fatalf("Cannot update event: %v", err)
	}

	updateEventMessage(s, &event)
	ReplyEphemeralMessage(s, i, "**EVENTO ATUALIZADO.**", 5*time.Second)
}

func HandleReactionAdd(GuildID string, s *discordgo.Session, db types.Database) func(s *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	return func(s *discordgo.Session, i *discordgo.MessageReactionAdd) {
		if GuildID != i.GuildID {
			return
		}

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

		if EventSlots[event.Type] != "" {
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
		} else {
			event.PlayerSlots = append(event.PlayerSlots, i.UserID)
		}

		_, err = db.Collection(types.EventsCollection).UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"player_slots": event.PlayerSlots}})
		if err != nil {
			return
		}

		updateEventMessage(s, &event)
	}
}

func setupEventsChannel(
	ctx context.Context,
	dg *discordgo.Session,
	db types.Database,
	everyoneRole *discordgo.Role,
	memberRole *discordgo.Role,
) *discordgo.Channel {

	events_channel, err := dg.ChannelEdit(EVENTS_CHANNEL_ID, &discordgo.ChannelEdit{
		Name:   EVENTS_CHANNEL_NAME,
		Locked: Some(true),
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

	_, err = dg.ChannelMessageSendComplex(events_channel.ID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       "Eventos Ativos",
			Description: EVENTS_CHANNEL_INIT_MESSAGE,
			Color:       0xcccccc,
		},
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Label:    "Criar Evento",
						Style:    discordgo.PrimaryButton,
						Emoji:    &discordgo.ComponentEmoji{Name: "🎮"},
						CustomID: "create_event",
					},
				},
			},
		},
	})
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

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s - %s", getEventName(event.Type), event.Title),
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
