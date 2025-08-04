package register

import (
	"log"
	"nwmanager/discordbot/common"

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

		// Check if current step is a text input step
		processor := GetStepProcessor()
		step := processor.GetStepByIndex(state.StepIndex)
		if step != nil && step.Type == StepTypeTextInput {
			err := processor.HandleStepResponse(ctx, state, state.StepIndex, m)
			if err != nil {
				log.Printf("Error handling step %d: %v", state.StepIndex, err)
			}
		}
	}
}
