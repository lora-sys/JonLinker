package proto

import (
	"testing"
	"time"
)

func TestWebSocketFrameSerializer_MarshalUnmarshal(t *testing.T) {
	serializer := NewWebSocketFrameSerializer()

	t.Run("marshal and unmarshal roundtrip", func(t *testing.T) {
		original := &WebSocketFrame{
			MessageType:   MessageType_TEXT,
			Payload:      []byte("hello world"),
			SequenceNum:   42,
			Timestamp:     time.Now().UnixMilli(),
			CacheKey:      "tool:query_jobs:abc123",
			CorrelationId: "corr-123",
			SchemaVersion: 1,
		}

		data, err := serializer.MarshalFrame(original)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		decoded, err := serializer.UnmarshalFrame(data)
		if err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if decoded.MessageType != original.MessageType {
			t.Errorf("MessageType = %v, want %v", decoded.MessageType, original.MessageType)
		}
		if string(decoded.Payload) != string(original.Payload) {
			t.Errorf("Payload = %q, want %q", string(decoded.Payload), string(original.Payload))
		}
		if decoded.SequenceNum != original.SequenceNum {
			t.Errorf("SequenceNum = %d, want %d", decoded.SequenceNum, original.SequenceNum)
		}
		if decoded.CacheKey != original.CacheKey {
			t.Errorf("CacheKey = %q, want %q", decoded.CacheKey, original.CacheKey)
		}
		if decoded.CorrelationId != original.CorrelationId {
			t.Errorf("CorrelationId = %q, want %q", decoded.CorrelationId, original.CorrelationId)
		}
	})
}

func TestWebSocketFrameSerializer_DetectFrameFormat(t *testing.T) {
	serializer := NewWebSocketFrameSerializer()

	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "empty data defaults to json",
			data: []byte{},
			want: "json",
		},
		{
			name: "protobuf varint field 1",
			data: []byte{0x08},
			want: "protobuf",
		},
		{
			name: "protobuf varint field 1 second byte",
			data: []byte{0x0F},
			want: "protobuf",
		},
		{
			name: "protobuf length delimited field 1",
			data: []byte{0x0A},
			want: "protobuf",
		},
		{
			name: "json object",
			data: []byte(`{"key": "value"}`),
			want: "json",
		},
		{
			name: "json array",
			data: []byte(`[1, 2, 3]`),
			want: "json",
		},
		{
			name: "unknown format",
			data: []byte("plain text"),
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := serializer.DetectFrameFormat(tt.data)
			if got != tt.want {
				t.Errorf("DetectFrameFormat(%v) = %q, want %q", tt.data, got, tt.want)
			}
		})
	}
}

func TestWebSocketFrameSerializer_SerializeAsJSON(t *testing.T) {
	serializer := NewWebSocketFrameSerializer()

	t.Run("serialize as JSON", func(t *testing.T) {
		frame := &WebSocketFrame{
			MessageType: MessageType_TOOL_CALL,
			Payload:     []byte(`{"name": "test"}`),
			SequenceNum: 1,
		}

		data, err := serializer.SerializeAsJSON(frame)
		if err != nil {
			t.Fatalf("failed to serialize as JSON: %v", err)
		}

		if len(data) == 0 {
			t.Error("expected non-empty data")
		}
	})

	t.Run("nil frame returns error", func(t *testing.T) {
		_, err := serializer.SerializeAsJSON(nil)
		if err == nil {
			t.Error("expected error for nil frame")
		}
	})
}

func TestGetMessageTypeName(t *testing.T) {
	tests := []struct {
		mt   MessageType
		want string
	}{
		{MessageType_TEXT, "TEXT"},
		{MessageType_TOOL_CALL, "TOOL_CALL"},
		{MessageType_TOOL_RESULT, "TOOL_RESULT"},
		{MessageType_STATE_CHANGE, "STATE_CHANGE"},
		{MessageType_MEMORY_UPDATE, "MEMORY_UPDATE"},
		{MessageType_HEARTBEAT, "HEARTBEAT"},
		{MessageType_ERROR, "ERROR"},
		{MessageType_MESSAGE_TYPE_UNSPECIFIED, "UNSPECIFIED"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := GetMessageTypeName(tt.mt)
			if got != tt.want {
				t.Errorf("GetMessageTypeName(%v) = %q, want %q", tt.mt, got, tt.want)
			}
		})
	}
}

func TestParseMessageType(t *testing.T) {
	tests := []struct {
		name string
		want MessageType
	}{
		{"TEXT", MessageType_TEXT},
		{"TOOL_CALL", MessageType_TOOL_CALL},
		{"TOOL_RESULT", MessageType_TOOL_RESULT},
		{"STATE_CHANGE", MessageType_STATE_CHANGE},
		{"MEMORY_UPDATE", MessageType_MEMORY_UPDATE},
		{"HEARTBEAT", MessageType_HEARTBEAT},
		{"ERROR", MessageType_ERROR},
		{"INVALID", MessageType_MESSAGE_TYPE_UNSPECIFIED},
		{"", MessageType_MESSAGE_TYPE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMessageType(tt.name)
			if got != tt.want {
				t.Errorf("ParseMessageType(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestWebSocketFrameSerializer_UnmarshalFrame_Errors(t *testing.T) {
	serializer := NewWebSocketFrameSerializer()

	t.Run("empty data returns error", func(t *testing.T) {
		_, err := serializer.UnmarshalFrame([]byte{})
		if err == nil {
			t.Error("expected error for empty data")
		}
	})

	t.Run("invalid protobuf returns error", func(t *testing.T) {
		_, err := serializer.UnmarshalFrame([]byte{0xFF, 0xFF, 0xFF})
		if err == nil {
			t.Error("expected error for invalid protobuf data")
		}
	})
}