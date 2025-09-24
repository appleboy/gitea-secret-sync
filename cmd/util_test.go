package main

import (
	"os"
	"reflect"
	"testing"
)

func TestToBool(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", false},
		{"true lowercase", "true", true},
		{"false lowercase", "false", false},
		{"true uppercase", "TRUE", true},
		{"false uppercase", "FALSE", false},
		{"true mixed case", "True", true},
		{"false mixed case", "False", false},
		{"numeric true", "1", true},
		{"numeric false", "0", false},
		{"single char true", "t", true},
		{"single char false", "f", false},
		{"single char true uppercase", "T", true},
		{"single char false uppercase", "F", false},
		{"invalid string", "foo", false},
		{"invalid number", "2", false},
		{"whitespace", " ", false},
		{"yes", "yes", false}, // strconv.ParseBool doesn't accept "yes"
		{"no", "no", false},   // strconv.ParseBool doesn't accept "no"
		// Additional edge cases for better coverage
		{"unicode true", "tｒue", false},   // Should fail parsing
		{"unicode false", "fａlse", false}, // Should fail parsing
		{"tab character", "\t", false},
		{"carriage return", "\r", false},
		{"newline", "\n", false},
		{"mixed whitespace", " \t\n ", false},
		{"leading/trailing spaces on true", " true ", false}, // strconv.ParseBool doesn't trim
		{"leading/trailing spaces on false", " false ", false},
		{"null character", "\x00", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toBool(tt.input)
			if result != tt.expected {
				t.Errorf("toBool(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetGlobalValue(t *testing.T) {
	// Save original env vars to restore later
	originalEnv := make(map[string]string)
	testKeys := []string{"TEST_KEY1", "TEST_KEY2", "INPUT_TEST_KEY2", "INPUT_TEST_KEY3", "EMPTY_KEY"}

	for _, key := range testKeys {
		if val, exists := os.LookupEnv(key); exists {
			originalEnv[key] = val
		}
	}

	// Cleanup function to restore environment
	cleanup := func() {
		for _, key := range testKeys {
			if err := os.Unsetenv(key); err != nil {
				panic("failed to unset env key " + key + ": " + err.Error())
			}
		}
		for key, val := range originalEnv {
			if err := os.Setenv(key, val); err != nil {
				panic("failed to reset env key " + key + ": " + err.Error())
			}
		}
	}
	defer cleanup()

	tests := []struct {
		name        string
		setupEnv    map[string]string
		key         string
		expected    string
		description string
	}{
		{
			name:        "basic env var",
			setupEnv:    map[string]string{"TEST_KEY1": "value1"},
			key:         "test_key1",
			expected:    "value1",
			description: "should return value from basic env var with case conversion",
		},
		{
			name:        "INPUT prefix priority",
			setupEnv:    map[string]string{"TEST_KEY2": "value2", "INPUT_TEST_KEY2": "input_value2"},
			key:         "test_key2",
			expected:    "input_value2",
			description: "should prioritize INPUT_ prefix over basic env var",
		},
		{
			name:        "INPUT prefix only",
			setupEnv:    map[string]string{"INPUT_TEST_KEY3": "input_value3"},
			key:         "test_key3",
			expected:    "input_value3",
			description: "should return INPUT_ prefixed value when only INPUT_ exists",
		},
		{
			name:        "case insensitive key",
			setupEnv:    map[string]string{"TEST_KEY1": "value1"},
			key:         "Test_Key1",
			expected:    "value1",
			description: "should handle mixed case keys",
		},
		{
			name:        "non-existent key",
			setupEnv:    map[string]string{},
			key:         "NON_EXISTENT_KEY",
			expected:    "",
			description: "should return empty string for non-existent key",
		},
		{
			name:        "empty env var",
			setupEnv:    map[string]string{"EMPTY_KEY": ""},
			key:         "empty_key",
			expected:    "",
			description: "should return empty string when env var is set to empty",
		},
		// Additional test cases for edge cases
		{
			name:        "key with numbers",
			setupEnv:    map[string]string{"TEST_KEY_WITH_NUMBERS123": "value123"},
			key:         "test_key_with_numbers123",
			expected:    "value123",
			description: "should handle keys with numbers",
		},
		{
			name:        "empty string key",
			setupEnv:    map[string]string{},
			key:         "",
			expected:    "",
			description: "should return empty string for empty key",
		},
		{
			name:        "key with special characters",
			setupEnv:    map[string]string{"INPUT_HYPHEN-KEY": "hyphen_value"},
			key:         "hyphen-key",
			expected:    "hyphen_value",
			description: "should handle special characters in keys",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing env vars
			cleanup()

			// Set up environment for this test
			for key, value := range tt.setupEnv {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("Failed to set %s: %v", key, err)
				}
			}

			result := getGlobalValue(tt.key)
			if result != tt.expected {
				t.Errorf("getGlobalValue(%q) = %q, expected %q. %s", tt.key, result, tt.expected, tt.description)
			}
		})
	}
}

func TestGetDataFromEnv(t *testing.T) {
	// Save original env vars to restore later
	originalEnv := make(map[string]string)
	testKeys := []string{"DATA_KEY1", "DATA_KEY2", "INPUT_DATA_KEY3", "EMPTY_KEY"}

	for _, key := range testKeys {
		if val, exists := os.LookupEnv(key); exists {
			originalEnv[key] = val
		}
	}

	// Cleanup function to restore environment
	cleanup := func() {
		for _, key := range testKeys {
			if err := os.Unsetenv(key); err != nil {
				panic("failed to unset env key " + key + ": " + err.Error())
			}
		}
		for key, val := range originalEnv {
			if err := os.Setenv(key, val); err != nil {
				panic("failed to reset env key " + key + ": " + err.Error())
			}
		}
	}
	defer cleanup()

	tests := []struct {
		name     string
		setupEnv map[string]string
		keys     []string
		expected map[string]string
	}{
		{
			name: "multiple existing keys",
			setupEnv: map[string]string{
				"DATA_KEY1":       "value1",
				"DATA_KEY2":       "value2",
				"INPUT_DATA_KEY3": "value3",
			},
			keys:     []string{"data_key1", "data_key2", "data_key3"},
			expected: map[string]string{"data_key1": "value1", "data_key2": "value2", "data_key3": "value3"},
		},
		{
			name:     "empty keys slice",
			setupEnv: map[string]string{"DATA_KEY1": "value1"},
			keys:     []string{},
			expected: map[string]string{},
		},
		{
			name:     "non-existent keys",
			setupEnv: map[string]string{},
			keys:     []string{"non_existent1", "non_existent2"},
			expected: map[string]string{},
		},
		{
			name: "mixed existing and non-existent keys",
			setupEnv: map[string]string{
				"DATA_KEY1": "value1",
				"EMPTY_KEY": "",
			},
			keys:     []string{"data_key1", "non_existent", "empty_key"},
			expected: map[string]string{"data_key1": "value1"},
		},
		{
			name:     "nil keys",
			setupEnv: map[string]string{"DATA_KEY1": "value1"},
			keys:     nil,
			expected: map[string]string{},
		},
		// Additional edge cases
		{
			name:     "duplicate keys in slice",
			setupEnv: map[string]string{"DATA_KEY1": "value1"},
			keys:     []string{"data_key1", "data_key1", "data_key1"},
			expected: map[string]string{"data_key1": "value1"},
		},
		{
			name: "INPUT prefix priority in batch",
			setupEnv: map[string]string{
				"BATCH_KEY1":       "normal_value1",
				"INPUT_BATCH_KEY1": "input_value1",
				"BATCH_KEY2":       "normal_value2",
			},
			keys:     []string{"batch_key1", "batch_key2"},
			expected: map[string]string{"batch_key1": "input_value1", "batch_key2": "normal_value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing env vars

			cleanup()

			// Set up environment for this test
			for key, value := range tt.setupEnv {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("Failed to set %s: %v", key, err)
				}
			}

			result := getDataFromEnv(tt.keys)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("getDataFromEnv(%v) = %v, expected %v", tt.keys, result, tt.expected)
			}
		})
	}
}

func TestSplitByCommaOrNewline(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil, // Updated to match new function behavior
		},
		{
			name:     "comma separated",
			input:    "TEST_SECRET_1,TEST_SECRET_2,API_KEY",
			expected: []string{"TEST_SECRET_1", "TEST_SECRET_2", "API_KEY"},
		},
		{
			name:     "newline separated",
			input:    "TEST_SECRET_1\nTEST_SECRET_2\nAPI_KEY",
			expected: []string{"TEST_SECRET_1", "TEST_SECRET_2", "API_KEY"},
		},
		{
			name:     "mixed comma and newline",
			input:    "TEST_SECRET_1,TEST_SECRET_2\nAPI_KEY",
			expected: []string{"TEST_SECRET_1", "TEST_SECRET_2", "API_KEY"},
		},
		{
			name:     "with extra spaces",
			input:    " TEST_SECRET_1 , TEST_SECRET_2 \n API_KEY ",
			expected: []string{"TEST_SECRET_1", "TEST_SECRET_2", "API_KEY"},
		},
		{
			name:     "with empty items",
			input:    "TEST_SECRET_1,,TEST_SECRET_2,\nAPI_KEY,",
			expected: []string{"TEST_SECRET_1", "TEST_SECRET_2", "API_KEY"},
		},
		{
			name:     "single item",
			input:    "TEST_SECRET_1",
			expected: []string{"TEST_SECRET_1"},
		},
		{
			name:     "trailing newline",
			input:    "TEST_SECRET_1\nTEST_SECRET_2\n",
			expected: []string{"TEST_SECRET_1", "TEST_SECRET_2"},
		},
		{
			name:     "only whitespace",
			input:    "   \n  \n  ",
			expected: nil, // Updated behavior: whitespace-only strings return nil
		},
		{
			name:     "only separators",
			input:    ",,,\n\n",
			expected: nil, // Updated behavior: only separators return nil
		},
		{
			name:     "leading and trailing separators",
			input:    ",\nTEST_SECRET_1,TEST_SECRET_2\n,",
			expected: []string{"TEST_SECRET_1", "TEST_SECRET_2"},
		},
		{
			name:     "mixed whitespace and separators",
			input:    " , \n , TEST_SECRET_1 , \n ",
			expected: []string{"TEST_SECRET_1"},
		},
		// Additional edge cases for splitByCommaOrNewline
		{
			name:     "carriage return and newline",
			input:    "item1\r\nitem2\r\nitem3",
			expected: []string{"item1", "item2", "item3"},
		},
		{
			name:     "tabs as whitespace",
			input:    "\titem1\t,\titem2\t\n\titem3\t",
			expected: []string{"item1", "item2", "item3"},
		},
		{
			name:     "unicode characters",
			input:    "测试1,測試2\n테스트3",
			expected: []string{"测试1", "測試2", "테스트3"},
		},
		{
			name:     "very long item names",
			input:    "VERY_LONG_SECRET_NAME_WITH_MANY_CHARACTERS_1,VERY_LONG_SECRET_NAME_WITH_MANY_CHARACTERS_2\nVERY_LONG_SECRET_NAME_WITH_MANY_CHARACTERS_3",
			expected: []string{"VERY_LONG_SECRET_NAME_WITH_MANY_CHARACTERS_1", "VERY_LONG_SECRET_NAME_WITH_MANY_CHARACTERS_2", "VERY_LONG_SECRET_NAME_WITH_MANY_CHARACTERS_3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitByCommaOrNewline(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("splitByCommaOrNewline(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkToBool(b *testing.B) {
	testValues := []string{"true", "false", "1", "0", "invalid"}

	for i := 0; i < b.N; i++ {
		value := testValues[i%len(testValues)]
		toBool(value)
	}
}

func BenchmarkGetGlobalValue(b *testing.B) {
	// Set up a test environment variable
	if err := os.Setenv("BENCH_TEST_KEY", "test_value"); err != nil {
		b.Fatalf("failed to set env: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("BENCH_TEST_KEY"); err != nil {
			b.Fatalf("failed to unset env: %v", err)
		}
	}()

	for i := 0; i < b.N; i++ {
		getGlobalValue("bench_test_key")
	}
}

func BenchmarkGetDataFromEnv(b *testing.B) {
	// Set up test environment variables
	testKeys := []string{"BENCH_KEY1", "BENCH_KEY2", "BENCH_KEY3"}
	for _, key := range testKeys {
		os.Setenv(key, "value_"+key)
	}
	defer func() {
		for _, key := range testKeys {
			os.Unsetenv(key)
		}
	}()

	keys := []string{"bench_key1", "bench_key2", "bench_key3"}

	for i := 0; i < b.N; i++ {
		getDataFromEnv(keys)
	}
}

// Additional benchmark test specifically for splitByCommaOrNewline with various input patterns
func BenchmarkSplitByCommaOrNewlinePatterns(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{"short_comma", "a,b,c"},
		{"short_newline", "a\nb\nc"},
		{"mixed_short", "a,b\nc"},
		{"long_comma", "item1,item2,item3,item4,item5,item6,item7,item8"},
		{"long_newline", "item1\nitem2\nitem3\nitem4\nitem5\nitem6\nitem7\nitem8"},
		{"long_mixed", "item1,item2\nitem3,item4\nitem5,item6\nitem7,item8"},
		{"with_spaces", " item1 , item2 \n item3 , item4 "},
		{"empty", ""},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				splitByCommaOrNewline(tc.input)
			}
		})
	}
}
