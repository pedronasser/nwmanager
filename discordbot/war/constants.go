package war

import "time"

const (
	// Command IDs
	COMMAND_CREATE_WAR = "/criar-guerra"
	COMMAND_CANCEL_WAR = "/cancelar-guerra"

	// Modal IDs
	MODAL_CREATE_WAR = "modal:create_war"

	// Button IDs
	BUTTON_PARTICIPATE_YES   = "war_participate:yes"
	BUTTON_PARTICIPATE_NO    = "war_participate:no"
	BUTTON_PARTICIPATE_MAYBE = "war_participate:maybe"
	BUTTON_PARTICIPATE_BENCH = "war_participate:bench"
	BUTTON_EDIT_WAR          = "war_edit"

	// Select menu IDs
	SELECT_WAR_TYPE = "select:war_type"

	// Cleanup routine interval
	CLEANUP_INTERVAL = 1 * time.Minute

	// War participation emojis
	EMOJI_YES   = "✅"
	EMOJI_NO    = "❌"
	EMOJI_MAYBE = "🤔"
	EMOJI_BENCH = "🪑"

	// War type emojis
	EMOJI_ATTACK  = "⚔️"
	EMOJI_DEFENSE = "🛡️"
)
