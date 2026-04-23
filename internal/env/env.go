package env

import "os"

// GetEnv retrieves the value of the environment variable named by the key. If the variable is not present in the environment, then it returns the defaultValue.
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
