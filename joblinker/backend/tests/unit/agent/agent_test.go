package agent_test

import (
	"testing"
	"time"

	"joblinker/internal/agent"
)

func TestSalaryNegotiator_NewNegotiator(t *testing.T) {
	neg := agent.NewSalaryNegotiator(80000, 120000, 100000)

	if neg == nil {
		t.Fatal("expected non-nil negotiator")
	}
}

func TestSalaryNegotiator_EvaluateOffer(t *testing.T) {
	neg := agent.NewSalaryNegotiator(80000, 120000, 100000)

	tests := []struct {
		offer    int
		expected agent.OfferDecision
	}{
		{85000, agent.OfferDecisionCounter},
		{95000, agent.OfferDecisionCounter},
		{100000, agent.OfferDecisionAccept},
		{110000, agent.OfferDecisionAccept},
		{120000, agent.OfferDecisionAccept},
		{75000, agent.OfferDecisionReject},
		{50000, agent.OfferDecisionReject},
	}

	for _, tt := range tests {
		result := neg.EvaluateOffer(tt.offer)
		if result != tt.expected {
			t.Errorf("EvaluateOffer(%d): expected %v, got %v", tt.offer, tt.expected, result)
		}
	}
}

func TestSalaryNegotiator_SimulateNegotiation_Accept(t *testing.T) {
	neg := agent.NewSalaryNegotiator(80000, 120000, 100000)

	result := neg.SimulateNegotiation(95000)

	if !result.Reached {
		t.Error("expected negotiation to reach agreement")
	}
}

func TestSalaryNegotiator_SimulateNegotiation_Reject(t *testing.T) {
	neg := agent.NewSalaryNegotiator(80000, 120000, 100000)

	result := neg.SimulateNegotiation(50000)

	if result.Reached {
		t.Error("expected negotiation to fail")
	}
}

func TestSalaryNegotiator_SimulateNegotiation_CounterOffers(t *testing.T) {
	neg := agent.NewSalaryNegotiator(80000, 120000, 100000)

	result := neg.SimulateNegotiation(85000)

	if result.CounterOffers == 0 {
		t.Error("expected at least one counter offer")
	}
}

func TestSalaryNegotiator_CalculateAcceptanceProbability(t *testing.T) {
	neg := agent.NewSalaryNegotiator(80000, 120000, 100000)

	prob := neg.CalculateAcceptanceProbability(90000)
	if prob < 0 || prob > 1 {
		t.Errorf("expected probability between 0 and 1, got %f", prob)
	}

	prob = neg.CalculateAcceptanceProbability(50000)
	if prob != 0 {
		t.Errorf("expected 0 for below minimum, got %f", prob)
	}

	prob = neg.CalculateAcceptanceProbability(120000)
	if prob != 1 {
		t.Errorf("expected 1 for above target, got %f", prob)
	}
}

func TestSalaryNegotiator_AddBonus(t *testing.T) {
	neg := agent.NewSalaryNegotiator(80000, 120000, 100000)

	result := neg.AddBonus(100000, 0.10)

	if result != 110000 {
		t.Errorf("expected 110000, got %d", result)
	}
}

func TestSalaryNegotiator_GenerateCounterOffer(t *testing.T) {
	neg := agent.NewSalaryNegotiator(80000, 120000, 100000)

	counter := neg.GenerateCounterOffer(90000, 1)

	if counter < 90000 {
		t.Errorf("expected counter >= 90000, got %d", counter)
	}
}

func TestSalaryNegotiator_GenerateCounterOffer_WalkAway(t *testing.T) {
	neg := agent.NewSalaryNegotiator(80000, 120000, 100000)

	counter := neg.GenerateCounterOffer(50000, 1)

	if counter != -1 {
		t.Errorf("expected -1 (walk away), got %d", counter)
	}
}

func TestScheduler_New(t *testing.T) {
	scheduler := agent.NewScheduler()

	if scheduler == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestScheduler_GenerateTimeSlots(t *testing.T) {
	scheduler := agent.NewScheduler()

	slots := scheduler.GenerateTimeSlots(time.Now(), 5, 60)

	if len(slots) == 0 {
		t.Error("expected time slots to be generated")
	}
}

func TestOfferGenerator_New(t *testing.T) {
	gen := agent.NewOfferGenerator()

	if gen == nil {
		t.Fatal("expected non-nil generator")
	}
}

func TestOfferGenerator_GenerateOffer(t *testing.T) {
	gen := agent.NewOfferGenerator()

	pkg := gen.GenerateOffer(100000, "senior")

	if pkg.BaseSalary <= 0 {
		t.Error("expected positive base salary")
	}

	if pkg.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", pkg.Currency)
	}
}

func TestOfferGenerator_GenerateOfferJSON(t *testing.T) {
	gen := agent.NewOfferGenerator()

	jsonStr, err := gen.GenerateOfferJSON(100000, "senior")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if jsonStr == "" {
		t.Error("expected non-empty JSON string")
	}
}

func TestOfferNegotiation_CreateNegotiation(t *testing.T) {
	gen := agent.NewOfferGenerator()
	initial := gen.GenerateOffer(100000, "senior")

	neg := gen.CreateNegotiation(initial, 110000)

	if neg == nil {
		t.Fatal("expected non-nil negotiation")
	}

	// Senior level adjusts by 1.3x
	if neg.InitialOffer != 130000 {
		t.Errorf("expected initial 130000 (senior adjustment), got %d", neg.InitialOffer)
	}
}

func TestOfferNegotiation_MakeCounterOffer(t *testing.T) {
	gen := agent.NewOfferGenerator()
	initial := gen.GenerateOffer(100000, "senior")

	neg := gen.CreateNegotiation(initial, 110000)

	// Senior: 130000, target: 110000
	// Counter 120000 is >= target, should succeed
	success := neg.MakeCounterOffer(120000)
	if !success {
		t.Error("expected successful counter offer")
	}
}

func TestOfferNegotiation_MakeCounterOffer_BelowTarget(t *testing.T) {
	gen := agent.NewOfferGenerator()
	initial := gen.GenerateOffer(100000, "senior")

	neg := gen.CreateNegotiation(initial, 110000)

	success := neg.MakeCounterOffer(90000)
	if success {
		t.Error("expected failure for counter below target")
	}
}

func TestOfferNegotiation_IsExpired(t *testing.T) {
	gen := agent.NewOfferGenerator()
	initial := gen.GenerateOffer(100000, "senior")

	neg := gen.CreateNegotiation(initial, 110000)

	if neg.IsExpired() {
		t.Error("expected negotiation to not be expired immediately")
	}
}

func TestFormatCompensation(t *testing.T) {
	result := agent.FormatCompensation(100000, 10000, "0.1%")

	if result == "" {
		t.Error("expected non-empty string")
	}
}

