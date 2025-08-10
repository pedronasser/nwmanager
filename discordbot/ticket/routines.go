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
	"time"

	"github.com/bwmarrin/discordgo"
)

func memberRoleMonitoringRoutine(ctx *common.ModuleContext) {
	config := GetModuleConfig(ctx)
	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

	routineExportPlayersCSV(ctx, ctx.DB())

	ticker := time.NewTicker(time.Duration(config.CheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Checking member roles for ticket management...")

			// Get all guild members
			members, err := ctx.Session().GuildMembers(globalConfig.GuildID, "", 1000)
			if err != nil {
				log.Printf("Error fetching guild members: %v", err)
				continue
			}

			// Get all active tickets
			activeTickets, err := getAllActiveTickets(ctx)
			if err != nil {
				log.Printf("Error fetching active tickets: %v", err)
				continue
			}

			// Check each member's role status
			for _, member := range members {
				err := processMemberRoleStatus(ctx, member, activeTickets, globalConfig.MemberRoleID)
				if err != nil {
					log.Printf("Error processing member %s: %v", member.User.ID, err)
				}
			}

			routineExportPlayersCSV(ctx, ctx.DB())

		case <-ctx.Context.Done():
			return
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

	if hasMemberRole {
		// Member has role but no ticket - create one
		if existingTicket == nil {
			err := createTicketForMember(ctx, member)
			if err != nil {
				return fmt.Errorf("failed to create ticket for member %s: %w", member.User.ID, err)
			}
			log.Printf("Created ticket for member %s (%s)", member.User.Username, member.User.ID)
		}
	} else {
		// Member doesn't have role but has ticket - delete it
		if existingTicket != nil {
			err := removeTicketForMember(ctx, existingTicket)
			if err != nil {
				return fmt.Errorf("failed to remove ticket for member %s: %w", member.User.ID, err)
			}
			log.Printf("Removed ticket for member %s (%s)", member.User.Username, member.User.ID)
		}
	}

	return nil
}

func createTicketForMember(ctx *common.ModuleContext, member *discordgo.Member) error {
	// Get player data
	player, err := types.GetPlayerByDiscordID(ctx.Context, ctx.DB(), member.User.ID)
	if err != nil {
		return fmt.Errorf("failed to get player data: %w", err)
	}
	if player == nil {
		return fmt.Errorf("no player data found for member %s", member.User.ID)
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
