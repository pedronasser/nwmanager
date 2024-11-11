package register

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"nwmanager/discordbot/constants"
	. "nwmanager/discordbot/helpers"
	"nwmanager/types"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
)

func completeRegister(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database) {
	ctx := context.Background()

	res := db.Collection(types.RegisterCollection).FindOne(ctx, bson.M{
		"discord_id": i.Member.User.ID,
	})
	if res.Err() != nil {
		log.Fatalf("Cannot find register: %v", res.Err())
	}

	var register types.Register
	err := res.Decode(&register)
	if err != nil {
		log.Fatalf("Cannot decode register: %v", err)
	}

	channels, err := s.GuildChannels(i.GuildID)
	if err != nil {
		log.Fatalf("Cannot get guild channels: %v", err)
	}

	var channel_id string
	for _, channel := range channels {
		if strings.ToLower(channel.Name) == strings.ToLower(register.InGameName) {
			channel_id = channel.ID
			break
		}
	}

	if channel_id == "" {
		channel, err := s.GuildChannelCreate(i.GuildID, register.InGameName, discordgo.ChannelTypeGuildText)
		if err != nil {
			log.Fatalf("Cannot create channel: %v", err)
		}

		channel_id = channel.ID
	}

	s.ChannelEdit(channel_id, &discordgo.ChannelEdit{
		ParentID: BOT_CHANNELS_CATEGORY,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:   EVERYONE_ROLE_ID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: discordgo.PermissionViewChannel,
			},
			{
				ID:    i.Member.User.ID,
				Type:  discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel,
			},
			{
				ID:    OFFICER_ROLE_ID,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: discordgo.PermissionViewChannel,
			},
		},
	})

	SendInteractiveMessage(s, i, "register:confirm", fmt.Sprintf("**INSCRIÇÃO REALIZADA.**\nAguarde o contato do bot ou de um de nossos oficiais em seu ticket: <#%s>", channel_id))
	s.ChannelMessageSend(channel_id, CreateRegisterInfoMessage(&register))
	s.ChannelMessageSend(channel_id, MentionUser(i.Member.User)+", este é seu TICKET.")
	if register.Approved {
		s.ChannelMessageSend(channel_id, "Seu registro foi aprovado.")
	} else {
		s.ChannelMessageSend(channel_id, "Seu registro está **pendente de aprovação** pois não encontramos seu nome de jogo presente na guild ou nos pedidos.")
	}
}

func sendGuildRule(s *discordgo.Session, i *discordgo.InteractionCreate, ruleNo int) {
	SendInteractiveMessage(s, i, "register:rule", fmt.Sprintf("\n**Regra %d**\n\n%s\n", ruleNo+1, GUILD_RULES[ruleNo]),
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: "register:rule:" + strconv.Itoa(ruleNo),
					Label:    "Entendi!",
					Style:    discordgo.PrimaryButton,
				},
			},
		},
	)
}

func getGuildRuleNumber(i *discordgo.InteractionCreate) int {
	customID := i.Interaction.MessageComponentData().CustomID
	parts := strings.Split(customID, ":")
	last := parts[len(parts)-1]
	ruleNo, _ := strconv.Atoi(last)
	return ruleNo
}

func CreateRegisterInfoMessage(r *types.Register) string {
	return fmt.Sprintf("**INSCRIÇÃO**\n\n"+
		"Armas: **%s\n**"+
		"Horários Ativo: **%s**\n"+
		"Dias da Semana Ativo: **%s**\n\n", printWeapons(r.Weapons), printTimes(r.Hours), printWeekDays(r.WeekDays))
}

func printWeapons(weapons []string) string {
	for i, weapon_name := range weapons {
		weapons[i] = constants.WEAPONS[weapon_name]
	}
	return strings.Join(weapons, ", ")
}

func printTimes(times []string) string {
	for i, time := range times {
		times[i] = constants.TIMES[time]
	}
	return strings.Join(times, ", ")
}

func printWeekDays(days []string) string {
	for i, day := range days {
		days[i] = constants.WEEKDAYS[day]
	}
	return strings.Join(days, ", ")
}
