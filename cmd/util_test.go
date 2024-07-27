package main

import (
	"os"
	"testing"
)

func TestToBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"true", true},
		{"false", false},
		{"foo", false},
	}

	for _, test := range tests {
		result := toBool(test.input)
		if result != test.expected {
			t.Errorf("ToBool(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestGetGlobalValue(t *testing.T) {
	os.Setenv("KEY1", "value1")
	os.Setenv("KEY2", "value2")
	os.Setenv("KEY3", "value3")
	os.Setenv("INPUT_KEY4", "value4")

	// test KEY1
	if val := getGlobalValue("KEY1"); val != "value1" {
		t.Errorf("Expected 'value1', but got '%s'", val)
	}

	// test KEY2
	if val := getGlobalValue("KEY2"); val != "value2" {
		t.Errorf("Expected 'value2', but got '%s'", val)
	}

	// test KEY3
	if val := getGlobalValue("KEY3"); val != "value3" {
		t.Errorf("Expected 'value3', but got '%s'", val)
	}

	// test KEY4
	if val := getGlobalValue("KEY4"); val != "value4" {
		t.Errorf("Expected 'value4', but got '%s'", val)
	}

	// test不存在的鍵
	if val := getGlobalValue("NON_EXISTENT_KEY"); val != "" {
		t.Errorf("Expected '', but got '%s'", val)
	}
}
