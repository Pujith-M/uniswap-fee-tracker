package utils

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestParseInt64(t *testing.T) {
	tests := []struct {
		input          interface{}
		expectedOutput int64
		expectError    bool
	}{
		// Test cases
		{"123", 123, false},       // Valid string input
		{"-123", -123, false},     // Valid negative string input
		{123.0, 123, false},       // Valid float64 input
		{"invalid", 0, true},      // Invalid string input
		{123.456, 123, false},     // Float with decimals (truncated)
		{true, 0, true},           // Invalid type
		{nil, 0, true},            // Nil input
		{[]int{1, 2, 3}, 0, true}, // Invalid type
	}

	for _, test := range tests {
		output, err := ParseInt64(test.input)
		if test.expectError {
			if err == nil {
				t.Errorf("Expected an error for input %v, but got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Did not expect an error for input %v, but got: %v", test.input, err)
			}
			if output != test.expectedOutput {
				t.Errorf("Expected output %d for input %v, but got %d", test.expectedOutput, test.input, output)
			}
		}
	}
}

func TestParseBigFloat(t *testing.T) {
	tests := []struct {
		input          interface{}
		expectedOutput string
		expectError    bool
	}{
		{"123.456", "123.456", false},   // Valid string input
		{"-123.456", "-123.456", false}, // Valid negative string input
		{123.456, "123.456", false},     // Valid float64 input
		{123.0, "123", false},           // Float without decimals
		{"invalid", "", true},           // Invalid string input
		{true, "", true},                // Invalid type
		{nil, "", true},                 // Nil input
		{[]string{"1.2"}, "", true},     // Invalid type
	}

	for _, test := range tests {
		output, err := ParseBigFloat(test.input)
		if test.expectError {
			if err == nil {
				t.Errorf("Expected an error for input %v, but got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Did not expect an error for input %v, but got: %v", test.input, err)
			}
			expected, ok := new(big.Float).SetString(test.expectedOutput)
			if !ok {
				t.Fatalf("Failed to parse expected output %s as *big.Float", test.expectedOutput)
			}
			assert.Equal(t, expected.String(), output.String(), "Expected output %s for input %v, but got %s", test.expectedOutput, test.input, output.String())
		}
	}
}
