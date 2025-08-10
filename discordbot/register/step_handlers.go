package register

import (
	"fmt"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/globals"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// IGN Step - Text Input
func createIGNStep(ctx *common.ModuleContext, state *RegistrationState) (*discordgo.MessageSend, error) {
	processor := GetStepProcessor()
	embed := &discordgo.MessageEmbed{
		Title:       processor.GetStepTitle(state.StepIndex),
		Description: "**Qual é o seu nome no jogo (IGN)?**\n\nPor favor, digite seu nome exatamente como aparece no New World.",
		Color:       0x0099ff,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Digite sua resposta na próxima mensagem",
		},
	}

	return &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
	}, nil
}

func handleIGNStep(ctx *common.ModuleContext, state *RegistrationState, interaction interface{}) error {
	if msg, ok := interaction.(*discordgo.MessageCreate); ok {
		// Validate IGN
		ign := strings.TrimSpace(msg.Content)
		if len(ign) < 2 || len(ign) > 20 {
			dg := ctx.Session()
			_, err := dg.ChannelMessageSend(msg.ChannelID, "❌ Nome inválido. O nome deve ter entre 2 e 20 caracteres.")
			return err
		}

		// Store IGN and move to next step
		state.IGN = ign
		state.StepIndex++

		// Process next step
		processor := GetStepProcessor()
		return processor.ProcessStep(ctx, state, state.StepIndex)
	}
	return nil
}

func validateIGN(input interface{}) error {
	if msg, ok := input.(*discordgo.MessageCreate); ok {
		ign := strings.TrimSpace(msg.Content)
		if len(ign) < 2 || len(ign) > 20 {
			return fmt.Errorf("nome inválido. O nome deve ter entre 2 e 20 caracteres")
		}
	}
	return nil
}

// PVP Classes Step - Select Menu
func createPVPClassesStep(ctx *common.ModuleContext, state *RegistrationState) (*discordgo.MessageSend, error) {
	processor := GetStepProcessor()
	globalCfg, _ := ctx.Config("globals").(*globals.GlobalsConfig)
	var classOptions []discordgo.SelectMenuOption
	for _, classKey := range PVP_CLASS_OPTIONS {
		className := globals.PVP_CLASS_NAMES[classKey]
		classEmoji := globalCfg.ClassEmojiIDs[string(classKey)]
		option := discordgo.SelectMenuOption{
			Label: className,
			Value: string(classKey),
		}

		if classEmoji != "" {
			option.Emoji = &discordgo.ComponentEmoji{
				Name: classEmoji,
			}
		}
		classOptions = append(classOptions, option)
	}

	embed := &discordgo.MessageEmbed{
		Title:       processor.GetStepTitle(state.StepIndex),
		Description: "**Qual das seguintes classes você pretende jogar Guerra?**",
		Color:       0x0099ff,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "select:pvp_classes",
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: "Selecione sua classe de Guerra",
					MinValues:   &[]int{1}[0],
					MaxValues:   1,
					Options:     classOptions,
				},
			},
		},
	}

	return &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	}, nil
}

func handlePVPClassesStep(ctx *common.ModuleContext, state *RegistrationState, interaction interface{}) error {
	if i, ok := interaction.(*discordgo.InteractionCreate); ok {
		// Store selected classes
		values := i.MessageComponentData().Values
		pvpClasses := make([]globals.PVPClassType, len(values))
		for idx, value := range values {
			pvpClasses[idx] = globals.PVPClassType(value)
		}
		state.PVPClasses = pvpClasses
		state.StepIndex++

		// Move to next step
		processor := GetStepProcessor()
		return processor.ProcessStep(ctx, state, state.StepIndex)
	}
	return nil
}

// Times Step - Select Menu
func createTimesStep(ctx *common.ModuleContext, state *RegistrationState) (*discordgo.MessageSend, error) {
	processor := GetStepProcessor()
	var timeOptions []discordgo.SelectMenuOption
	for _, timeKey := range TIME_OPTIONS {
		timeName := TIMES[timeKey]
		timeOptions = append(timeOptions, discordgo.SelectMenuOption{
			Label: timeName,
			Value: timeKey,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       processor.GetStepTitle(state.StepIndex),
		Description: "**Quais horários você costuma jogar?**\n\nSelecione todos os horários em que você geralmente está disponível.",
		Color:       0x0099ff,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "select:times",
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: "Selecione seus horários disponíveis",
					MinValues:   &[]int{1}[0],
					MaxValues:   len(timeOptions),
					Options:     timeOptions,
				},
			},
		},
	}

	return &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	}, nil
}

func handleTimesStep(ctx *common.ModuleContext, state *RegistrationState, interaction interface{}) error {
	if i, ok := interaction.(*discordgo.InteractionCreate); ok {
		// Store selected times
		state.Times = i.MessageComponentData().Values
		state.StepIndex++

		// Move to next step
		processor := GetStepProcessor()
		return processor.ProcessStep(ctx, state, state.StepIndex)
	}
	return nil
}

// Weekdays Step - Select Menu
func createWeekdaysStep(ctx *common.ModuleContext, state *RegistrationState) (*discordgo.MessageSend, error) {
	processor := GetStepProcessor()
	var weekdayOptions []discordgo.SelectMenuOption
	for _, weekdayKey := range WEEKDAY_OPTIONS {
		weekdayName := WEEKDAYS[weekdayKey]
		weekdayOptions = append(weekdayOptions, discordgo.SelectMenuOption{
			Label: weekdayName,
			Value: weekdayKey,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       processor.GetStepTitle(state.StepIndex),
		Description: "**Quais dias da semana você costuma jogar?**\n\nSelecione todos os dias em que você geralmente joga.",
		Color:       0x0099ff,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "select:weekdays",
					MenuType:    discordgo.StringSelectMenu,
					Placeholder: "Selecione os dias da semana",
					MinValues:   &[]int{1}[0],
					MaxValues:   len(weekdayOptions),
					Options:     weekdayOptions,
				},
			},
		},
	}

	return &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	}, nil
}

func handleWeekdaysStep(ctx *common.ModuleContext, state *RegistrationState, interaction interface{}) error {
	if i, ok := interaction.(*discordgo.InteractionCreate); ok {
		// Store selected weekdays
		state.Weekdays = i.MessageComponentData().Values
		state.StepIndex++

		// Move to next step
		processor := GetStepProcessor()
		return processor.ProcessStep(ctx, state, state.StepIndex)
	}
	return nil
}

// War Experience Step - Button/Modal
func createWarExperienceStep(ctx *common.ModuleContext, state *RegistrationState) (*discordgo.MessageSend, error) {
	processor := GetStepProcessor()

	embed := &discordgo.MessageEmbed{
		Title:       processor.GetStepTitle(state.StepIndex),
		Description: "**Você já participou de Guerras no New World? Se sim, em qual guild você jogou sua última guerra?**\n\nClique em uma das opções abaixo:",
		Color:       0x0099ff,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Sim, já participei",
					Style:    discordgo.PrimaryButton,
					CustomID: "btn:war_yes",
				},
				discordgo.Button{
					Label:    "Não participei",
					Style:    discordgo.SecondaryButton,
					CustomID: "btn:war_no",
				},
			},
		},
	}

	return &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	}, nil
}

func handleWarExperienceStep(ctx *common.ModuleContext, state *RegistrationState, interaction interface{}) error {
	if i, ok := interaction.(*discordgo.InteractionCreate); ok {
		customID := i.MessageComponentData().CustomID

		switch customID {
		case "btn:war_yes":
			// User has war experience, show modal for guild name
			state.HasWarExperience = true

			modal := discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "modal:guild_name",
					Title:    "Experiência em Guerras",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "guild_name_input",
									Label:       "Em qual guild você lutou sua última guerra?",
									Style:       discordgo.TextInputShort,
									Placeholder: "Digite o nome da guild",
									Required:    true,
									MaxLength:   100,
								},
							},
						},
					},
				},
			}

			dg := ctx.Session()
			return dg.InteractionRespond(i.Interaction, &modal)

		case "btn:war_no":
			// User has no war experience
			state.HasWarExperience = false
			state.PreviousGuildName = ""

			// This is the last step, complete registration
			completeRegistration(ctx, state, i)
			return nil
		}
	}
	return nil
}
