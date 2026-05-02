package proto

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// ParseError represents a detailed Protobuf parse error with context
type ParseError struct {
	Message   string
	ByteOffset int
	Field     string
	Suggestion string
}

func (e *ParseError) Error() string {
	if e.ByteOffset >= 0 {
		return fmt.Sprintf("%s at byte %d (field: %s). %s", e.Message, e.ByteOffset, e.Field, e.Suggestion)
	}
	return fmt.Sprintf("%s (field: %s). %s", e.Message, e.Field, e.Suggestion)
}

// NewParseError creates a new parse error with context
func NewParseError(msg string, offset int, field, suggestion string) *ParseError {
	return &ParseError{
		Message:    msg,
		ByteOffset: offset,
		Field:      field,
		Suggestion: suggestion,
	}
}

// ContentType represents the serialization format
type ContentType string

const (
	ContentTypeJSON     ContentType = "application/json"
	ContentTypeProtobuf ContentType = "application/x-protobuf"
)

// DetectContentType determines if incoming data is Protobuf by peeking first byte
// Protobuf varint encoding: field 1 wire type 0 (varint) starts with 0x08-0x0F
// field 1 wire type 2 (length-delimited) starts with 0x0A
func DetectContentType(data []byte) ContentType {
	if len(data) == 0 {
		return ContentTypeJSON
	}
	firstByte := data[0]
	// Protobuf-encoded messages have predictable first bytes for field 1
	// 0x08-0x0F: field 1 varint (wire type 0)
	// 0x0A-0x0B: field 1 length-delimited (wire type 2)
	// 0x12-0x13: field 2 length-delimited
	if (firstByte >= 0x08 && firstByte <= 0x0F) || (firstByte >= 0x0A && firstByte <= 0x0B) || (firstByte >= 0x12 && firstByte <= 0x13) {
		return ContentTypeProtobuf
	}
	return ContentTypeJSON
}

// ParseAcceptHeader parses the Accept header and returns the preferred content type
func ParseAcceptHeader(accept string) ContentType {
	if accept == "" {
		return ContentTypeJSON
	}
	if containsProtoAccept(accept) {
		return ContentTypeProtobuf
	}
	return ContentTypeJSON
}

func containsProtoAccept(accept string) bool {
	return contains(accept, "application/x-protobuf") || contains(accept, "application/vnd.google.protobuf")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetWireType returns the Protobuf wire type from the first byte
func GetWireType(data []byte) (fieldNumber int, wireType int, err error) {
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("empty data")
	}
	firstByte := data[0]
	fieldNumber = int(firstByte) >> 3
	wireType = int(firstByte) & 0x07
	return fieldNumber, wireType, nil
}

// PeekMessageSize peeks at the varint length prefix to estimate message size
func PeekMessageSize(data []byte) (size int, ok bool) {
	if len(data) < 1 {
		return 0, false
	}
	// Simple length-delimited message: first byte is tag, second byte is length varint
	if data[0] == 0x0A || data[0] == 0x12 { // length-delimited field 1 or 2
		if len(data) >= 2 {
			size = int(data[1])
			return size, true
		}
	}
	return 0, false
}

// BinaryMarshal serializes a message to Protobuf binary format
func BinaryMarshal(msg interface{}) ([]byte, error) {
	// This is a placeholder - actual marshaling uses generated proto code
	// The generated .pb.go files have Marshal methods
	return nil, fmt.Errorf("use generated proto marshaling methods")
}

// BinaryUnmarshal deserializes a message from Protobuf binary format
func BinaryUnmarshal(buf []byte, msg interface{}) error {
	// This is a placeholder - actual unmarshaling uses generated proto code
	return fmt.Errorf("use generated proto unmarshaling methods")
}

// ValidateSchemaVersion checks if the schema version is compatible
func ValidateSchemaVersion(version uint32, supported []uint32) bool {
	for _, v := range supported {
		if v == version {
			return true
		}
	}
	return false
}

// ValidateSchemaVersionWithError returns an error with details if version is not supported
func ValidateSchemaVersionWithError(version uint32, supported []uint32, schemaName string) error {
	if !ValidateSchemaVersion(version, supported) {
		return NewParseError(
			fmt.Sprintf("unsupported schema version %d for %s", version, schemaName),
			-1,
			"schema_version",
			fmt.Sprintf("supported versions: %v; upgrade client or server", supported),
		)
	}
	return nil
}

// SafeUnmarshal attempts to unmarshal Protobuf data and returns a detailed error on failure
func SafeUnmarshal(data []byte, msg proto.Message) error {
	if len(data) == 0 {
		return NewParseError("empty data", 0, "root", "ensure message is not empty")
	}

	// Try to detect common issues
	fieldNum, wireType, err := GetWireType(data)
	if err != nil {
		return NewParseError(err.Error(), -1, "header", "verify message uses Protobuf encoding")
	}

	// Check for wire type mismatches (common error)
	if wireType == 5 || wireType == 7 { // Fixed32/Fixed64 - rarely used in our messages
		return NewParseError(
			"unexpected wire type for common fields",
			0,
			fmt.Sprintf("field %d wire type %d", fieldNum, wireType),
			"verify .proto schema matches encoded data",
		)
	}

	// Attempt unmarshal
	if err := proto.Unmarshal(data, msg); err != nil {
		return NewParseError(
			fmt.Sprintf("unmarshal failed: %v", err),
			0,
			"root",
			"fallback to JSON if Protobuf parsing fails; check schema version compatibility",
		)
	}

	return nil
}

// DetectAndReportError analyzes a failed parse and returns actionable error information
func DetectAndReportError(data []byte) *ParseError {
	if len(data) == 0 {
		return NewParseError("empty data", 0, "root", "provide valid Protobuf message")
	}

	firstByte := data[0]

	// Check if it looks like JSON
	if firstByte == '{' || firstByte == '[' {
		return NewParseError(
			"message appears to be JSON, not Protobuf",
			0,
			"header",
			"use Content-Type: application/json or check if Protobuf encoding was intended",
		)
	}

	// Check if it looks like XML
	if firstByte == '<' {
		return NewParseError(
			"message appears to be XML, not Protobuf",
			0,
			"header",
			"use XML parser or check if Protobuf encoding was intended",
		)
	}

	// Protobuf parsing failed
	fieldNum, wireType, _ := GetWireType(data)
	return NewParseError(
		"invalid Protobuf encoding",
		0,
		fmt.Sprintf("field %d wire type %d", fieldNum, wireType),
		"verify message was encoded with same .proto schema; consider falling back to JSON",
	)
}
