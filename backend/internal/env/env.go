package env

import (
	"os"
	"strconv"
)

func GetString(key, fallback string) string {
	val := os.Getenv(key)

	if val == "" {
		return fallback
	}

	return val
}

func GetInt(key string, fallback int) int {
	valAsString := os.Getenv(key)
	valAsInt, err := strconv.Atoi(valAsString)

	if (err != nil) {
		return fallback
	}

	return valAsInt
}
