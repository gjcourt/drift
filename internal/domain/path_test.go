package domain

import (
	"math"
	"testing"
)

func TestFinal(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"empty path returns zero", nil, 0},
		{"single value", []float64{100}, 100},
		{"returns last element", []float64{100, 110, 95, 120}, 120},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := SimulatedPath{Values: tc.values}
			if got := p.Final(); got != tc.want {
				t.Errorf("Final() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMaxDrawdown(t *testing.T) {
	tests := []struct {
		name      string
		values    []float64
		wantClose float64
	}{
		{"empty path returns zero", nil, 0},
		{"single value returns zero", []float64{100}, 0},
		{"two values no drawdown", []float64{100, 110}, 0},
		{"two values 50pct drawdown", []float64{100, 50}, -0.5},
		{"recovery after drawdown", []float64{100, 50, 200, 100}, -0.5},
		{"new peak resets drawdown", []float64{100, 120, 60, 130, 65}, -0.5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := SimulatedPath{Values: tc.values}
			got := p.MaxDrawdown()
			if math.Abs(got-tc.wantClose) > 1e-9 {
				t.Errorf("MaxDrawdown() = %v, want ~%v", got, tc.wantClose)
			}
			if len(tc.values) >= 2 && got > 0 {
				t.Errorf("MaxDrawdown() = %v > 0; drawdown must be non-positive", got)
			}
		})
	}
}
