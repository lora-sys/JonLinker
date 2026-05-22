package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	"joblinker/internal/eino/prompt/templates"
)

func NewChatModelAgent(ctx context.Context, name string, description string, systemPrompt string, chatModel model.ToolCallingChatModel, tools []tool.BaseTool) (adk.Agent, error) {
	cfg := &adk.ChatModelAgentConfig{
		Name:        name,
		Description: description,
		Instruction: systemPrompt,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		},
		MaxIterations: 5,
	}

	a, err := adk.NewChatModelAgent(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("new chat model agent %s: %w", name, err)
	}

	log.Printf("ChatModelAgent created: name=%s, tools=%d", name, len(tools))
	return a, nil
}

func NewSeekerChatModelAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools []tool.BaseTool) (adk.Agent, error) {
	return NewChatModelAgent(ctx,
		"seeker",
		"Job seeker agent that evaluates opportunities and negotiates terms",
		templates.SeekerSystemPrompt,
		chatModel,
		tools,
	)
}

func NewRecruiterChatModelAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools []tool.BaseTool) (adk.Agent, error) {
	return NewChatModelAgent(ctx,
		"recruiter",
		"Recruiter agent that posts jobs and evaluates candidates",
		templates.RecruiterSystemPrompt,
		chatModel,
		tools,
	)
}

// NewGeneralRecruiterAgent creates a general-purpose recruiter ChatModelAgent
// for use in the routing supervisor as the fallback agent.
func NewGeneralRecruiterAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools []tool.BaseTool) (adk.Agent, error) {
	a, err := NewRecruiterChatModelAgent(ctx, chatModel, tools)
	if err != nil {
		return nil, err
	}
	log.Printf("GeneralRecruiterAgent created")
	return a, nil
}
