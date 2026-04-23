package agent

import (
	"encoding/json"
	"time"
)

type OfferGenerator struct {
	baseSalary    int
	bonusPercent  int
	equityPercent float64
	benefits      []string
}

func NewOfferGenerator() *OfferGenerator {
	return &OfferGenerator{
		baseSalary:    100000,
		bonusPercent:  10,
		equityPercent: 0.01,
		benefits: []string{
			"Health Insurance",
			"401k Match",
			"Unlimited PTO",
			"Remote Work Options",
		},
	}
}

type CompensationPackage struct {
	BaseSalary     int      `json:"base_salary"`
	SigningBonus   int      `json:"signing_bonus"`
	AnnualBonus    int      `json:"annual_bonus"`
	EquityGrant    float64  `json:"equity_grant"`
	EquityType     string   `json:"equity_type"`
	Benefits       []string `json:"benefits"`
	Currency       string   `json:"currency"`
	TotalCompensation int   `json:"total_compensation"`
}

func (g *OfferGenerator) GenerateOffer(baseSalary int, level string) *CompensationPackage {
	adjustedSalary := g.adjustSalaryByLevel(baseSalary, level)
	bonus := int(float64(adjustedSalary) * float64(g.bonusPercent) / 100.0)

	return &CompensationPackage{
		BaseSalary:        adjustedSalary,
		SigningBonus:      int(float64(adjustedSalary) * 0.05),
		AnnualBonus:       bonus,
		EquityGrant:       g.equityPercent * float64(adjustedSalary),
		EquityType:        "ISO",
		Benefits:          g.benefits,
		Currency:          "USD",
		TotalCompensation: adjustedSalary + bonus + int(float64(adjustedSalary)*g.equityPercent*4),
	}
}

func (g *OfferGenerator) adjustSalaryByLevel(baseSalary int, level string) int {
	multipliers := map[string]float64{
		"junior":  0.8,
		"mid":     1.0,
		"senior":  1.3,
		"lead":    1.5,
		"principal": 1.8,
	}

	if mult, ok := multipliers[level]; ok {
		return int(float64(baseSalary) * mult)
	}
	return baseSalary
}

func (g *OfferGenerator) GenerateOfferJSON(baseSalary int, level string) (string, error) {
	pkg := g.GenerateOffer(baseSalary, level)
	data, err := json.Marshal(pkg)
	if err != nil {
		return "{}", err
	}
	return string(data), nil
}

type OfferNegotiation struct {
	CurrentOffer    *CompensationPackage
	TargetSalary    int
	InitialOffer    int
	CounterOffers   int
	MaxCounters     int
	Deadline        time.Time
}

func (g *OfferGenerator) CreateNegotiation(initial *CompensationPackage, target int) *OfferNegotiation {
	return &OfferNegotiation{
		CurrentOffer:  initial,
		TargetSalary:  target,
		InitialOffer:  initial.BaseSalary,
		CounterOffers: 0,
		MaxCounters:   3,
		Deadline:      time.Now().Add(7 * 24 * time.Hour),
	}
}

func (n *OfferNegotiation) MakeCounterOffer(newSalary int) bool {
	if n.CounterOffers >= n.MaxCounters {
		return false
	}
	if newSalary < n.TargetSalary {
		return false
	}

	n.CurrentOffer.BaseSalary = newSalary
	n.CounterOffers++
	return true
}

func (n *OfferNegotiation) IsExpired() bool {
	return time.Now().After(n.Deadline)
}
