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
	AgentRecruiter
)

func Route(hasResumeAgent bool, profileJSON []byte, message string, lastAgent string, hasApplication bool) AgentKind {
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

	interviewKeywords := []string{"面试", "练习", "recruiter", "interview", "practice", "聊"}
	for _, kw := range interviewKeywords {
		if strings.Contains(lower, kw) {
			if hasApplication {
				return AgentRecruiter
			}
		}
	}

	searchKeywords := []string{"搜索", "查找", "找", "search", "find", "job", "职位", "工作", "机会"}
	for _, kw := range searchKeywords {
		if strings.Contains(lower, kw) {
			return AgentSearch
		}
	}

	switch lastAgent {
	case "resume":
		return AgentResume
	case "recruiter":
		return AgentRecruiter
	default:
		return AgentSearch
	}
}
