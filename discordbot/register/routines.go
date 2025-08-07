package register

import (
	"fmt"
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	PLAYER_CLEANUP_INTERVAL = 1 * time.Hour // Check once per hour
)

// playerCleanupRoutine checks all players in the database and archives those
// who are no longer in the Discord guild or don't have the member role
func playerCleanupRoutine(ctx *common.ModuleContext) {
	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

	// Execute immediately on startup
	performPlayerCleanup(ctx, globalConfig)

	ticker := time.NewTicker(PLAYER_CLEANUP_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			performPlayerCleanup(ctx, globalConfig)

		case <-ctx.Context.Done():
			return
		}
	}
}

// performPlayerCleanup performs the actual player cleanup logic
func performPlayerCleanup(ctx *common.ModuleContext, globalConfig *globals.GlobalsConfig) {
	log.Println("Starting player cleanup routine...")

	// Get all players from database
	players, err := types.GetActivePlayers(ctx.Context, ctx.DB())
	if err != nil {
		log.Printf("Error fetching players from database: %v", err)
		return
	}

	archivedCount := 0
	checkedCount := 0

	// Check each player
	for _, player := range players {
		checkedCount++

		shouldArchive, reason, err := shouldArchivePlayer(ctx, &player, globalConfig.GuildID, globalConfig.MemberRoleID)
		if err != nil {
			log.Printf("Error checking player %s (%s): %v", player.IGN, player.DiscordID, err)
			continue
		}

		if shouldArchive {
			err = types.ArchivePlayer(ctx.Context, ctx.DB(), &player)
			if err != nil {
				log.Printf("Error archiving player %s (%s): %v", player.IGN, player.DiscordID, err)
				continue
			}

			archivedCount++
			log.Printf("Archived player %s (%s): %s", player.IGN, player.DiscordID, reason)
		}
	}

	log.Printf("Player cleanup routine completed. Checked %d players, archived %d players", checkedCount, archivedCount)
}

// shouldArchivePlayer checks if a player should be archived based on their Discord status
func shouldArchivePlayer(ctx *common.ModuleContext, player *types.Player, guildID string, memberRoleID string) (bool, string, error) {
	session := ctx.Session()

	// Try to get the member from the guild
	member, err := session.GuildMember(guildID, player.DiscordID)
	if err != nil {
		// Check if it's a "not found" error (member left the guild)
		if restErr, ok := err.(*discordgo.RESTError); ok && restErr.Message.Code == discordgo.ErrCodeUnknownMember {
			return true, "member left the guild", nil
		}
		// For other errors, don't archive (could be temporary API issues)
		return false, "", fmt.Errorf("error fetching guild member: %w", err)
	}

	// Check if member has the required role
	if !discordutils.HasRole(member, memberRoleID) {
		return true, "member does not have the required member role", nil
	}

	// Player is still in guild and has the member role
	return false, "", nil
}
