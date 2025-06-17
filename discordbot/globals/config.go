package globals

import (
	"context"
	"fmt"
	"log"
	"nwmanager/helpers"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const (
	SEPARATOR = "・"
)

var (
	ADMIN_ROLE_ID,
	DB_PREFIX,
	EVERYONE_ROLE_NAME,
	MEMBER_ROLE_NAME,
	ADMIN_ROLE_NAME,
	CONSUL_ROLE_NAME,
	OFFICER_ROLE_NAME,
	OPR_ROLE_NAME,
	RAID_DEVOUR_ROLE,
	RAID_GORGON_ROLE,
	BRUISER_ROLE_NAME,
	MAGE_ROLE_NAME,
	ASSASSIN_ROLE_NAME,
	ARCO_MOSQUETE_ROLE_NAME,
	DISRUPTOR_ROLE_NAME,
	HEALER_ROLE_NAME,
	DEBUFFER_ROLE_NAME,
	TANK_ROLE_NAME,
	ARCHIVE_CATEGORY,
	RECRUIT_ROLE_NAME string
)

var ACCESS_ROLE_IDS map[string]string
var CLASS_ROLE_IDS map[string]string
var CLASS_CATEGORY_IDS map[string]string

func init() {
	_ = godotenv.Load()
	DB_PREFIX = helpers.LoadOrDefault("DB_PREFIX", "")

	EVERYONE_ROLE_NAME = helpers.LoadOrDefault("EVERYONE_ROLE_NAME", "@everyone")
	// MEMBER_ROLE_NAME = helpers.LoadOrDefault("MEMBER_ROLE_NAME", "👥・Membro")
	// ADMIN_ROLE_NAME = helpers.LoadOrDefault("ADMIN_ROLE_NAME", "👑 Governador")
	// CONSUL_ROLE_NAME = helpers.LoadOrDefault("CONSUL_ROLE_NAME", "💎Consul")
	// OFFICER_ROLE_NAME = helpers.LoadOrDefault("OFFICER_ROLE_NAME", "🏆Oficial")
	// OPR_ROLE_NAME = helpers.LoadOrDefault("OPR_ROLE_NAME", "⚔️・OPR")
	// RAID_DEVOUR_ROLE = helpers.LoadOrDefault("RAID_DEVOUR_ROLE", "🪱・Devorador")
	// RAID_GORGON_ROLE = helpers.LoadOrDefault("RAID_GORGON_ROLE", "🗿・Gorgonas")
	// BRUISER_ROLE_NAME = helpers.LoadOrDefault("BRUISER_ROLE_NAME", "🪓・Bruiser")
	// MAGE_ROLE_NAME = helpers.LoadOrDefault("MAGE_ROLE_NAME", "🧙・Mago")
	// ASSASSIN_ROLE_NAME = helpers.LoadOrDefault("ASSASSIN_ROLE_NAME", "😈・Assassino")
	// ARCO_MOSQUETE_ROLE_NAME = helpers.LoadOrDefault("ARCO_MOSQUETE_ROLE_NAME", "🏹・Arco/Mosquete")
	// DISRUPTOR_ROLE_NAME = helpers.LoadOrDefault("DISRUPTOR_ROLE_NAME", "👻・Disruptor")
	// HEALER_ROLE_NAME = helpers.LoadOrDefault("HEALER_ROLE_NAME", "🚑・Healer")
	// DEBUFFER_ROLE_NAME = helpers.LoadOrDefault("DEBUFFER_ROLE_NAME", "🏴・Debuffer")
	// TANK_ROLE_NAME = helpers.LoadOrDefault("TANK_ROLE_NAME", "🔰・Tank")
	// ARCHIVE_CATEGORY = helpers.LoadOrDefault("ARCHIVE_CATEGORY", "📚・Arquivo")
	// RECRUIT_ROLE_NAME = helpers.LoadOrDefault("RECRUIT_ROLE_NAME", "🌱・Recruta")

	ADMIN_ROLE_ID = helpers.LoadOrDefault("ADMIN_ROLE_ID", "")

	ACCESS_ROLE_IDS = map[string]string{
		EVERYONE_ROLE_NAME: "",
		MEMBER_ROLE_NAME:   "",
		ADMIN_ROLE_NAME:    "",
		CONSUL_ROLE_NAME:   "",
		OFFICER_ROLE_NAME:  "",
	}

	CLASS_ROLE_IDS = map[string]string{
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

	CLASS_CATEGORY_IDS = map[string]string{
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
}

// var BUILD_LEADER_ROLE_NAMES = map[string]string{
// 	BRUISER_ROLE_NAME:       "・BL Bruiser",
// 	MAGE_ROLE_NAME:          "・BL Mago",
// 	ASSASSIN_ROLE_NAME:      "・BL Assassino",
// 	HEALER_ROLE_NAME:        "・BL Healer",
// 	DEBUFFER_ROLE_NAME:      "・BL Debuffer",
// 	TANK_ROLE_NAME:          "・BL Tank",
// 	DISRUPTOR_ROLE_NAME:     "・BL Disruptor",
// 	ARCO_MOSQUETE_ROLE_NAME: "・BL Arco/Mosquete",
// }

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

func Setup(ctx context.Context, dg *discordgo.Session, AppID, GuildID *string) {
	fmt.Println("Loading globals")
	guild, err := dg.State.Guild(*GuildID)
	if err != nil {
		log.Fatalf("Cannot get guild: %v", err)
	}

	for roleName := range ACCESS_ROLE_IDS {
		role := GetRoleByName(guild, roleName)
		if role == nil {
			fmt.Println("Role not found:", roleName)
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

	if ADMIN_ROLE_ID != "" {
		ACCESS_ROLE_IDS[ADMIN_ROLE_NAME] = ADMIN_ROLE_ID
		fmt.Println("Overriden Admin Role ID")
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
