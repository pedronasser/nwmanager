package register

import (
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"os"
	"slices"
	"strings"
)

const ModuleName = "register"

type RegisterConfig struct {
	Enabled                bool   `json:"enabled"`
	WelcomeChannelID       string `json:"welcome_channel_id"`
	AdminRoleID            string `json:"admin_role_id"`
	MemberRoleID           string `json:"member_role_id"`
	WelcomeMessage         string `json:"welcome_message"`
	RegistrationCategoryID string `json:"registration_category_id"`
}

type RegisterModule struct{}

// Store ongoing registrations in memory
var RegisterData = make(map[string]*RegistrationState)

type RegistrationState struct {
	DiscordID  string
	TopicID    string // Registration channel ID
	Step       int
	IGN        string
	PVPClasses []string
	Times      []string
	Weekdays   []string
	MessageID  string
}

func (s *RegisterModule) Name() string {
	return ModuleName
}

func (s *RegisterModule) Setup(ctx *common.ModuleContext, config any) (bool, error) {
	var cfg = config.(*RegisterConfig)
	if !cfg.Enabled {
		return false, nil
	}
	log.Println("Register module is enabled, setting up...")

	global, _ := ctx.Config("globals").(*globals.GlobalsConfig)
	dg := ctx.Session()

	// Setup welcome channel with registration button
	if cfg.WelcomeChannelID != "" {
		err := setupWelcomeChannel(ctx, cfg.WelcomeChannelID)
		if err != nil {
			log.Printf("Cannot setup welcome channel: %v", err)
		}

		// Add handler for welcome channel
		channel, err := dg.Channel(cfg.WelcomeChannelID)
		if err != nil {
			log.Printf("Could not retrieve welcome channel: %v", err)
		} else {
			dg.AddHandler(discordutils.CreateChannelHandler(ctx, channel, handlers))
		}
	}

	// Add interaction handlers
	dg.AddHandler(HandleRegistrationAction(ctx, global.GuildID))
	dg.AddHandler(HandleRegistrationMessage(ctx))

	// Add message handler for welcome channel cleanup
	dg.AddHandler(HandleWelcomeChannelMessages(ctx))

	return true, nil
}

func (s *RegisterModule) DefaultConfig() any {
	var IsModuleEnabledFromEnv = slices.Contains(strings.Split(os.Getenv("MODULES"), ","), ModuleName)

	welcomeMessage := "Para se tornar um membro oficial da nossa guild no New World, " +
		"clique no botão abaixo para iniciar o processo de registro.\n\n" +
		"Após completar o registro, nossa equipe irá revisar e aprovar sua entrada!"

	return &RegisterConfig{
		Enabled:                IsModuleEnabledFromEnv,
		WelcomeChannelID:       os.Getenv("WELCOME_CHANNEL_ID"),
		AdminRoleID:            os.Getenv("ADMIN_ROLE_ID"),
		MemberRoleID:           os.Getenv("MEMBER_ROLE_ID"),
		WelcomeMessage:         welcomeMessage,
		RegistrationCategoryID: os.Getenv("REGISTRATION_CATEGORY_ID"),
	}
}
