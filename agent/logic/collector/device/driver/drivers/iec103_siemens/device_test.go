package iec103_siemens

import (
	"testing"
)

// TestParseAddr 测试解析地址
func TestParseAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint32
		hasError bool
	}{
		{
			name:     "valid address 0x201",
			input:    "0x201",
			expected: 0x201,
			hasError: false,
		},
		{
			name:     "valid address 0x102",
			input:    "0x102",
			expected: 0x102,
			hasError: false,
		},
		{
			name:     "valid address 0xFFF",
			input:    "0xFFF",
			expected: 0xFFF,
			hasError: false,
		},
		{
			name:     "not 0x prefix",
			input:    "201",
			expected: 201,
			hasError: false,
		},
		{
			name:     "invalid format - empty string",
			input:    "",
			expected: 0,
			hasError: true,
		},
		{
			name:     "invalid format - too short",
			input:    "0x",
			expected: 0,
			hasError: true,
		},
		{
			name:     "invalid hex characters",
			input:    "0xG12",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAddr(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("parseAddr(%s) expected error, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("parseAddr(%s) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("parseAddr(%s) = 0x%X, want 0x%X", tt.input, result, tt.expected)
				}
			}
		})
	}
}
