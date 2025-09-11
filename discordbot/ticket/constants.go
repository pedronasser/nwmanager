package ticket

// Ticket module constants

const (
	// Default emojis for ticket actions
	EmojiSendBuild      = "📸"
	EmojiSendQuestion   = "❓"
	EmojiChangeClass    = "⚔️"
	EmojiChangeName     = "✏️"
	EmojiViewBuild      = "👁️"
	EmojiConfirmBuild   = "✅"
	EmojiCloseThread    = "❌"
	EmojiQuestionSolved = "✅"

	// Default class emoji if not found in config
	DefaultClassEmoji = "⚔️"
)

// Colors for embeds
const (
	ColorTicketMain     = 0x00ff00 // Green
	ColorBuildThread    = 0x0099ff // Blue
	ColorQuestionThread = 0xff9900 // Orange
)
