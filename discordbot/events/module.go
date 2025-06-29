package events

import (
	"fmt"
	"log"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/discordutils"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const ModuleName = "events"

type EventsConfig struct {
	Enabled bool `json:"enabled"`

	ChannelIDs   []string `json:"channel_ids"`
	RequireAdmin bool     `json:"require_admin"`
	GuideMessage bool     `json:"guide_message"`
	ChannelName  string   `json:"channel_name"`
	InitMessage  string   `json:"init_message"`

	EventTypeEmojis    map[types.EventType]string   `json:"event_type_emojis"`
	EventSlots         map[types.EventType]string   `json:"event_slots"`
	EventSlotsCount    map[types.EventType]int      `json:"event_slots_count"`
	EventSlotRoleName  map[EventSlotRole]string     `json:"event_slot_role_name"`
	EventSlotRoleEmoji map[EventSlotRole]string     `json:"event_slot_role_emoji"`
	EventTypeOptions   []common.EventSelectorOption `json:"event_type_options"`
	EventTypeNames     map[types.EventType]string   `json:"event_type_names"`
	EventNameMap       map[types.EventType]string   `json:"event_name_map"`
}

type EventsModule struct{}

func (s *EventsModule) Name() string {
	return ModuleName
}

func (s *EventsModule) Setup(ctx *common.ModuleContext, config any) (bool, error) {
	var cfg = config.(*EventsConfig)
	if !cfg.Enabled {
		return false, nil
	}
	fmt.Println("Events module is enabled, setting up...")

	global, _ := ctx.Config("globals").(*globals.GlobalsConfig)

	dg := ctx.Session()

	for _, channel_id := range cfg.ChannelIDs {
		_, err := setupEventsChannel(ctx, channel_id)
		if err != nil {
			log.Printf("Cannot setup events channel %s: %v", channel_id, err)
			continue
		}

		channel, err := ctx.Session().Channel(channel_id)
		if err != nil {
			log.Printf("Could not retrieve channel for event: %v", err)
			continue
		}

		dg.AddHandler(discordutils.CreateChannelHandler(ctx, channel, handlers))
	}

	_, err := dg.ApplicationCommandCreate(global.AppID, global.GuildID, &discordgo.ApplicationCommand{
		Name:        "evento",
		Description: "Iniciar um novo evento",
	})
	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	dg.ApplicationCommandDelete(global.AppID, global.GuildID, "encerrar")

	dg.AddHandler(HandleMessages(ctx, global.GuildID))
	dg.AddHandler(HandleEventAction(ctx, global.GuildID))

	go eventsCheckRoutine(ctx)

	return true, nil
}

func (s *EventsModule) DefaultConfig() any {
	EVENTS_CHANNEL_IDS = strings.Split(os.Getenv("EVENTS_CHANNEL_IDS"), ",")

	if os.Getenv("EVENTS_REQUIRE_ADMIN") != "" {
		EVENTS_REQUIRE_ADMIN = true
	}

	if v := os.Getenv("EVENTS_GUIDE_MESSAGE"); v != "" {
		EVENTS_GUIDE_MESSAGE = v == "true"
	}

	return &EventsConfig{
		Enabled: true,

		ChannelIDs:         EVENTS_CHANNEL_IDS,
		RequireAdmin:       EVENTS_REQUIRE_ADMIN,
		GuideMessage:       EVENTS_GUIDE_MESSAGE,
		ChannelName:        "",
		InitMessage:        EVENTS_CHANNEL_INIT_MESSAGE,
		EventTypeEmojis:    EventTypeEmojis,
		EventSlots:         EventSlots,
		EventSlotsCount:    EventSlotsCount,
		EventSlotRoleName:  EventSlotRoleName,
		EventSlotRoleEmoji: EventSlotRoleEmoji,
		EventTypeOptions:   EVENT_TYPE_OPTIONS,
		EventTypeNames:     EventTypeNames,
		EventNameMap: map[types.EventType]string{
			types.EventTypeDungeonNormal: EventNameDungeonNormal,
			types.EventTypeDungeonM1:     EventNameDungeonM1,
			types.EventTypeDungeonM2:     EventNameDungeonM2,
			types.EventTypeDungeonM3:     EventNameDungeonM3,
			types.EventTypeRaidGorgon:    EventNameRaidGorgon,
			types.EventTypeRaidDevour:    EventNameRaidDevour,
			types.EventTypeOPR:           EventNameOPR,
			types.EventTypeArena:         EventNameArena,
			types.EventTypeInfluenceRace: EventNameInfluenceRace,
			types.EventTypeLootRoute:     EventNameLootRoute,
			types.EventTypeWar:           EventNameWar,
		},
	}
}

func GetModuleConfig(ctx *common.ModuleContext) *EventsConfig {
	if module, ok := ctx.Config(ModuleName).(*EventsConfig); ok {
		return module
	}
	return nil
}
