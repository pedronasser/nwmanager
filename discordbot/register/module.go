package register

import (
	"fmt"
	"nwmanager/discordbot/common"
	"os"
	"slices"
	"strings"
)

const ModuleName = "register"

type RegisterConfig struct {
	Enabled bool `json:"enabled"`
}

type RegisterModule struct{}

func (s *RegisterModule) Name() string {
	return ModuleName
}

func (s *RegisterModule) Setup(ctx *common.ModuleContext, config any) (bool, error) {
	var cfg = config.(*RegisterConfig)
	if !cfg.Enabled {
		return false, nil
	}
	fmt.Println("Register module is enabled, setting up...")

	return true, nil
}

func (s *RegisterModule) DefaultConfig() any {
	var IsModuleEnabledFromEnv = slices.Contains(strings.Split(os.Getenv("MODULES"), ","), ModuleName)
	return &RegisterConfig{
		Enabled: IsModuleEnabledFromEnv,
	}
}
