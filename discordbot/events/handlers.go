package events

import (
	"context"
	"fmt"
	"log"
	"nwmanager/discordbot/discordutils"
	. "nwmanager/helpers"
	"nwmanager/types"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database){
	"/evento": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		ctx := context.Background()
		if ownerHasEvent(ctx, db, i.Member.User) {
			discordutils.ReplyEphemeralMessage(s, i, "Você já possui um evento em andamento.", 5*time.Second)
			return
		}

		discordutils.SendInteractiveMessage(s, i, "evento:create", "Qual tipo de evento gostaria de iniciar?",
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
			discordutils.ReplyEphemeralMessage(s, i, "Você já possui um evento em andamento.", 5*time.Second)
			return
		}

		discordutils.SendInteractiveMessage(s, i, "evento:create", "Qual tipo de evento gostaria de iniciar?",
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
			discordutils.ReplyEphemeralMessage(s, i, "Você não possui um evento em andamento.", 5*time.Second)
			return
		}

		var event types.Event
		err := res.Decode(&event)
		if err != nil {
			log.Fatalf("Cannot decode event: %v", err)
		}

		removeEvent(ctx, db, s, &event)
		discordutils.ReplyEphemeralMessage(s, i, "**EVENTO ENCERRADO.**", 5*time.Second)
	},

	"msg:evento:create": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		data := i.MessageComponentData()
		tipo := types.EventType(data.Values[0])
		EventsData[i.Member.User.ID] = &types.Event{
			Owner: i.Member.User.ID,
			Type:  tipo,
		}
		discordutils.SendModal(s, i, "create", "Criação de evento",
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "title",
						Label:       "Título do evento",
						Placeholder: "Exemplo: Gorgonas dos CLT",
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
						Placeholder: "Exemplo: GS 700+",
						Style:       discordgo.TextInputParagraph,
						Required:    false,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "date",
						Label:       "Dia do Evento (formato: DD/MM/AAAA)",
						Placeholder: "Exemplo: 25/11/2024 (deixe vazio se não for agendado)",
						Style:       discordgo.TextInputShort,
						MinLength:   10,
						MaxLength:   10,
						Required:    false,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "hour",
						Label:       "Horário do Evento (formato: HH:mm)",
						Placeholder: "Exemplo: 16:59 (deixe vazio se não for agendado)",
						Style:       discordgo.TextInputShort,
						MinLength:   5,
						MaxLength:   5,
						Required:    false,
					},
				},
			},
		)

		s.InteractionResponseDelete(i.Interaction)
	},

	"modal:create": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		fmt.Println("modal:create", EventsData)
		data := i.ModalSubmitData()
		if v, ok := EventsData[i.Member.User.ID]; ok {
			title := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			description := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			date := data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			hour := data.Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			var scheduledAt *time.Time
			if date != "" && hour != "" {
				d, err := time.Parse("02/01/2006 15:04", date+" "+hour)
				if err != nil {
					discordutils.ReplyEphemeralMessage(s, i, "A data/hora digitada está inválida.", 5*time.Second)
					return
				}
				scheduledAt = &d
			}
			go createEvent(s, i, db, v.Type, title, description, scheduledAt)
			discordutils.ReplyEphemeralMessage(s, i, fmt.Sprintf("**EVENTO CRIADO.**\n\nPara encerrar o evento envie **/encerrar**.\nVeja mais informações em <#%s>.", EVENTS_CHANNEL_ID), 5*time.Second)
		}
		delete(EventsData, i.Member.User.ID)
	},
}

var EventsData = map[string]*types.Event{}

func HandleMessages(GuildID string, s *discordgo.Session, db types.Database) func(s *discordgo.Session, i *discordgo.MessageCreate) {
	return func(s *discordgo.Session, i *discordgo.MessageCreate) {
		if GuildID != i.GuildID || i.ChannelID != EVENTS_CHANNEL_ID {
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
		if GuildID != i.GuildID || i.ChannelID != EVENTS_CHANNEL_ID {
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

	if event.Owner != i.Member.User.ID && !discordutils.IsMemberAdmin(i.Member) {
		discordutils.ReplyEphemeralMessage(s, i, "Você não é o organizador do evento.", 5*time.Second)
		return
	}

	closeEvent(ctx, db, s, &event)
	discordutils.ReplyEphemeralMessage(s, i, "**EVENTO ENCERRADO.**", 5*time.Second)
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
		discordutils.ReplyEphemeralMessage(s, i, "Você não é o organizador do evento.", 5*time.Second)
		return
	}

	var hora, data string
	if event.ScheduledAt != nil {
		hora = event.ScheduledAt.Format("15:04")
		data = event.ScheduledAt.Format("02/01/2006")
	}

	discordutils.SendModal(s, i, "edit_"+event.ID.Hex(), "Editando evento",
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
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "date",
					Label:       "Dia do Evento (formato: DD/MM/AAAA)",
					Placeholder: "Exemplo: 25/11/2024 (deixe vazio se não for agendado)",
					Style:       discordgo.TextInputShort,
					MinLength:   10,
					MaxLength:   10,
					Required:    false,
					Value:       data,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "hour",
					Label:       "Horário do Evento (formato: HH:mm)",
					Placeholder: "Exemplo: 16:59 (deixe vazio se não for agendado)",
					Style:       discordgo.TextInputShort,
					MinLength:   5,
					MaxLength:   5,
					Required:    false,
					Value:       hora,
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
		discordutils.ReplyEphemeralMessage(s, i, "Você não é o organizador do evento.", 5*time.Second)
		return
	}

	data := i.ModalSubmitData()
	title := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	description := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	date := data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	hour := data.Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	var scheduledAt *time.Time
	if date != "" && hour != "" {
		d, err := time.Parse("02/01/2006 15:04", date+" "+hour)
		if err != nil {
			discordutils.ReplyEphemeralMessage(s, i, "A data/hora digitada está inválida.", 5*time.Second)
			return
		}
		scheduledAt = &d
	}

	event.Title = title
	event.Description = description
	event.ScheduledAt = scheduledAt
	_, err = db.Collection(types.EventsCollection).UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"title": title, "description": description}},
	)
	if err != nil {
		log.Fatalf("Cannot update event: %v", err)
	}

	updateEventMessage(s, &event)
	discordutils.ReplyEphemeralMessage(s, i, "**EVENTO ATUALIZADO.**", 5*time.Second)
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
	everyoneRole string,
	memberRole string,
) *discordgo.Channel {

	events_channel, err := dg.ChannelEdit(EVENTS_CHANNEL_ID, &discordgo.ChannelEdit{
		Name:   EVENTS_CHANNEL_NAME,
		Locked: Some(true),
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:    everyoneRole,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: 0,
				Deny:  discordgo.PermissionReadMessageHistory | discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionAddReactions,
			},
			{
				ID:    memberRole,
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
