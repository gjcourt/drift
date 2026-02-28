package domain

import (
	"math"
	"testing"
)

func makeConstantPaths(n int, start, terminal float64) []SimulatedPath {
	paths := make([]SimulatedPath, n)
	for i := range paths {
		paths[i] = SimulatedPath{Values: []float64{start, terminal}}
	}
	return paths
}

func TestComputeStatsEmpty(t *testing.T) {
	stats := ComputeStats(nil, 100_000, 1)
	if stats != (ResultStats{}) {
		t.Errorf("ComputeStats(nil) should return zero struct, got %+v", stats)
	}
}

func TestComputeStatsConstantGain(t *testing.T) {
	paths := makeConstantPaths(1000, 100_000, 110_000)
	stats := ComputeStats(paths, 100_000, 1)
	if stats.ProbabilityOfLoss != 0 {
		t.Errorf("ProbabilityOfLoss = %v, want 0", stats.ProbabilityOfLoss)
	}
	for _, p := range []float64{stats.P5, stats.P25, stats.P50, stats.P75, stats.P95} {
		if math.Abs(p-110_000) > 1 {
			t.Errorf("percentile = %v, want ~110000", p)
		}
	}
}

func TestComputeStatsConstantLoss(t *testing.T) {
	paths := makeConstantPaths(1000, 100_000, 90_000)
	stats := ComputeStats(paths, 100_000, 1)
	if math.Abs(stats.ProbabilityOfLoss-1.0) > 1e-9 {
		t.Errorf("ProbabilityOfLoss = %v, want 1.0", stats.ProbabilityOfLoss)
	}
}

func TestComputeStatsMean(t *testing.T) {
	paths := []SimulatedPath{
		{Values: []float64{100_000, 200_000}},
		{Values: []float64{100_000, 50_000}},
	}
	stats := ComputeStats(paths, 100_000, 1)
	if math.Abs(stats.Mean-125_000) > 1 {
		t.Errorf("Mean = %v, want 125000", stats.Mean)
	}
}

func TestComputeStatsCAGR(t *testing.T) {
	paths := makeConstantPaths(1000, 100_000, 200_000)
	stats := ComputeStats(paths, 100_000, 5)
	expectedCAGR := math.Pow(2, 1.0/5) - 1
	if math.Abs(stats.MedianCAGR-expectedCAGR) > 1e-6 {
		t.Errorf("MedianCAGR = %v, want ~%v", stats.MedianCAGR, expectedCAGR)
	}
}
