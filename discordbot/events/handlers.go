package events

import (
	"context"
	"fmt"
	"log"
	"nwmanager/discordbot/discordutils"
	. "nwmanager/helpers"
	"nwmanager/types"
	"slices"
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

		EventsData[i.Member.User.ID] = &types.Event{
			IsInviteOnly: false,
			Owner:        i.Member.User.ID,
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

		EventsData[i.Member.User.ID] = &types.Event{
			IsInviteOnly: false,
			Owner:        i.Member.User.ID,
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

	"msg:create_closed_event": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		ctx := context.Background()
		if ownerHasEvent(ctx, db, i.Member.User) {
			discordutils.ReplyEphemeralMessage(s, i, "Você já possui um evento em andamento.", 5*time.Second)
			return
		}

		EventsData[i.Member.User.ID] = &types.Event{
			IsInviteOnly: true,
			Owner:        i.Member.User.ID,
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
		if _, ok := EventsData[i.Member.User.ID]; !ok {
			return
		}
		data := i.MessageComponentData()
		tipo := types.EventType(data.Values[0])
		EventsData[i.Member.User.ID].Owner = i.Member.User.ID
		EventsData[i.Member.User.ID].Type = tipo
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

	"msg:join_": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		id := strings.TrimPrefix(i.MessageComponentData().CustomID, "join_")
		parts := strings.Split(id, "_")
		if len(parts) != 2 {
			return
		}

		classRune := parts[1]
		role := EventSlotRole(classRune[0])
		eventId := parts[0]

		ctx := context.Background()
		event, err := types.GetEventByID(ctx, db, eventId)
		if err != nil {
			return
		}

		if event.Status != types.EventStatusOpen {
			discordutils.ReplyEphemeralMessage(s, i, "Este evento não está mais aberto para inscrições.", 5*time.Second)
			return
		}

		if event.IsInviteOnly && event.Owner != i.Member.User.ID {
			discordutils.ReplyEphemeralMessage(s, i, "Um pedido de entrada foi feito ao organizador deste evento.\nAguarde a aprovação.", 5*time.Second)
			sendJoinRequest(ctx, db, s, event, i.Member.User, role)
			return
		}

		if isUserAlreadyInEvent(event, i.Member.User.ID) {
			err = updateEventPlayerRole(i.Member.User, db, event, role)
		} else {
			err = addPlayerToEvent(i.Member.User.ID, db, event, role)
		}
		if err != nil {
			discordutils.ReplyEphemeralMessage(s, i, err.Error(), 5*time.Second)
			return
		}
		updateEventMessage(s, event)

		discordutils.ReplyEphemeralMessage(s, i, "Você foi inscrito no evento.", 5*time.Second)
	},

	"msg:leave_": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		eventId := strings.TrimPrefix(i.MessageComponentData().CustomID, "leave_")

		ctx := context.Background()
		event, err := types.GetEventByID(ctx, db, eventId)
		if err != nil {
			return
		}

		if event.Status != types.EventStatusOpen {
			return
		}

		if !slices.Contains(event.PlayerSlots, i.Member.User.ID) {
			discordutils.ReplyEphemeralMessage(s, i, "Você não está inscrito neste evento.", 5*time.Second)
			return
		}

		removePlayerFromEvent(i.Member.User.ID, db, event)
		updateEventMessage(s, event)

		discordutils.ReplyEphemeralMessage(s, i, "Você foi removido do evento.", 5*time.Second)
	},

	"msg:approve_": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		id := strings.TrimPrefix(i.MessageComponentData().CustomID, "approve_")
		parts := strings.Split(id, "_")
		fmt.Println(parts)
		if len(parts) != 3 {
			return
		}

		eventId := parts[0]
		userId := parts[1]
		slot := EventSlotRole(parts[2][0])

		ctx := context.Background()
		event, err := types.GetEventByID(ctx, db, eventId)
		if err != nil {
			discordutils.ReplyEphemeralMessage(s, i, "Este evento não existe.", 5*time.Second)
			return
		}

		if event.Status != types.EventStatusOpen {
			discordutils.ReplyEphemeralMessage(s, i, "Este evento já foi encerrado.", 5*time.Second)
			return
		}

		err = addPlayerToEvent(userId, db, event, slot)
		if err != nil {
			discordutils.ReplyEphemeralMessage(s, i, err.Error(), 5*time.Second)
			return
		}

		updateEventMessage(s, event)
		discordutils.ReplyEphemeralMessage(s, i, "Jogador aprovado no evento.", 5*time.Second)

		discordutils.SendMemberDM(s, userId, fmt.Sprintf("Seu pedido de entrada no evento %s foi aprovado pelo organizador.", event.Title))

		s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
	},

	"msg:reject_": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		id := strings.TrimPrefix(i.MessageComponentData().CustomID, "reject_")
		parts := strings.Split(id, "_")
		if len(parts) != 2 {
			return
		}

		eventId := parts[0]
		userId := parts[1]
		// slot := EventSlotRole(parts[2][0])

		ctx := context.Background()
		event, err := types.GetEventByID(ctx, db, eventId)
		if err != nil {
			return
		}

		if event.Status != types.EventStatusOpen {
			return
		}

		err = discordutils.SendModal(s, i, fmt.Sprintf("reject_confirm_%s_%s", eventId, userId), "Confirmação de recusa",
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID: "motive",
						Label:    "Motivo da recusa (opcional)",
						Style:    discordgo.TextInputParagraph,
						Required: false,
					},
				},
			},
		)
		if err != nil {
			log.Fatalf("Cannot send modal: %v", err)
		}

	},

	"modal:reject_confirm_": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		id := strings.TrimPrefix(i.ModalSubmitData().CustomID, "reject_confirm_")
		parts := strings.Split(id, "_")
		if len(parts) != 2 {
			return
		}

		// eventId := parts[0]
		userId := parts[1]
		motive := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

		if motive == "" {
			motive = "Não informado."
		}

		discordutils.ReplyEphemeralMessage(s, i, "Jogador recusado no evento.", 5*time.Second)
		discordutils.SendMemberDM(s, userId, fmt.Sprintf("Seu pedido de entrada no evento foi recusado pelo organizador.\nMotivo: **%s**", motive))
		s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
	},

	"modal:create": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		data := i.ModalSubmitData()
		if v, ok := EventsData[i.Member.User.ID]; ok {
			title := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			description := data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			date := data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			hour := data.Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			var scheduledAt *time.Time
			if date != "" || hour != "" {
				if date == "" {
					date = time.Now().Format("02/01/2006")
				}
				if hour == "" {
					hour = time.Now().Format("15:04")
				}
				d, err := time.Parse("02/01/2006 15:04", date+" "+hour)
				if err != nil {
					discordutils.ReplyEphemeralMessage(s, i, "A data/hora digitada está inválida.", 5*time.Second)
					return
				}
				scheduledAt = &d
			}
			go createEvent(s, i, db, v.Type, title, description, scheduledAt, v.IsInviteOnly)
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

// func HandleReactionAdd(GuildID string, s *discordgo.Session, db types.Database) func(s *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
// 	return func(s *discordgo.Session, i *discordgo.MessageReactionAdd) {
// 		if GuildID != i.GuildID {
// 			return
// 		}

// 		if i.ChannelID != EVENTS_CHANNEL_ID || i.UserID == s.State.User.ID {
// 			return
// 		}

// 		_ = s.MessageReactionRemove(i.ChannelID, i.MessageID, i.Emoji.Name, i.UserID)

// 		ctx := context.Background()
// 		res := db.Collection(types.EventsCollection).FindOne(ctx, bson.M{"message_id": i.MessageID})
// 		if res.Err() != nil {
// 			return
// 		}

// 		var event types.Event
// 		err := res.Decode(&event)
// 		if err != nil {
// 			log.Fatalf("Cannot decode event: %v", err)
// 		}

// 		isPlayerInEvent := isUserAlreadyInEvent(&event, i.UserID)
// 		if i.Emoji.Name == "❌" && isPlayerInEvent {
// 			removePlayerFromEvent(i.Member.User, db, &event)
// 			updateEventMessage(s, &event)
// 			return
// 		}

// 		if isPlayerInEvent {
// 			return
// 		}

// 		role := resolveEventSlotFromEmoji(i.Emoji.Name)
// 		if role == EventSlotAny {
// 			return
// 		}

// 		if event.Status != types.EventStatusOpen {
// 			return
// 		}

// 		if event.IsInviteOnly {
// 			msg, err := discordutils.SendMemberDM(s, event.Owner, fmt.Sprintf("O jogador **%s** tentou se inscrever no evento **%s**. Clique em ✅ para aceitar ou ❌ para recusar.", i.Member.User.Mention(), event.Title))
// 			if err != nil {
// 				return
// 			}

// 			_ = s.MessageReactionAdd(msg.ChannelID, msg.ID, "✅")
// 			_ = s.MessageReactionAdd(msg.ChannelID, msg.ID, "❌")

// 			discordutils.SendMemberDM(s, i.Member.User.ID, fmt.Sprintf("Você tentou se inscrever no evento %s. Aguarde a aprovação do organizador.", event.Title))
// 			return
// 		}

// 		if EventSlots[event.Type] != "" {
// 			freeSlots := getEventFreeSlotsByRole(&event, role)
// 			if len(freeSlots) == 0 {
// 				return
// 			}

// 			slot := freeSlots[0]
// 			event.PlayerSlots[slot] = i.UserID

// 			remainingSlots := getEventFreeSlots(&event)
// 			if len(remainingSlots) == 0 {
// 				event.Status = types.EventStatusCompleted
// 				event.CompletedAt = Some(time.Now())
// 			}
// 		} else {
// 			event.PlayerSlots = append(event.PlayerSlots, i.UserID)
// 		}

// 		_, err = db.Collection(types.EventsCollection).UpdateOne(ctx, bson.M{"_id": event.ID}, bson.M{"$set": bson.M{"player_slots": event.PlayerSlots}})
// 		if err != nil {
// 			return
// 		}

// 		updateEventMessage(s, &event)
// 	}
// }

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
						Label:    "Novo Evento",
						Style:    discordgo.PrimaryButton,
						Emoji:    &discordgo.ComponentEmoji{Name: "🎮"},
						CustomID: "create_event",
					},
					&discordgo.Button{
						Label:    "Novo Evento Fechado",
						Style:    discordgo.SecondaryButton,
						Emoji:    &discordgo.ComponentEmoji{Name: "🔒"},
						CustomID: "create_closed_event",
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
