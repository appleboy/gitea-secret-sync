// Package main provides utility functions for environment variable handling
// and string processing operations.
package main

import (
	"os"
	"strconv"
	"strings"
)

// toBool converts a string to a boolean using Go's standard parsing.
// Accepts values like "true", "false", "1", "0", "t", "f", "T", "F", etc.
// Returns false for any invalid input.
func toBool(value string) bool {
	result, err := strconv.ParseBool(value)
	if err != nil {
		return false // Default to false for invalid inputs
	}
	return result
}

// getGlobalValue retrieves environment variable with INPUT_ prefix fallback.
// First checks INPUT_<KEY>, then falls back to <KEY>.
// Returns empty string if neither exists.
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

// getDataFromEnv retrieves multiple environment variables and returns them as a map.
// Only includes keys that have non-empty values.
// Pre-allocates the map with the expected capacity for better performance.
func getDataFromEnv(keys []string) map[string]string {
	// Pre-allocate map with known capacity for better performance
	envVars := make(map[string]string, len(keys))

	for _, key := range keys {
		if val := getGlobalValue(key); val != "" {
			envVars[key] = val
		}
	}

	return envVars
}

// splitByCommaOrNewline splits a string by both commas and newlines,
// trims whitespace from each item, and filters out empty strings.
func splitByCommaOrNewline(input string) []string {
	if input == "" {
		return nil // Return nil instead of empty slice for consistency with Go idioms
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

	// Return nil if no valid items found (all were empty after trimming)
	if len(result) == 0 {
		return nil
	}

	return result
}
