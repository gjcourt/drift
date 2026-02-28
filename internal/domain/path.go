package domain

// SimulatedPath represents one Monte Carlo scenario — daily portfolio values
// from day 0 (= StartValue) through day HorizonDays.
type SimulatedPath struct {
	Values []float64 // length = HorizonDays + 1
}

// Final returns the terminal portfolio value.
func (p SimulatedPath) Final() float64 {
	if len(p.Values) == 0 {
		return 0
	}
	return p.Values[len(p.Values)-1]
}

// MaxDrawdown returns the worst peak-to-trough drawdown across the path (negative fraction).
func (p SimulatedPath) MaxDrawdown() float64 {
	if len(p.Values) < 2 {
		return 0
	}
	peak := p.Values[0]
	maxDD := 0.0
	for _, v := range p.Values[1:] {
		if v > peak {
			peak = v
		}
		if peak > 0 {
			dd := (v - peak) / peak
			if dd < maxDD {
				maxDD = dd
			}
		}
	}
	return maxDD
}
