package register

import (
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
	ID         string
	Name       string
	Type       StepType
	StepNumber int
	TotalSteps int
	Creator    StepCreator
	Handler    StepHandler
	Validator  StepValidator
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
				ID:         "ign",
				Name:       "In-Game Name",
				Type:       StepTypeTextInput,
				StepNumber: 1,
				TotalSteps: 4,
				Creator:    createIGNStep,
				Handler:    handleIGNStep,
				Validator:  validateIGN,
			},
			{
				ID:         "pvp_classes",
				Name:       "PVP Classes",
				Type:       StepTypeSelectMenu,
				StepNumber: 2,
				TotalSteps: 4,
				Creator:    createPVPClassesStep,
				Handler:    handlePVPClassesStep,
			},
			{
				ID:         "times",
				Name:       "Available Times",
				Type:       StepTypeSelectMenu,
				StepNumber: 3,
				TotalSteps: 4,
				Creator:    createTimesStep,
				Handler:    handleTimesStep,
			},
			{
				ID:         "weekdays",
				Name:       "Weekdays",
				Type:       StepTypeSelectMenu,
				StepNumber: 4,
				TotalSteps: 4,
				Creator:    createWeekdaysStep,
				Handler:    handleWeekdaysStep,
			},
		},
	}
}

// GetStepByID returns a step by its ID
func (sp *StepProcessor) GetStepByID(id string) *StepDefinition {
	for _, step := range sp.steps {
		if step.ID == id {
			return &step
		}
	}
	return nil
}

// GetStepByNumber returns a step by its number (1-based)
func (sp *StepProcessor) GetStepByNumber(number int) *StepDefinition {
	for _, step := range sp.steps {
		if step.StepNumber == number {
			return &step
		}
	}
	return nil
}

// GetNextStep returns the next step after the given step
func (sp *StepProcessor) GetNextStep(currentStepID string) *StepDefinition {
	for i, step := range sp.steps {
		if step.ID == currentStepID && i+1 < len(sp.steps) {
			return &sp.steps[i+1]
		}
	}
	return nil
}

// GetAllSteps returns all steps
func (sp *StepProcessor) GetAllSteps() []StepDefinition {
	return sp.steps
}

// IsLastStep checks if the given step is the last one
func (sp *StepProcessor) IsLastStep(stepID string) bool {
	if len(sp.steps) == 0 {
		return true
	}
	lastStep := sp.steps[len(sp.steps)-1]
	return lastStep.ID == stepID
}

// ProcessStep executes a step
func (sp *StepProcessor) ProcessStep(ctx *common.ModuleContext, state *RegistrationState, stepID string) error {
	step := sp.GetStepByID(stepID)
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

// HandleStepResponse handles a response to a step
func (sp *StepProcessor) HandleStepResponse(ctx *common.ModuleContext, state *RegistrationState, stepID string, interaction interface{}) error {
	step := sp.GetStepByID(stepID)
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
