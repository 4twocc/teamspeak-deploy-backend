package utils

import "os"

// GetEnv retrieves the value of the environment variable named by the key.
// If the variable is empty, the fallback value is returned instead.
func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
