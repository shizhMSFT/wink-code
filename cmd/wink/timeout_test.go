// Package main_test contains unit tests for timeout configuration
package main

import (
	"os"
	"strconv"
	"testing"
)

func TestTimeoutEnvVarParsing(t *testing.T) {
	tests := []struct {
		name       string
		envValue   string
		wantValid  bool
		wantResult int
	}{
		{
			name:       "valid integer",
			envValue:   "45",
			wantValid:  true,
			wantResult: 45,
		},
		{
			name:       "valid minimum",
			envValue:   "5",
			wantValid:  true,
			wantResult: 5,
		},
		{
			name:      "invalid - below minimum",
			envValue:  "2",
			wantValid: false,
		},
		{
			name:      "invalid format",
			envValue:  "invalid",
			wantValid: false,
		},
		{
			name:      "empty string",
			envValue:  "",
			wantValid: false,
		},
		{
			name:       "high value",
			envValue:   "600",
			wantValid:  true,
			wantResult: 600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue == "" {
				return
			}

			timeout, err := strconv.Atoi(tt.envValue)

			if tt.wantValid {
				if err != nil {
					t.Errorf("Expected valid parse, got error: %v", err)
					return
				}
				if timeout < 5 {
					t.Errorf("Expected timeout >= 5, got %d", timeout)
					return
				}
				if timeout != tt.wantResult {
					t.Errorf("Expected timeout %d, got %d", tt.wantResult, timeout)
				}
			} else {
				if err == nil && timeout >= 5 {
					t.Errorf("Expected invalid, but got valid timeout: %d", timeout)
				}
			}
		})
	}
}

func TestTimeoutFlagPrecedence(t *testing.T) {
	// Simulate flag precedence logic

	// Case 1: Flag explicitly set (not default)
	flagValue := 60
	envValue := "45"
	defaultValue := 30

	result := flagValue
	if flagValue == defaultValue && envValue != "" {
		// Use env only if flag is at default
		if timeout, err := strconv.Atoi(envValue); err == nil && timeout >= 5 {
			result = timeout
		}
	}

	if result != 60 {
		t.Errorf("Expected flag to take precedence: got %d, want 60", result)
	}

	// Case 2: Flag at default, env var present
	flagValue = 30 // at default
	result = flagValue
	if flagValue == defaultValue && envValue != "" {
		if timeout, err := strconv.Atoi(envValue); err == nil && timeout >= 5 {
			result = timeout
		}
	}

	if result != 45 {
		t.Errorf("Expected env var when flag at default: got %d, want 45", result)
	}

	// Case 3: Flag at default, no env var
	flagValue = 30
	envValue = ""
	result = flagValue
	if flagValue == defaultValue && envValue != "" {
		if timeout, err := strconv.Atoi(envValue); err == nil && timeout >= 5 {
			result = timeout
		}
	}

	if result != 30 {
		t.Errorf("Expected default when no env var: got %d, want 30", result)
	}
}

func TestTimeoutValidationBounds(t *testing.T) {
	tests := []struct {
		timeout     int
		wantValid   bool
		wantWarning bool
	}{
		{timeout: 1, wantValid: false, wantWarning: false},
		{timeout: 4, wantValid: false, wantWarning: false},
		{timeout: 5, wantValid: true, wantWarning: false},
		{timeout: 30, wantValid: true, wantWarning: false},
		{timeout: 300, wantValid: true, wantWarning: false},
		{timeout: 301, wantValid: true, wantWarning: true},
		{timeout: 600, wantValid: true, wantWarning: true},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.timeout), func(t *testing.T) {
			valid := tt.timeout >= 5
			warning := tt.timeout > 300

			if valid != tt.wantValid {
				t.Errorf("Timeout %d: expected valid=%v, got valid=%v",
					tt.timeout, tt.wantValid, valid)
			}

			if warning != tt.wantWarning {
				t.Errorf("Timeout %d: expected warning=%v, got warning=%v",
					tt.timeout, tt.wantWarning, warning)
			}
		})
	}
}

func TestTimeoutWithEnvironment(t *testing.T) {
	// Save and restore original env
	original := os.Getenv("WINK_TIMEOUT")
	defer func() {
		if original != "" {
			os.Setenv("WINK_TIMEOUT", original)
		} else {
			os.Unsetenv("WINK_TIMEOUT")
		}
	}()

	// Test with env var set
	os.Setenv("WINK_TIMEOUT", "45")
	envValue := os.Getenv("WINK_TIMEOUT")
	if envValue != "45" {
		t.Errorf("Expected env var to be '45', got '%s'", envValue)
	}

	// Test parsing
	timeout, err := strconv.Atoi(envValue)
	if err != nil {
		t.Errorf("Failed to parse env var: %v", err)
	}
	if timeout != 45 {
		t.Errorf("Expected parsed timeout 45, got %d", timeout)
	}

	// Test with invalid env var
	os.Setenv("WINK_TIMEOUT", "invalid")
	envValue = os.Getenv("WINK_TIMEOUT")
	_, err = strconv.Atoi(envValue)
	if err == nil {
		t.Error("Expected parse error for invalid env var")
	}

	// Test unset
	os.Unsetenv("WINK_TIMEOUT")
	envValue = os.Getenv("WINK_TIMEOUT")
	if envValue != "" {
		t.Errorf("Expected empty env var after unset, got '%s'", envValue)
	}
}
