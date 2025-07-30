package register

import (
	"log"
	"nwmanager/discordbot/common"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Message handler for IGN input and other text interactions
func HandleRegistrationMessage(ctx *common.ModuleContext) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore messages from bots
		if m.Author.Bot {
			return
		}

		// Check if user has an ongoing registration
		state, exists := RegisterData[m.Author.ID]
		if !exists {
			return
		}

		// Check if message is in the correct channel
		if m.ChannelID != state.TopicID {
			return
		}

		// Handle based on current step
		switch state.Step {
		case STEP_IGN:
			err := handleIGNInput(ctx, state, m)
			if err != nil {
				log.Printf("Error handling IGN input: %v", err)
			}
		}
	}
}

func handleIGNInput(ctx *common.ModuleContext, state *RegistrationState, m *discordgo.MessageCreate) error {
	dg := ctx.Session()

	// Validate IGN (basic validation)
	ign := strings.TrimSpace(m.Content)
	if len(ign) < 2 || len(ign) > 20 {
		_, err := dg.ChannelMessageSend(m.ChannelID, "❌ Nome inválido. O nome deve ter entre 2 e 20 caracteres.")
		return err
	}

	// Store IGN and move to next step
	state.IGN = ign
	state.Step = STEP_CLASSES

	// Ask for PVP classes
	return askForClasses(ctx, state.TopicID, state.DiscordID)
}

func askForClasses(ctx *common.ModuleContext, channelID, userID string) error {
	dg := ctx.Session()

	// Create class options from constants
	var classOptions []discordgo.SelectMenuOption
	for classKey, className := range PVP_CLASSES {
		classOptions = append(classOptions, discordgo.SelectMenuOption{
			Label: className,
			Value: classKey,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⚔️ Registro - Passo 2/4",
		Description: "**Quais das seguintes classes PVP você joga?**\n\nVocê pode selecionar várias opções.",
		Color:       0x0099ff,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "select:pvp_classes",
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: "Selecione suas classes de PvP",
					MinValues:   &[]int{1}[0],
					MaxValues:   len(classOptions),
					Options:     classOptions,
				},
			},
		},
	}

	_, err := dg.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})

	return err
}
