package events

import (
	"fmt"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"slices"
)

type Config struct {
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

func DefaultConfig() Config {
	return Config{
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

var (
	EVENTS_CHANNEL_IDS   = []string{}
	EVENTS_REQUIRE_ADMIN = false
	EVENTS_GUIDE_MESSAGE = true
	EVENTS_CHANNEL_NAME  = ""
)

const (
	EVENTS_CHANNEL_INIT_MESSAGE = "**Clique no botão abaixo** ou **envie /evento** para criar um evento.\n\nPara encerrar um evento, **clique no botão de encerrar**x."
)

var EventSlotsCount = map[types.EventType]int{
	types.EventTypeDungeonNormal: 5,
	types.EventTypeDungeonM1:     5,
	types.EventTypeDungeonM2:     5,
	types.EventTypeDungeonM3:     5,
	types.EventTypeRaidGorgon:    10,
	types.EventTypeRaidDevour:    20,
	types.EventTypeOPR:           5,
	types.EventTypeArena:         3,
}

// Event slots
var EventSlots = map[types.EventType]string{
	types.EventTypeDungeonNormal: "THDDD",
	types.EventTypeDungeonM1:     "THDDD",
	types.EventTypeDungeonM2:     "THDDD",
	types.EventTypeDungeonM3:     "THDDD",
	types.EventTypeRaidGorgon:    "THDDD HDDDD",
	types.EventTypeRaidDevour:    "T5HFS 12223 12224 RPPPP",
	types.EventTypeOPR:           "THDDD",
	types.EventTypeArena:         "",
	types.EventTypeInfluenceRace: "",
	types.EventTypeLootRoute:     "",
}

var EventSlotRoleName = map[EventSlotRole]string{
	EventSlotTank:         "Tank",
	EventSlotDPS:          "DPS",
	EventSlotHeal:         "Heal",
	EventSlotRangedTank:   "R. Tank",
	EventSlotDPSBlood:     "R. Blood",
	EventSlotDPSEvade:     "R. Evade",
	EventSlotDPSSpear:     "Lança",
	EventSlotDPSFire:      "Fire DPS",
	EventSlotDPSSerenity:  "Serenity",
	EventSlotDPSRendBot:   "Rend Bot",
	EventSlotDPSSnS:       "SnS DPS",
	EventSlotDPSPadLight:  "Pad",
	EventSlotSupportFlail: "Flail",
}

var EventSlotRoleEmoji = map[EventSlotRole]string{
	EventSlotTank:         "🔰",
	EventSlotDPS:          "⚔️",
	EventSlotHeal:         "🌿",
	EventSlotRangedTank:   "🎯",
	EventSlotDPSBlood:     "🩸",
	EventSlotDPSEvade:     "⚡",
	EventSlotDPSSpear:     "⚜️",
	EventSlotDPSRendBot:   "🤖",
	EventSlotDPSSerenity:  "🗡",
	EventSlotDPSSnS:       "🛡️",
	EventSlotDPSFire:      "🔥",
	EventSlotDPSPadLight:  "💡",
	EventSlotSupportFlail: "🌀",
}

var EventTypeEmojis = map[types.EventType]string{
	types.EventTypeDungeonNormal: EventTypeEmojiDungeonNormal,
	types.EventTypeDungeonM1:     EventTypeEmojiDungeonM1,
	types.EventTypeDungeonM2:     EventTypeEmojiDungeonM2,
	types.EventTypeDungeonM3:     EventTypeEmojiDungeonM3,
	types.EventTypeRaidGorgon:    EventTypeEmojiRaidGorgon,
	types.EventTypeRaidDevour:    EventTypeEmojiRaidDevour,
	types.EventTypeOPR:           EventTypeEmojiOPR,
	types.EventTypeArena:         EventTypeEmojiArena,
	types.EventTypeInfluenceRace: EventTypeEmojiInfluenceRace,
	types.EventTypeLootRoute:     EventTypeEmojiLootRoute,
}

var EventTypeNames = map[types.EventType]string{
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
}

var (
	EVENT_TYPE_OPTIONS = []common.EventSelectorOption{
		{
			Label: EventNameLootRoute,
			Value: string(types.EventTypeLootRoute),
			Emoji: EventTypeEmojiLootRoute,
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameRaidGorgon, EventSlotsCount[types.EventTypeRaidGorgon]),
			Value: string(types.EventTypeRaidGorgon),
			Emoji: EventTypeEmojiRaidGorgon,
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameRaidDevour, EventSlotsCount[types.EventTypeRaidDevour]),
			Value: string(types.EventTypeRaidDevour),
			Emoji: EventTypeEmojiRaidDevour,
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameOPR, EventSlotsCount[types.EventTypeOPR]),
			Value: string(types.EventTypeOPR),
			Emoji: EventTypeEmojiOPR,
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameArena, EventSlotsCount[types.EventTypeArena]),
			Value: string(types.EventTypeArena),
			Emoji: EventTypeEmojiArena,
		},
		{
			Label: fmt.Sprintf("%s", EventNameInfluenceRace),
			Value: string(types.EventTypeInfluenceRace),
			Emoji: EventTypeEmojiInfluenceRace,
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameDungeonNormal, EventSlotsCount[types.EventTypeDungeonNormal]),
			Value: string(types.EventTypeDungeonNormal),
			Emoji: EventTypeEmojiDungeonNormal,
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameDungeonM1, EventSlotsCount[types.EventTypeDungeonM1]),
			Value: string(types.EventTypeDungeonM1),
			Emoji: EventTypeEmojiDungeonM1,
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameDungeonM2, EventSlotsCount[types.EventTypeDungeonM2]),
			Value: string(types.EventTypeDungeonM2),
			Emoji: EventTypeEmojiDungeonM2,
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameDungeonM3, EventSlotsCount[types.EventTypeDungeonM3]),
			Value: string(types.EventTypeDungeonM3),
			Emoji: EventTypeEmojiDungeonM3,
		},
	}
)

func getEventTypeName(eventType types.EventType) string {
	if name, ok := EventTypeNames[eventType]; ok {
		return fmt.Sprintf("%s %s", EventTypeEmojis[eventType], name)
	}

	return ""
}

func getEventSlotCount(eventType types.EventType) int {
	if slots, ok := EventSlots[eventType]; ok && slots == "" {
		return EventSlotsCount[eventType]
	}

	slots := EventSlots[eventType]
	slotCount := 0
	for _, slot := range slots {
		if slot != ' ' {
			slotCount++
		}
	}

	return slotCount
}

func getEventFreeSlotsCount(event *types.Event) int {
	if slots, ok := EventSlots[event.Type]; ok && slots == "" {
		totalSlots := getEventSlotCount(event.Type)
		return totalSlots - len(event.PlayerSlots)
	} else {
		freeSlots := 0
		for _, slot := range event.PlayerSlots {
			if slot == "" {
				freeSlots++
			}
		}

		return freeSlots
	}
}

func getEventSlotsCountByRole(eventType types.EventType, role EventSlotRole) int {
	slots := EventSlots[eventType]
	eventSlots := 0
	for _, slot := range slots {
		if slot == rune(role) {
			eventSlots++
		}
	}

	return eventSlots
}

func getEventRoleByPosition(eventType types.EventType, position int) EventSlotRole {
	slots := EventSlots[eventType]
	if position >= len(slots) {
		return EventSlotAny
	}

	letters := 0
	var role byte
	for i := 0; i < len(slots); i++ {
		if slots[i] == ' ' {
			continue
		}

		if letters == position {
			role = slots[i]
			break
		}
		letters++
	}

	return EventSlotRole(role)
}

func getEventRoleNameByPosition(eventType types.EventType, position int) string {
	slots := EventSlots[eventType]
	if position >= len(slots) {
		return ""
	}

	letters := 0
	var role byte
	for i := 0; i < len(slots); i++ {
		if slots[i] == ' ' {
			continue
		}

		if letters == position {
			role = slots[i]
			break
		}
		letters++
	}

	if roleName, ok := EventSlotRoleName[EventSlotRole(role)]; ok {
		return EventSlotRoleEmoji[EventSlotRole(role)] + globals.SEPARATOR + roleName
	}

	return ""
}

func resolveEventSlotFromEmoji(emoji string) EventSlotRole {
	for role, emojiRole := range EventSlotRoleEmoji {
		if emojiRole == emoji {
			return role
		}
	}

	return EventSlotAny
}

func getEventFreeSlots(event *types.Event) []int {
	freeSlots := []int{}
	for i, slot := range event.PlayerSlots {
		if slot == "" {
			freeSlots = append(freeSlots, i)
		}
	}

	return freeSlots
}

func getEventFreeSlotsByRole(event *types.Event, targetRole EventSlotRole) []int {
	freeSlots := []int{}
	for i, slot := range event.PlayerSlots {
		if slot != "" {
			continue
		}
		role := getEventRoleByPosition(event.Type, i)
		if role == targetRole || role == EventSlotAny {
			freeSlots = append(freeSlots, i)
		}
	}

	return freeSlots
}

func getEventSlotTypes(event *types.Event) []EventSlotRole {
	slotTypes := []EventSlotRole{}

	for i, _ := range event.PlayerSlots {
		role := getEventRoleByPosition(event.Type, i)
		if role == EventSlotAny {
			continue
		}

		if slices.Contains(slotTypes, role) {
			continue
		}

		slotTypes = append(slotTypes, role)
	}

	return slotTypes
}
