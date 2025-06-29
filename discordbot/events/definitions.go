package events

import (
	"fmt"
	"nwmanager/discordbot/common"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"slices"
)

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
	types.EventTypeRaidGorgon:    "TL223 S1H22",
	types.EventTypeRaidDevour:    "R5HFS 1222E 12223 RPPPP",
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
	EventSlotDPSPadLight:  "Pad Light",
	EventSlotSupportFlail: "Flail",
	EventSlotHealFlail:    "Heal/Flail",
	EventSlotDPSEvadeFire: "Evade Fire",
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
	EventSlotHealFlail:    "🪄",
	EventSlotDPSEvadeFire: "⚡",
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

func getEventTypeName(config *EventsConfig, eventType types.EventType) string {
	if name, ok := config.EventTypeNames[eventType]; ok {
		return fmt.Sprintf("%s %s", EventTypeEmojis[eventType], name)
	}

	return ""
}

func getEventSlotCount(config *EventsConfig, eventType types.EventType) int {
	if slots, ok := config.EventSlots[eventType]; ok && slots == "" {
		return config.EventSlotsCount[eventType]
	}

	slots := config.EventSlots[eventType]
	slotCount := 0
	for _, slot := range slots {
		if slot != ' ' {
			slotCount++
		}
	}

	return slotCount
}

func getEventFreeSlotsCount(config *EventsConfig, event *types.Event) int {
	if slots, ok := config.EventSlots[event.Type]; ok && slots == "" {
		totalSlots := getEventSlotCount(config, event.Type)
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

func getEventSlotsCountByRole(config *EventsConfig, eventType types.EventType, role EventSlotRole) int {
	slots := config.EventSlots[eventType]
	eventSlots := 0
	for _, slot := range slots {
		if slot == rune(role) {
			eventSlots++
		}
	}

	return eventSlots
}

func getEventRoleByPosition(config *EventsConfig, eventType types.EventType, position int) EventSlotRole {
	slots := config.EventSlots[eventType]
	if position >= len(slots) {
		return EventSlotAny
	}

	letters := 0
	var role byte
	for i := range len(slots) {
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

func getEventRoleNameByPosition(config *EventsConfig, eventType types.EventType, position int) string {
	slots := config.EventSlots[eventType]
	if position >= len(slots) {
		return ""
	}

	letters := 0
	var role byte
	for i := range len(slots) {
		if slots[i] == ' ' {
			continue
		}

		if letters == position {
			role = slots[i]
			break
		}
		letters++
	}

	if roleName, ok := config.EventSlotRoleName[EventSlotRole(role)]; ok {
		return config.EventSlotRoleEmoji[EventSlotRole(role)] + globals.SEPARATOR + roleName
	}

	return ""
}

func resolveEventSlotFromEmoji(config *EventsConfig, emoji string) EventSlotRole {
	for role, emojiRole := range config.EventSlotRoleEmoji {
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

func getEventFreeSlotsByRole(config *EventsConfig, event *types.Event, targetRole EventSlotRole) []int {
	freeSlots := []int{}
	for i, slot := range event.PlayerSlots {
		if slot != "" {
			continue
		}
		role := getEventRoleByPosition(config, event.Type, i)
		if role == targetRole || role == EventSlotAny {
			freeSlots = append(freeSlots, i)
		}
	}

	return freeSlots
}

func getEventSlotTypes(config *EventsConfig, event *types.Event) []EventSlotRole {
	slotTypes := []EventSlotRole{}

	for i := range event.PlayerSlots {
		role := getEventRoleByPosition(config, event.Type, i)
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
