# Registration Module - Step System Refactoring

This document describes the refactoring of the registration module to use a dynamic step-based system.

## Overview

The registration module has been refactored from a hardcoded step system to a flexible, dynamic step-based architecture. This allows for easy addition, modification, and reordering of registration steps.

## Architecture

### Core Components

1. **StepDefinition**: Defines a single registration step with its properties and handlers
2. **StepProcessor**: Manages the overall step flow and execution
3. **StepCreator**: Function type that creates the UI for a step (messages, components)
4. **StepHandler**: Function type that handles user responses to a step
5. **StepValidator**: Optional function type that validates step input

### Step Types

- `StepTypeTextInput`: Steps that require text input from users (like IGN)
- `StepTypeSelectMenu`: Steps that use Discord select menus for choices
- `StepTypeButton`: Steps that use buttons for interaction
- `StepTypeCompletion`: Final step type for completion actions

## Current Step Flow

1. **IGN Step** (`ign`): Text input for in-game name
2. **PVP Classes Step** (`pvp_classes`): Select menu for PVP class selection
3. **Times Step** (`times`): Select menu for available playing times
4. **Weekdays Step** (`weekdays`): Select menu for available weekdays

## Key Features

### Dynamic Step Management
- Steps are defined in a vector and processed dynamically
- Easy to add new steps or reorder existing ones
- Each step knows its position and total count

### Flexible Step Types
- Different interaction types supported (text, select menus, buttons)
- Each step can have custom UI creation logic
- Validation can be applied to any step

### Backward Compatibility
- Maintains the old `Step` field in `RegistrationState` for compatibility
- Gradual migration path from old system to new system

## Adding New Steps

To add a new step:

1. Create a step creator function following the `StepCreator` signature
2. Create a step handler function following the `StepHandler` signature
3. Optionally create a validator function following the `StepValidator` signature
4. Add the step definition to the `steps` slice in `NewStepProcessor()`

Example:
```go
{
    ID:         "new_step",
    Name:       "New Step",
    Type:       StepTypeSelectMenu,
    StepNumber: 5,
    TotalSteps: 5,
    Creator:    createNewStep,
    Handler:    handleNewStep,
    Validator:  validateNewStep, // optional
}
```

## File Structure

- `steps.go`: Core step system definitions and processor
- `step_handlers.go`: Individual step creators and handlers
- `module.go`: Updated to include step processor instance
- `handlers.go`: Updated to use new step system
- `helpers.go`: Updated message handling for text input steps

## Migration Notes

- The old hardcoded functions (`askForIGN`, `askForTimes`, etc.) have been removed
- The new system maintains compatibility with existing database structures
- Step constants (`STEP_IGN`, `STEP_CLASSES`, etc.) are still used for backward compatibility
