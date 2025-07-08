package voice_channel

import (
	"fmt"
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/globals"
	. "nwmanager/helpers"
	"nwmanager/types"
	"os"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const ModuleName = "voice_channel"
const VOICE_CHANNEL_COMMAND_CREATE = "vc-setup"

type VoiceChannelConfig struct {
	Enabled  bool     `json:"enabled"`
	Creators []string `json:"creators"` // List of voice channel creator IDs
}

type VoiceChannelModule struct{}

func (s *VoiceChannelModule) Name() string {
	return ModuleName
}

func (s *VoiceChannelModule) Setup(ctx *common.ModuleContext, config any) (bool, error) {
	var cfg = config.(*VoiceChannelConfig)
	if !cfg.Enabled {
		return false, nil
	}
	fmt.Println("VoiceChannel module is enabled, setting up...")

	setupVoiceChannelCreators(ctx)

	return true, nil
}

func (s *VoiceChannelModule) DefaultConfig() any {
	var IsModuleEnabledFromEnv = slices.Contains(strings.Split(os.Getenv("MODULES"), ","), ModuleName)
	return &VoiceChannelConfig{
		Enabled: IsModuleEnabledFromEnv,
	}
}

func GetModuleConfig(ctx *common.ModuleContext) *VoiceChannelConfig {
	if module, ok := ctx.Config(ModuleName).(*VoiceChannelConfig); ok {
		return module
	}
	return nil
}

func setupVoiceChannelCreators(ctx *common.ModuleContext) {
	globalCfg := globals.GetModuleConfig(ctx)
	ds := ctx.Session()
	db := ctx.DB()

	var voiceChannels []types.VoiceChannel
	col := db.Collection(globalCfg.DBPrefix + types.VoiceChannelCollection)
	cursor, err := col.Find(ctx.Context, bson.M{})
	if err != nil {
		log.Printf("Failed to retrieve voice channels from database: %v\n", err)
		return
	}

	err = cursor.All(ctx.Context, &voiceChannels)
	if err != nil {
		log.Printf("Failed to decode voice channels from database: %v\n", err)
		return
	}

	for _, vc := range voiceChannels {
		channel, err := ds.Channel(vc.ChannelID)
		if err != nil {
			channel, err = createVoiceChannelCreator(ctx)
			if err != nil {
				log.Printf("Failed to create new voice channel for owner %s: %v\n", vc.OwnerID, err)
				continue
			}
		}

		ds.AddHandler(func(s *discordgo.Session, vce *discordgo.VoiceStateUpdate) {
			if vce.ChannelID == channel.ID && vce.UserID != s.State.User.ID {
				newChannel, err := createNewVoiceChannel(ctx, vc.OwnerID)
				if err != nil {
					log.Printf("Failed to create new voice channel: %v\n", err)
					return
				}

				// Move the user to the new channel
				err = s.GuildMemberMove(newChannel.GuildID, vce.UserID, Some(newChannel.ID))
				if err != nil {
					log.Printf("Failed to move user to new voice channel: %v\n", err)
					return
				}
			}
		})
	}

	_, err = ds.ApplicationCommandCreate(globalCfg.AppID, globalCfg.GuildID, &discordgo.ApplicationCommand{
		Name:        VOICE_CHANNEL_COMMAND_CREATE,
		Description: "Cria um novo canal de voz para você",
		Type:        discordgo.ChatApplicationCommand,
	})
	if err != nil {
		log.Printf("Failed to create application command for voice channel creator: %v\n", err)
	}

	ds.AddHandler(func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
		if ic.Type != discordgo.InteractionApplicationCommand {
			return
		}

		if ic.ApplicationCommandData().Name == VOICE_CHANNEL_COMMAND_CREATE {
			_, err := createVoiceChannelCreator(ctx)
			if err != nil {
				log.Printf("Failed to create voice channel creator: %v\n", err)
				return
			}
		}
	})
}

func createNewVoiceChannel(ctx *common.ModuleContext, ownerID string) (*discordgo.Channel, error) {
	globalCfg := globals.GetModuleConfig(ctx)
	ds := ctx.Session()

	member, err := ds.GuildMember(globalCfg.GuildID, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild member for owner ID %s: %w", ownerID, err)
	}

	channel, err := ds.GuildChannelCreate(globalCfg.GuildID, fmt.Sprintf("Canal de %s", member.DisplayName()), discordgo.ChannelTypeGuildVoice)
	if err != nil {
		return nil, fmt.Errorf("failed to create new voice channel: %w", err)
	}
	return channel, nil
}

func createVoiceChannelCreator(ctx *common.ModuleContext) (*discordgo.Channel, error) {
	globalCfg := globals.GetModuleConfig(ctx)
	ds := ctx.Session()
	db := ctx.DB()

	channel, err := ds.GuildChannelCreate(globalCfg.GuildID, "Criar Canal de Voz", discordgo.ChannelTypeGuildVoice)
	if err != nil {
		return nil, fmt.Errorf("failed to create voice channel creator: %w", err)
	}

	opts := options.Update().SetUpsert(true)
	res, err := db.Collection(common.ConfigCollectionName).UpdateOne(ctx.Context, bson.M{"_id": globalCfg.GuildID}, bson.M{
		"$push": bson.M{
			"voice_channels.creators": channel.ID,
		},
	}, opts)
	voiceConfig := GetModuleConfig(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to update voice channel creators in database: %w", err)
	}

	voiceConfig.Creators = append(voiceConfig.Creators, channel.ID)
	// TODO: Figure out better way to update the config in the context
	// and let the config be updated automatically
	// It must be thread-safe

	if res.MatchedCount == 0 {
		return nil, fmt.Errorf("no matching guild found in database for ID %s", globalCfg.GuildID)
	}

	return channel, nil
}
