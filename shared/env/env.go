package env

import (
	"os"
	"strconv"
)

// GetString gets an environment variable as a string or returns the fallback value.
func GetString(key string, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return fallback
}

// GetInt gets an environment variable as an integer or returns the fallback value.
func GetInt(key string, fallback int) int {
	if val, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return fallback
}

// GetBool gets an environment variable as a boolean or returns the fallback value.
func GetBool(key string, fallback bool) bool {
	if val, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return fallback
}
