package voice_channel

import (
	"fmt"
	"nwmanager/discordbot/common"
	"os"
	"slices"
	"strings"
)

const ModuleName = "voice_channel"

type VoiceChannelConfig struct {
	Enabled bool `json:"enabled"`
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

	return true, nil
}

func (s *VoiceChannelModule) DefaultConfig() any {
	var IsModuleEnabledFromEnv = slices.Contains(strings.Split(os.Getenv("MODULES"), ","), ModuleName)
	return &VoiceChannelConfig{
		Enabled: IsModuleEnabledFromEnv,
	}
}
