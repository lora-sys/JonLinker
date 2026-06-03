package agent

import (
	"encoding/json"
	"testing"

	"github.com/lora-sys/JonLinker/internal/resume"
)

func TestRouteEmptyProfile(t *testing.T) {
	kind := Route(false, nil, "hello")
	if kind != AgentResume {
		t.Errorf("expected AgentResume, got %v", kind)
	}
}

func TestRouteEmptyProfileJSON(t *testing.T) {
	kind := Route(false, []byte(""), "hello")
	if kind != AgentResume {
		t.Errorf("expected AgentResume, got %v", kind)
	}
}

func TestRouteProfileWithErrors(t *testing.T) {
	kind := Route(false, []byte("{invalid json}"), "hello")
	if kind != AgentResume {
		t.Errorf("expected AgentResume, got %v", kind)
	}
}

func TestRouteIncompleteProfileMissingName(t *testing.T) {
	p := resume.CandidateProfile{Title: "Engineer", Skills: []string{"Go"}}
	b, _ := json.Marshal(p)
	kind := Route(false, b, "hello")
	if kind != AgentResume {
		t.Errorf("expected AgentResume, got %v", kind)
	}
}

func TestRouteIncompleteProfileMissingTitle(t *testing.T) {
	p := resume.CandidateProfile{Name: "Alice", Skills: []string{"Go"}}
	b, _ := json.Marshal(p)
	kind := Route(false, b, "hello")
	if kind != AgentResume {
		t.Errorf("expected AgentResume, got %v", kind)
	}
}

func TestRouteIncompleteProfileMissingSkills(t *testing.T) {
	p := resume.CandidateProfile{Name: "Alice", Title: "Engineer"}
	b, _ := json.Marshal(p)
	kind := Route(false, b, "hello")
	if kind != AgentResume {
		t.Errorf("expected AgentResume, got %v", kind)
	}
}

func TestRouteCompleteProfileWithResumeKeyword(t *testing.T) {
	p := resume.CandidateProfile{Name: "Alice", Title: "Engineer", Skills: []string{"Go"}}
	b, _ := json.Marshal(p)
	kind := Route(false, b, "修改简历")
	if kind != AgentResume {
		t.Errorf("expected AgentResume, got %v", kind)
	}
}

func TestRouteCompleteProfileWithUpdateKeyword(t *testing.T) {
	p := resume.CandidateProfile{Name: "Alice", Title: "Engineer", Skills: []string{"Go"}}
	b, _ := json.Marshal(p)
	kind := Route(false, b, "更新简历")
	if kind != AgentResume {
		t.Errorf("expected AgentResume, got %v", kind)
	}
}

func TestRouteCompleteProfileSearch(t *testing.T) {
	p := resume.CandidateProfile{Name: "Alice", Title: "Engineer", Skills: []string{"Go"}}
	b, _ := json.Marshal(p)
	kind := Route(false, b, "帮我找前端工作")
	if kind != AgentSearch {
		t.Errorf("expected AgentSearch, got %v", kind)
	}
}

func TestRouteEmptyString(t *testing.T) {
	kind := Route(false, nil, "")
	if kind != AgentResume {
		t.Errorf("expected AgentResume, got %v", kind)
	}
}

func TestRouteWhitespaceOnly(t *testing.T) {
	kind := Route(false, nil, "   ")
	if kind != AgentResume {
		t.Errorf("expected AgentResume, got %v", kind)
	}
}
