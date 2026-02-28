package domain

// SimulationModel identifies the stochastic model used to generate paths.
type SimulationModel string

// Supported simulation model identifiers.
const (
	ModelGBM            SimulationModel = "gbm"
	ModelBootstrap      SimulationModel = "bootstrap"
	ModelBlockBootstrap SimulationModel = "block_bootstrap"
)

// SimulationConfig holds all parameters that define a single simulation run.
type SimulationConfig struct {
	Model        SimulationModel
	NumPaths     int
	HorizonDays  int
	LookbackDays int
	StartValue   float64
	Seed         *int64 // nil means non-deterministic

	// Optional cash-flow parameters
	AnnualContribution float64
	WithdrawalRate     float64
}

// Validate returns an error string if the config is invalid, or empty string if valid.
func (c SimulationConfig) Validate() string {
	if c.NumPaths <= 0 {
		return "num_paths must be positive"
	}
	if c.HorizonDays <= 0 {
		return "horizon_days must be positive"
	}
	if c.LookbackDays <= 0 {
		return "lookback_days must be positive"
	}
	if c.StartValue <= 0 {
		return "start_value must be positive"
	}
	return ""
}
