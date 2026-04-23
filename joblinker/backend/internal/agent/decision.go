package agent

import (
	"math/rand"
	"time"
)

type DecisionMaker struct {
	confidence float64
}

func NewDecisionMaker() *DecisionMaker {
	return &DecisionMaker{confidence: 0.8}
}

type Decision struct {
	Action    string
	Reason    string
	Confidence float64
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
