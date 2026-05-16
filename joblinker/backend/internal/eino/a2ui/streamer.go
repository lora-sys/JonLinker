package a2ui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func StreamToWriter(ctx context.Context, w io.Writer, events *adk.AsyncIterator[*adk.AgentEvent], sessionID string) error {
	begin := Message{
		BeginRendering: &BeginRenderingMsg{
			SessionID: sessionID,
			RootID:    sessionID,
		},
	}
	if err := writeSSE(w, begin); err != nil {
		return fmt.Errorf("a2ui: write begin: %w", err)
	}

	state := NewSurfaceState(sessionID, sessionID)
	msgCount := 0

	for {
		event, ok := events.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			log.Printf("A2UI AgentEvent error: %v", event.Err)
			return event.Err
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			mo := event.Output.MessageOutput

			if mo.IsStreaming && mo.MessageStream != nil {
				if err := handleStream(ctx, w, state, mo.MessageStream, &msgCount); err != nil {
					return err
				}
			} else if mo.Message != nil {
				if err := handleMsg(ctx, w, state, mo.Message); err != nil {
					return err
				}
			}
		}

		if event.Action != nil && event.Action.TransferToAgent != nil {
			comp := Component{
				ID:   fmt.Sprintf("transfer-%d", msgCount),
				Type: "card",
				Props: map[string]interface{}{
					"content": fmt.Sprintf("Transferring to agent: %s", event.Action.TransferToAgent.DestAgentName),
					"hint":    "system",
				},
			}
			if err := writeSSE(w, Message{SurfaceUpdate: &SurfaceUpdateMsg{Components: []Component{comp}}}); err != nil {
				return err
			}
		}
	}

	return nil
}

func handleStream(ctx context.Context, w io.Writer, state *SurfaceState, stream *schema.StreamReader[*schema.Message], msgCount *int) error {
	compID := fmt.Sprintf("assistant-%d", *msgCount)
	state.AddComponent(Component{
		ID:      compID,
		Type:    "text",
		Props:   map[string]interface{}{"hint": "assistant"},
		DataKey: compID,
	})

	if err := writeSSE(w, Message{
		SurfaceUpdate: &SurfaceUpdateMsg{Components: []Component{state.Components[compID]}},
	}); err != nil {
		return err
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if msg.Content != "" {
			if err := writeSSE(w, Message{
				DataModelUpdate: &DataModelUpdateMsg{Key: compID, Delta: msg.Content},
			}); err != nil {
				return err
			}
		}
	}

	*msgCount++
	return nil
}

func handleMsg(ctx context.Context, w io.Writer, state *SurfaceState, msg *schema.Message) error {
	hint := ComponentHint(HintAssistant)
	if msg.Role == schema.Tool {
		hint = HintToolResult
	} else if len(msg.ToolCalls) > 0 {
		hint = HintToolCall
	}

	comp := AgentEventToComponent("msg", msg, hint)
	state.AddComponent(comp)

	return writeSSE(w, Message{
		SurfaceUpdate: &SurfaceUpdateMsg{Components: []Component{comp}},
	})
}

func writeSSE(w io.Writer, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "data: %s\n\n", data)
	return err
}
