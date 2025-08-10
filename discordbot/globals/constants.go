package globals

type PVPClassType string

const (
	// User can select multiple classes
	PVP_CLASS_DISRUPTOR        PVPClassType = "disruptor"
	PVP_CLASS_DISRUPTOR_MEDIUM PVPClassType = "disruptor_medium"
	PVP_CLASS_HEALER           PVPClassType = "healer"
	PVP_CLASS_HEALER_SIEGE     PVPClassType = "healer_siege"
	PVP_CLASS_TANK             PVPClassType = "tank"
	PVP_CLASS_FIRE_ICE         PVPClassType = "fire_ice"
	PVP_CLASS_ICE_FLAIL        PVPClassType = "ice_flail"
	PVP_CLASS_VOID_ICE         PVPClassType = "void_ice"
	PVP_CLASS_VOID_BLADE       PVPClassType = "void_blade"
	PVP_CLASS_BRUISER          PVPClassType = "bruiser"
	PVP_CLASS_RANGED_ASSASSIN  PVPClassType = "ranged_assassin"
	PVP_CLASS_HEALER_AOE       PVPClassType = "healer_aoe"
	PVP_CLASS_OFF_META         PVPClassType = "off_meta"
)

var PVP_CLASS_NAMES = map[PVPClassType]string{
	PVP_CLASS_DISRUPTOR_MEDIUM: "Disruptor (Médio)",
	PVP_CLASS_DISRUPTOR:        "Disruptor (Pesado)",
	PVP_CLASS_HEALER:           "Healer (Ponto)",
	PVP_CLASS_HEALER_SIEGE:     "Healer (Siege)",
	PVP_CLASS_HEALER_AOE:       "Healer AOE",
	PVP_CLASS_TANK:             "Tank",
	PVP_CLASS_FIRE_ICE:         "Fire/Ice (Mago)",
	PVP_CLASS_VOID_ICE:         "Void/Ice (Debuffer)",
	PVP_CLASS_VOID_BLADE:       "Void Blade",
	PVP_CLASS_ICE_FLAIL:        "Ice/Flail (Cleanser)",
	PVP_CLASS_BRUISER:          "Bruiser",
	PVP_CLASS_RANGED_ASSASSIN:  "Ranged Assassin (Bow/Musket)",
	PVP_CLASS_OFF_META:         "Off Meta (Outras classes)",
}
