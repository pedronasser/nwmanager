package register

import (
	"fmt"
	"nwmanager/discordbot/common"

	"github.com/bwmarrin/discordgo"
)

// StepType represents the type of step interaction
type StepType string

const (
	StepTypeTextInput  StepType = "text_input"
	StepTypeSelectMenu StepType = "select_menu"
	StepTypeButton     StepType = "button"
	StepTypeCompletion StepType = "completion"
)

// StepDefinition defines a registration step
type StepDefinition struct {
	Name      string
	Type      StepType
	Creator   StepCreator
	Handler   StepHandler
	Validator StepValidator
}

// StepCreator creates the step message/components
type StepCreator func(ctx *common.ModuleContext, state *RegistrationState) (*discordgo.MessageSend, error)

// StepHandler handles the step response
type StepHandler func(ctx *common.ModuleContext, state *RegistrationState, interaction interface{}) error

// StepValidator validates the step input (optional)
type StepValidator func(input interface{}) error

// StepProcessor manages the step flow
type StepProcessor struct {
	steps []StepDefinition
}

// NewStepProcessor creates a new step processor with the registration steps
func NewStepProcessor() *StepProcessor {
	return &StepProcessor{
		steps: []StepDefinition{
			{
				Name:      "In-Game Name",
				Type:      StepTypeTextInput,
				Creator:   createIGNStep,
				Handler:   handleIGNStep,
				Validator: validateIGN,
			},
			{
				Name:    "PVP Classes",
				Type:    StepTypeSelectMenu,
				Creator: createPVPClassesStep,
				Handler: handlePVPClassesStep,
			},
			{
				Name:    "Available Times",
				Type:    StepTypeSelectMenu,
				Creator: createTimesStep,
				Handler: handleTimesStep,
			},
			{
				Name:    "Weekdays",
				Type:    StepTypeSelectMenu,
				Creator: createWeekdaysStep,
				Handler: handleWeekdaysStep,
			},
			{
				Name:    "War Experience",
				Type:    StepTypeButton,
				Creator: createWarExperienceStep,
				Handler: handleWarExperienceStep,
			},
		},
	}
}

// GetTotalSteps returns the total number of steps
func (sp *StepProcessor) GetTotalSteps() int {
	return len(sp.steps)
}

// GetStepTitle generates the step title dynamically
func (sp *StepProcessor) GetStepTitle(stepIndex int) string {
	if stepIndex < 0 || stepIndex >= len(sp.steps) {
		return "Unknown Step"
	}
	return fmt.Sprintf("📝 Registro - Passo %d/%d", stepIndex+1, len(sp.steps))
}

// GetStepByIndex returns a step by its index (0-based)
func (sp *StepProcessor) GetStepByIndex(index int) *StepDefinition {
	if index < 0 || index >= len(sp.steps) {
		return nil
	}
	return &sp.steps[index]
}

// GetAllSteps returns all steps
func (sp *StepProcessor) GetAllSteps() []StepDefinition {
	return sp.steps
}

// IsLastStep checks if the given step index is the last one
func (sp *StepProcessor) IsLastStep(stepIndex int) bool {
	return stepIndex == len(sp.steps)-1
}

// ProcessStep executes a step by index
func (sp *StepProcessor) ProcessStep(ctx *common.ModuleContext, state *RegistrationState, stepIndex int) error {
	step := sp.GetStepByIndex(stepIndex)
	if step == nil {
		return nil // Step not found, ignore
	}

	// Create and send the step message
	messageData, err := step.Creator(ctx, state)
	if err != nil {
		return err
	}

	dg := ctx.Session()
	_, err = dg.ChannelMessageSendComplex(state.TopicID, messageData)
	return err
}

// HandleStepResponse handles a response to a step by index
func (sp *StepProcessor) HandleStepResponse(ctx *common.ModuleContext, state *RegistrationState, stepIndex int, interaction interface{}) error {
	step := sp.GetStepByIndex(stepIndex)
	if step == nil {
		return nil // Step not found, ignore
	}

	// Validate input if validator exists
	if step.Validator != nil {
		if err := step.Validator(interaction); err != nil {
			return err
		}
	}

	// Handle the step response
	return step.Handler(ctx, state, interaction)
}
