package globals

type PVPClassType string

const (
	// User can select multiple classes
	PVP_CLASS_DISRUPTOR       PVPClassType = "disruptor"
	PVP_CLASS_HEALER          PVPClassType = "healer"
	PVP_CLASS_TANK            PVPClassType = "tank"
	PVP_CLASS_FIRE_ICE        PVPClassType = "fire_ice"
	PVP_CLASS_ICE_FLAIL       PVPClassType = "ice_flail"
	PVP_CLASS_VOID_ICE        PVPClassType = "void_ice"
	PVP_CLASS_BRUISER         PVPClassType = "bruiser"
	PVP_CLASS_RANGED_ASSASSIN PVPClassType = "ranged_assassin"
	PVP_CLASS_HEALER_AOE      PVPClassType = "healer_aoe"
	PVP_CLASS_OFF_META        PVPClassType = "off_meta"
)

var PVP_CLASS_NAMES = map[PVPClassType]string{
	PVP_CLASS_DISRUPTOR:       "Disruptor",
	PVP_CLASS_HEALER:          "Healer",
	PVP_CLASS_HEALER_AOE:      "Healer AOE",
	PVP_CLASS_TANK:            "Tank",
	PVP_CLASS_FIRE_ICE:        "Fire/Ice (Mago)",
	PVP_CLASS_VOID_ICE:        "Void/Ice (Debuffer)",
	PVP_CLASS_ICE_FLAIL:       "Ice/Flail (Cleanser)",
	PVP_CLASS_BRUISER:         "Bruiser",
	PVP_CLASS_RANGED_ASSASSIN: "Ranged Assassin (Bow/Musket)",
	PVP_CLASS_OFF_META:        "Off Meta (Outras classes)",
}
