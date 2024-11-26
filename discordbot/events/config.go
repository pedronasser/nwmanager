package events

import (
	"fmt"
	"nwmanager/discordbot/helpers"
	"nwmanager/types"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	EVENTS_CHANNEL_ID = ""
)

func init() {
	_ = godotenv.Load()
	EVENTS_CHANNEL_ID = os.Getenv("EVENTS_CHANNEL_ID")
	if EVENTS_CHANNEL_ID == "" {
		fmt.Println("EVENTS_CHANNEL_ID is not set")
		os.Exit(1)
	}
}

const (
	EVENTS_CHANNEL_NAME         = "🟢・eventos"
	EVENTS_CHANNEL_INIT_MESSAGE = "**Clique no botão abaixo** ou **envie /evento** para criar um evento.\n\nPara encerrar um evento, **clique no botão de encerrar**x."

	MEMBER_ROLE_NAME  = "👥・Membro"
	ADMIN_ROLE_NAME   = "👑 Governador"
	CONSUL_ROLE_NAME  = "💎Consul"
	OFFICER_ROLE_NAME = "🏆Oficial"
	OPR_ROLE_NAME     = "⚔️・OPR"
	RAID_DEVOUR_ROLE  = "🪱・Devorador"
	RAID_GORGON_ROLE  = "🗿・Gorgonas"
)

var (
	ADMIN_ROLE_ID   = ""
	CONSUL_ROLE_ID  = ""
	OFFICER_ROLE_ID = ""
)

var (
	EVENT_CLEANUP_INTERVAL         time.Duration = 30 * time.Second
	EVENT_COMPLETE_EXPIRE_DURATION time.Duration = 15 * time.Minute
	EVENT_MAX_DURATION             time.Duration = 60 * time.Minute
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

const (
	EventTypeSlotsDungeonNormal = 5
	EventTypeSlotsDungeonM1     = 5
	EventTypeSlotsDungeonM2     = 5
	EventTypeSlotsDungeonM3     = 5
	EventTypeSlotsRaidGorgon    = 10
	EventTypeSlotsRaidDevour    = 20
	EventTypeSlotsOPR           = 5
	EventTypeSlotsArena         = 3
	EventTypeSlotsInfluenceRace = -1
	EventTypeSlotsWar           = -1
)

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
	EventTypeEmojiWar           = "⚔️"
	EventTypeEmojiLootRoute     = "💎"
)

// Event slots
var EventSlots = map[types.EventType]string{
	types.EventTypeDungeonNormal: "THDDD",
	types.EventTypeDungeonM1:     "THDDD",
	types.EventTypeDungeonM2:     "THDDD",
	types.EventTypeDungeonM3:     "THDDD",
	types.EventTypeRaidGorgon:    "THDDD HDDDD",
	types.EventTypeRaidDevour:    "THDDD HDDDD DDDDD DDDDD",
	types.EventTypeOPR:           "THDDD",
	types.EventTypeArena:         "AAA",
	types.EventTypeInfluenceRace: "",
	types.EventTypeWar:           "",
	types.EventTypeLootRoute:     "",
}

// Event roles
var EventRoles = map[types.EventType]string{
	types.EventTypeOPR:        OPR_ROLE_NAME,
	types.EventTypeRaidDevour: RAID_DEVOUR_ROLE,
	types.EventTypeRaidGorgon: RAID_GORGON_ROLE,
}

var EventRoleIDs = map[types.EventType]string{}

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
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameRaidGorgon, EventTypeSlotsRaidGorgon),
			Value: string(types.EventTypeRaidGorgon),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiRaidGorgon,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameRaidDevour, EventTypeSlotsRaidDevour),
			Value: string(types.EventTypeRaidDevour),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiRaidDevour,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameOPR, EventTypeSlotsOPR),
			Value: string(types.EventTypeOPR),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiOPR,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameArena, EventTypeSlotsArena),
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
			Label: fmt.Sprintf("%s", EventNameWar),
			Value: string(types.EventTypeWar),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiWar,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameDungeonNormal, EventTypeSlotsDungeonNormal),
			Value: string(types.EventTypeDungeonNormal),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiDungeonNormal,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameDungeonM1, EventTypeSlotsDungeonM1),
			Value: string(types.EventTypeDungeonM1),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiDungeonM1,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameDungeonM2, EventTypeSlotsDungeonM2),
			Value: string(types.EventTypeDungeonM2),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiDungeonM2,
			},
		},
		{
			Label: fmt.Sprintf("%s [Vagas: %d]", EventNameDungeonM3, EventTypeSlotsDungeonM3),
			Value: string(types.EventTypeDungeonM3),
			Emoji: &discordgo.ComponentEmoji{
				Name: EventTypeEmojiDungeonM3,
			},
		},
	}
)

func getEventName(eventType types.EventType) string {
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
	case types.EventTypeWar:
		return EventTypeEmojiWar + " " + EventNameWar
	case types.EventTypeLootRoute:
		return EventTypeEmojiLootRoute + " " + EventNameLootRoute
	}

	return ""
}

type EventSlotRole rune

const (
	EventSlotTank EventSlotRole = 'T'
	EventSlotHeal EventSlotRole = 'H'
	EventSlotDPS  EventSlotRole = 'D'
	EventSlotAny  EventSlotRole = 'A'
)

func getEventSlotCount(eventType types.EventType) int {
	slots := EventSlots[eventType]
	slotCount := 0
	for _, slot := range slots {
		if slot != ' ' {
			slotCount++
		}
	}

	return slotCount
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

	switch EventSlotRole(role) {
	case EventSlotTank:
		return "🛡️ Tank"
	case EventSlotHeal:
		return "🌿 Heal"
	case EventSlotDPS:
		return "⚔️ DPS"
	case EventSlotAny:
		return "Jogador"
	}

	return ""
}

func resolveEventSlotFromEmoji(emoji string) EventSlotRole {
	switch emoji {
	case "🛡️":
		return EventSlotTank
	case "🌿":
		return EventSlotHeal
	case "⚔️":
		return EventSlotDPS
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

func getEventRoleID(guild *discordgo.Guild, event *types.Event) string {
	if roleName, ok := EventRoles[event.Type]; ok {
		role := helpers.GetRoleByName(guild, roleName)
		if role == nil {
			return ""
		}
		return role.ID
	}

	return ""
}

func isMemberAdmin(member *discordgo.Member) bool {
	for _, role := range member.Roles {
		if ADMIN_ROLE_ID != "" && role == ADMIN_ROLE_ID {
			return true
		}
		if CONSUL_ROLE_ID != "" && role == CONSUL_ROLE_ID {
			return true
		}
		if OFFICER_ROLE_ID != "" && role == OFFICER_ROLE_ID {
			return true
		}
	}

	return false
}
