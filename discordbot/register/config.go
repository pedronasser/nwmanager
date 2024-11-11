package register

import (
	"nwmanager/discordbot/constants"

	"github.com/bwmarrin/discordgo"
)

var (
	WELCOME_CHANNEL_ID    = "1271479520297353309"
	OFFICER_ROLE_ID       = "1271561606710689938"
	MEMBER_ROLE_ID        = "1271559707198357655"
	BOT_CHANNELS_CATEGORY = "1271479520297353307"
	EVERYONE_ROLE_ID      = "1271479520297353306"

	MIN_WEAPON_OPTIONS = 2

	GUILD_RULES = []string{
		"**USO DO DISCORD OBRIGATÓRIO** se estiver logado no jogo.",
		"**INATIVIDADE MÁXIMA** permitida (sem aviso antecipado): :seven: **DIAS**",
		"Mantenha seu **TICKET¹** sempre respondido.",
	}

	WEAPON_OPTIONS = []discordgo.SelectMenuOption{
		{
			Label: constants.WEAPON_NAME_SWORD_SHIELD,
			Value: constants.WEAPON_SWORD_SHIELD,
			Emoji: &discordgo.ComponentEmoji{
				Name: "🛡️",
			},
		},
		{
			Label: constants.WEAPON_NAME_TWO_HANDED_SWORD,
			Value: constants.WEAPON_TWO_HANDED_SWORD,
			Emoji: &discordgo.ComponentEmoji{
				Name: "⚔️",
			},
		},
		{
			Label: constants.WEAPON_NAME_BOW,
			Value: constants.WEAPON_BOW,
			Emoji: &discordgo.ComponentEmoji{
				Name: "🏹",
			},
		},
		{
			Label: constants.WEAPON_NAME_AXE,
			Value: constants.WEAPON_AXE,
			Emoji: &discordgo.ComponentEmoji{
				Name: "🪓",
			},
		},
		{
			Label: constants.WEAPON_NAME_STAFF,
			Value: constants.WEAPON_STAFF,
			Emoji: &discordgo.ComponentEmoji{
				Name: "🧙",
			},
		},
		{
			Label: constants.WEAPON_NAME_DAGGER,
			Value: constants.WEAPON_DAGGER,
			Emoji: &discordgo.ComponentEmoji{
				Name: "🗡️",
			},
		},
		{
			Label: constants.WEAPON_NAME_WAND,
			Value: constants.WEAPON_WAND,
			Emoji: &discordgo.ComponentEmoji{
				Name: "🪄",
			},
		},
	}

	WEEK_DAYS_OPTIONS = []discordgo.SelectMenuOption{
		{
			Label: constants.WEEKDAY_NAME_MONDAY,
			Value: constants.WEEKDAY_MONDAY,
		},
		{
			Label: constants.WEEKDAY_NAME_TUESDAY,
			Value: constants.WEEKDAY_TUESDAY,
		},
		{
			Label: constants.WEEKDAY_NAME_WEDNESDAY,
			Value: constants.WEEKDAY_WEDNESDAY,
		},
		{
			Label: constants.WEEKDAY_NAME_THURSDAY,
			Value: constants.WEEKDAY_THURSDAY,
		},
		{
			Label: constants.WEEKDAY_NAME_FRIDAY,
			Value: constants.WEEKDAY_FRIDAY,
		},
		{
			Label: constants.WEEKDAY_NAME_SATURDAY,
			Value: constants.WEEKDAY_SATURDAY,
		},
		{
			Label: constants.WEEKDAY_NAME_SUNDAY,
			Value: constants.WEEKDAY_SUNDAY,
		},
	}

	TIME_OPTIONS = []discordgo.SelectMenuOption{
		{
			Label: constants.TIME_NAME_MORNING,
			Value: constants.TIME_MORNING,
		},
		{
			Label: constants.TIME_NAME_AFTERNOON,
			Value: constants.TIME_AFTERNOON,
		},
		{
			Label: constants.TIME_NAME_18,
			Value: constants.TIME_18,
		},
		{
			Label: constants.TIME_NAME_19,
			Value: constants.TIME_19,
		},
		{
			Label: constants.TIME_NAME_20,
			Value: constants.TIME_20,
		},
		{
			Label: constants.TIME_NAME_21,
			Value: constants.TIME_21,
		},
		{
			Label: constants.TIME_NAME_22,
			Value: constants.TIME_22,
		},
		{
			Label: constants.TIME_NAME_23,
			Value: constants.TIME_23,
		},
		{
			Label: constants.TIME_NAME_DAWN,
			Value: constants.TIME_DAWN,
		},
	}
)
