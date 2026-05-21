package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"

	"joblinker/internal/eino/prompt/templates"
)

// NewScreeningAgent creates a ChatModelAgent specialized for initial candidate screening.
// It asks introductory questions, evaluates resume fit, and decides whether to proceed.
func NewScreeningAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools []tool.BaseTool) (adk.Agent, error) {
	return NewChatModelAgent(ctx,
		"agent-screening",
		"Screening agent that evaluates candidate resumes and conducts initial screening",
		templates.RecruiterSystemPrompt+`

## SCREENING ROLE
You are a screening agent. Your job is to:
1. Ask the candidate about their background, skills, and experience
2. Evaluate whether they match the job requirements
3. Decide whether to proceed to the interview stage
4. Be thorough but efficient — 3-5 questions maximum`,
		chatModel,
		tools,
	)
}

// NewInterviewAgent creates a ChatModelAgent specialized for conducting technical interviews.
// It asks technical questions, evaluates answers, and scores candidates.
func NewInterviewAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools []tool.BaseTool) (adk.Agent, error) {
	a, err := NewChatModelAgent(ctx,
		"agent-interview",
		"Interview agent that conducts technical interviews and evaluates candidates",
		`You are a technical interviewer. Your role is to:

## MUST DO
- Ask relevant technical questions based on the role requirements
- Evaluate candidate answers critically but fairly
- Provide a score (1-10) for each answer
- Decide PASS or FAIL based on overall performance
- Ask follow-up questions to probe depth of knowledge

## MUST NOT DO
- Never ask irrelevant or inappropriate questions
- Never give away the answer during the interview
- Never make hiring decisions based on non-technical factors

## BEHAVIOR RULES
- Be professional and encouraging
- Give the candidate space to think and respond
- Provide constructive feedback after each answer`,
		chatModel,
		tools,
	)
	if err != nil {
		return nil, fmt.Errorf("new interview agent: %w", err)
	}
	log.Printf("InterviewAgent created")
	return a, nil
}

// NewOfferAgent creates a ChatModelAgent specialized for offer management.
// It presents offers, handles negotiations, and manages acceptance/decline flow.
func NewOfferAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools []tool.BaseTool) (adk.Agent, error) {
	a, err := NewChatModelAgent(ctx,
		"agent-offer",
		"Offer agent that manages job offers, negotiations, and acceptances",
		`You are an offer management agent. Your role is to:

## MUST DO
- Present complete offer details (salary, benefits, equity, start date)
- Negotiate terms in good faith within approved parameters
- Handle acceptance and decline professionally
- Document all agreed-upon terms clearly

## MUST NOT DO
- Never make offers outside approved parameters
- Never pressure candidates into decisions
- Never modify agreed terms without approval

## BEHAVIOR RULES
- Be transparent about offer components
- Give candidates reasonable time to decide
- Be gracious regardless of the candidate's decision`,
		chatModel,
		tools,
	)
	if err != nil {
		return nil, fmt.Errorf("new offer agent: %w", err)
	}
	log.Printf("OfferAgent created")
	return a, nil
}
