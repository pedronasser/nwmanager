package ticket

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

func memberRoleMonitoringRoutine(ctx *common.ModuleContext) {
	config := GetModuleConfig(ctx)
	globalConfig := ctx.Config("globals").(*globals.GlobalsConfig)

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
		log.Printf("No player data found for member %s, skipping ticket creation", member.User.ID)
		return nil
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

	// Initially create in the main ticket category
	// Will be moved to class-specific category when player selects/changes class
	channel, err := ctx.Session().GuildChannelCreateComplex(globalConfig.GuildID, discordgo.GuildChannelCreateData{
		Name:     player.IGN,
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: config.TicketCategoryID, // Main ticket category
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

	// Create and send the main ticket message
	err = setupTicketMessage(ctx, channel, player)
	if err != nil {
		// Clean up channel if message setup fails
		ctx.Session().ChannelDelete(channel.ID)
		return nil, fmt.Errorf("failed to setup ticket message: %w", err)
	}

	return channel, nil
}
