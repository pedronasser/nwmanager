package register

import (
	"fmt"
	"nwmanager/discordbot/common"
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
	return &RegisterConfig{
		Enabled: true,
	}
}
