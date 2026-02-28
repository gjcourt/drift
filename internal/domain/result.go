package domain

import (
	"math"
	"sort"
	"time"
)

// SimulationResult aggregates statistics over all generated paths for one run.
type SimulationResult struct {
	ID           string
	ExperimentID string
	RanAt        time.Time
	NumPaths     int
	Config       SimulationConfig

	Paths []SimulatedPath // retained in memory; persisted separately if requested

	// Pre-computed statistics
	Stats ResultStats
}

// ResultStats contains aggregated percentile statistics over all paths.
type ResultStats struct {
	P5                float64
	P25               float64
	P50               float64
	P75               float64
	P95               float64
	Mean              float64
	StdDev            float64
	ProbabilityOfLoss float64
	MedianMaxDrawdown float64
	P95MaxDrawdown    float64
	MedianCAGR        float64
}

// ComputeStats derives ResultStats from the completed set of simulated paths.
func ComputeStats(paths []SimulatedPath, startValue, horizonYears float64) ResultStats {
	if len(paths) == 0 {
		return ResultStats{}
	}
	finals := make([]float64, len(paths))
	drawdowns := make([]float64, len(paths))
	sum := 0.0
	losses := 0

	for i, p := range paths {
		f := p.Final()
		finals[i] = f
		drawdowns[i] = p.MaxDrawdown()
		sum += f
		if f < startValue {
			losses++
		}
	}

	sort.Float64s(finals)
	sort.Float64s(drawdowns)

	n := len(finals)
	mean := sum / float64(n)

	variance := 0.0
	for _, f := range finals {
		d := f - mean
		variance += d * d
	}
	variance /= float64(n)

	p50 := finals[n/2]
	medianCAGR := 0.0
	if startValue > 0 && horizonYears > 0 && p50 > 0 {
		medianCAGR = math.Pow(p50/startValue, 1.0/horizonYears) - 1
	}

	return ResultStats{
		P5:                finals[int(float64(n)*0.05)],
		P25:               finals[int(float64(n)*0.25)],
		P50:               p50,
		P75:               finals[int(float64(n)*0.75)],
		P95:               finals[int(float64(n)*0.95)],
		Mean:              mean,
		StdDev:            math.Sqrt(variance),
		ProbabilityOfLoss: float64(losses) / float64(n),
		MedianMaxDrawdown: drawdowns[n/2],
		P95MaxDrawdown:    drawdowns[int(float64(n)*0.95)],
		MedianCAGR:        medianCAGR,
	}
}
