package globals

type PVPClassType string

const (
	// User can select multiple classes
	PVP_CLASS_DISRUPTOR PVPClassType = "disruptor"
	PVP_CLASS_HEALER    PVPClassType = "healer"
	PVP_CLASS_TANK      PVPClassType = "tank"
	PVP_CLASS_FIRE      PVPClassType = "fire"
	PVP_CLASS_ICE_FLAIL PVPClassType = "ice_flail"
	PVP_CLASS_BRUISER   PVPClassType = "bruiser"
	PVP_CLASS_BOW       PVPClassType = "bow"
	PVP_CLASS_ASSASSIN  PVPClassType = "assassin"
)

var PVP_CLASS_NAMES = map[PVPClassType]string{
	PVP_CLASS_DISRUPTOR: "Disruptor",
	PVP_CLASS_HEALER:    "Healer",
	PVP_CLASS_TANK:      "Tank",
	PVP_CLASS_FIRE:      "Fire",
	PVP_CLASS_ICE_FLAIL: "Ice Flail",
	PVP_CLASS_BRUISER:   "Bruiser",
	PVP_CLASS_BOW:       "Bow",
	PVP_CLASS_ASSASSIN:  "Assassin",
}
