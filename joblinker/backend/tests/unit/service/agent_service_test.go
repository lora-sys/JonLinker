package service_test

import (
	"testing"

	"joblinker/internal/agent"
)

func TestOfferGenerator_GenerateOffer(t *testing.T) {
	gen := agent.NewOfferGenerator()

	pkg := gen.GenerateOffer(100000, "senior")

	if pkg.BaseSalary <= 0 {
		t.Error("expected positive base salary")
	}

	if pkg.Currency != "USD" {
		t.Errorf("expected currency USD, got %s", pkg.Currency)
	}

	if len(pkg.Benefits) == 0 {
		t.Error("expected benefits to be populated")
	}
}

func TestOfferGenerator_LevelAdjustment(t *testing.T) {
	gen := agent.NewOfferGenerator()

	tests := []struct {
		level       string
		baseSalary  int
		minExpected int
	}{
		{"junior", 100000, 70000},
		{"mid", 100000, 90000},
		{"senior", 100000, 120000},
		{"lead", 100000, 140000},
	}

	for _, tt := range tests {
		pkg := gen.GenerateOffer(tt.baseSalary, tt.level)
		if pkg.BaseSalary < tt.minExpected {
			t.Errorf("level %s: expected >= %d, got %d",
				tt.level, tt.minExpected, pkg.BaseSalary)
		}
	}
}

func TestOfferNegotiation_CreateNegotiation(t *testing.T) {
	gen := agent.NewOfferGenerator()
	initial := gen.GenerateOffer(100000, "senior")

	neg := gen.CreateNegotiation(initial, 110000)

	if neg == nil {
		t.Fatal("expected non-nil negotiation")
	}

	// Senior level adjusts salary by 1.3x, so initial is 130000
	if neg.InitialOffer != 130000 {
		t.Errorf("expected initial 130000 (adjusted), got %d", neg.InitialOffer)
	}

	if neg.TargetSalary != 110000 {
		t.Errorf("expected target 110000, got %d", neg.TargetSalary)
	}
}

func TestOfferNegotiation_MakeCounterOffer_Success(t *testing.T) {
	gen := agent.NewOfferGenerator()
	initial := gen.GenerateOffer(100000, "senior")

	neg := gen.CreateNegotiation(initial, 110000)

	success := neg.MakeCounterOffer(120000)
	if !success {
		t.Error("expected successful counter offer")
	}

	if neg.CurrentOffer.BaseSalary != 120000 {
		t.Errorf("expected current offer 120000, got %d", neg.CurrentOffer.BaseSalary)
	}
}

func TestOfferNegotiation_MakeCounterOffer_MaxCounters(t *testing.T) {
	gen := agent.NewOfferGenerator()
	initial := gen.GenerateOffer(100000, "senior")

	neg := gen.CreateNegotiation(initial, 110000)
	neg.MaxCounters = 2

	neg.MakeCounterOffer(102000)
	neg.MakeCounterOffer(104000)

	success := neg.MakeCounterOffer(106000)
	if success {
		t.Error("expected failure at max counters")
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
		t.Error("expected not expired immediately after creation")
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

	if jsonStr[0] != '{' {
		t.Error("expected JSON object")
	}
}

func TestCompensationPackage_TotalCompensation(t *testing.T) {
	gen := agent.NewOfferGenerator()

	pkg := gen.GenerateOffer(100000, "senior")

	if pkg.TotalCompensation == 0 {
		t.Error("expected total compensation to be calculated")
	}
}

func TestCompensationPackage_EquityType(t *testing.T) {
	gen := agent.NewOfferGenerator()

	pkg := gen.GenerateOffer(100000, "senior")

	if pkg.EquityType != "ISO" {
		t.Errorf("expected equity type ISO, got %s", pkg.EquityType)
	}
}
