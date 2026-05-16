package runner

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"
)

type ADKRunner struct {
	runner    *adk.Runner
	store     compose.CheckPointStore
}

func NewADKRunner(ctx context.Context, agent adk.Agent, store compose.CheckPointStore) *ADKRunner {
	r := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: true,
		CheckPointStore: store,
	})
	return &ADKRunner{
		runner: r,
		store:  store,
	}
}

func (r *ADKRunner) Query(ctx context.Context, query string, opts ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	return r.runner.Query(ctx, query, opts...)
}

func (r *ADKRunner) Resume(ctx context.Context, checkPointID string, opts ...adk.AgentRunOption) (*adk.AsyncIterator[*adk.AgentEvent], error) {
	return r.runner.Resume(ctx, checkPointID, opts...)
}
