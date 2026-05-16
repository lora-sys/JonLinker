package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/model"
)

func NewDeepRecruiterAgent(ctx context.Context, chatModel model.BaseChatModel, instruction string, subAgents []adk.Agent) (adk.ResumableAgent, error) {
	d, err := deep.New(ctx, &deep.Config{
		Name:        "deep-recruiter",
		Description: "Deep recruiter agent that decomposes complex hiring tasks",
		ChatModel:   chatModel,
		Instruction: instruction,
		SubAgents:   subAgents,
	})
	if err != nil {
		return nil, fmt.Errorf("new deep recruiter: %w", err)
	}
	log.Printf("DeepRecruiterAgent created")
	return d, nil
}
