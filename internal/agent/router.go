package agent

import (
	"encoding/json"
	"strings"

	"github.com/lora-sys/JonLinker/internal/resume"
)

type AgentKind int

const (
	AgentResume AgentKind = iota
	AgentSearch
)

func Route(hasResumeAgent bool, profileJSON []byte, message string) AgentKind {
	if len(profileJSON) > 0 {
		var profile resume.CandidateProfile
		if err := json.Unmarshal(profileJSON, &profile); err != nil {
			return AgentResume
		}
		if profile.Name == "" || profile.Title == "" || len(profile.Skills) == 0 {
			return AgentResume
		}
	} else {
		return AgentResume
	}

	lower := strings.ToLower(message)
	resumeKeywords := []string{"修改", "完善", "更新", "补充", "改", "update", "edit", "change", "modify"}
	for _, kw := range resumeKeywords {
		if strings.Contains(lower, kw) {
			profileRelated := []string{"简历", "资料", "技能", "profile", "resume", "skill"}
			for _, pr := range profileRelated {
				if strings.Contains(lower, pr) {
					return AgentResume
				}
			}
		}
	}

	return AgentSearch
}
