package globals

import (
	"github.com/joho/godotenv"
)

const (
	SEPARATOR = "・"
)

var (
	ADMIN_ROLE_ID,
	DB_PREFIX string
)

var ACCESS_ROLE_IDS map[string]string
var CLASS_ROLE_IDS map[string]string
var CLASS_CATEGORY_IDS map[string]string

func init() {
	_ = godotenv.Load()

	ACCESS_ROLE_IDS = map[string]string{}
	CLASS_ROLE_IDS = map[string]string{}
	CLASS_CATEGORY_IDS = map[string]string{}
}
