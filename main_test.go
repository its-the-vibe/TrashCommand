package main

import (
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"DEBUG", LogLevelDebug},
		{"debug", LogLevelDebug},
		{"INFO", LogLevelInfo},
		{"info", LogLevelInfo},
		{"ERROR", LogLevelError},
		{"error", LogLevelError},
		{"", LogLevelInfo},
		{"unknown", LogLevelInfo},
	}

	for _, tt := range tests {
		got := parseLogLevel(tt.input)
		if got != tt.expected {
			t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestIsBot(t *testing.T) {
	tests := []struct {
		name     string
		event    ReactionEvent
		expected bool
	}{
		{
			name: "non-bot user",
			event: ReactionEvent{
				Event: Event{User: "U111"},
				Authorizations: []Auth{
					{UserID: "U111", IsBot: false},
				},
			},
			expected: false,
		},
		{
			name: "bot user",
			event: ReactionEvent{
				Event: Event{User: "U222"},
				Authorizations: []Auth{
					{UserID: "U222", IsBot: true},
				},
			},
			expected: true,
		},
		{
			name: "user not in authorizations",
			event: ReactionEvent{
				Event: Event{User: "U333"},
				Authorizations: []Auth{
					{UserID: "U444", IsBot: true},
				},
			},
			expected: false,
		},
		{
			name:     "no authorizations",
			event:    ReactionEvent{Event: Event{User: "U555"}},
			expected: false,
		},
		{
			name: "multiple authorizations, reactor is bot",
			event: ReactionEvent{
				Event: Event{User: "U666"},
				Authorizations: []Auth{
					{UserID: "U777", IsBot: false},
					{UserID: "U666", IsBot: true},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBot(tt.event)
			if got != tt.expected {
				t.Errorf("isBot() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	const key = "TEST_GETENV_VAR"

	t.Run("returns default when unset", func(t *testing.T) {
		got := getEnv(key, "default")
		if got != "default" {
			t.Errorf("getEnv() = %q, want %q", got, "default")
		}
	})

	t.Run("returns env value when set", func(t *testing.T) {
		t.Setenv(key, "hello")
		got := getEnv(key, "default")
		if got != "hello" {
			t.Errorf("getEnv() = %q, want %q", got, "hello")
		}
	})
}

func TestGetEnvInt(t *testing.T) {
	const key = "TEST_GETENVINT_VAR"

	t.Run("returns default when unset", func(t *testing.T) {
		got := getEnvInt(key, 42)
		if got != 42 {
			t.Errorf("getEnvInt() = %d, want %d", got, 42)
		}
	})

	t.Run("returns parsed int when set", func(t *testing.T) {
		t.Setenv(key, "99")
		got := getEnvInt(key, 42)
		if got != 99 {
			t.Errorf("getEnvInt() = %d, want %d", got, 99)
		}
	})

	t.Run("returns default on invalid int", func(t *testing.T) {
		t.Setenv(key, "notanumber")
		got := getEnvInt(key, 42)
		if got != 42 {
			t.Errorf("getEnvInt() = %d, want %d", got, 42)
		}
	})
}
