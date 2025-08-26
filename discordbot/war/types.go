package war

import "nwmanager/types"

// WarData stores temporary war creation data during the modal process
type WarData struct {
	FortName      string
	WarType       types.WarType
	OpponentGuild string
	ScheduledDate string
	ScheduledTime string
}

// Store ongoing war creation data in memory (if needed for complex flows)
var WarCreationData = make(map[string]*WarData)
