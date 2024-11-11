package register

import (
	"context"
	"fmt"
	"log"
	. "nwmanager/discordbot/helpers"
	. "nwmanager/helpers"
	"nwmanager/types"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var guildRulesMsg string

func init() {
	for no, rule := range GUILD_RULES {
		guildRulesMsg += fmt.Sprintf(":number_%d: • %s\n", no+1, rule)
	}
}

func Setup(dg *discordgo.Session, AppID, GuildID *string, db types.Database) {
	guild, err := dg.State.Guild(*GuildID)
	if err != nil {
		log.Fatalf("Cannot get guild: %v", err)
	}

	var everyoneRoleID string
	for _, role := range guild.Roles {
		if role.Name == "@everyone" {
			everyoneRoleID = role.ID
			break
		}
	}

	welcome_channel, err := dg.ChannelEdit(WELCOME_CHANNEL_ID, &discordgo.ChannelEdit{
		Name:     "👋・entrada",
		Locked:   Some(true),
		Position: Some(1),
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:    everyoneRoleID,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: discordgo.PermissionReadMessageHistory,
				Deny:  discordgo.PermissionSendMessages | discordgo.PermissionAddReactions,
			},
		},
	})
	if err != nil {
		log.Fatalf("Cannot edit welcome channel: %v", err)
	}

	msgs, err := dg.ChannelMessages(welcome_channel.ID, 100, "", "", "")
	if err != nil {
		log.Fatalf("Cannot get welcome channel messages: %v", err)
	}
	for _, msg := range msgs {
		err = dg.ChannelMessageDelete(welcome_channel.ID, msg.ID)
		if err != nil {
			log.Fatalf("Cannot delete welcome channel message: %v", err)
		}
	}

	_, err = dg.ChannelMessageSend(welcome_channel.ID, "Digite **/inscrever** para se inscrever na guild.")
	if err != nil {
		log.Fatalf("Cannot send welcome message: %v", err)
	}

	_, err = dg.ApplicationCommandCreate(*AppID, *GuildID, &discordgo.ApplicationCommand{
		Name:        "inscrever",
		Description: "Inscrever-se na guild",
	})
	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	dg.AddHandler(CreateHandler(handlers, db))
}

var handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database){
	"/inscrever": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		SendModal(s, i, "register", "Inscrição da Guild",
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:  "name",
						Label:     "Qual é o seu nome in-game?",
						Style:     discordgo.TextInputShort,
						Required:  true,
						MaxLength: 15,
						MinLength: 3,
					},
				},
			},
		)
	},

	"modal:register": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		data := i.ModalSubmitData()
		ctx := context.Background()

		db.Collection(types.RegisterCollection).DeleteOne(ctx, bson.M{"discord_id": i.Member.User.ID})

		nickname := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		register := &types.Register{
			ID:         primitive.NewObjectID(),
			DiscordID:  i.Member.User.ID,
			InGameName: nickname,
		}
		res, err := db.Collection(types.RegisterCollection).InsertOne(ctx, register)
		if err != nil {
			log.Fatalf("Cannot insert register: %v", err)
		}

		register.ID = res.InsertedID.(primitive.ObjectID)

		SendInteractiveMessage(s, i, "register:weapons", fmt.Sprintf("Olá, **%s**.\n\nSelecione as **armas de seu personagem**: :arrow_down:", nickname),
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "register:weapons",
						MenuType:    discordgo.StringSelectMenu,
						MinValues:   &MIN_WEAPON_OPTIONS,
						MaxValues:   len(WEAPON_OPTIONS),
						Placeholder: "Clique aqui para selecionar suas armas",
						Options:     WEAPON_OPTIONS,
					},
				},
			},
		)
	},

	"msg:register:weapons": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		data := i.MessageComponentData()

		res, err := db.Collection(types.RegisterCollection).UpdateOne(context.Background(), bson.M{"discord_id": i.Member.User.ID}, bson.M{
			"$set": bson.M{"weapons": data.Values},
		})
		if err != nil {
			log.Fatalf("Cannot update register: %v", err)
		}

		if res.ModifiedCount == 0 {
			log.Fatalf("No register updated")
		}

		SendInteractiveMessage(s, i, "register:time", "Quais **horários** você tem disponibilidade de jogar? :arrow_down:",
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "register:time",
						MenuType:    discordgo.StringSelectMenu,
						MaxValues:   len(TIME_OPTIONS),
						Placeholder: "Clique aqui para selecionar seus horários",
						Options:     TIME_OPTIONS,
					},
				},
			},
		)
	},

	"msg:register:time": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		data := i.MessageComponentData()

		res, err := db.Collection(types.RegisterCollection).UpdateOne(context.Background(), bson.M{"discord_id": i.Member.User.ID}, bson.M{
			"$set": bson.M{"hours": data.Values},
		})
		if err != nil {
			log.Fatalf("Cannot update register: %v", err)
		}

		if res.ModifiedCount == 0 {
			log.Fatalf("Cannot update register: %v", err)
		}

		SendInteractiveMessage(s, i, "register:time", "Quais **dias da semana** você tem disponibilidade de jogar? :arrow_down:\n",
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.SelectMenu{
						CustomID:    "register:days",
						MenuType:    discordgo.StringSelectMenu,
						MaxValues:   len(WEEK_DAYS_OPTIONS),
						Placeholder: "Clique aqui para selecionar seus dias da semana",
						Options:     WEEK_DAYS_OPTIONS,
					},
				},
			},
		)
	},

	"msg:register:days": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		data := i.MessageComponentData()

		res, err := db.Collection(types.RegisterCollection).UpdateOne(context.Background(), bson.M{"discord_id": i.Member.User.ID}, bson.M{
			"$set": bson.M{"week_days": data.Values},
		})
		if err != nil {
			log.Fatalf("Cannot update register: %v", err)
		}

		if res.ModifiedCount == 0 {
			log.Fatalf("Cannot update register: %v", err)
		}

		SendInteractiveMessage(s, i, "register:time", fmt.Sprintf(
			"**REGRAS DA GUILD**\n\nLeia as regras com atenção:\n%s\n¹ Ticket é um canal de texto no discord com o nome do seu personagem que será criado após sua inscrição realizada.\n\nVocê aceita as regras acima?",
			guildRulesMsg),
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						CustomID: "register:rules:accept",
						Label:    "Aceito as regras",
						Style:    discordgo.SuccessButton,
					},
					discordgo.Button{
						CustomID: "register:rules:reject",
						Label:    "Que regras?",
						Style:    discordgo.DangerButton,
					},
				},
			},
		)
	},

	"msg:register:rule": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		ruleNo := getGuildRuleNumber(i)
		ruleNo++
		if ruleNo >= len(GUILD_RULES) {
			completeRegister(s, i, db)
			return
		}
		sendGuildRule(s, i, ruleNo)
	},

	"msg:register:rules:accept": func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
		completeRegister(s, i, db)
	},
}
