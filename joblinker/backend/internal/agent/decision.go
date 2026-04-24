package agent

import (
	"fmt"
	"joblinker/pkg/ai"
	"math/rand"
	"strings"
	"time"
)

type DecisionMaker struct {
	confidence float64
	aiClient   *ai.Client
}

func NewDecisionMaker() *DecisionMaker {
	return &DecisionMaker{confidence: 0.8}
}

func NewDecisionMakerWithAI(client *ai.Client) *DecisionMaker {
	return &DecisionMaker{confidence: 0.85, aiClient: client}
}

type Decision struct {
	Action     string
	Reason     string
	Confidence float64
	AIUsed     bool
}

const (
	ActionExpressInterest   = "EXPRESS_INTEREST"
	ActionDecline          = "DECLINE"
	ActionNegotiate        = "NEGOTIATE"
	ActionAccept           = "ACCEPT"
	ActionScheduleInterview = "SCHEDULE_INTERVIEW"
	ActionWithdraw         = "WITHDRAW"
	ActionWait             = "WAIT"
)

func (dm *DecisionMaker) EvaluateIntroduction(msg *Message) Decision {
	return Decision{
		Action:    ActionExpressInterest,
		Reason:    "Positive introduction response based on role match",
		Confidence: dm.confidence,
	}
}

func (dm *DecisionMaker) EvaluateInterest(msg *Message) Decision {
	level := ""
	if msg.Payload.Parameters != nil {
		level = msg.Payload.Parameters.InterestLevel
	}
	if level == "high" {
		return Decision{
			Action:    ActionExpressInterest,
			Reason:    "High interest level warrants reciprocal interest",
			Confidence: dm.confidence + 0.1,
		}
	}
	if level == "medium" {
		return Decision{
			Action:    ActionExpressInterest,
			Reason:    "Medium interest - will express cautious interest",
			Confidence: dm.confidence,
		}
	}
	return Decision{
		Action:    ActionDecline,
		Reason:    "Low interest level - not a good match",
		Confidence: dm.confidence,
	}
}

func (dm *DecisionMaker) EvaluateOffer(msg *Message) Decision {
	neg := msg.Payload.Negotiation
	if neg == nil || neg.Compensation == nil {
		return Decision{
			Action:    ActionNegotiate,
			Reason:    "No compensation details - requesting clarification",
			Confidence: 0.5,
		}
	}

	comp := neg.Compensation
	if comp.BaseSalary < 50000 {
		return Decision{
			Action:    ActionNegotiate,
			Reason:    "Salary below acceptable threshold",
			Confidence: dm.confidence,
		}
	}

	if comp.BaseSalary > 100000 {
		return Decision{
			Action:    ActionAccept,
			Reason:    "Excellent compensation package",
			Confidence: dm.confidence + 0.15,
		}
	}

	return Decision{
		Action:    ActionNegotiate,
		Reason:    "Compensation acceptable but room for negotiation",
		Confidence: dm.confidence,
	}
}

func (dm *DecisionMaker) EvaluateNegotiation(msg *Message) Decision {
	if dm.aiClient != nil {
		context := fmt.Sprintf("Negotiation for match ID: %s, Intent: %s", msg.Payload.Parameters.MatchID, msg.Payload.Intent)
		if msg.Payload.Negotiation != nil && msg.Payload.Negotiation.Compensation != nil {
			context += fmt.Sprintf(", Salary: %d", msg.Payload.Negotiation.Compensation.BaseSalary)
		}
		agentType := "recruiter"
		if msg.Payload.Parameters.Role == "employer" {
			agentType = "job_seeker"
		}
		response, err := dm.aiClient.GenerateAgentResponse(context, agentType)
		if err == nil && response != "" {
			return Decision{
				Action:     dm.parseAIResponseAction(response),
				Reason:     response,
				Confidence: 0.85,
				AIUsed:     true,
			}
		}
	}

	r := rand.Float64()
	if r < 0.3 {
		return Decision{
			Action:    ActionAccept,
			Reason:    "Accepting current terms",
			Confidence: 0.7,
		}
	}
	if r < 0.6 {
		return Decision{
			Action:    ActionNegotiate,
			Reason:    "Counter-offering with improved terms",
			Confidence: 0.6,
		}
	}
	return Decision{
		Action:    ActionWait,
		Reason:    "Waiting for updated offer",
		Confidence: 0.5,
	}
}

func (dm *DecisionMaker) parseAIResponseAction(response string) string {
	lower := strings.ToLower(response)
	if strings.Contains(lower, "accept") || strings.Contains(lower, "agree") {
		return ActionAccept
	}
	if strings.Contains(lower, "counter") || strings.Contains(lower, "negotiate") {
		return ActionNegotiate
	}
	if strings.Contains(lower, "decline") || strings.Contains(lower, "reject") {
		return ActionDecline
	}
	if strings.Contains(lower, "wait") || strings.Contains(lower, "delay") {
		return ActionWait
	}
	return ActionNegotiate
}

func (dm *DecisionMaker) MakeDecision(msg *Message) Decision {
	switch msg.Payload.Intent {
	case IntentIntroduction:
		return dm.EvaluateIntroduction(msg)
	case IntentInterest:
		return dm.EvaluateInterest(msg)
	case IntentOffer:
		return dm.EvaluateOffer(msg)
	case IntentNegotiation:
		return dm.EvaluateNegotiation(msg)
	case IntentSchedule:
		return Decision{
			Action:    ActionScheduleInterview,
			Reason:    "Scheduling interview as requested",
			Confidence: 0.9,
		}
	default:
		return Decision{
			Action:    ActionWait,
			Reason:    "Unknown intent - awaiting clarification",
			Confidence: 0.3,
		}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (dm *DecisionMaker) EvaluateMatchWithAI(seekerProfile, jobDescription string) (float64, string, error) {
	if dm.aiClient == nil {
		return 0.5, "AI client not available", nil
	}
	return dm.aiClient.EvaluateMatch(seekerProfile, jobDescription)
}
