package env

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)


func GetString(key, fallback string) string {
	err := godotenv.Load()

	if err != nil {
		return fallback
	}

	val := os.Getenv(key)

	return val
}

func GetInt(key string, fallback int) int {
	err := godotenv.Load()

	if err != nil {
		return fallback
	}

	valAsString := os.Getenv(key)
	valAsInt, err := strconv.Atoi(valAsString)

	if err != nil {
		return fallback
	}

	return valAsInt
}