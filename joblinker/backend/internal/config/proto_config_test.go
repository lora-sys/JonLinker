package config

import (
	"fmt"
	"os"
	"testing"
)

func TestGetProtobufMode(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		want    ProtobufMode
	}{
		{"default none", "", ProtobufModeNone},
		{"explicit none", "none", ProtobufModeNone},
		{"internal mode", "internal", ProtobufModeInternal},
		{"websocket mode", "websocket", ProtobufModeWebSocket},
		{"all mode", "all", ProtobufModeAll},
		{"uppercase none", "NONE", ProtobufModeNone},
		{"uppercase internal", "INTERNAL", ProtobufModeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal == "" {
				os.Unsetenv("PROTOBUF_ENABLED")
			} else {
				os.Setenv("PROTOBUF_ENABLED", tt.envVal)
			}
			got := GetProtobufMode()
			if got != tt.want {
				t.Errorf("GetProtobufMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsInternalProtobufEnabled(t *testing.T) {
	tests := []struct {
		envVal string
		want   bool
	}{
		{"none", false},
		{"internal", true},
		{"websocket", false},
		{"all", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.envVal, func(t *testing.T) {
			if tt.envVal == "" {
				os.Unsetenv("PROTOBUF_ENABLED")
			} else {
				os.Setenv("PROTOBUF_ENABLED", tt.envVal)
			}
			got := IsInternalProtobufEnabled()
			if got != tt.want {
				t.Errorf("IsInternalProtobufEnabled() with PROTOBUF_ENABLED=%q = %v, want %v", tt.envVal, got, tt.want)
			}
		})
	}
}

func TestIsWebSocketProtobufEnabled(t *testing.T) {
	tests := []struct {
		envVal string
		want   bool
	}{
		{"none", false},
		{"internal", false},
		{"websocket", true},
		{"all", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.envVal, func(t *testing.T) {
			if tt.envVal == "" {
				os.Unsetenv("PROTOBUF_ENABLED")
			} else {
				os.Setenv("PROTOBUF_ENABLED", tt.envVal)
			}
			got := IsWebSocketProtobufEnabled()
			if got != tt.want {
				t.Errorf("IsWebSocketProtobufEnabled() with PROTOBUF_ENABLED=%q = %v, want %v", tt.envVal, got, tt.want)
			}
		})
	}
}

func TestIsQueueProtobufEnabled(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		want    bool
	}{
		{"queue enabled explicitly", "true", true},
		{"queue disabled explicitly", "false", false},
		{"queue not set", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal == "" {
				os.Unsetenv("QUEUE_PROTOBUF_ENABLED")
			} else {
				os.Setenv("QUEUE_PROTOBUF_ENABLED", tt.envVal)
			}
			got := IsQueueProtobufEnabled()
			if got != tt.want {
				t.Errorf("IsQueueProtobufEnabled() with QUEUE_PROTOBUF_ENABLED=%q = %v, want %v", tt.envVal, got, tt.want)
			}
		})
	}
}

func TestValidateProtobufMode(t *testing.T) {
	tests := []struct {
		mode string
		want error
	}{
		{"none", nil},
		{"internal", nil},
		{"websocket", nil},
		{"all", nil},
		{"invalid", fmt.Errorf("invalid")},
		{"NONE", nil},
		{"Internal", nil},
		{"", fmt.Errorf("invalid")},
		{"partial", fmt.Errorf("invalid")},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			got := ValidateProtobufMode(tt.mode)
			if (got != nil) != (tt.want != nil) {
				t.Errorf("ValidateProtobufMode(%q) error = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}