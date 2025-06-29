package helpers

import (
	"log"
	"os"
)

func LoadOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Println("Using default value for", key, ":", defaultValue)
		return defaultValue
	}
	log.Println("Loaded", key, ":", value)
	return value
}
