package proto

import (
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// WebSocketFrameSerializer handles marshaling/unmarshaling of WebSocket frames
type WebSocketFrameSerializer struct{}

// NewWebSocketFrameSerializer creates a new WebSocketFrame serializer
func NewWebSocketFrameSerializer() *WebSocketFrameSerializer {
	return &WebSocketFrameSerializer{}
}

// MarshalFrame serializes a WebSocketFrame to protobuf binary
func (s *WebSocketFrameSerializer) MarshalFrame(frame *WebSocketFrame) ([]byte, error) {
	if frame == nil {
		return nil, fmt.Errorf("frame is nil")
	}
	return proto.Marshal(frame)
}

// UnmarshalFrame deserializes a WebSocketFrame from protobuf binary
func (s *WebSocketFrameSerializer) UnmarshalFrame(data []byte) (*WebSocketFrame, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	frame := &WebSocketFrame{}
	if err := proto.Unmarshal(data, frame); err != nil {
		return nil, fmt.Errorf("failed to unmarshal WebSocketFrame: %w", err)
	}
	return frame, nil
}

// DetectFrameFormat detects if data is Protobuf or JSON based on first byte
func (s *WebSocketFrameSerializer) DetectFrameFormat(data []byte) string {
	if len(data) == 0 {
		return "json" // Default to JSON for empty data
	}
	firstByte := data[0]
	// Protobuf varint encoding for field 1 (MessageType field 1, wire type 0)
	// Field 1 wire type 0: 0x08-0x0F (field_number << 3 | wire_type = 1 << 3 | 0 = 8-15)
	// Field 1 wire type 2 (length-delimited): 0x0A-0x0B
	if (firstByte >= 0x08 && firstByte <= 0x0F) || (firstByte >= 0x0A && firstByte <= 0x0B) {
		return "protobuf"
	}
	// Check for JSON object/array
	if firstByte == '{' || firstByte == '[' {
		return "json"
	}
	return "unknown"
}

// SerializeAsProtobuf marshals the frame as Protobuf
func (s *WebSocketFrameSerializer) SerializeAsProtobuf(frame *WebSocketFrame) ([]byte, error) {
	return s.MarshalFrame(frame)
}

// SerializeAsJSON marshals the frame as JSON (for fallback/debugging)
func (s *WebSocketFrameSerializer) SerializeAsJSON(frame *WebSocketFrame) ([]byte, error) {
	if frame == nil {
		return nil, fmt.Errorf("frame is nil")
	}
	return protojson.Marshal(frame)
}

// DeserializeFromProtobuf deserializes frame from Protobuf binary
func (s *WebSocketFrameSerializer) DeserializeFromProtobuf(data []byte) (*WebSocketFrame, error) {
	return s.UnmarshalFrame(data)
}

// GetMessageTypeName returns the string name of a MessageType
func GetMessageTypeName(mt MessageType) string {
	switch mt {
	case MessageType_TEXT:
		return "TEXT"
	case MessageType_TOOL_CALL:
		return "TOOL_CALL"
	case MessageType_TOOL_RESULT:
		return "TOOL_RESULT"
	case MessageType_STATE_CHANGE:
		return "STATE_CHANGE"
	case MessageType_MEMORY_UPDATE:
		return "MEMORY_UPDATE"
	case MessageType_HEARTBEAT:
		return "HEARTBEAT"
	case MessageType_ERROR:
		return "ERROR"
	default:
		return "UNSPECIFIED"
	}
}

// ParseMessageType parses a MessageType from its string name
func ParseMessageType(name string) MessageType {
	switch name {
	case "TEXT":
		return MessageType_TEXT
	case "TOOL_CALL":
		return MessageType_TOOL_CALL
	case "TOOL_RESULT":
		return MessageType_TOOL_RESULT
	case "STATE_CHANGE":
		return MessageType_STATE_CHANGE
	case "MEMORY_UPDATE":
		return MessageType_MEMORY_UPDATE
	case "HEARTBEAT":
		return MessageType_HEARTBEAT
	case "ERROR":
		return MessageType_ERROR
	default:
		return MessageType_MESSAGE_TYPE_UNSPECIFIED
	}
}