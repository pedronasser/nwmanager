package common

const (
	WEAPON_SWORD_SHIELD     = "sword_shield"
	WEAPON_TWO_HANDED_SWORD = "two_handed_sword"
	WEAPON_BOW              = "bow"
	WEAPON_AXE              = "axe"
	WEAPON_STAFF            = "staff"
	WEAPON_DAGGER           = "dagger"
	WEAPON_WAND             = "wand"

	WEAPON_NAME_SWORD_SHIELD     = "Espada e Escudo"
	WEAPON_NAME_TWO_HANDED_SWORD = "Espada de duas mãos"
	WEAPON_NAME_BOW              = "Arco"
	WEAPON_NAME_AXE              = "Machado"
	WEAPON_NAME_STAFF            = "Cajado"
	WEAPON_NAME_DAGGER           = "Adaga"
	WEAPON_NAME_WAND             = "Varinha"

	TIME_MORNING   = "morning"
	TIME_AFTERNOON = "afternoon"
	TIME_18        = "18"
	TIME_19        = "19"
	TIME_20        = "20"
	TIME_21        = "21"
	TIME_22        = "22"
	TIME_23        = "23"
	TIME_DAWN      = "dawn"

	TIME_NAME_MORNING   = "Manhã"
	TIME_NAME_AFTERNOON = "Tarde"
	TIME_NAME_18        = "18h-19h"
	TIME_NAME_19        = "19h-20h"
	TIME_NAME_20        = "20h-21h"
	TIME_NAME_21        = "21h-22h"
	TIME_NAME_22        = "22h-23h"
	TIME_NAME_23        = "23h-00h"
	TIME_NAME_DAWN      = "Madrugada"

	WEEKDAY_MONDAY    = "monday"
	WEEKDAY_TUESDAY   = "tuesday"
	WEEKDAY_WEDNESDAY = "wednesday"
	WEEKDAY_THURSDAY  = "thursday"
	WEEKDAY_FRIDAY    = "friday"
	WEEKDAY_SATURDAY  = "saturday"
	WEEKDAY_SUNDAY    = "sunday"

	WEEKDAY_NAME_MONDAY    = "Segunda-feira"
	WEEKDAY_NAME_TUESDAY   = "Terça-feira"
	WEEKDAY_NAME_WEDNESDAY = "Quarta-feira"
	WEEKDAY_NAME_THURSDAY  = "Quinta-feira"
	WEEKDAY_NAME_FRIDAY    = "Sexta-feira"
	WEEKDAY_NAME_SATURDAY  = "Sábado"
	WEEKDAY_NAME_SUNDAY    = "Domingo"
)

var (
	WEAPONS = map[string]string{
		WEAPON_SWORD_SHIELD:     WEAPON_NAME_SWORD_SHIELD,
		WEAPON_TWO_HANDED_SWORD: WEAPON_NAME_TWO_HANDED_SWORD,
		WEAPON_BOW:              WEAPON_NAME_BOW,
		WEAPON_AXE:              WEAPON_NAME_AXE,
		WEAPON_STAFF:            WEAPON_NAME_STAFF,
		WEAPON_DAGGER:           WEAPON_NAME_DAGGER,
		WEAPON_WAND:             WEAPON_NAME_WAND,
	}

	TIMES = map[string]string{
		TIME_MORNING:   TIME_NAME_MORNING,
		TIME_AFTERNOON: TIME_NAME_AFTERNOON,
		TIME_18:        TIME_NAME_18,
		TIME_19:        TIME_NAME_19,
		TIME_20:        TIME_NAME_20,
		TIME_21:        TIME_NAME_21,
		TIME_22:        TIME_NAME_22,
		TIME_23:        TIME_NAME_23,
		TIME_DAWN:      TIME_NAME_DAWN,
	}

	WEEKDAYS = map[string]string{
		WEEKDAY_MONDAY:    WEEKDAY_NAME_MONDAY,
		WEEKDAY_TUESDAY:   WEEKDAY_NAME_TUESDAY,
		WEEKDAY_WEDNESDAY: WEEKDAY_NAME_WEDNESDAY,
		WEEKDAY_THURSDAY:  WEEKDAY_NAME_THURSDAY,
		WEEKDAY_FRIDAY:    WEEKDAY_NAME_FRIDAY,
		WEEKDAY_SATURDAY:  WEEKDAY_NAME_SATURDAY,
		WEEKDAY_SUNDAY:    WEEKDAY_NAME_SUNDAY,
	}
)
