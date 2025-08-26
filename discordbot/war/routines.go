package war

import (
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/types"
	"time"
)

// warCleanupRoutine runs periodically to archive wars that have reached their scheduled time
func warCleanupRoutine(ctx *common.ModuleContext) {
	ticker := time.NewTicker(CLEANUP_INTERVAL)
	defer ticker.Stop()

	log.Printf("War cleanup routine started with interval: %v", CLEANUP_INTERVAL)

	for {
		select {
		case <-ticker.C:
			err := processWarCleanup(ctx)
			if err != nil {
				log.Printf("Error in war cleanup routine: %v", err)
			}
		case <-ctx.Context.Done():
			log.Println("War cleanup routine stopped")
			return
		}
	}
}

// processWarCleanup checks for wars that need to be archived
func processWarCleanup(ctx *common.ModuleContext) error {
	// Get all active wars
	wars, err := types.GetActiveWars(ctx.Context, ctx.DB())
	if err != nil {
		return err
	}

	now := time.Now()
	
	for _, war := range wars {
		// Check if war time has passed
		if war.ScheduledAt != nil && war.ScheduledAt.Before(now) {
			err := CancelWar(ctx, war)
			if err != nil {
				log.Printf("Error archiving war %s: %v", war.ID.Hex(), err)
				continue
			}
			
			log.Printf("Successfully archived war: %s (%s vs %s)", 
				war.FortName, war.FortName, war.OpponentGuild)
		}
	}

	return nil
}
