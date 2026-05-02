package testdoubles

// ServerDeps aggregates all outbound-port fakes for unit tests.
// Add one field per outbound port as migration progresses.
// Current outbound ports are in internal/ports/outbound/:
//   - outbound.AssetRepository
//   - outbound.ExperimentRepository
//   - outbound.SimulationRepository
type ServerDeps struct{}

// NewServerDeps returns a ServerDeps with all fakes initialised to safe zero-value defaults.
func NewServerDeps() *ServerDeps {
	return &ServerDeps{}
}
