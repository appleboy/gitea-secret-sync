package main

import (
	"reflect"
	"testing"
)

func TestSplitByCommaOrNewline(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
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
