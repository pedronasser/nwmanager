package management

import (
	"context"
	"fmt"
	"log"
	"nwmanager/database"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/forPelevin/gomoji"
)

func processPlayerArchiving(ctx context.Context, dg *discordgo.Session, db database.Database, player *types.Player) {
	err := types.ArchivePlayer(ctx, db, player)
	if err != nil {
		log.Fatalf("Cannot delete player: %v", err)
	}

	if player.TicketChannel != "" {
		archiveTicketChannel(dg, player)
	}

	log.Printf("Player '%s' has been archived", player.IGN)
}

func processPlayerUnarchiving(ctx context.Context, dg *discordgo.Session, db database.Database, player *types.Player) {
	err := types.UnarchivePlayer(ctx, db, player)
	if err != nil {
		log.Fatalf("Cannot delete player: %v", err)
	}

	if player.TicketChannel != "" {
		updateTicketChannel(dg, player, player.WarClass, player.IGN)
	}

	log.Printf("Player '%s' has been unarchived", player.IGN)
}

func processPlayerDeleting(ctx context.Context, dg *discordgo.Session, db database.Database, player *types.Player) {
	if player.TicketChannel != "" {
		_, _ = dg.ChannelDelete(player.TicketChannel)
	}

	err := types.DeletePlayer(ctx, db, player)
	if err != nil {
		log.Fatalf("Cannot delete player: %v", err)
	}
}

func getMemberWarClass(member *discordgo.Member) string {
	for _, roleID := range member.Roles {
		for warClass, classID := range globals.CLASS_ROLE_IDS {
			if roleID == classID {
				return warClass
			}
		}
	}

	return ""
}

func getClassEmoji(className string) string {
	emojis := gomoji.FindAll(className)
	if len(emojis) == 0 {
		return ""
	}

	return emojis[0].Character
}

func createTicketChannel(dg *discordgo.Session, GuildID *string, memberID *string, className string, title string) (*discordgo.Channel, error) {
	category, err := dg.Channel(globals.CLASS_CATEGORY_IDS[className])
	if err != nil {
		return nil, fmt.Errorf("Cannot get category: %v", err)
	}

	ticketChannel, err := dg.GuildChannelCreateComplex(*GuildID, discordgo.GuildChannelCreateData{
		Name:     title,
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: globals.CLASS_CATEGORY_IDS[className],
		PermissionOverwrites: append(category.PermissionOverwrites,
			&discordgo.PermissionOverwrite{
				ID:   *memberID,
				Type: discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionSendMessages |
					discordgo.PermissionViewChannel |
					discordgo.PermissionReadMessageHistory |
					discordgo.PermissionAddReactions |
					discordgo.PermissionAttachFiles,
			},
		),
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot create ticket channel: %v", err)
	}

	// Send message
	_, err = dg.ChannelMessageSendComplex(ticketChannel.ID, &discordgo.MessageSend{
		Content: fmt.Sprintf("Olá, <@%s>! Este é seu ticket! Qualquer dúvida ou problema, fique à vontade para perguntar aqui e marque um <@&%s>. :smiley:", *memberID, globals.ACCESS_ROLE_IDS[globals.OFFICER_ROLE_NAME]),
	})

	return ticketChannel, nil
}

func updateTicketChannel(dg *discordgo.Session, player *types.Player, className string, title string) {
	if player.TicketChannel == "" {
		return
	}

	category, err := dg.Channel(globals.CLASS_CATEGORY_IDS[className])
	if err != nil {
		return
	}

	_, _ = dg.ChannelEditComplex(player.TicketChannel, &discordgo.ChannelEdit{
		Name:     title,
		ParentID: globals.CLASS_CATEGORY_IDS[className],
		PermissionOverwrites: append(category.PermissionOverwrites,
			&discordgo.PermissionOverwrite{
				ID:   player.DiscordID,
				Type: discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionSendMessages |
					discordgo.PermissionViewChannel |
					discordgo.PermissionReadMessageHistory |
					discordgo.PermissionAddReactions |
					discordgo.PermissionAttachFiles,
			},
		),
	})
}

func deleteTicketChannel(dg *discordgo.Session, player *types.Player) {
	if player.TicketChannel == "" {
		return
	}

	_, _ = dg.ChannelDelete(player.TicketChannel)
}

func archiveTicketChannel(dg *discordgo.Session, player *types.Player) {
	category, err := dg.Channel(globals.CLASS_CATEGORY_IDS[globals.ARCHIVE_CATEGORY])
	if err != nil {
		fmt.Println(err)
		return
	}

	_, _ = dg.ChannelEditComplex(player.TicketChannel, &discordgo.ChannelEdit{
		Name:                 fmt.Sprintf("%s-arquivado", player.IGN),
		ParentID:             category.ID,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{},
	})
}

func findTicketChannel(channels []*discordgo.Channel, playerName string) *discordgo.Channel {
	for _, channel := range channels {
		parts := strings.Split(channel.Name, globals.SEPARATOR)
		if len(parts) != 2 {
			continue
		}
		withoutPrefix := parts[1]
		if strings.ToLower(withoutPrefix) == strings.Join(strings.Split(strings.ToLower(playerName), " "), "-") {
			return channel
		}
	}

	return nil
}

func shouldHaveTicket(member *discordgo.Member) bool {
	class := getMemberWarClass(member)
	if class == globals.RECRUIT_ROLE_NAME {
		return false
	}

	return true
}

func IsTicketChannel(channel *discordgo.Channel) bool {
	for _, id := range globals.CLASS_CATEGORY_IDS {
		if channel.ParentID == id {
			return true
		}
	}

	return false
}
