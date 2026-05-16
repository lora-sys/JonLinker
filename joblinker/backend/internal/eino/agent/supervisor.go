package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
)

func NewA2ASupervisor(ctx context.Context, seekerAgent, recruiterAgent adk.Agent) (adk.ResumableAgent, error) {
	// Use seeker as the supervisor that delegates to recruiter when needed
	s, err := supervisor.New(ctx, &supervisor.Config{
		Supervisor: seekerAgent,
		SubAgents:  []adk.Agent{recruiterAgent},
	})
	if err != nil {
		return nil, fmt.Errorf("new a2a supervisor: %w", err)
	}
	log.Printf("A2A Supervisor created: seeker=%s, recruiter=%s",
		seekerAgent.Name(ctx), recruiterAgent.Name(ctx))
	return s, nil
}
