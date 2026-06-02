package memory

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/schema"
)

type MemoryStore interface {
	Read(ctx context.Context, sessionID string) ([]*schema.Message, error)
	Write(ctx context.Context, sessionID string, msgs []*schema.Message) error
}

func EncodeMessages(msgs []*schema.Message) ([]byte, error) {
	return json.Marshal(msgs)
}

func DecodeMessages(b []byte) ([]*schema.Message, error) {
	if len(b) == 0 {
		return nil, nil
	}
	var msgs []*schema.Message
	if err := json.Unmarshal(b, &msgs); err != nil {
		return nil, err
	}
	return msgs, nil
}
