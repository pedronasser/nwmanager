package events

import (
	"fmt"
	"nwmanager/discordbot/globals"
	"nwmanager/types"
	"slices"

	"github.com/bwmarrin/discordgo"
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

const (
	EventNameDungeonNormal = "Dungeon Normal"
	EventNameDungeonM1     = "Dungeon Mutada (M1)"
	EventNameDungeonM2     = "Dungeon Mutada (M2)"
	EventNameDungeonM3     = "Dungeon Mutada (M3)"
	EventNameRaidGorgon    = "Raid Gorgonas"
	EventNameRaidDevour    = "Raid Devorador"
	EventNameOPR           = "Outpost Rush (OPR)"
	EventNameArena         = "Arena"
	EventNameInfluenceRace = "Corrida de Influência"
	EventNameWar           = "Guerra"
	EventNameLootRoute     = "Rota"
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

const (
	EventTypeEmojiDungeonNormal = "🧌"
	EventTypeEmojiDungeonM1     = "1️⃣"
	EventTypeEmojiDungeonM2     = "2️⃣"
	EventTypeEmojiDungeonM3     = "3️⃣"
	EventTypeEmojiRaidGorgon    = "🗿"
	EventTypeEmojiRaidDevour    = "🪱"
	EventTypeEmojiOPR           = "⚔️"
	EventTypeEmojiArena         = "🏹"
	EventTypeEmojiInfluenceRace = "🏁"
	EventTypeEmojiLootRoute     = "💎"
)

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

type EventSlotRole rune

const (
	EventSlotTank         EventSlotRole = 'T'
	EventSlotDPS          EventSlotRole = 'D'
	EventSlotAny          EventSlotRole = 'A'
	EventSlotHeal         EventSlotRole = 'H'
	EventSlotRangedTank   EventSlotRole = '0' // Ranged Tank
	EventSlotDPSBlood     EventSlotRole = '1' // Rapier Blood
	EventSlotDPSEvade     EventSlotRole = '2' // Rapier Evade
	EventSlotDPSSpear     EventSlotRole = '3' // Lança
	EventSlotDPSSerenity  EventSlotRole = '4' // Serenidade
	EventSlotDPSFire      EventSlotRole = '5' // Fire DPS
	EventSlotDPSRendBot   EventSlotRole = 'R' // Rend Bot
	EventSlotDPSSnS       EventSlotRole = 'S' // SnS DPS
	EventSlotDPSPadLight  EventSlotRole = 'P' // Arco Pad
	EventSlotSupportFlail EventSlotRole = 'F' // Flail/Suporte
)

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

var (
	EVENT_TYPE_OPTIONS = []discordgo.SelectMenuOption{
		{
			Label: EventNameLootRoute,
			Value: string(types.EventTypeLootRoute),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiLootRoute,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameRaidGorgon, EventSlotsCount[types.EventTypeRaidGorgon]),
			Value: string(types.EventTypeRaidGorgon),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiRaidGorgon,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameRaidDevour, EventSlotsCount[types.EventTypeRaidDevour]),
			Value: string(types.EventTypeRaidDevour),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiRaidDevour,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameOPR, EventSlotsCount[types.EventTypeOPR]),
			Value: string(types.EventTypeOPR),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiOPR,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameArena, EventSlotsCount[types.EventTypeArena]),
			Value: string(types.EventTypeArena),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiArena,
			},
		},
		{
			Label: fmt.Sprintf("%s", EventNameInfluenceRace),
			Value: string(types.EventTypeInfluenceRace),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiInfluenceRace,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameDungeonNormal, EventSlotsCount[types.EventTypeDungeonNormal]),
			Value: string(types.EventTypeDungeonNormal),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiDungeonNormal,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameDungeonM1, EventSlotsCount[types.EventTypeDungeonM1]),
			Value: string(types.EventTypeDungeonM1),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiDungeonM1,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameDungeonM2, EventSlotsCount[types.EventTypeDungeonM2]),
			Value: string(types.EventTypeDungeonM2),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiDungeonM2,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameDungeonM3, EventSlotsCount[types.EventTypeDungeonM3]),
			Value: string(types.EventTypeDungeonM3),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiDungeonM3,
			},
		},
	}
)

func getEventTypeName(eventType types.EventType) string {
	switch eventType {
	case types.EventTypeDungeonNormal:
		return EventTypeEmojiDungeonNormal + " " + EventNameDungeonNormal
	case types.EventTypeDungeonM1:
		return EventTypeEmojiDungeonM1 + " " + EventNameDungeonM1
	case types.EventTypeDungeonM2:
		return EventTypeEmojiDungeonM2 + " " + EventNameDungeonM2
	case types.EventTypeDungeonM3:
		return EventTypeEmojiDungeonM3 + " " + EventNameDungeonM3
	case types.EventTypeRaidGorgon:
		return EventTypeEmojiRaidGorgon + " " + EventNameRaidGorgon
	case types.EventTypeRaidDevour:
		return EventTypeEmojiRaidDevour + " " + EventNameRaidDevour
	case types.EventTypeOPR:
		return EventTypeEmojiOPR + " " + EventNameOPR
	case types.EventTypeArena:
		return EventTypeEmojiArena + " " + EventNameArena
	case types.EventTypeInfluenceRace:
		return EventTypeEmojiInfluenceRace + " " + EventNameInfluenceRace
	case types.EventTypeLootRoute:
		return EventTypeEmojiLootRoute + " " + EventNameLootRoute
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
