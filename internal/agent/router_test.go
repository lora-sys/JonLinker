package agent

import (
	"encoding/json"
	"testing"
)

func TestRoute(t *testing.T) {
	completeProfile, _ := json.Marshal(map[string]any{
		"name":   "Alice",
		"title":  "Engineer",
		"skills": []string{"Go"},
	})
	incompleteProfile, _ := json.Marshal(map[string]any{
		"name": "Alice",
	})

	tests := []struct {
		name           string
		hasResumeAgent bool
		profileJSON    []byte
		message        string
		lastAgent      string
		hasApplication bool
		want           AgentKind
	}{
		// AgentResume
		{
			name:    "no profile",
			message: "hello",
			want:    AgentResume,
		},
		{
			name:        "incomplete profile",
			profileJSON: incompleteProfile,
			message:     "hello",
			want:        AgentResume,
		},
		{
			name:        "resume keyword",
			profileJSON: completeProfile,
			message:     "修改简历",
			want:        AgentResume,
		},
		// AgentSearch
		{
			name:        "search keyword",
			profileJSON: completeProfile,
			message:     "搜索职位",
			want:        AgentSearch,
		},
		{
			name:        "default fallback to search",
			profileJSON: completeProfile,
			message:     "你好",
			lastAgent:   "search",
			want:        AgentSearch,
		},
		{
			name:        "no lastAgent defaults to search",
			profileJSON: completeProfile,
			message:     "你好",
			want:        AgentSearch,
		},
		// AgentRecruiter
		{
			name:           "interview keyword with application",
			profileJSON:    completeProfile,
			message:        "我想练习面试",
			hasApplication: true,
			want:           AgentRecruiter,
		},
		{
			name:           "no keyword fallback to recruiter",
			profileJSON:    completeProfile,
			message:        "你好",
			lastAgent:      "recruiter",
			hasApplication: true,
			want:           AgentRecruiter,
		},
		// Edge: interview keyword without application → fall through
		{
			name:        "interview keyword no application",
			profileJSON: completeProfile,
			message:     "面试",
			lastAgent:   "search",
			want:        AgentSearch,
		},
		// Edge: has application but no interview keyword → lastAgent
		{
			name:           "application no interview keyword",
			profileJSON:    completeProfile,
			message:        "你好",
			lastAgent:      "search",
			hasApplication: true,
			want:           AgentSearch,
		},
		// Edge: lastAgent resume fallback
		{
			name:        "no keyword fallback to resume",
			profileJSON: completeProfile,
			message:     "你好",
			lastAgent:   "resume",
			want:        AgentResume,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Route(tt.hasResumeAgent, tt.profileJSON, tt.message, tt.lastAgent, tt.hasApplication)
			if got != tt.want {
				t.Errorf("Route() = %d (%s), want %d", got, agentKindName(got), tt.want)
			}
		})
	}
}

func agentKindName(k AgentKind) string {
	switch k {
	case AgentResume:
		return "resume"
	case AgentSearch:
		return "search"
	case AgentRecruiter:
		return "recruiter"
	default:
		return "unknown"
	}
}
