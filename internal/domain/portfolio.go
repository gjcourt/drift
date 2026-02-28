package domain

// Portfolio aggregates a set of weighted assets that are simulated together.
type Portfolio struct {
	Assets    []PortfolioAsset
	Rebalance RebalanceFrequency
}

// PortfolioAsset is a symbol + fractional weight (must sum to 1.0 across portfolio).
type PortfolioAsset struct {
	Symbol string
	Weight float64
}

// RebalanceFrequency controls how often the portfolio is rebalanced during simulation.
type RebalanceFrequency string

// Rebalance frequency options.
const (
	RebalanceNone      RebalanceFrequency = "none"
	RebalanceMonthly   RebalanceFrequency = "monthly"
	RebalanceQuarterly RebalanceFrequency = "quarterly"
	RebalanceAnnual    RebalanceFrequency = "annual"
)

// TotalWeight returns the sum of all asset weights (should equal 1.0 for a valid portfolio).
func (p Portfolio) TotalWeight() float64 {
	total := 0.0
	for _, a := range p.Assets {
		total += a.Weight
	}
	return total
}
