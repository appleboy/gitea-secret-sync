package main

import (
	"os"
	"strings"
)

func toBool(value string) bool {
	return strings.ToLower(value) == "true"
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

// splitByCommaOrNewline splits a string by both commas and newlines,
// trims whitespace from each item, and filters out empty strings
func splitByCommaOrNewline(input string) []string {
	if input == "" {
		return []string{}
	}

	var result []string

	// Replace newlines with commas to normalize the input
	normalized := strings.ReplaceAll(input, "\n", ",")

	// Split by commas
	parts := strings.Split(normalized, ",")

	// Trim whitespace and filter out empty strings
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
