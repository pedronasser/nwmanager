package management

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"nwmanager/types"

	"github.com/bwmarrin/discordgo"
	"github.com/forPelevin/gomoji"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Setup(ctx context.Context, dg *discordgo.Session, AppID, GuildID *string, db types.Database) {
	fmt.Println("Loading management")

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			registerNewPlayers(ctx, dg, GuildID, db)
			archiveUnavailablePlayers(ctx, dg, GuildID, db)
			deleteArchivedPlayers(ctx, dg, db)
		}
	}()
}

func registerNewPlayers(ctx context.Context, dg *discordgo.Session, GuildID *string, db types.Database) {
	members, err := dg.GuildMembers(*GuildID, "", 500)
	if err != nil {
		log.Fatalf("Cannot get guild members: %v", err)
	}

	channels, err := dg.GuildChannels(*GuildID)
	if err != nil {
		log.Fatalf("Cannot get guild channels: %v", err)
	}

	for _, member := range members {
		if slices.Contains(member.Roles, globals.ACCESS_ROLE_IDS[globals.MEMBER_ROLE_NAME]) {
			class := getMemberWarClass(member)
			if class == "" {
				log.Printf("Member %s has no class", member.User.ID)
				continue
			}

			if member.Nick == "" {
				log.Printf("Member %s has no nickname", member.User.ID)
				continue
			}

			if discordutils.IsMemberAdmin(member) {
				log.Printf("Member %s is an admin", member.User.ID)
				continue
			}

			classEmoji := getClassEmoji(class)
			nickWithoutEmoji := strings.Trim(gomoji.RemoveEmojis(member.Nick), " ")
			if member.Nick != classEmoji+nickWithoutEmoji {
				_ = dg.GuildMemberNickname(*GuildID, member.User.ID, classEmoji+nickWithoutEmoji)
			}

			player, err := types.GetPlayerByDiscordID(ctx, db, member.User.ID)
			if err != nil {
				log.Fatalf("Cannot get player: %v", err)
				continue
			}
			if player == nil {
				ign := gomoji.RemoveEmojis(discordutils.GetMemberName(member))
				log.Printf("Registering player %s", ign)
				newPlayer := types.Player{
					ID:        primitive.NewObjectID(),
					DiscordID: member.User.ID,
					IGN:       ign,
					WarClass:  getMemberWarClass(member),
				}

				if shouldHaveTicket(member) {
					existantTicket := findTicketChannel(channels, nickWithoutEmoji)
					if existantTicket != nil {
						log.Printf("Player %s has a ticket channel", nickWithoutEmoji)
						newPlayer.TicketChannel = existantTicket.ID
						updateTicketChannel(dg, &newPlayer, class, strings.Join([]string{classEmoji, nickWithoutEmoji}, globals.SEPARATOR))
					} else {
						ticketChannel, _ := createTicketChannel(dg, GuildID, &(member.User.ID), class, strings.Join([]string{classEmoji, nickWithoutEmoji}, globals.SEPARATOR))
						newPlayer.TicketChannel = ticketChannel.ID
					}
				}

				err = types.InsertPlayer(ctx, db, &newPlayer)
				if err != nil {
					log.Fatalf("Cannot create player: %v", err)
					continue
				}
			} else {
				updatedPlayer := false
				if !shouldHaveTicket(member) && player.TicketChannel != "" {
					archiveTicketChannel(dg, player)
					player.TicketChannel = ""
					updatedPlayer = true
				}

				if player.ArchivedAt != nil {
					updateTicketChannel(dg, player, class, strings.Join([]string{classEmoji, nickWithoutEmoji}, globals.SEPARATOR))
					player.ArchivedAt = nil
					updatedPlayer = true
				}

				if player.WarClass != class {
					updateTicketChannel(dg, player, class, strings.Join([]string{classEmoji, nickWithoutEmoji}, globals.SEPARATOR))
					player.WarClass = class
					updatedPlayer = true
				}

				if updatedPlayer {
					err = types.UpdatePlayer(ctx, db, player)
					if err != nil {
						log.Fatalf("Cannot update player: %v", err)
					}
				}
			}
		}
	}
}

func archiveUnavailablePlayers(ctx context.Context, dg *discordgo.Session, GuildID *string, db types.Database) {
	players, err := types.GetPlayers(ctx, db)
	if err != nil {
		log.Fatalf("Cannot get players: %v", err)
	}

	for _, player := range players {
		if player.ArchivedAt != nil {
			continue
		}
		member, _ := dg.GuildMember(*GuildID, player.DiscordID)
		if member == nil {
			processPlayerArchiving(ctx, dg, db, &player)
		} else {
			if !slices.Contains(member.Roles, globals.ACCESS_ROLE_IDS[globals.MEMBER_ROLE_NAME]) {
				processPlayerArchiving(ctx, dg, db, &player)
			}
		}
	}
}

func deleteArchivedPlayers(ctx context.Context, dg *discordgo.Session, db types.Database) {
	players, err := types.GetArchivedPlayers(ctx, db)
	if err != nil {
		log.Fatalf("Cannot get archived players: %v", err)
	}

	for _, player := range players {
		if player.ArchivedAt == nil {
			continue
		}

		archived := (*player.ArchivedAt)
		if archived.Add(24 * time.Hour).Before(time.Now()) {
			processPlayerDeleting(ctx, dg, db, &player)
		}
	}
}

func processPlayerArchiving(ctx context.Context, dg *discordgo.Session, db types.Database, player *types.Player) {
	err := types.ArchivePlayer(ctx, db, player)
	if err != nil {
		log.Fatalf("Cannot delete player: %v", err)
	}

	if player.TicketChannel != "" {
		archiveTicketChannel(dg, player)
	}

	log.Printf("Player '%s' has been archived", player.IGN)
}

func processPlayerDeleting(ctx context.Context, dg *discordgo.Session, db types.Database, player *types.Player) {
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

	_, _ = dg.ChannelEditComplex(player.TicketChannel, &discordgo.ChannelEdit{
		Name:     title,
		ParentID: globals.CLASS_CATEGORY_IDS[className],
	})
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
	if discordutils.IsMemberAdmin(member) {
		return false
	}

	class := getMemberWarClass(member)
	if class == globals.RECRUIT_ROLE_NAME {
		return false
	}

	return true
}
