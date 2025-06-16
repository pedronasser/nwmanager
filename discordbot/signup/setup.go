package signup

import (
	"context"
	"fmt"
	"nwmanager/database"
	"nwmanager/discordbot/discordutils"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	WELCOME_CHANNEL_ID = ""
)

func init() {
	_ = godotenv.Load()
	WELCOME_CHANNEL_ID = os.Getenv("WELCOME_CHANNEL_ID")
}

func Setup(ctx context.Context, dg *discordgo.Session, AppID, GuildID *string, db database.Database) {
	if WELCOME_CHANNEL_ID == "" {
		fmt.Println("signup: WELCOME_CHANNEL_ID is not set")
		return
	}

	discordutils.ClearChannel(dg, WELCOME_CHANNEL_ID)
	dg.ChannelMessageSendComplex(WELCOME_CHANNEL_ID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       "Bem-vindo(a) à **CLAVE**.",
			Description: "Por favor, selecione umas das opções abaixo:",
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						CustomID: "signup_recruit",
						Label:    "Gostaria de ser recrutado!",
						Style:    discordgo.SecondaryButton,
						Emoji:    &discordgo.ComponentEmoji{Name: "🤝"},
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						CustomID: "signup_complete",
						Label:    "Estou aqui como complete de guerra",
						Style:    discordgo.SecondaryButton,
						Emoji:    &discordgo.ComponentEmoji{Name: "🧩"},
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						CustomID: "signup_friend",
						Label:    "Quero jogar com um amigo",
						Style:    discordgo.SecondaryButton,
						Emoji:    &discordgo.ComponentEmoji{Name: "👥"},
					},
				},
			},
		},
	})
}
