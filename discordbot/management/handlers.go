package management

// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"nwmanager/database"
// 	"nwmanager/discordbot/discordutils"
// 	"nwmanager/discordbot/globals"
// 	"nwmanager/types"
// 	"slices"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/bwmarrin/discordgo"
// 	"go.mongodb.org/mongo-driver/bson"
// )

// var ImageTypeSelections = []string{
// 	"Equipamentos de Guerra",
// 	"Print de OPR",
// }

// func HandleTicketMessages(ctx context.Context, dg *discordgo.Session, GuildID *string, db database.Database) func(s *discordgo.Session, i *discordgo.MessageCreate) {
// 	return func(s *discordgo.Session, i *discordgo.MessageCreate) {
// 		if i.GuildID != *GuildID {
// 			return
// 		}

// 		channel, err := s.Channel(i.ChannelID)
// 		if err != nil {
// 			// fmt.Println("Error getting channel: ", err)
// 			return
// 		}

// 		if !IsTicketChannel(channel) {
// 			// fmt.Println("Not a ticket channel")
// 			return
// 		}

// 		if i.Message.Author.ID == s.State.User.ID {
// 			// fmt.Println("Message from bot")
// 			return
// 		}

// 		player, err := types.GetPlayerByTicketChannel(ctx, db, i.ChannelID)
// 		if err != nil || player == nil || player.TicketChannel == "" || player.DiscordID != i.Author.ID {
// 			// fmt.Println("Error getting player: ", err)
// 			return
// 		}

// 		if player.TicketChannel != i.ChannelID {
// 			// fmt.Println("Player is not the owner")
// 			return
// 		}

// 		db.Collection(globals.DB_PREFIX+types.PlayerCollection).
// 			UpdateOne(ctx, bson.M{"_id": player.ID}, bson.M{
// 				"$set": bson.M{"stats.last_ticket_message_at": time.Now()},
// 			})

// 		if i.Message.Attachments == nil || len(i.Message.Attachments) == 0 {
// 			// fmt.Println("No attachments")
// 			return
// 		}

// 		if !slices.Contains(validImageTypes, i.Attachments[0].ContentType) {
// 			// fmt.Println("Invalid image type", i.Attachments[0].ContentType)
// 			return
// 		}

// 		HandlePlayerTicketImageUpload(ctx, s, i, player)
// 	}
// }

// var validImageTypes = []string{
// 	"image/jpeg",
// 	"image/png",
// 	"image/gif",
// 	"image/jpg",
// 	"image/bmp",
// }

// func HandlePlayerTicketImageUpload(ctx context.Context, s *discordgo.Session, i *discordgo.MessageCreate, player *types.Player) {
// 	// if player.IGN != "PedroNC" {
// 	// 	fmt.Println("Not PedroNC")
// 	// 	return
// 	// }

// 	options := make([]discordgo.MessageComponent, 0)
// 	for j, selection := range ImageTypeSelections {
// 		options = append(options, discordgo.Button{
// 			CustomID: fmt.Sprintf("ticket-image_%d_%s", j, i.Message.ID),
// 			Label:    selection,
// 			Style:    discordgo.PrimaryButton,
// 		})
// 	}

// 	_, err := s.ChannelMessageSendComplex(player.TicketChannel, &discordgo.MessageSend{
// 		Content: fmt.Sprintf("<@%s>, por favor identifique a imagem que acabou de enviar:", player.DiscordID),
// 		Flags:   discordgo.MessageFlagsEphemeral,
// 		Components: []discordgo.MessageComponent{
// 			discordgo.ActionsRow{
// 				Components: options,
// 			},
// 		},
// 	})
// 	if err != nil {
// 		fmt.Println("Error sending message: ", err)
// 	}
// }

// func HandleTicketInteractions(ctx context.Context, dg *discordgo.Session, GuildID *string, db database.Database) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
// 	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
// 		if i.GuildID != *GuildID {
// 			return
// 		}

// 		if i.Type != discordgo.InteractionMessageComponent {
// 			return
// 		}

// 		if i.MessageComponentData().CustomID == "" {
// 			return
// 		}

// 		if strings.HasPrefix(i.MessageComponentData().CustomID, "ticket-image_") {
// 			HandleTicketImageInteraction(ctx, s, i, db)
// 		}
// 	}
// }

// func HandleTicketImageInteraction(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, db database.Database) {
// 	// ticket-image_0_<message_id>
// 	// ticket-image_1_<message_id>

// 	parts := strings.Split(i.MessageComponentData().CustomID, "_")
// 	if len(parts) != 3 {
// 		return
// 	}
// 	selection, _ := strconv.ParseInt(parts[1], 10, 64)
// 	message_id := parts[2]

// 	player, err := types.GetPlayerByTicketChannel(ctx, db, i.ChannelID)
// 	if err != nil {
// 		fmt.Println("Error getting player: ", err)
// 		return
// 	}

// 	original, err := s.ChannelMessage(i.ChannelID, message_id)
// 	if err != nil {
// 		fmt.Println("Error getting original message: ", err)
// 		return
// 	}

// 	if selection == 1 {
// 		resp, err := http.Get(original.Attachments[0].URL)
// 		if err != nil {
// 			fmt.Println("Error getting image: ", err)
// 		} else {
// 			defer resp.Body.Close()
// 			buffer, _ := io.ReadAll(resp.Body)
// 			s.ChannelFileSendWithMessage(OPR_PRINTS_CHANNEL_ID, fmt.Sprintf("**%s** enviada por <@%s>", ImageTypeSelections[selection], player.DiscordID), original.Attachments[0].Filename, bytes.NewBuffer(buffer))

// 			db.Collection(globals.DB_PREFIX+types.PlayerCollection).
// 				UpdateOne(ctx, bson.M{"_id": player.ID}, bson.M{
// 					"$set": bson.M{
// 						"stats.last_opr_print_at": time.Now(),
// 						"stats.opr_prints":        player.Stats.OPRPrints + 1,
// 					},
// 				})
// 		}
// 		discordutils.ReplyEphemeralMessage(s, i, "Agredecemos pelo envio. Sua print foi armazenda em nossos registros.", 5*time.Second)
// 		s.ChannelMessageDelete(i.ChannelID, message_id)
// 	} else {
// 		// Check
// 		// s.MessageReactionAdd(i.ChannelID, message_id, "✅")
// 		discordutils.ReplyEphemeralMessage(s, i, "Agradecemos pelo envio. Notificaremos um build leader para fazer a avaliação.", 15*time.Second)
// 		s.ChannelMessageSend(OPR_PRINTS_CHANNEL_ID, fmt.Sprintf("<@&%s>, **<@%s>** enviou uma print de seus equipamentos em <#%s>", globals.CLASS_LEADER_ROLE_IDS[player.WarClass], player.DiscordID, player.TicketChannel))
// 		db.Collection(globals.DB_PREFIX+types.PlayerCollection).
// 			UpdateOne(ctx, bson.M{"_id": player.ID}, bson.M{
// 				"$set": bson.M{
// 					"stats.last_equip_update": time.Now(),
// 				},
// 			})
// 	}

// 	s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
// }
