package ingestion

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gjcourt/drift/internal/domain"
)

// ExperimentConfig is the JSON schema for staging an experiment programmatically.
type ExperimentConfig struct {
	Version    string       `json:"version"`
	Experiment ExpMeta      `json:"experiment"`
	Portfolio  PortfolioCfg `json:"portfolio"`
	Simulation SimCfg       `json:"simulation"`
	Parameters ParamCfg     `json:"parameters"`
}

// ExpMeta contains the name and description fields of an experiment JSON config.
type ExpMeta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PortfolioCfg describes portfolio assets and rebalance frequency in JSON config.
type PortfolioCfg struct {
	Assets    []AssetCfg `json:"assets"`
	Rebalance string     `json:"rebalance"`
}

// AssetCfg is a symbol–weight pair in a portfolio JSON config.
type AssetCfg struct {
	Symbol string  `json:"symbol"`
	Weight float64 `json:"weight"`
}

// SimCfg holds simulation parameters in a JSON experiment config.
type SimCfg struct {
	Model        string  `json:"model"`
	NumPaths     int     `json:"num_paths"`
	HorizonDays  int     `json:"horizon_days"`
	LookbackDays int     `json:"lookback_days"`
	StartValue   float64 `json:"start_value"`
	Seed         *int64  `json:"seed"`
}

// ParamCfg holds optional cash-flow parameters in a JSON experiment config.
type ParamCfg struct {
	AnnualContribution float64  `json:"annual_contribution"`
	WithdrawalRate     *float64 `json:"withdrawal_rate"`
}

// ParseExperimentJSON parses a JSON experiment config into domain objects.
func ParseExperimentJSON(r io.Reader) (domain.Experiment, error) {
	var cfg ExperimentConfig
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return domain.Experiment{}, fmt.Errorf("decode experiment JSON: %w", err)
	}

	assets := make([]domain.PortfolioAsset, len(cfg.Portfolio.Assets))
	for i, a := range cfg.Portfolio.Assets {
		assets[i] = domain.PortfolioAsset{Symbol: a.Symbol, Weight: a.Weight}
	}

	rebalance := domain.RebalanceFrequency(cfg.Portfolio.Rebalance)
	if rebalance == "" {
		rebalance = domain.RebalanceNone
	}

	model := domain.SimulationModel(cfg.Simulation.Model)
	if model == "" {
		model = domain.ModelGBM
	}

	withdrawalRate := 0.0
	if cfg.Parameters.WithdrawalRate != nil {
		withdrawalRate = *cfg.Parameters.WithdrawalRate
	}

	return domain.Experiment{
		Name:        cfg.Experiment.Name,
		Description: cfg.Experiment.Description,
		Portfolio: domain.Portfolio{
			Assets:    assets,
			Rebalance: rebalance,
		},
		Config: domain.SimulationConfig{
			Model:              model,
			NumPaths:           cfg.Simulation.NumPaths,
			HorizonDays:        cfg.Simulation.HorizonDays,
			LookbackDays:       cfg.Simulation.LookbackDays,
			StartValue:         cfg.Simulation.StartValue,
			Seed:               cfg.Simulation.Seed,
			AnnualContribution: cfg.Parameters.AnnualContribution,
			WithdrawalRate:     withdrawalRate,
		},
	}, nil
}
