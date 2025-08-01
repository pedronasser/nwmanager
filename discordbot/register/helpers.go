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

		// Also handle using new step system if CurrentStepID is set
		if state.CurrentStepID != "" {
			processor := GetStepProcessor()
			step := processor.GetStepByID(state.CurrentStepID)
			if step != nil && step.Type == StepTypeTextInput {
				err := processor.HandleStepResponse(ctx, state, state.CurrentStepID, m)
				if err != nil {
					log.Printf("Error handling step %s: %v", state.CurrentStepID, err)
				}
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
	state.CurrentStepID = "pvp_classes"

	// Use the new step system to process the next step
	processor := GetStepProcessor()
	return processor.ProcessStep(ctx, state, "pvp_classes")
}
