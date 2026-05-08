// Package eino provides Eino framework integration for JobLinker A2A recruitment system.
//
// This package wraps the CloudWeGo Eino framework to provide AI agent capabilities:
//
//   - ChatModel interface for AI client integration (see chatmodel package)
//   - Prompt template management via ChatTemplate (see prompt package)
//   - Tool registration via InvokableTool interface (see tools package)
//   - Memory management for conversation context (see memory package)
//   - Multi-agent coordination via DeepRecruiter (see agent package)
//   - Agent pooling via Runner (see runner package)
//
// Subpackages:
//   - agent: SeekerAgent, RecruiterAgent, DeepRecruiter
//   - chatmodel: EinoChatModel wrapper
//   - memory: AgentMemory, VectorRetriever
//   - prompt: Loader, ChatTemplate management
//   - runner: AgentRunner pool
//   - tools: Job, Candidate, Offer, Interview tools
//
// Example usage:
//
//	import (
//	    "joblinker/internal/eino/agent"
//	    "joblinker/internal/eino/runner"
//	    "joblinker/pkg/ai"
//	)
//
//	aiClient := ai.NewClient()
//	seeker := agent.NewSeekerAgent(aiClient)
//	runner := runner.NewAgentRunner(aiClient, nil)
package eino
