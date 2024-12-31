package globals

import (
	"context"
	"fmt"
	"log"
	"nwmanager/types"

	"github.com/bwmarrin/discordgo"
)

const (
	SEPARATOR = "・"
)

var (
	EVERYONE_ROLE_NAME      = "@everyone"
	MEMBER_ROLE_NAME        = "👥・Membro"
	ADMIN_ROLE_NAME         = "👑 Governador"
	CONSUL_ROLE_NAME        = "💎Consul"
	OFFICER_ROLE_NAME       = "🏆Oficial"
	OPR_ROLE_NAME           = "⚔️・OPR"
	RAID_DEVOUR_ROLE        = "🪱・Devorador"
	RAID_GORGON_ROLE        = "🗿・Gorgonas"
	BRUISER_ROLE_NAME       = "🪓・Bruiser"
	MAGE_ROLE_NAME          = "🧙・Mago"
	ASSASSIN_ROLE_NAME      = "😈・Assassino"
	ARCO_MOSQUETE_ROLE_NAME = "🏹・Arco/Mosquete"
	DISRUPTOR_ROLE_NAME     = "👻・Disruptor"
	HEALER_ROLE_NAME        = "🚑・Healer"
	DEBUFFER_ROLE_NAME      = "🏴・Debuffer"
	TANK_ROLE_NAME          = "🔰・Tank"
	ARCHIVE_CATEGORY        = "📚・Arquivo"
	RECRUIT_ROLE_NAME       = "🌱・Recruta"
)

var ACCESS_ROLE_IDS = map[string]string{
	EVERYONE_ROLE_NAME: "",
	MEMBER_ROLE_NAME:   "",
	ADMIN_ROLE_NAME:    "",
	CONSUL_ROLE_NAME:   "",
	OFFICER_ROLE_NAME:  "",
}

var CLASS_ROLE_IDS = map[string]string{
	BRUISER_ROLE_NAME:       "",
	MAGE_ROLE_NAME:          "",
	ASSASSIN_ROLE_NAME:      "",
	HEALER_ROLE_NAME:        "",
	DEBUFFER_ROLE_NAME:      "",
	TANK_ROLE_NAME:          "",
	DISRUPTOR_ROLE_NAME:     "",
	ARCO_MOSQUETE_ROLE_NAME: "",
	RECRUIT_ROLE_NAME:       "",
}

var CLASS_CATEGORY_IDS = map[string]string{
	BRUISER_ROLE_NAME:       "",
	MAGE_ROLE_NAME:          "",
	ASSASSIN_ROLE_NAME:      "",
	HEALER_ROLE_NAME:        "",
	DEBUFFER_ROLE_NAME:      "",
	TANK_ROLE_NAME:          "",
	DISRUPTOR_ROLE_NAME:     "",
	ARCO_MOSQUETE_ROLE_NAME: "",
	RECRUIT_ROLE_NAME:       "",
	ARCHIVE_CATEGORY:        "",
}

var BUILD_LEADER_ROLE_NAMES = map[string]string{
	BRUISER_ROLE_NAME:       "・BL Bruiser",
	MAGE_ROLE_NAME:          "・BL Mago",
	ASSASSIN_ROLE_NAME:      "・BL Assassino",
	HEALER_ROLE_NAME:        "・BL Healer",
	DEBUFFER_ROLE_NAME:      "・BL Debuffer",
	TANK_ROLE_NAME:          "・BL Tank",
	DISRUPTOR_ROLE_NAME:     "・BL Disruptor",
	ARCO_MOSQUETE_ROLE_NAME: "・BL Arco/Mosquete",
}

var CLASS_LEADER_ROLE_IDS = map[string]string{
	BRUISER_ROLE_NAME:       "",
	MAGE_ROLE_NAME:          "",
	ASSASSIN_ROLE_NAME:      "",
	HEALER_ROLE_NAME:        "",
	DEBUFFER_ROLE_NAME:      "",
	TANK_ROLE_NAME:          "",
	DISRUPTOR_ROLE_NAME:     "",
	ARCO_MOSQUETE_ROLE_NAME: "",
}

func Setup(ctx context.Context, dg *discordgo.Session, AppID, GuildID *string, db types.Database) {
	fmt.Println("Loading globals")
	guild, err := dg.State.Guild(*GuildID)
	if err != nil {
		log.Fatalf("Cannot get guild: %v", err)
	}

	for roleName := range ACCESS_ROLE_IDS {
		role := GetRoleByName(guild, roleName)
		if role == nil {
			continue
		}
		ACCESS_ROLE_IDS[roleName] = role.ID
		fmt.Printf("Found Access Role %s: %s\n", roleName, role.ID)
	}

	for roleName := range CLASS_ROLE_IDS {
		role := GetRoleByName(guild, roleName)
		if role == nil {
			continue
		}
		CLASS_ROLE_IDS[roleName] = role.ID
		fmt.Printf("Found Class Role %s: %s\n", roleName, role.ID)
	}

	for roleName := range BUILD_LEADER_ROLE_NAMES {
		role := GetRoleByName(guild, BUILD_LEADER_ROLE_NAMES[roleName])
		if role == nil {
			continue
		}
		CLASS_LEADER_ROLE_IDS[roleName] = role.ID
		fmt.Printf("Found Build Leader Role %s: %s\n", roleName, role.ID)
	}

	channels, err := dg.GuildChannels(*GuildID)
	if err != nil {
		log.Fatalf("Cannot get guild channels: %v", err)
	}

	for roleName := range CLASS_CATEGORY_IDS {
		for _, channel := range channels {
			if channel.Name == roleName {
				CLASS_CATEGORY_IDS[roleName] = channel.ID
				fmt.Printf("Found Class Category %s: %s\n", roleName, channel.ID)
				break
			}
		}
	}
}

var knownRoles = map[string]discordgo.Role{}

func GetRoleByName(guild *discordgo.Guild, roleName string) *discordgo.Role {
	if role, ok := knownRoles[roleName]; ok {
		return &role
	}
	for _, role := range guild.Roles {
		if role.Name == roleName {
			knownRoles[roleName] = *role
			return role
		}
	}

	return nil
}
