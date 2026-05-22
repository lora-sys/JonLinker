package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/filesystem"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

func NewDeepRecruiterAgent(ctx context.Context, chatModel model.BaseChatModel, instruction string, subAgents []adk.Agent, tools []tool.BaseTool, backend filesystem.Backend) (adk.ResumableAgent, error) {
	cfg := &deep.Config{
		Name:        "deep-recruiter",
		Description: "Deep recruiter agent that decomposes complex hiring tasks",
		ChatModel:   chatModel,
		Instruction: instruction,
		SubAgents:   subAgents,
		MaxIteration: 50,
	}
	if len(tools) > 0 {
		cfg.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		}
	}
	if backend != nil {
		cfg.Backend = backend
	}
	d, err := deep.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("new deep recruiter: %w", err)
	}
	log.Printf("DeepRecruiterAgent created with tools=%d backend=%v", len(tools), backend != nil)
	return d, nil
}
