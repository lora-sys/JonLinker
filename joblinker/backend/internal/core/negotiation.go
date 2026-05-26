package core

import "fmt"

// SalaryNegotiator handles salary and compensation negotiations between agents.
// Not safe for concurrent use; wrap in a mutex if shared across goroutines.
type SalaryNegotiator struct {
	minSalary    int
	maxSalary    int
	targetSalary int
	currentOffer int
	strategy     NegotiationStrategy
}

// NegotiationStrategy defines the concession rate during negotiation.
type NegotiationStrategy string

const (
	StrategyAggressive   NegotiationStrategy = "aggressive"
	StrategyModerate     NegotiationStrategy = "moderate"
	StrategyConservative NegotiationStrategy = "conservative"
)

// OfferDecision represents the result of evaluating an offer.
type OfferDecision int

const (
	OfferDecisionAccept  OfferDecision = iota
	OfferDecisionCounter
	OfferDecisionReject
)

// NegotiationResult holds the outcome of a simulated negotiation.
type NegotiationResult struct {
	AgreedSalary  int
	FinalOffer    int
	CounterOffers int
	Conceded      bool
	Reached       bool
}

// NewSalaryNegotiator creates a new SalaryNegotiator.
// If minSalary == targetSalary, EvaluateOffer immediately accepts at >= min.
func NewSalaryNegotiator(min, max, target int) *SalaryNegotiator {
	if target < min {
		target = min
	}
	if max < min {
		max = min
	}
	return &SalaryNegotiator{
		minSalary:    min,
		maxSalary:    max,
		targetSalary: target,
		currentOffer: max,
		strategy:     StrategyModerate,
	}
}

// GenerateCounterOffer creates a counter offer based on strategy and negotiation state.
// Returns -1 if the offer is below the minimum (walk away).
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

// EvaluateOffer checks if an offer meets minimum requirements.
// Returns Accept if offer >= targetSalary, Reject if offer < minSalary,
// and Counter for offers in between.
func (n *SalaryNegotiator) EvaluateOffer(offer int) OfferDecision {
	if offer < n.minSalary {
		return OfferDecisionReject
	}
	if offer >= n.targetSalary {
		return OfferDecisionAccept
	}
	return OfferDecisionCounter
}

// SimulateNegotiation runs a complete salary negotiation up to 5 rounds.
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

// CalculateAcceptanceProbability returns the probability of accepting an offer (0.0 to 1.0).
// Returns 1.0 when minSalary == targetSalary (no range to negotiate).
func (n *SalaryNegotiator) CalculateAcceptanceProbability(offer int) float64 {
	if offer < n.minSalary {
		return 0.0
	}
	if offer >= n.targetSalary {
		return 1.0
	}

	rangeSize := n.targetSalary - n.minSalary
	if rangeSize <= 0 {
		return 1.0 // min == target, already meeting expectation
	}

	position := offer - n.minSalary
	return float64(position) / float64(rangeSize)
}

// AddBonus adds a signing bonus to a base salary.
func (n *SalaryNegotiator) AddBonus(baseSalary int, bonusPercent float64) int {
	return baseSalary + int(float64(baseSalary)*bonusPercent)
}

// FormatCompensation creates a formatted compensation string.
func FormatCompensation(salary int, bonus int, equity string) string {
	return fmt.Sprintf("Base: $%d, Bonus: $%d, Equity: %s", salary, bonus, equity)
}
