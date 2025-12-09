package config

import (
	"os"
	"testing"
)

func TestGetEnvAsString(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		setEnv       bool
		defaultValue string
		expected     string
	}{
		{
			name:         "env variable exists",
			envKey:       "TEST_STRING",
			envValue:     "hello",
			setEnv:       true,
			defaultValue: "default",
			expected:     "hello",
		},
		{
			name:         "env variable does not exist",
			envKey:       "TEST_STRING_MISSING",
			setEnv:       false,
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "env variable is empty string",
			envKey:       "TEST_STRING_EMPTY",
			envValue:     "",
			setEnv:       true,
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "env variable with spaces",
			envKey:       "TEST_STRING_SPACES",
			envValue:     "  value with spaces  ",
			setEnv:       true,
			defaultValue: "default",
			expected:     "  value with spaces  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := GetEnvAsString(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		setEnv       bool
		defaultValue int
		expected     int
	}{
		{
			name:         "valid positive integer",
			envKey:       "TEST_INT_POSITIVE",
			envValue:     "42",
			setEnv:       true,
			defaultValue: 0,
			expected:     42,
		},
		{
			name:         "valid negative integer",
			envKey:       "TEST_INT_NEGATIVE",
			envValue:     "-100",
			setEnv:       true,
			defaultValue: 0,
			expected:     -100,
		},
		{
			name:         "valid zero",
			envKey:       "TEST_INT_ZERO",
			envValue:     "0",
			setEnv:       true,
			defaultValue: 99,
			expected:     0,
		},
		{
			name:         "env variable does not exist",
			envKey:       "TEST_INT_MISSING",
			setEnv:       false,
			defaultValue: 123,
			expected:     123,
		},
		{
			name:         "invalid integer - not a number",
			envKey:       "TEST_INT_INVALID",
			envValue:     "not_a_number",
			setEnv:       true,
			defaultValue: 456,
			expected:     456,
		},
		{
			name:         "invalid integer - float",
			envKey:       "TEST_INT_FLOAT",
			envValue:     "123.456",
			setEnv:       true,
			defaultValue: 789,
			expected:     789,
		},
		{
			name:         "invalid integer - empty string",
			envKey:       "TEST_INT_EMPTY",
			envValue:     "",
			setEnv:       true,
			defaultValue: 999,
			expected:     999,
		},
		{
			name:         "large integer",
			envKey:       "TEST_INT_LARGE",
			envValue:     "2147483647",
			setEnv:       true,
			defaultValue: 0,
			expected:     2147483647,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := GetEnvAsInt(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		setEnv       bool
		defaultValue bool
		expected     bool
	}{
		{
			name:         "valid true - lowercase",
			envKey:       "TEST_BOOL_TRUE_LOWER",
			envValue:     "true",
			setEnv:       true,
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "valid true - uppercase",
			envKey:       "TEST_BOOL_TRUE_UPPER",
			envValue:     "TRUE",
			setEnv:       true,
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "valid true - number 1",
			envKey:       "TEST_BOOL_TRUE_ONE",
			envValue:     "1",
			setEnv:       true,
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "valid false - lowercase",
			envKey:       "TEST_BOOL_FALSE_LOWER",
			envValue:     "false",
			setEnv:       true,
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "valid false - uppercase",
			envKey:       "TEST_BOOL_FALSE_UPPER",
			envValue:     "FALSE",
			setEnv:       true,
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "valid false - number 0",
			envKey:       "TEST_BOOL_FALSE_ZERO",
			envValue:     "0",
			setEnv:       true,
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "env variable does not exist",
			envKey:       "TEST_BOOL_MISSING",
			setEnv:       false,
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "invalid boolean - not a bool",
			envKey:       "TEST_BOOL_INVALID",
			envValue:     "not_a_bool",
			setEnv:       true,
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "invalid boolean - empty string",
			envKey:       "TEST_BOOL_EMPTY",
			envValue:     "",
			setEnv:       true,
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "invalid boolean - number 2",
			envKey:       "TEST_BOOL_TWO",
			envValue:     "2",
			setEnv:       true,
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "valid true - mixed case",
			envKey:       "TEST_BOOL_TRUE_MIXED",
			envValue:     "True",
			setEnv:       true,
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "valid false - mixed case",
			envKey:       "TEST_BOOL_FALSE_MIXED",
			envValue:     "False",
			setEnv:       true,
			defaultValue: true,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := GetEnvAsBool(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
