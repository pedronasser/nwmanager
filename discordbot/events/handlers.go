package events

import (
	"fmt"
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	. "nwmanager/helpers"
	"nwmanager/types"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var handlers = map[string]func(ctx *common.ModuleContext, i *discordgo.InteractionCreate){
	"/evento": func(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
		cfg := GetModuleConfig(ctx)
		if !slices.Contains(cfg.ChannelIDs, i.ChannelID) {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Este comando só pode ser usado em canal de eventos.", 5*time.Second)
			return
		}

		if !canCreateEvent(ctx, i.Member) {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você não possui permissão para criar eventos.", 5*time.Second)
			return
		}

		if ownerHasEvent(ctx, i.Member.User) {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você já possui um evento em andamento.", 5*time.Second)
			return
		}

		EventsData[i.Member.User.ID] = &types.Event{
			IsInviteOnly: false,
			Owner:        i.Member.User.ID,
			ChannelID:    i.ChannelID,
		}
		discordutils.SendInteractiveMessage(ctx.Session(), i, "evento:create", "Qual tipo de evento gostaria de iniciar?",
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "evento:create",
						MenuType:    discordgo.StringSelectMenu,
						Placeholder: "Clique aqui para selecionar o tipo de evento",
						MinValues:   Some(1),
						MaxValues:   *Some(1),
						Options:     discordutils.CreateSelectMenus(EVENT_TYPE_OPTIONS...),
					},
				},
			},
		)
	},

	"msg:create_event": func(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
		if !canCreateEvent(ctx, i.Member) {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você não possui permissão para criar eventos.", 5*time.Second)
			return
		}

		if ownerHasEvent(ctx, i.Member.User) {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você já possui um evento em andamento.", 5*time.Second)
			return
		}

		EventsData[i.Member.User.ID] = &types.Event{
			IsInviteOnly: false,
			Owner:        i.Member.User.ID,
			ChannelID:    i.ChannelID,
		}

		discordutils.SendInteractiveMessage(ctx.Session(), i, "evento:create", "Qual tipo de evento gostaria de iniciar?",
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "evento:create",
						MenuType:    discordgo.StringSelectMenu,
						Placeholder: "Clique aqui para selecionar o tipo de evento",
						MinValues:   Some(1),
						MaxValues:   *Some(1),
						Options:     discordutils.CreateSelectMenus(EVENT_TYPE_OPTIONS...),
					},
				},
			},
		)
	},

	"msg:create_closed_event": func(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
		if !canCreateEvent(ctx, i.Member) {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você não possui permissão para criar eventos.", 5*time.Second)
			return
		}

		if ownerHasEvent(ctx, i.Member.User) {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você já possui um evento em andamento.", 5*time.Second)
			return
		}

		EventsData[i.Member.User.ID] = &types.Event{
			IsInviteOnly: true,
			Owner:        i.Member.User.ID,
		}
		discordutils.SendInteractiveMessage(ctx.Session(), i, "evento:create", "Qual tipo de evento gostaria de iniciar?",
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "evento:create",
						MenuType:    discordgo.StringSelectMenu,
						Placeholder: "Clique aqui para selecionar o tipo de evento",
						MinValues:   Some(1),
						MaxValues:   *Some(1),
						Options:     discordutils.CreateSelectMenus(EVENT_TYPE_OPTIONS...),
					},
				},
			},
		)
	},

	// "/encerrar": func(s *discordgo.Session, i *discordgo.InteractionCreate, db database.Database) {
	// 	if !canCreateEvent(i.Member) {
	// 		discordutils.ReplyEphemeralMessage(s, i, "Você não possui permissão para encerrar eventos.", 5*time.Second)
	// 		return
	// 	}

	// 	ctx := context.Background()
	// 	res := db.Collection(globals.DB_PREFIX+types.EventsCollection).FindOne(ctx, bson.M{"owner": i.Member.User.ID, "status": types.EventStatusOpen})
	// 	if res.Err() != nil {
	// 		discordutils.ReplyEphemeralMessage(s, i, "Você não possui um evento em andamento.", 5*time.Second)
	// 		return
	// 	}

	// 	var event types.Event
	// 	err := res.Decode(&event)
	// 	if err != nil {
	// 		log.Fatalf("Cannot decode event: %v", err)
	// 	}

	// 	removeEvent(ctx, db, s, &event)
	// 	discordutils.ReplyEphemeralMessage(s, i, "**EVENTO ENCERRADO.**", 5*time.Second)
	// },

	"msg:evento:create": func(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
		if _, ok := EventsData[i.Member.User.ID]; !ok {
			return
		}
		data := i.MessageComponentData()
		tipo := types.EventType(data.Values[0])
		EventsData[i.Member.User.ID].Owner = i.Member.User.ID
		EventsData[i.Member.User.ID].Type = tipo
		discordutils.SendModal(ctx.Session(), i, "create", "Criação de evento",
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

		ctx.Session().InteractionResponseDelete(i.Interaction)
	},

	"msg:join_": func(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
		id := strings.TrimPrefix(i.MessageComponentData().CustomID, "join_")
		parts := strings.Split(id, "_")
		if len(parts) != 2 {
			return
		}

		classRune := parts[1]
		role := EventSlotRole(classRune[0])
		eventId := parts[0]

		event, err := types.GetEventByID(ctx, eventId)
		if err != nil {
			return
		}

		if event.Status != types.EventStatusOpen {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Este evento não está mais aberto para inscrições.", 5*time.Second)
			return
		}

		if event.IsInviteOnly && event.Owner != i.Member.User.ID {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Um pedido de entrada foi feito ao organizador deste evento.\nAguarde a aprovação.", 5*time.Second)
			sendJoinRequest(ctx, event, i.Member.User, role)
			return
		}

		if isUserAlreadyInEvent(event, i.Member.User.ID) {
			err = updateEventPlayerRole(ctx, i.Member.User, event, role)
		} else {
			err = addPlayerToEvent(ctx, i.Member.User.ID, event, role)
		}
		if err != nil {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, err.Error(), 5*time.Second)
			return
		}

		err = updateEventMessage(ctx, event)
		if err != nil {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Ocorreu um erro ao atualizar o evento.", 5*time.Second)
			log.Printf("Cannot update event message: %v", err)
			return
		}

		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você foi inscrito no evento.", 5*time.Second)
	},

	"msg:leave_": func(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
		eventId := strings.TrimPrefix(i.MessageComponentData().CustomID, "leave_")

		event, err := types.GetEventByID(ctx, eventId)
		if err != nil {
			return
		}

		if event.Status != types.EventStatusOpen {
			return
		}

		if !slices.Contains(event.PlayerSlots, i.Member.User.ID) {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você não está inscrito neste evento.", 5*time.Second)
			return
		}

		removePlayerFromEvent(ctx, i.Member.User.ID, event)
		updateEventMessage(ctx, event)

		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você foi removido do evento.", 5*time.Second)
	},

	"msg:approve_": func(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
		id := strings.TrimPrefix(i.MessageComponentData().CustomID, "approve_")
		parts := strings.Split(id, "_")
		if len(parts) != 3 {
			return
		}

		eventId := parts[0]
		userId := parts[1]
		slot := EventSlotRole(parts[2][0])

		event, err := types.GetEventByID(ctx, eventId)
		if err != nil {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Este evento não existe.", 5*time.Second)
			return
		}

		if event.Status != types.EventStatusOpen {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Este evento já foi encerrado.", 5*time.Second)
			return
		}

		err = addPlayerToEvent(ctx, userId, event, slot)
		if err != nil {
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, err.Error(), 5*time.Second)
			return
		}

		updateEventMessage(ctx, event)
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Jogador aprovado no evento.", 5*time.Second)

		discordutils.SendMemberDM(ctx.Session(), userId, fmt.Sprintf("Seu pedido de entrada no evento %s foi aprovado pelo organizador.", event.Title))

		ctx.Session().ChannelMessageDelete(i.ChannelID, i.Message.ID)
	},

	"msg:reject_": func(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
		id := strings.TrimPrefix(i.MessageComponentData().CustomID, "reject_")
		parts := strings.Split(id, "_")
		if len(parts) != 2 {
			return
		}

		eventId := parts[0]
		userId := parts[1]
		// slot := EventSlotRole(parts[2][0])

		event, err := types.GetEventByID(ctx, eventId)
		if err != nil {
			return
		}

		if event.Status != types.EventStatusOpen {
			return
		}

		err = discordutils.SendModal(ctx.Session(), i, fmt.Sprintf("reject_confirm_%s_%s", eventId, userId), "Confirmação de recusa",
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

	"modal:reject_confirm_": func(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
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

		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Jogador recusado no evento.", 5*time.Second)
		discordutils.SendMemberDM(ctx.Session(), userId, fmt.Sprintf("Seu pedido de entrada no evento foi recusado pelo organizador.\nMotivo: **%s**", motive))
		ctx.Session().ChannelMessageDelete(i.ChannelID, i.Message.ID)
	},

	"modal:create": func(ctx *common.ModuleContext, i *discordgo.InteractionCreate) {
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
					discordutils.ReplyEphemeralMessage(ctx.Session(), i, "A data/hora digitada está inválida.", 5*time.Second)
					return
				}
				scheduledAt = &d
			}

			go func() {
				err := createEvent(ctx, i, v.Type, v.ChannelID, title, description, scheduledAt, v.IsInviteOnly)
				if err != nil {
					log.Printf("Cannot create event: %v", err)
					discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Ocorreu um erro ao criar o evento.", 5*time.Second)
					return
				}
				discordutils.ReplyEphemeralMessage(ctx.Session(), i, fmt.Sprintf("**EVENTO CRIADO.**\n\nPara encerrar o evento envie **/encerrar**.\nVeja mais informações em <#%s>.", v.ChannelID), 5*time.Second)
			}()
		}
		delete(EventsData, i.Member.User.ID)
	},
}

var EventsData = map[string]*types.Event{}

func HandleMessages(ctx *common.ModuleContext, GuildID string) func(s *discordgo.Session, i *discordgo.MessageCreate) {
	cfg := GetModuleConfig(ctx)
	return func(s *discordgo.Session, i *discordgo.MessageCreate) {
		if GuildID != i.GuildID {
			return
		}

		if i.Author.ID == s.State.User.ID {
			return
		}

		if slices.Contains(cfg.ChannelIDs, i.ChannelID) {
			_ = s.ChannelMessageDelete(i.ChannelID, i.ID)
		}
	}
}

func HandleEventAction(ctx *common.ModuleContext, GuildID string) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cfg := GetModuleConfig(ctx)
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if GuildID != i.GuildID || !slices.Contains(cfg.ChannelIDs, i.ChannelID) {
			return
		}

		switch i.Type {
		case discordgo.InteractionMessageComponent:
			if id, found := strings.CutPrefix(i.MessageComponentData().CustomID, "close_event_"); found {
				handleEventClose(ctx, i, id)
			}

			if id, found := strings.CutPrefix(i.MessageComponentData().CustomID, "edit_"); found {
				handleEventEditPrompt(ctx, i, id)
				return
			}
		case discordgo.InteractionModalSubmit:
			if id, found := strings.CutPrefix(i.ModalSubmitData().CustomID, "edit_"); found {
				handleEventEdit(ctx, i, id)
				return
			}
		default:
		}
	}
}

func handleEventClose(ctx *common.ModuleContext, i *discordgo.InteractionCreate, id string) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Cannot convert id to ObjectID: %v", err)
	}
	res := ctx.DB().Collection(globals.DB_PREFIX+types.EventsCollection).FindOne(ctx.Context, bson.M{"_id": oid})
	if res.Err() != nil {
		log.Fatalf("Cannot find event: %v", res.Err())
		return
	}

	var event types.Event
	err = res.Decode(&event)
	if err != nil {
		log.Fatalf("Cannot decode event: %v", err)
	}

	if event.Owner != i.Member.User.ID && !globals.IsMemberAdmin(ctx, i.Member) {
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você não é o organizador do evento.", 5*time.Second)
		return
	}

	err = closeEvent(ctx, &event)
	if err != nil {
		log.Printf("Cannot close event: %v", err)
		return
	}

	discordutils.ReplyEphemeralMessage(ctx.Session(), i, "**EVENTO ENCERRADO.**", 5*time.Second)
}

func handleEventEditPrompt(ctx *common.ModuleContext, i *discordgo.InteractionCreate, id string) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Cannot convert id to ObjectID: %v", err)
	}
	res := ctx.DB().Collection(globals.DB_PREFIX+types.EventsCollection).FindOne(ctx.Context, bson.M{"_id": oid})
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
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você não é o organizador do evento.", 5*time.Second)
		return
	}

	var hora, data string
	if event.ScheduledAt != nil {
		hora = event.ScheduledAt.Format("15:04")
		data = event.ScheduledAt.Format("02/01/2006")
	}

	discordutils.SendModal(ctx.Session(), i, "edit_"+event.ID.Hex(), "Editando evento",
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

func handleEventEdit(ctx *common.ModuleContext, i *discordgo.InteractionCreate, id string) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Cannot convert id to ObjectID: %v", err)
	}
	res := ctx.DB().Collection(globals.DB_PREFIX+types.EventsCollection).FindOne(ctx.Context, bson.M{"_id": oid})
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
		discordutils.ReplyEphemeralMessage(ctx.Session(), i, "Você não é o organizador do evento.", 5*time.Second)
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
			discordutils.ReplyEphemeralMessage(ctx.Session(), i, "A data/hora digitada está inválida.", 5*time.Second)
			return
		}
		scheduledAt = &d
	}

	event.Title = title
	event.Description = description
	event.ScheduledAt = scheduledAt
	_, err = ctx.DB().Collection(globals.DB_PREFIX+types.EventsCollection).UpdateOne(ctx.Context,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"title": title, "description": description}},
	)
	if err != nil {
		log.Fatalf("Cannot update event: %v", err)
	}

	updateEventMessage(ctx, &event)
	discordutils.ReplyEphemeralMessage(ctx.Session(), i, "**EVENTO ATUALIZADO.**", 5*time.Second)
}

func setupEventsChannel(
	ctx *common.ModuleContext,
	channel_id string,
) (*discordgo.Channel, error) {
	events_channel, err := ctx.Session().Channel(channel_id)
	if err != nil {
		return nil, fmt.Errorf("Cannot get events channel: %v", err)
	}

	// if EVENTS_GUIDE_MESSAGE {
	// 	_, err = dg.ChannelMessageSendComplex(events_channel.ID, &discordgo.MessageSend{
	// 		Embed: &discordgo.MessageEmbed{
	// 			Title:       "Eventos Ativos",
	// 			Description: EVENTS_CHANNEL_INIT_MESSAGE,
	// 			Color:       0xcccccc,
	// 		},
	// 		Components: []discordgo.MessageComponent{
	// 			&discordgo.ActionsRow{
	// 				Components: []discordgo.MessageComponent{
	// 					&discordgo.Button{
	// 						Label:    "Novo Evento",
	// 						Style:    discordgo.PrimaryButton,
	// 						Emoji:    &discordgo.ComponentEmoji{Name: "🎮"},
	// 						CustomID: "create_event",
	// 					},
	// 					&discordgo.Button{
	// 						Label:    "Novo Evento Fechado",
	// 						Style:    discordgo.SecondaryButton,
	// 						Emoji:    &discordgo.ComponentEmoji{Name: "🔒"},
	// 						CustomID: "create_closed_event",
	// 					},
	// 				},
	// 			},
	// 		},
	// 	})
	// 	if err != nil {
	// 		log.Fatalf("Cannot send setup message: %v", err)
	// 	}
	// }

	channel_msgs, err := ctx.Session().ChannelMessages(events_channel.ID, 100, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("Cannot get channel messages: %v", err)
	}

	c, err := ctx.DB().Collection(globals.DB_PREFIX+types.EventsCollection).Find(ctx.Context, bson.M{
		"status":     types.EventStatusOpen,
		"channel_id": events_channel.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot query events: %v", err)
	}

	var events []types.Event
	if err = c.All(ctx.Context, &events); err != nil {
		return nil, fmt.Errorf("Cannot decode events: %v", err)
	}

	msgIDs := []string{}
	for _, event := range events {
		if event.MessageID != "" {
			msgIDs = append(msgIDs, event.MessageID)
		}
		if _, err := ctx.Session().ChannelMessage(events_channel.ID, event.MessageID); err == nil {
			continue
		}

		fmt.Println("Creating message for event:", event.ID.Hex())
		message, err := createEventMessage(ctx, events_channel, &event)
		if err != nil {
			return nil, fmt.Errorf("Cannot create event message: %v", err)
		}

		_, err = ctx.DB().Collection(globals.DB_PREFIX+types.EventsCollection).UpdateOne(ctx.Context, bson.M{
			"_id": event.ID,
		}, bson.M{
			"$set": bson.M{
				"message_id": message.ID,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("Cannot update event message id: %v", err)
		}
	}

	for _, msg := range channel_msgs {
		if slices.Contains(msgIDs, msg.ID) {
			continue
		}

		fmt.Println("Deleting message:", msg.ID, "from channel:", events_channel.ID)
		ctx.Session().ChannelMessageDelete(events_channel.ID, msg.ID)
	}

	return events_channel, nil
}
