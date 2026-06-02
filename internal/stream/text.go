package stream

import (
	"context"
	"io"
	"time"
)

func WriteChunked(w io.Writer, text string, sleep time.Duration, ctx context.Context) error {
	runes := []rune(text)
	chunkSize := 3
	for i := 0; i < len(runes); i += chunkSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		if _, err := io.WriteString(w, string(runes[i:end])); err != nil {
			return err
		}
		time.Sleep(sleep)
	}
	return nil
}
