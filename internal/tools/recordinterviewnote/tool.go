package recordinterviewnote

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/lora-sys/JonLinker/internal/session"
)

type InterviewNote struct {
	JobURL            string   `json:"job_url"`
	QuestionsAsked    []string `json:"questions_asked"`
	CandidateAnswers  string   `json:"candidate_answers"`
	OverallAssessment string   `json:"overall_assessment"`
	CreatedAt         string   `json:"created_at"`
}

type recordInterviewNoteArgs struct {
	JobURL            string   `json:"job_url"`
	QuestionsAsked    []string `json:"questions_asked"`
	CandidateAnswers  string   `json:"candidate_answers"`
	OverallAssessment string   `json:"overall_assessment"`
}

type Tool struct {
	cpStore interface {
		Set(ctx context.Context, key string, data []byte) error
	}
}

func NewTool(store interface{ Set(ctx context.Context, key string, data []byte) error }) *Tool {
	return &Tool{cpStore: store}
}

func (t *Tool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "record_interview_note",
		Desc: "记录面试过程中的问答和评价。在面试完成后调用此工具保存面试记录。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"job_url": {
				Type:     schema.String,
				Desc:     "面试的职位URL",
				Required: true,
			},
			"questions_asked": {
				Type:     schema.Array,
				Desc:     "面试中提问的问题列表",
				Required: true,
			},
			"candidate_answers": {
				Type:     schema.String,
				Desc:     "候选人对问题的回答摘要",
				Required: true,
			},
			"overall_assessment": {
				Type:     schema.String,
				Desc:     "对候选人面试表现的综合评价",
				Required: true,
			},
		}),
	}, nil
}

func (t *Tool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var args recordInterviewNoteArgs
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	var missing []string
	if args.JobURL == "" {
		missing = append(missing, "job_url")
	}
	if len(args.QuestionsAsked) == 0 {
		missing = append(missing, "questions_asked")
	}
	if args.CandidateAnswers == "" {
		missing = append(missing, "candidate_answers")
	}
	if args.OverallAssessment == "" {
		missing = append(missing, "overall_assessment")
	}
	if len(missing) > 0 {
		return "", fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	sessionID := session.SessionIDFromContext(ctx)
	if sessionID == "" {
		return "session not found", nil
	}

	note := InterviewNote{
		JobURL:            args.JobURL,
		QuestionsAsked:    args.QuestionsAsked,
		CandidateAnswers:  args.CandidateAnswers,
		OverallAssessment: args.OverallAssessment,
		CreatedAt:         time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(note)
	if err != nil {
		return "", fmt.Errorf("marshal note: %w", err)
	}

	key := sessionID + ":interview_" + args.JobURL
	if err := t.cpStore.Set(ctx, key, data); err != nil {
		return "", fmt.Errorf("save note: %w", err)
	}

	return "面试记录已保存", nil
}
