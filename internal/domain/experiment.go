package domain

import "time"

// ExperimentStatus tracks the lifecycle of an experiment run.
type ExperimentStatus string

// Lifecycle status values for ExperimentStatus.
const (
	StatusDraft    ExperimentStatus = "draft"
	StatusRunning  ExperimentStatus = "running"
	StatusComplete ExperimentStatus = "complete"
	StatusFailed   ExperimentStatus = "failed"
)

// Experiment is a named simulation configuration that can be run one or many times.
type Experiment struct {
	ID          string
	Name        string
	Description string
	Portfolio   Portfolio
	Config      SimulationConfig
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Run is a single execution of an Experiment, capturing the result snapshot.
type Run struct {
	ID           string
	ExperimentID string
	StartedAt    time.Time
	FinishedAt   *time.Time
	Status       ExperimentStatus
	Error        string
	Stats        ResultStats
}
