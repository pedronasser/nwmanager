package discordutils

import (
	"fmt"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func SendModal(s *discordgo.Session, i *discordgo.InteractionCreate, id string, title string, components ...discordgo.MessageComponent) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:      title,
			CustomID:   id,
			Components: components,
		},
	})
	if err != nil {
		panic(err)
	}
}

func SendInteractiveMessage(s *discordgo.Session, i *discordgo.InteractionCreate, id string, content string, components ...discordgo.MessageComponent) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    fmt.Sprintf(":robot:\n\n%s", content),
			CustomID:   id,
			Flags:      discordgo.MessageFlagsEphemeral,
			Components: components,
		},
	})
	if err != nil {
		panic(err)
	}
}

func CreateHandler(guildID string, channelID string, handlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database), db types.Database) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.GuildID != guildID || i.ChannelID != channelID {
			return
		}

		// d, _ := json.MarshalIndent(i, "", "\t")
		// fmt.Println(string(d))
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

	time.Sleep(destroyDelay)

	err = s.InteractionResponseDelete(i.Interaction)
	if err != nil {
		panic(err)
	}
}

func SendMemberDM(s *discordgo.Session, userID string, content string) error {
	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(channel.ID, content)
	if err != nil {
		return err
	}

	return nil
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
