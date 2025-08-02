package ticket

import (
	"context"
	"fmt"
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Helper functions

func isTicketOwner(ctx *common.ModuleContext, channelID, userID string) bool {
	// Get ticket from database to check ownership
	ticket, err := getTicketByChannelID(ctx, channelID)
	if err != nil || ticket == nil {
		return false
	}
	return ticket.DiscordID == userID
}

func createClassSelectOptions(classEmojis map[string]string) []discordgo.SelectMenuOption {
	var options []discordgo.SelectMenuOption

	for class, emoji := range classEmojis {
		options = append(options, discordgo.SelectMenuOption{
			Label: globals.PVP_CLASS_NAMES[globals.PVPClassType(class)],
			Value: class,
			Emoji: &discordgo.ComponentEmoji{Name: emoji},
		})
	}

	return options
}

func getPlayerByTicketChannel(ctx *common.ModuleContext, channelID string) (*types.Player, error) {
	// Get ticket first
	ticket, err := getTicketByChannelID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}
	if ticket == nil {
		return nil, fmt.Errorf("ticket not found")
	}

	// Get player by discord ID
	player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), ticket.DiscordID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	return player, nil
}

func updatePlayerClass(ctx *common.ModuleContext, player *types.Player, selectedClass string) error {
	// Update player's war class
	player.WarClass = selectedClass

	// Update in database
	err := types.UpdatePlayer(context.Background(), ctx.DB(), player)
	if err != nil {
		return fmt.Errorf("failed to update player class: %w", err)
	}

	return nil
}

func getCurrentBuildInfo(ctx *common.ModuleContext, player *types.Player) string {
	log.Printf("getCurrentBuildInfo called for player %s, BuildImages count: %d", player.IGN, len(player.BuildImages))

	// Basic player info
	classDisplay := player.WarClass
	if classDisplay == "" {
		classDisplay = "Não definida"
	}
	buildInfo := fmt.Sprintf("**Classe:** %s\n", classDisplay)

	// Check for last update time
	if player.Stats != nil && player.Stats.LastEquipUpdate != nil {
		buildInfo += fmt.Sprintf("**Última atualização:** %s\n", player.Stats.LastEquipUpdate.Format("02/01/2006 15:04"))
	}

	// Check for build images (BuildImages is directly on Player struct)
	if len(player.BuildImages) > 0 {
		log.Printf("Found %d build images for player %s", len(player.BuildImages), player.IGN)
		buildInfo += "\n**Screenshots da Build:**\n"
		for i, imageURL := range player.BuildImages {
			buildInfo += fmt.Sprintf("[Imagem %d](%s)\n", i+1, imageURL)
		}
	} else {
		log.Printf("No build images found for player %s", player.IGN)
		buildInfo += "\n**Nenhuma screenshot enviada ainda**"
	}

	return buildInfo
}

func setupBuildThread(ctx *common.ModuleContext, threadID, playerIGN string) {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📸 Build - %s", playerIGN),
		Description: "**Envie as screenshots da sua build aqui!**\n\nPode enviar várias imagens mostrando:\n• Equipamentos\n• Atributos\n• Skills\n• Armas\n\n*Suas imagens serão automaticamente salvas quando você confirmar a build.*\n\nQuando terminar, clique no botão abaixo para confirmar.",
		Color:       0x0099ff,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "✅ Confirmar Build",
					Style:    discordgo.SuccessButton,
					CustomID: "ticket:submit_build",
				},
				discordgo.Button{
					Label:    "❌ Fechar Thread",
					Style:    discordgo.SecondaryButton,
					CustomID: "ticket:close_thread",
				},
			},
		},
	}

	_, err := ctx.Session().ChannelMessageSendComplex(threadID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		log.Printf("Error setting up build thread: %v", err)
	}
}

func setupQuestionThread(ctx *common.ModuleContext, threadID, playerIGN string) {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("❓ Dúvidas - %s", playerIGN),
		Description: "**Envie suas dúvidas aqui!**\n\nPode perguntar sobre:\n• Builds e equipamentos\n• Estratégias de PvP\n• Eventos da guild\n• Qualquer outra coisa relacionada ao jogo\n\nQuando sua dúvida for resolvida, clique no botão para fechar a thread.",
		Color:       0xff9900,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "✅ Dúvida Resolvida",
					Style:    discordgo.SuccessButton,
					CustomID: "ticket:close_thread",
				},
			},
		},
	}

	_, err := ctx.Session().ChannelMessageSendComplex(threadID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		log.Printf("Error setting up question thread: %v", err)
	}
}

func setupTicketMessage(ctx *common.ModuleContext, channel *discordgo.Channel, player *types.Player) error {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎫 Ticket - %s", player.IGN),
		Description: fmt.Sprintf("Bem-vindo(a) ao seu ticket pessoal, **%s**!", player.IGN),
		Color:       0x00ff00,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use os botões abaixo para interagir com o ticket",
		},
	}

	components := []discordgo.MessageComponent{
		// discordgo.ActionsRow{
		// 	Components: []discordgo.MessageComponent{
		// 		discordgo.Button{
		// 			Label:    "Enviar Build",
		// 			Style:    discordgo.PrimaryButton,
		// 			CustomID: "ticket:send_build",
		// 			Emoji: &discordgo.ComponentEmoji{
		// 				Name: "📸",
		// 			},
		// 		},
		// 		discordgo.Button{
		// 			Label:    "Enviar Dúvida",
		// 			Style:    discordgo.SecondaryButton,
		// 			CustomID: "ticket:send_question",
		// 			Emoji: &discordgo.ComponentEmoji{
		// 				Name: "❓",
		// 			},
		// 		},
		// 	},
		// },
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Trocar Classe Guerra",
					Style:    discordgo.SecondaryButton,
					CustomID: "ticket:change_class",
					Emoji: &discordgo.ComponentEmoji{
						Name: "⚔️",
					},
				},
				// discordgo.Button{
				// 	Label:    "Ver Build Atual",
				// 	Style:    discordgo.SecondaryButton,
				// 	CustomID: "ticket:view_build",
				// 	Emoji: &discordgo.ComponentEmoji{
				// 		Name: "👁️",
				// 	},
				// },
			},
		},
	}

	message, err := ctx.Session().ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		return fmt.Errorf("failed to send ticket message: %w", err)
	}

	// Store ticket in database
	ticket := &Ticket{
		DiscordID:     player.DiscordID,
		ChannelID:     channel.ID,
		MessageID:     message.ID,
		PlayerIGN:     player.IGN,
		PlayerClass:   player.WarClass,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
		IsActive:      true,
	}

	err = insertTicket(ctx, ticket)
	if err != nil {
		log.Printf("Error storing ticket in database: %v", err)
		// Don't return error as the ticket channel was created successfully
	}

	return nil
}
