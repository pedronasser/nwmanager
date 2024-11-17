package helpers

import (
	"fmt"
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

func CreateHandler(guildID string, handlers map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, db types.Database), db types.Database) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.GuildID != guildID {
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
