package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
)

// NewRoutingSupervisor creates a routing supervisor that delegates to specialized
// sub-agents based on intent. The supervisor agent's instruction prompt defines
// routing rules, and the LLM decides which sub-agent to invoke.
//
// Routing logic (defined in the supervisor's system prompt):
//   - "screening / initial evaluation"     → screeningAgent
//   - "interview / technical assessment"   → interviewAgent
//   - "offer / salary negotiation"         → offerAgent
//   - "general recruitment"                → generalAgent (recruiter)
//
// The supervisor also handles the conversation itself when no sub-agent is needed.
func NewRoutingSupervisor(ctx context.Context,
	screeningAgent, interviewAgent, offerAgent, generalAgent adk.Agent,
) (adk.ResumableAgent, error) {

	routingInstruction := `You are a recruitment supervisor that routes user requests to the right specialist agent.

## Available Agents
- **agent-screening**: Handles initial candidate screening, resume evaluation, and basic qualification questions.
- **agent-interview**: Conducts technical interviews, asks assessment questions, and evaluates candidate answers.
- **agent-offer**: Manages job offers, salary negotiations, acceptance, and decline handling.
- **recruiter**: General recruitment agent that handles all other recruitment-related tasks.

## Routing Rules
Analyze the user's request and delegate to the most appropriate agent:

1. If the user asks about initial screening, qualifications, resume review → route to **agent-screening**
2. If the user needs a technical interview, skill assessment, or evaluation → route to **agent-interview**
3. If the user discusses salary, offers, acceptance, or decline → route to **agent-offer**
4. For all other recruitment tasks → route to **recruiter**

## Behavior
- Always route to one agent — do not try to handle the request yourself
- If the request spans multiple domains, route to the most relevant agent
- Provide brief context when delegating`
	routingAgent, err := NewChatModelAgent(ctx,
		"supervisor-router",
		"Routing supervisor that delegates to specialist recruitment agents",
		routingInstruction,
		nil, // no tools needed for the supervisor itself
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("new routing supervisor agent: %w", err)
	}

	s, err := supervisor.New(ctx, &supervisor.Config{
		Supervisor: routingAgent,
		SubAgents: []adk.Agent{
			screeningAgent,
			interviewAgent,
			offerAgent,
			generalAgent,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("new routing supervisor: %w", err)
	}

	log.Printf("RoutingSupervisor created with sub-agents: screening, interview, offer, general")
	return s, nil
}
