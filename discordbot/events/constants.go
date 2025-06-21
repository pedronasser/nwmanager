package events

type EventSlotRole rune

const (
	EventNameDungeonNormal = "Dungeon Normal"
	EventNameDungeonM1     = "Dungeon Mutada (M1)"
	EventNameDungeonM2     = "Dungeon Mutada (M2)"
	EventNameDungeonM3     = "Dungeon Mutada (M3)"
	EventNameRaidGorgon    = "Gorgonas"
	EventNameRaidDevour    = "Devorador"
	EventNameOPR           = "Outpost Rush (OPR)"
	EventNameArena         = "Arena"
	EventNameInfluenceRace = "Corrida de Influência"
	EventNameWar           = "Guerra"
	EventNameLootRoute     = "Rota"
)

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
