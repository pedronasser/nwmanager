package ticket

import (
	"fmt"
	"log"
	"nwmanager/database"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"os"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
)

func memberRoleMonitoringRoutine(ctx *common.ModuleContext) {
	config := GetModuleConfig(ctx)
	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

	performPeriodicMemberCheck(ctx, globalConfig)
	routineExportPlayersCSV(ctx, ctx.DB())

	ticker := time.NewTicker(time.Duration(config.CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			routineExportPlayersCSV(ctx, ctx.DB())

		case <-ctx.Context.Done():
			return
		}
	}
}

// getAllGuildMembers fetches all members from a guild using pagination
func getAllGuildMembers(ctx *common.ModuleContext, guildID string) ([]*discordgo.Member, error) {
	var allMembers []*discordgo.Member
	const limit = 1000 // Discord API limit per request
	after := ""

	for {
		members, err := ctx.Session().GuildMembers(guildID, after, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch guild members: %w", err)
		}

		allMembers = append(allMembers, members...)

		// If we got fewer members than the limit, we've reached the end
		if len(members) < limit {
			break
		}

		// Set the cursor for the next page to the ID of the last member
		after = members[len(members)-1].User.ID
	}

	log.Printf("Successfully fetched %d total guild members", len(allMembers))
	return allMembers, nil
}

func performPeriodicMemberCheck(ctx *common.ModuleContext, globalConfig *globals.GlobalsConfig) {
	log.Println("Running periodic member role check (backup routine)...")

	// Get all guild members with pagination
	members, err := getAllGuildMembers(ctx, globalConfig.GuildID)
	if err != nil {
		log.Printf("Error fetching guild members: %v", err)
		return
	}

	// Get all active tickets
	activeTickets, err := getAllActiveTickets(ctx)
	if err != nil {
		log.Printf("Error fetching active tickets: %v", err)
		return
	}

	// Create a map of current member IDs for fast lookup
	memberIDs := make(map[string]bool)
	for _, member := range members {
		memberIDs[member.User.ID] = true
	}

	// Check each member's role status
	for _, member := range members {
		err := processMemberRoleStatus(ctx, member, activeTickets, globalConfig.MemberRoleID)
		if err != nil {
			log.Printf("Error processing member %s: %v", member.User.ID, err)
		}
	}

	// Cleanup: Remove tickets for members who are no longer in the guild
	for _, ticket := range activeTickets {
		if !memberIDs[ticket.DiscordID] {
			// Double-check by attempting to retrieve the member directly from Discord API
			// This ensures we don't have false positives due to pagination limits
			_, err := ctx.Session().GuildMember(globalConfig.GuildID, ticket.DiscordID)
			if err != nil {
				// Member not found in guild, remove the ticket
				log.Printf("Confirmed orphaned ticket for user %s who is no longer in the guild: %v", ticket.DiscordID, err)
				err := removeTicketForMember(ctx, &ticket)
				if err != nil {
					log.Printf("Error removing orphaned ticket for user %s: %v", ticket.DiscordID, err)
				} else {
					log.Printf("Successfully removed orphaned ticket for departed user %s", ticket.DiscordID)
				}
			} else {
				// Member exists but wasn't in our members list (probably due to pagination)
				log.Printf("Ticket owner %s exists in guild but wasn't in member list (pagination limit)", ticket.DiscordID)
			}
		}
	}
}

func processMemberRoleStatus(ctx *common.ModuleContext, member *discordgo.Member, activeTickets []Ticket, memberRoleID string) error {
	hasMemberRole := discordutils.HasRole(member, memberRoleID)

	// Find existing ticket for this member
	var existingTicket *Ticket
	for _, ticket := range activeTickets {
		if ticket.DiscordID == member.User.ID {
			existingTicket = &ticket
			break
		}
	}

	log.Printf("Processing member %s (%s): hasMemberRole=%v, hasExistingTicket=%v",
		member.User.Username, member.User.ID, hasMemberRole, existingTicket != nil)

	if hasMemberRole {
		// Member has role but no ticket - create one
		if existingTicket == nil {
			log.Printf("Creating ticket for member %s who has role but no ticket", member.User.ID)
			err := createTicketForMember(ctx, member)
			if err != nil {
				return fmt.Errorf("failed to create ticket for member %s: %w", member.User.ID, err)
			}
			log.Printf("Created ticket for member %s (%s)", member.User.Username, member.User.ID)
		} else {
			log.Printf("Member %s already has ticket, no action needed", member.User.ID)
		}
	} else {
		// Member doesn't have role but has ticket - delete it
		if existingTicket != nil {
			log.Printf("Removing ticket for member %s who no longer has role", member.User.ID)
			err := removeTicketForMember(ctx, existingTicket)
			if err != nil {
				return fmt.Errorf("failed to remove ticket for member %s: %w", member.User.ID, err)
			}
			log.Printf("Removed ticket for member %s (%s)", member.User.Username, member.User.ID)
		} else {
			log.Printf("Member %s has no role and no ticket, no action needed", member.User.ID)
		}
	}

	return nil
}

func createTicketForMember(ctx *common.ModuleContext, member *discordgo.Member) error {
	// Get player data with retry logic in case of timing issues
	var player *types.Player
	var err error

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		player, err = types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), member.User.ID)
		if err != nil {
			if i == maxRetries-1 {
				return fmt.Errorf("failed to get player data after %d retries: %w", maxRetries, err)
			}
			log.Printf("Attempt %d failed to get player data for %s, retrying: %v", i+1, member.User.ID, err)
			time.Sleep(time.Second * time.Duration(i+1)) // Progressive delay
			continue
		}
		if player != nil {
			break // Successfully found player
		}
		if i == maxRetries-1 {
			return fmt.Errorf("no player data found for member %s after %d retries", member.User.ID, maxRetries)
		}
		log.Printf("Attempt %d: no player data found for %s, retrying", i+1, member.User.ID)
		time.Sleep(time.Second * time.Duration(i+1)) // Progressive delay
	}

	// Create ticket channel
	channel, err := createTicketChannel(ctx, member, player)
	if err != nil {
		return fmt.Errorf("failed to create ticket channel: %w", err)
	}

	log.Printf("Created ticket channel %s for player %s", channel.ID, player.IGN)
	return nil
}

func removeTicketForMember(ctx *common.ModuleContext, ticket *Ticket) error {
	// Delete the Discord channel
	_, err := ctx.Session().ChannelDelete(ticket.ChannelID)
	if err != nil {
		log.Printf("Error deleting ticket channel %s: %v", ticket.ChannelID, err)
		// Continue with database cleanup even if Discord deletion fails
	}

	// Mark ticket as inactive in database
	err = deleteTicket(ctx, ticket)
	if err != nil {
		return fmt.Errorf("failed to deactivate ticket in database: %w", err)
	}

	// Clear player.ticket_channel field
	player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), ticket.DiscordID)
	if err != nil {
		log.Printf("Error getting player to clear ticket_channel for discord ID %s: %v", ticket.DiscordID, err)
		// Don't return error as ticket cleanup was successful
	} else if player != nil {
		player.TicketChannel = ""
		err = types.UpdatePlayer(ctx.Context, ctx.DB(), player)
		if err != nil {
			log.Printf("Error clearing ticket_channel for player %s: %v", player.IGN, err)
			// Don't return error as ticket cleanup was successful
		} else {
			log.Printf("Successfully cleared ticket_channel for player %s", player.IGN)
		}
	}

	return nil
}

func createTicketChannel(ctx *common.ModuleContext, member *discordgo.Member, player *types.Player) (*discordgo.Channel, error) {
	config := GetModuleConfig(ctx)
	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

	// Determine which category to create the ticket in
	var categoryID string
	if player.WarClass != "" {
		// Player has a class defined, create in class-specific category
		if classCategory, exists := globalConfig.ClassCategoryIDs[player.WarClass]; exists {
			categoryID = classCategory
		} else {
			// Fallback to main category if class category doesn't exist
			categoryID = config.TicketCategoryID
		}
	} else {
		// No class defined, create in main ticket category
		categoryID = config.TicketCategoryID
	}

	channel, err := ctx.Session().GuildChannelCreateComplex(globalConfig.GuildID, discordgo.GuildChannelCreateData{
		Name:     player.IGN,
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: categoryID,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:   globalConfig.GuildID, // @everyone role (guild ID)
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: discordgo.PermissionViewChannel,
			},
			{
				ID:    member.User.ID,
				Type:  discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
			},
			{
				ID:    globalConfig.AdminRoleID,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory | discordgo.PermissionManageMessages,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create ticket channel: %w", err)
	}

	// If player has a war class defined, add class role and update nickname
	if player.WarClass != "" {
		// Add PVP class role if it exists in the configuration
		if classRoleID, exists := globalConfig.ClassRoleIDs[player.WarClass]; exists {
			err = ctx.Session().GuildMemberRoleAdd(globalConfig.GuildID, player.DiscordID, classRoleID)
			if err != nil {
				log.Printf("Error adding class role %s to player %s: %v", classRoleID, player.IGN, err)
			} else {
				log.Printf("Successfully added class role %s to player %s", classRoleID, player.IGN)
			}
		}

		// Update Discord nickname with class emoji
		if classEmoji, exists := globalConfig.ClassEmojiIDs[player.WarClass]; exists {
			newNickname := fmt.Sprintf("%s %s", classEmoji, player.IGN)
			err = ctx.Session().GuildMemberNickname(globalConfig.GuildID, player.DiscordID, newNickname)
			if err != nil {
				log.Printf("Error updating nickname for player %s: %v", player.IGN, err)
			} else {
				log.Printf("Successfully updated nickname for player %s to %s", player.IGN, newNickname)
			}
		}
	}

	// Create and send the main ticket message
	err = setupTicketMessage(ctx, channel, player)
	if err != nil {
		// Clean up channel if message setup fails
		ctx.Session().ChannelDelete(channel.ID)
		return nil, fmt.Errorf("failed to setup ticket message: %w", err)
	}

	return channel, nil
}

func routineExportPlayersCSV(ctx *common.ModuleContext, db database.Database) {
	players, err := types.GetActivePlayers(ctx.Context, db)
	if err != nil {
		log.Fatalf("Cannot get players: %v", err)
	}

	csvFile, err := os.Create("players_new.csv")
	if err != nil {
		log.Fatalf("Cannot create file: %v", err)
	}
	defer csvFile.Close()

	_, _ = csvFile.WriteString("Name,Classe")
	for _, player := range players {
		_, _ = csvFile.WriteString("\n" + player.IGN + "," + player.WarClass)
	}

	os.Remove("static/players.csv")
	os.Rename("players_new.csv", "static/players.csv")

	log.Println("Exported players to players.csv")
}

// HandleGuildMemberUpdate handles real-time member role updates
func HandleGuildMemberUpdate(ctx *common.ModuleContext) func(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	return func(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
		globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

		// Only process events for our guild
		if m.GuildID != globalConfig.GuildID {
			return
		}

		// Get the old member state to compare roles
		if m.BeforeUpdate == nil {
			// If we don't have before state, we can't compare roles
			// But we can still check if the member has the role and needs a ticket
			log.Printf("Member update for %s (%s) without before state, checking current role status",
				m.Member.User.Username, m.Member.User.ID)

			// Get all active tickets to pass to processing function
			activeTickets, err := getAllActiveTickets(ctx)
			if err != nil {
				log.Printf("Error fetching active tickets for member update: %v", err)
				return
			}

			// Process based on current role status
			err = processMemberRoleStatus(ctx, m.Member, activeTickets, globalConfig.MemberRoleID)
			if err != nil {
				log.Printf("Error processing member role status for %s: %v", m.Member.User.ID, err)
			}
			return
		}

		memberRoleID := globalConfig.MemberRoleID

		// Check if member role status changed
		oldHasMemberRole := slices.Contains(m.BeforeUpdate.Roles, memberRoleID)
		newHasMemberRole := slices.Contains(m.Member.Roles, memberRoleID)

		// Only process if member role status actually changed
		if oldHasMemberRole == newHasMemberRole {
			return
		}

		log.Printf("Member %s (%s) role change detected: had member role: %v -> has member role: %v",
			m.Member.User.Username, m.Member.User.ID, oldHasMemberRole, newHasMemberRole)

		// Get all active tickets to pass to processing function
		activeTickets, err := getAllActiveTickets(ctx)
		if err != nil {
			log.Printf("Error fetching active tickets for member update: %v", err)
			return
		}

		// Process the member role status change
		err = processMemberRoleStatus(ctx, m.Member, activeTickets, memberRoleID)
		if err != nil {
			log.Printf("Error processing member role change for %s: %v", m.Member.User.ID, err)
		}
	}
}

// HandleGuildMemberRemove handles when members leave the guild entirely
func HandleGuildMemberRemove(ctx *common.ModuleContext) func(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
	return func(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
		globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

		// Only process events for our guild
		if m.GuildID != globalConfig.GuildID {
			return
		}

		log.Printf("Member %s (%s) left the guild, checking for active ticket",
			m.Member.User.Username, m.Member.User.ID)

		// Check if this member has an active ticket
		ticket, err := getTicketByDiscordID(ctx, m.Member.User.ID)
		if err != nil {
			log.Printf("Error checking for ticket when member left: %v", err)
			return
		}

		// If they have an active ticket, remove it
		if ticket != nil {
			err := removeTicketForMember(ctx, ticket)
			if err != nil {
				log.Printf("Error removing ticket for departed member %s: %v", m.Member.User.ID, err)
			} else {
				log.Printf("Successfully removed ticket for departed member %s (%s)",
					m.Member.User.Username, m.Member.User.ID)
			}
		}
	}
}
