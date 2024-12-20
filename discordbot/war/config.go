package war

import (
	"fmt"
	"nwmanager/types"
	"os"

	"github.com/joho/godotenv"
)

var (
	WAR_CHANNEL_ID = ""
)

func init() {
	_ = godotenv.Load()
	WAR_CHANNEL_ID = os.Getenv("WAR_CHANNEL_ID")
	if WAR_CHANNEL_ID == "" {
		fmt.Println("WAR_CHANNEL_ID is not set")
		os.Exit(1)
	}
}

const (
	EVENTS_CHANNEL_NAME = "🪖・guerra"

	MEMBER_ROLE_NAME  = "👥・Membro"
	ADMIN_ROLE_NAME   = "👑 Governador"
	CONSUL_ROLE_NAME  = "💎Consul"
	OFFICER_ROLE_NAME = "🏆Oficial"
)

var (
	MEMBER_ROLE_ID  = ""
	ADMIN_ROLE_ID   = ""
	CONSUL_ROLE_ID  = ""
	OFFICER_ROLE_ID = ""
)

// War Class Emojis
var WarClassEmojis = map[types.WarClass]string{
	types.WarClassBruiser:           "🪓",
	types.WarClassHealer:            "🌿",
	types.WarClassHealerAOE:         "🟢",
	types.WarClassTank:              "🛡️",
	types.WarClassVoidFlail:         "🔄",
	types.WarClassVoidIce:           "🏴",
	types.WarClassFlailIce:          "❄️",
	types.WarClassFireIce:           "🔥",
	types.WarClassFireAbyss:         "🌀",
	types.WarClassFireBlunder:       "🔫",
	types.WarClassFireRapier:        "🩸",
	types.WarClassDisruptorScorpion: "🦂",
	types.WarClassDisruptorPoison:   "👻",
	types.WarClassDisruptorHatchet:  "💀",
	types.WarClassDisruptorGS:       "⚔️",
	types.WarClassBow:               "🏹",
}

var WarClassNames = map[types.WarClass]string{
	types.WarClassBruiser:           "Bruiser",
	types.WarClassHealer:            "Healer Pocket",
	types.WarClassHealerAOE:         "Healer AOE",
	types.WarClassTank:              "Tank",
	types.WarClassVoidFlail:         "Void/Flail",
	types.WarClassVoidIce:           "Void/Ice",
	types.WarClassFlailIce:          "Flail/Ice",
	types.WarClassFireIce:           "Fire/Ice",
	types.WarClassFireAbyss:         "Fire/Abyss",
	types.WarClassFireBlunder:       "Fire/Bacamarte",
	types.WarClassFireRapier:        "Fire/Rapier",
	types.WarClassDisruptorScorpion: "Disrup./Escorpião",
	types.WarClassDisruptorHatchet:  "Disrup./Machadinha",
	types.WarClassDisruptorPoison:   "Disrup./Veneno",
	types.WarClassDisruptorGS:       "Disrup./GS",
	types.WarClassBow:               "Arco",
}

func getWarClassName(classType types.WarClass) string {
	return WarClassEmojis[classType] + " " + WarClassNames[classType]
}
