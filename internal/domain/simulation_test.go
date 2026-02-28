package domain

import "testing"

func TestSimulationConfigValidate(t *testing.T) {
	valid := SimulationConfig{
		Model:        ModelGBM,
		NumPaths:     1000,
		HorizonDays:  252,
		LookbackDays: 756,
		StartValue:   100_000,
	}

	tests := []struct {
		name    string
		mutate  func(c SimulationConfig) SimulationConfig
		wantErr bool
	}{
		{"valid config", func(c SimulationConfig) SimulationConfig { return c }, false},
		{"zero num_paths", func(c SimulationConfig) SimulationConfig { c.NumPaths = 0; return c }, true},
		{"negative num_paths", func(c SimulationConfig) SimulationConfig { c.NumPaths = -1; return c }, true},
		{"zero horizon_days", func(c SimulationConfig) SimulationConfig { c.HorizonDays = 0; return c }, true},
		{"negative horizon_days", func(c SimulationConfig) SimulationConfig { c.HorizonDays = -5; return c }, true},
		{"zero lookback_days", func(c SimulationConfig) SimulationConfig { c.LookbackDays = 0; return c }, true},
		{"zero start_value", func(c SimulationConfig) SimulationConfig { c.StartValue = 0; return c }, true},
		{"negative start_value", func(c SimulationConfig) SimulationConfig { c.StartValue = -1; return c }, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.mutate(valid)
			msg := cfg.Validate()
			hasErr := msg != ""
			if hasErr != tc.wantErr {
				t.Errorf("Validate() = %q, wantErr = %v", msg, tc.wantErr)
			}
		})
	}
}
