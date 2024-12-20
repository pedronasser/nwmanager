package war

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
	"/guerra": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		discordutils.SendModal(s, i, "create", "Criação de evento de guerra",
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "territory",
						Label:       "Território da Guerra",
						Placeholder: "Exemplo: Queda eterna",
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
						Placeholder: "Exemplo: 25/11/2024",
						Style:       discordgo.TextInputShort,
						MinLength:   10,
						MaxLength:   10,
						Required:    true,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "hour",
						Label:       "Horário do Evento (formato: HH:mm)",
						Placeholder: "Exemplo: 16:59",
						Style:       discordgo.TextInputShort,
						MinLength:   5,
						MaxLength:   5,
						Required:    true,
					},
				},
			},
		)
	},

	"modal:create": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		data := i.ModalSubmitData()
		territory := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
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
		go createWar(s, i, db, territory, description, scheduledAt)
		discordutils.ReplyEphemeralMessage(s, i, "GUERRA CRIADA.", 5*time.Second)
	},
}

func ClearChannelMessages(GuildID string, s *discordgo.Session, db types.Database) func(s *discordgo.Session, i *discordgo.MessageCreate) {
	return func(s *discordgo.Session, i *discordgo.MessageCreate) {
		if GuildID != i.GuildID {
			return
		}

		if i.Author.ID == s.State.User.ID {
			return
		}

		if i.ChannelID == WAR_CHANNEL_ID {
			_ = s.ChannelMessageDelete(i.ChannelID, i.ID)
		}
	}
}

func HandleWarInteractions(GuildID string, s *discordgo.Session, db types.Database) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.ChannelID != WAR_CHANNEL_ID {
			return
		}
		if i.Type == discordgo.InteractionMessageComponent {
			if strings.HasPrefix(i.MessageComponentData().CustomID, "close_event_") {
				handleEventClose(s, i, db)
				return
			}

			if strings.HasPrefix(i.MessageComponentData().CustomID, "edit_") {
				id := strings.TrimPrefix(i.MessageComponentData().CustomID, "edit_")
				handleEventEditPrompt(s, i, db, id)
				return
			}

			if strings.HasPrefix(i.MessageComponentData().CustomID, "join_") {
				id := strings.TrimPrefix(i.MessageComponentData().CustomID, "join_")
				ctx := context.Background()

				oid, err := primitive.ObjectIDFromHex(id)
				if err != nil {
					log.Fatalf("Cannot convert id to ObjectID: %v", err)
				}

				war := getWarData(ctx, db, oid)
				if war == nil {
					log.Fatalf("Cannot find event: %v", err)
					return
				}

				if isUserAlreadyInEvent(war, i.Member.User.ID) {
					discordutils.ReplyEphemeralMessage(s, i, "Você já está inscrito neste evento.", 5*time.Second)
					return
				}

				JoinInteractions[i.Member.User.ID] = i.Interaction
				discordutils.SendInteractiveMessage(s, i, "evento:create", "**Inscrição de Guerra**",
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								CustomID:    fmt.Sprintf("selection_%s", id),
								MenuType:    discordgo.StringSelectMenu,
								Placeholder: "Selecione sua classe",
								MinValues:   Some(1),
								MaxValues:   *Some(1),
								Options:     getWarClassesOptions(),
							},
						},
					},
				)
				return
			}

			if strings.HasPrefix(i.MessageComponentData().CustomID, "selection_") {
				id := strings.TrimPrefix(i.MessageComponentData().CustomID, "selection_")
				handleEventJoin(s, i, db, id)
				return
			}

			if strings.HasPrefix(i.MessageComponentData().CustomID, "leave_") {
				id := strings.TrimPrefix(i.MessageComponentData().CustomID, "leave_")
				handleEventLeave(s, i, db, id)
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

var JoinInteractions = make(map[string]*discordgo.Interaction)

func handleEventJoin(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database, id string) {
	ctx := context.Background()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Cannot convert id to ObjectID: %v", err)
	}

	war := getWarData(ctx, db, oid)
	if war == nil {
		log.Fatalf("Cannot find event: %v", err)
		return
	}

	if war.Status != types.EventStatusOpen {
		discordutils.ReplyEphemeralMessage(s, i, "Este evento não está mais aberto para inscrições.", 5*time.Second)
		return
	}

	if isUserAlreadyInEvent(war, i.Member.User.ID) {
		discordutils.ReplyEphemeralMessage(s, i, "Você já está inscrito neste evento.", 5*time.Second)
		return
	}

	data := i.MessageComponentData()
	class := data.Values[0]
	if len(class) == 0 {
		discordutils.ReplyEphemeralMessage(s, i, "Selecione uma classe.", 5*time.Second)
		return
	}

	war.Players[i.Member.User.ID] = types.WarClass(class)
	_, err = db.Collection(types.WarsCollection).UpdateOne(ctx, bson.M{
		"_id": war.ID,
	}, bson.M{
		"$set": bson.M{
			"players." + i.Member.User.ID: class,
		},
	})

	if err != nil {
		log.Fatalf("Cannot update event: %v", err)
	}

	fmt.Println("Updating message")
	go updateEventMessage(s, war)

	discordutils.ReplyEphemeralMessage(s, i, "Você foi inscrito na guerra.", 5*time.Second)

	if interaction, ok := JoinInteractions[i.Member.User.ID]; ok {
		s.InteractionResponseDelete(interaction)
		delete(JoinInteractions, i.Member.User.ID)
	}
}

func handleEventClose(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
	id := strings.TrimPrefix(i.MessageComponentData().CustomID, "close_event_")
	ctx := context.Background()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Cannot convert id to ObjectID: %v", err)
	}
	res := db.Collection(types.WarsCollection).FindOne(ctx, bson.M{"_id": oid})
	if res.Err() != nil {
		log.Fatalf("Cannot find event: %v", res.Err())
		return
	}

	var war types.War
	err = res.Decode(&war)
	if err != nil {
		log.Fatalf("Cannot decode war: %v", err)
	}

	if !discordutils.IsMemberAdmin(i.Member) {
		discordutils.ReplyEphemeralMessage(s, i, "Você não tem permissão para fazer fechar este evento.", 5*time.Second)
		return
	}

	closeWar(ctx, db, s, &war)
	discordutils.ReplyEphemeralMessage(s, i, "**GUERRA ENCERRADA.**", 5*time.Second)
}

func handleEventEditPrompt(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database, id string) {
	ctx := context.Background()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Cannot convert id to ObjectID: %v", err)
	}
	res := db.Collection(types.WarsCollection).FindOne(ctx, bson.M{"_id": oid})
	if res.Err() != nil {
		log.Fatalf("Cannot find event: %v", res.Err())
		return
	}

	var war types.War
	err = res.Decode(&war)
	if err != nil {
		log.Fatalf("Cannot decode event: %v", err)
	}

	if !discordutils.IsMemberAdmin(i.Member) {
		discordutils.ReplyEphemeralMessage(s, i, "Você não tem permissão para editar esta guerra", 5*time.Second)
		return
	}

	discordutils.SendModal(s, i, "edit_"+war.ID.Hex(), "Editando Guerra",
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "territory",
					Label:       "Território da Guerra",
					Placeholder: "Exemplo: Queda Eterna",
					Style:       discordgo.TextInputShort,
					Required:    true,
					MaxLength:   50,
					Value:       war.Territory,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "description",
					Label:       "Descrição/requisitos do evento",
					Placeholder: "Exemplo: Ticket Atualizado",
					Style:       discordgo.TextInputParagraph,
					Required:    false,
					Value:       war.Description,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "date",
					Label:       "Dia do Evento (formato: DD/MM/AAAA)",
					Placeholder: "Exemplo: 25/11/2024",
					Style:       discordgo.TextInputShort,
					MinLength:   10,
					MaxLength:   10,
					Required:    true,
					Value:       war.ScheduledAt.Format("02/01/2006"),
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    "hour",
					Label:       "Horário do Evento (formato: HH:mm)",
					Placeholder: "Exemplo: 16:59",
					Style:       discordgo.TextInputShort,
					MinLength:   5,
					MaxLength:   5,
					Required:    true,
					Value:       war.ScheduledAt.Format("15:04"),
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
	res := db.Collection(types.WarsCollection).FindOne(ctx, bson.M{"_id": oid})
	if res.Err() != nil {
		log.Fatalf("Cannot find event: %v", res.Err())
		return
	}

	var war types.War
	err = res.Decode(&war)
	if err != nil {
		log.Fatalf("Cannot decode event: %v", err)
	}

	if !discordutils.IsMemberAdmin(i.Member) {
		discordutils.ReplyEphemeralMessage(s, i, "Você nao tem permissão para editar a guerra", 5*time.Second)
		return
	}

	data := i.ModalSubmitData()
	territory := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
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

	war.Territory = territory
	war.Description = description
	war.ScheduledAt = scheduledAt
	_, err = db.Collection(types.WarsCollection).UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"territory": territory, "description": description}},
	)
	if err != nil {
		log.Fatalf("Cannot update event: %v", err)
	}

	updateEventMessage(s, &war)
	discordutils.ReplyEphemeralMessage(s, i, "**GUERRA ATUALIZADA.**", 5*time.Second)
}

func handleEventLeave(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database, id string) {
	ctx := context.Background()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Fatalf("Cannot convert id to ObjectID: %v", err)
	}

	war := getWarData(ctx, db, oid)
	if war == nil {
		log.Fatalf("Cannot find event: %v", err)
		return
	}

	if war.Status != types.EventStatusOpen {
		discordutils.ReplyEphemeralMessage(s, i, "Este evento não está mais aberto para inscrições.", 5*time.Second)
		return
	}

	if !isUserAlreadyInEvent(war, i.Member.User.ID) {
		discordutils.ReplyEphemeralMessage(s, i, "Você não está inscrito neste evento.", 5*time.Second)
		return
	}

	delete(war.Players, i.Member.User.ID)

	_, err = db.Collection(types.WarsCollection).UpdateOne(ctx, bson.M{
		"_id": war.ID,
	}, bson.M{
		"$unset": bson.M{
			"players." + i.Member.User.ID: "",
		},
	})

	if err != nil {
		log.Fatalf("Cannot update event: %v", err)
	}

	updateEventMessage(s, war)
	discordutils.ReplyEphemeralMessage(s, i, "Você foi removido da guerra.", 5*time.Second)
}

func setupEventsChannel(
	ctx context.Context,
	dg *discordgo.Session,
	db types.Database,
	everyoneRole string,
	memberRole string,
) *discordgo.Channel {

	wars_channel, err := dg.ChannelEdit(WAR_CHANNEL_ID, &discordgo.ChannelEdit{
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
		log.Fatalf("Cannot edit events channel: %v", err)
	}

	msgs, err := dg.ChannelMessages(wars_channel.ID, 100, "", "", "")
	if err != nil {
		log.Fatalf("Cannot get channel messages: %v", err)
	}
	for _, msg := range msgs {
		err = dg.ChannelMessageDelete(wars_channel.ID, msg.ID)
		if err != nil {
			log.Fatalf("Cannot delete channel message: %v", err)
		}
	}

	// Query all events
	c, err := db.Collection(types.WarsCollection).Find(ctx, bson.M{
		"status": types.EventStatusOpen,
	})
	if err != nil {
		log.Fatalf("Cannot query events: %v", err)
	}

	var wars []types.War
	if err = c.All(ctx, &wars); err != nil {
		log.Fatalf("Cannot decode events: %v", err)
	}

	for _, war := range wars {
		message := createWarMessage(dg, wars_channel, &war)

		_, err = db.Collection(types.WarsCollection).UpdateOne(ctx, bson.M{
			"_id": war.ID,
		}, bson.M{
			"$set": bson.M{
				"message_id": message.ID,
			},
		})
		if err != nil {
			log.Fatalf("Cannot update event message id: %v", err)
		}
	}

	return wars_channel
}
