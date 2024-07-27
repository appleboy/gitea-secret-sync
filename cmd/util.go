package main

import (
	"os"
	"strings"
)

func toBool(s string) bool {
	if s == "" {
		return false
	}
	return s == "true"
}

func getGlobalValue(key string) string {
	key = strings.ToUpper(key) // Convert key to uppercase

	// Check if there is an environment variable with the format "INPUT_<KEY>"
	if value := os.Getenv("INPUT_" + key); value != "" {
		return value // Return the value of the "INPUT_<KEY>" environment variable
	}

	// If the "INPUT_<KEY>" environment variable doesn't exist or is empty,
	// return the value of the "<KEY>" environment variable
	return os.Getenv(key)
}

func getDataFromEnv(keys []string) map[string]string {
	keysMap := make(map[string]string)
	for _, key := range keys {
		val := getGlobalValue(key)
		if val == "" {
			continue
		}
		keysMap[key] = val
	}
	return keysMap
}
