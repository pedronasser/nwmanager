package helpers

import (
	"fmt"
	"os"
)

func LoadOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		fmt.Println("Using default value for", key, ":", defaultValue)
		return defaultValue
	}
	fmt.Println("Loaded", key, ":", value)
	return value
}
