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
	"github.com/joho/godotenv"
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

	globals, _ := ctx.Config("globals").(*globals.GlobalsConfig)

	dg := ctx.Session()
	db := ctx.DB()

	_ = godotenv.Load()
	EVENTS_CHANNEL_IDS = strings.Split(os.Getenv("EVENTS_CHANNEL_IDS"), ",")
	if len(EVENTS_CHANNEL_IDS) == 0 {
		fmt.Println("EVENTS_CHANNEL_IDS is not set")
		os.Exit(1)
	}

	if os.Getenv("EVENTS_REQUIRE_ADMIN") != "" {
		EVENTS_REQUIRE_ADMIN = true
	}

	if v := os.Getenv("EVENTS_GUIDE_MESSAGE"); v != "" {
		EVENTS_GUIDE_MESSAGE = v == "true"
	}

	for _, channel_id := range EVENTS_CHANNEL_IDS {
		_, err := setupEventsChannel(ctx.Context, dg, db, channel_id)
		if err != nil {
			log.Printf("Cannot setup events channel %s: %v", channel_id, err)
			continue
		}
		dg.AddHandler(discordutils.CreateHandler(globals.GuildID, channel_id, handlers, db))
	}

	_, err := dg.ApplicationCommandCreate(globals.AppID, globals.GuildID, &discordgo.ApplicationCommand{
		Name:        "evento",
		Description: "Iniciar um novo evento",
	})
	if err != nil {
		log.Fatalf("Cannot create slash command: %v", err)
	}

	dg.ApplicationCommandDelete(globals.AppID, globals.GuildID, "encerrar")

	// _, err = dg.ApplicationCommandCreate(*AppID, *GuildID, &discordgo.ApplicationCommand{
	// 	Name:        "encerrar",
	// 	Description: "Encerre um evento",
	// })
	// if err != nil {
	// 	log.Fatalf("Cannot create slash command: %v", err)
	// }

	dg.AddHandler(HandleMessages(globals.GuildID, dg, db))
	dg.AddHandler(HandleEventClose(globals.GuildID, dg, db))

	go eventsCheckRoutine(db, dg)

	return true, nil
}

func (s *EventsModule) DefaultConfig() any {
	return &EventsConfig{
		Enabled: true,

		ChannelIDs:         []string{},
		RequireAdmin:       false,
		GuideMessage:       true,
		ChannelName:        "",
		InitMessage:        EVENTS_CHANNEL_INIT_MESSAGE,
		EventTypeEmojis:    EventTypeEmojis,
		EventSlots:         EventSlots,
		EventSlotsCount:    EventSlotsCount,
		EventSlotRoleName:  EventSlotRoleName,
		EventSlotRoleEmoji: EventSlotRoleEmoji,
		EventTypeOptions:   EVENT_TYPE_OPTIONS,
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
