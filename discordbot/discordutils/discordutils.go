package discordutils

import (
	"fmt"
	"log"
	"nwmanager/database"
	"nwmanager/discordbot/globals"
	"slices"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func SendModal(s *discordgo.Session, i *discordgo.InteractionCreate, id string, title string, components ...discordgo.MessageComponent) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:      title,
			CustomID:   id,
			Components: components,
		},
	})
}

func SendInteractiveMessage(s *discordgo.Session, i *discordgo.InteractionCreate, id string, content string, components ...discordgo.MessageComponent) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    fmt.Sprintf(":robot:\n\n%s", content),
			CustomID:   id,
			Flags:      discordgo.MessageFlagsEphemeral,
			Components: components,
		},
	})
}

func CreateHandler(guildID string, channelID string, handlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, db database.Database), db database.Database) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		channel, err := s.Channel(i.ChannelID)
		if err != nil {
			return
		}

		if channel.Type == discordgo.ChannelTypeGuildText && (i.ChannelID != channelID || i.GuildID != guildID) {
			return
		}

		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := handlers["/"+i.ApplicationCommandData().Name]; ok {
				h(s, i, db)
				return
			}

			for h, handler := range handlers {
				if strings.HasPrefix("/"+i.ApplicationCommandData().Name, h) {
					handler(s, i, db)
					return
				}
			}
		case discordgo.InteractionMessageComponent:
			if h, ok := handlers["msg:"+i.MessageComponentData().CustomID]; ok {
				h(s, i, db)
				return
			}

			for h, handler := range handlers {
				if strings.HasPrefix("msg:"+i.MessageComponentData().CustomID, h) {
					handler(s, i, db)
					return
				}
			}

		case discordgo.InteractionModalSubmit:
			key := "modal:" + i.ModalSubmitData().CustomID
			if h, ok := handlers[key]; ok {
				h(s, i, db)
				return
			}

			for h, handler := range handlers {
				if strings.HasPrefix(key, h) {
					handler(s, i, db)
					break
				}
			}
		}
	}
}

func MentionUser(user *discordgo.User) string {
	return fmt.Sprintf("<@%s>", user.ID)
}

var knownRoles = map[string]discordgo.Role{}

func GetRoleByName(guild *discordgo.Guild, roleName string) *discordgo.Role {
	if role, ok := knownRoles[roleName]; ok {
		return &role
	}
	for _, role := range guild.Roles {
		if role.Name == roleName {
			knownRoles[roleName] = *role
			return role
		}
	}

	return nil
}

func ReplyEphemeralMessage(s *discordgo.Session, i *discordgo.InteractionCreate, content string, destroyDelay time.Duration) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf(":robot:\n\n%s", content),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	if err != nil {
		panic(err)
	}

	if destroyDelay > 0 {
		time.Sleep(destroyDelay)
	}

	err = s.InteractionResponseDelete(i.Interaction)
	if err != nil {
		panic(err)
	}
}

func SendMemberDM(s *discordgo.Session, userID string, content string) (*discordgo.Message, error) {
	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		return nil, err
	}

	msg, err := s.ChannelMessageSend(channel.ID, content)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func GetGuildMembers(s *discordgo.Session, guildID string, memberRoleID string) ([]*discordgo.Member, error) {
	members := []*discordgo.Member{}
	after := ""
	for {
		chunk, err := s.GuildMembers(guildID, after, 1000)
		if err != nil {
			return nil, err
		}

		actualMembers := []*discordgo.Member{}
		for _, member := range chunk {
			for _, role := range member.Roles {
				if role == memberRoleID {
					actualMembers = append(actualMembers, member)
					break
				}
			}
		}

		members = append(members, actualMembers...)
		if len(chunk) < 1000 {
			break
		}

		after = chunk[len(chunk)-1].User.ID
	}

	return members, nil
}

func GetMemberName(member *discordgo.Member) string {
	if member.Nick != "" {
		return member.Nick
	}
	if member.User.GlobalName != "" {
		return member.User.GlobalName
	}
	return member.User.Username
}

func IsMemberAdmin(member *discordgo.Member) bool {
	for _, role := range member.Roles {
		if globals.ACCESS_ROLE_IDS[globals.ADMIN_ROLE_NAME] != "" && role == globals.ACCESS_ROLE_IDS[globals.ADMIN_ROLE_NAME] {
			return true
		}
		if globals.ACCESS_ROLE_IDS[globals.CONSUL_ROLE_NAME] != "" && role == globals.ACCESS_ROLE_IDS[globals.CONSUL_ROLE_NAME] {
			return true
		}
		if globals.ACCESS_ROLE_IDS[globals.OFFICER_ROLE_NAME] != "" && role == globals.ACCESS_ROLE_IDS[globals.OFFICER_ROLE_NAME] {
			return true
		}
	}

	return false
}

func IsMember(member *discordgo.Member) bool {
	if globals.ACCESS_ROLE_IDS[globals.MEMBER_ROLE_NAME] == "" {
		panic("MEMBER_ROLE_NAME is not set")
	}

	return slices.Contains(member.Roles, globals.ACCESS_ROLE_IDS[globals.MEMBER_ROLE_NAME])
}

func RetrieveAllMembers(dg *discordgo.Session, GuildID string) map[string]*discordgo.Member {
	var members = map[string]*discordgo.Member{}

	mem, err := dg.GuildMembers(GuildID, "", 1000, discordgo.WithRetryOnRatelimit(true))
	for len(mem) > 0 {
		if err != nil {
			log.Fatalf("Cannot get guild members: %v", err)
		}
		fmt.Println("Got", len(mem), "members")
		for _, m := range mem {
			members[m.User.ID] = m
		}
		mem, err = dg.GuildMembers(GuildID, mem[len(mem)-1].User.ID, 1000, discordgo.WithRetryOnRatelimit(true))
	}

	return members
}

func ClearChannel(s *discordgo.Session, channelID string) {
	allMessages := []*discordgo.Message{}

	messages, err := s.ChannelMessages(channelID, 100, "", "", "")
	for len(messages) > 0 {
		if err != nil {
			log.Fatalf("Cannot get messages: %v", err)
		}
		fmt.Println("Cleared", len(messages), "messages")
		allMessages = append(allMessages, messages...)
		messages, err = s.ChannelMessages(channelID, 100, messages[len(messages)-1].ID, "", "")
	}

	for _, message := range allMessages {
		err = s.ChannelMessageDelete(channelID, message.ID)
		if err != nil {
			log.Fatalf("Cannot delete message: %v", err)
		}
	}
}
