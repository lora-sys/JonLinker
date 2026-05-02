package config

import (
	"fmt"
	"os"
	"strings"
)

// ProtobufMode controls which channels use Protobuf serialization
type ProtobufMode string

const (
	ProtobufModeNone       ProtobufMode = "none"       // All JSON, no Protobuf
	ProtobufModeInternal  ProtobufMode = "internal"    // Only internal service communication (API Gateway ↔ backend)
	ProtobufModeWebSocket  ProtobufMode = "websocket"  // Only WebSocket agent dialogue
	ProtobufModeAll        ProtobufMode = "all"        // All channels use Protobuf
)

// PROTOBUF_ENABLED env var controls protocol mode
// Values: "none" (default, JSON for debugging), "internal" (REST only), "websocket" (WS only), "all" (full Protobuf)

func GetProtobufMode() ProtobufMode {
	mode := strings.ToLower(os.Getenv("PROTOBUF_ENABLED"))
	switch mode {
	case "internal":
		return ProtobufModeInternal
	case "websocket":
		return ProtobufModeWebSocket
	case "all":
		return ProtobufModeAll
	default:
		return ProtobufModeNone
	}
}

// IsInternalProtobufEnabled returns true if internal service communication should use Protobuf
func IsInternalProtobufEnabled() bool {
	mode := GetProtobufMode()
	return mode == ProtobufModeInternal || mode == ProtobufModeAll
}

// IsWebSocketProtobufEnabled returns true if WebSocket messages should use Protobuf
func IsWebSocketProtobufEnabled() bool {
	mode := GetProtobufMode()
	return mode == ProtobufModeWebSocket || mode == ProtobufModeAll
}

// QUEUE_PROTOBUF_ENABLED for RabbitMQ message encoding
func IsQueueProtobufEnabled() bool {
	return strings.ToLower(os.Getenv("QUEUE_PROTOBUF_ENABLED")) == "true"
}

// ValidateProtobufMode validates that the mode string is valid
func ValidateProtobufMode(mode string) error {
	switch ProtobufMode(strings.ToLower(mode)) {
	case ProtobufModeNone, ProtobufModeInternal, ProtobufModeWebSocket, ProtobufModeAll:
		return nil
	default:
		return fmt.Errorf("invalid PROTOBUF_ENABLED value: %s (valid: none, internal, websocket, all)", mode)
	}
}
