package a2ui

import "github.com/cloudwego/eino/schema"

type Message struct {
	BeginRendering   *BeginRenderingMsg   `json:"beginRendering,omitempty"`
	SurfaceUpdate    *SurfaceUpdateMsg    `json:"surfaceUpdate,omitempty"`
	DataModelUpdate  *DataModelUpdateMsg  `json:"dataModelUpdate,omitempty"`
	DeleteSurface    *DeleteSurfaceMsg    `json:"deleteSurface,omitempty"`
	InterruptRequest *InterruptRequestMsg `json:"interruptRequest,omitempty"`
}

type BeginRenderingMsg struct {
	SessionID string `json:"sessionId"`
	RootID    string `json:"rootId"`
}

type SurfaceUpdateMsg struct {
	Components []Component `json:"components"`
}

type Component struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Props    map[string]interface{} `json:"props"`
	Children []string               `json:"children,omitempty"`
	DataKey  string                 `json:"dataKey,omitempty"`
}

type DataModelUpdateMsg struct {
	Key   string `json:"key"`
	Delta string `json:"delta"`
}

type DeleteSurfaceMsg struct {
	SessionID string `json:"sessionId"`
}

type InterruptRequestMsg struct {
	InterruptID string   `json:"interruptId"`
	Question    string   `json:"question"`
	Options     []string `json:"options,omitempty"`
}

type SurfaceState struct {
	RootID     string
	Components map[string]Component
	DataModels map[string]string
	SessionID  string
}

func NewSurfaceState(sessionID, rootID string) *SurfaceState {
	return &SurfaceState{
		RootID:     rootID,
		Components: make(map[string]Component),
		DataModels: make(map[string]string),
		SessionID:  sessionID,
	}
}

func (s *SurfaceState) AddComponent(c Component) {
	s.Components[c.ID] = c
}

type ComponentHint string

const (
	HintToolCall   ComponentHint = "tool_call"
	HintToolResult ComponentHint = "tool_result"
	HintAssistant  ComponentHint = "assistant"
	HintSystem     ComponentHint = "system"
)

func AgentEventToComponent(eventType string, msg *schema.Message, hint ComponentHint) Component {
	content := msg.Content
	compID := eventType + "-msg"
	if len(content) > 8 {
		compID = eventType + "-" + content[:8]
	}

	props := map[string]interface{}{
		"content": content,
		"role":    string(msg.Role),
		"hint":    string(hint),
	}

	if len(msg.ToolCalls) > 0 {
		props["tool_calls"] = msg.ToolCalls
	}

	return Component{
		ID:    compID,
		Type:  "card",
		Props: props,
	}
}
