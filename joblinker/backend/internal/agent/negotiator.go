package agent

import (
	"fmt"
)

// SalaryNegotiator handles salary and compensation negotiations between agents
type SalaryNegotiator struct {
	minSalary    int
	maxSalary    int
	targetSalary int
	currentOffer int
	strategy     NegotiationStrategy
}

type NegotiationStrategy string

const (
	StrategyAggressive NegotiationStrategy = "aggressive"
	StrategyModerate   NegotiationStrategy = "moderate"
	StrategyConservative NegotiationStrategy = "conservative"
)

func NewSalaryNegotiator(min, max, target int) *SalaryNegotiator {
	return &SalaryNegotiator{
		minSalary:    min,
		maxSalary:    max,
		targetSalary: target,
		currentOffer: max,
		strategy:     StrategyModerate,
	}
}

type NegotiationResult struct {
	AgreedSalary  int
	FinalOffer    int
	CounterOffers int
	Conceded      bool
	Reached       bool
}

// GenerateCounterOffer creates a counter offer based on strategy and negotiation state
func (n *SalaryNegotiator) GenerateCounterOffer(currentOffer int, rounds int) int {
	if currentOffer < n.minSalary {
		return -1 // Walk away
	}

	// Reduce gap each round to simulate convergence
	gap := n.maxSalary - currentOffer
	concessionRate := 0.15 + (float64(rounds) * 0.05) // 15-30% concession rate
	if concessionRate > 0.35 {
		concessionRate = 0.35
	}

	counterOffer := currentOffer + int(float64(gap)*concessionRate)
	if counterOffer < n.targetSalary {
		counterOffer = n.targetSalary
	}

	return counterOffer
}

// EvaluateOffer checks if an offer meets minimum requirements
func (n *SalaryNegotiator) EvaluateOffer(offer int) OfferDecision {
	if offer < n.minSalary {
		return OfferDecisionReject
	}
	if offer >= n.targetSalary {
		return OfferDecisionAccept
	}
	return OfferDecisionCounter
}

type OfferDecision int

const (
	OfferDecisionAccept OfferDecision = iota
	OfferDecisionCounter
	OfferDecisionReject
)

// SimulateNegotiation runs a complete salary negotiation
func (n *SalaryNegotiator) SimulateNegotiation(initialOffer int) *NegotiationResult {
	result := &NegotiationResult{
		FinalOffer: initialOffer,
	}

	rounds := 0
	currentOffer := initialOffer

	for rounds < 5 {
		decision := n.EvaluateOffer(currentOffer)
		switch decision {
		case OfferDecisionAccept:
			result.Reached = true
			result.AgreedSalary = currentOffer
			return result
		case OfferDecisionReject:
			result.Reached = false
			return result
		case OfferDecisionCounter:
			rounds++
			result.CounterOffers++
			currentOffer = n.GenerateCounterOffer(currentOffer, rounds)
			if currentOffer < 0 {
				result.Reached = false
				return result
			}
			result.FinalOffer = currentOffer
		}
	}

	// Max rounds reached, accept last offer or walk
	if currentOffer >= n.minSalary {
		result.Reached = true
		result.Conceded = true
		result.AgreedSalary = currentOffer
	} else {
		result.Reached = false
	}
	return result
}

// CalculateAcceptanceProbability based on offer vs target
func (n *SalaryNegotiator) CalculateAcceptanceProbability(offer int) float64 {
	if offer < n.minSalary {
		return 0.0
	}
	if offer >= n.targetSalary {
		return 1.0
	}

	rangeSize := n.targetSalary - n.minSalary
	position := offer - n.minSalary
	return float64(position) / float64(rangeSize)
}

// AddBonus adds signing bonus to compensation package
func (n *SalaryNegotiator) AddBonus(baseSalary int, bonusPercent float64) int {
	return baseSalary + int(float64(baseSalary)*bonusPercent)
}

// FormatCompensation creates a formatted compensation string
func FormatCompensation(salary int, bonus int, equity string) string {
	return fmt.Sprintf("Base: $%d, Bonus: $%d, Equity: %s", salary, bonus, equity)
}