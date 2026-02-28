package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/gjcourt/drift/internal/domain"
)

// Store implements AssetRepository, SimulationRepository, and ExperimentRepository
// using a single SQLite database.
type Store struct {
	db *sql.DB
}

// New opens (or creates) a SQLite database at the given path and applies the schema.
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite is single-writer
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(schema)
	return err
}

const schema = `
PRAGMA journal_mode=WAL;

CREATE TABLE IF NOT EXISTS assets (
	id      TEXT PRIMARY KEY,
	symbol  TEXT NOT NULL UNIQUE,
	name    TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS price_records (
	symbol         TEXT NOT NULL,
	date           TEXT NOT NULL,
	open           REAL,
	high           REAL,
	low            REAL,
	close          REAL,
	volume         INTEGER,
	adjusted_close REAL NOT NULL,
	PRIMARY KEY (symbol, date)
);

CREATE TABLE IF NOT EXISTS experiments (
	id          TEXT PRIMARY KEY,
	name        TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	portfolio   TEXT NOT NULL,
	config      TEXT NOT NULL,
	created_at  TEXT NOT NULL,
	updated_at  TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS runs (
	id            TEXT PRIMARY KEY,
	experiment_id TEXT NOT NULL,
	started_at    TEXT NOT NULL,
	finished_at   TEXT,
	status        TEXT NOT NULL,
	error         TEXT NOT NULL DEFAULT '',
	stats         TEXT NOT NULL DEFAULT '{}'
);
`

// ──────────────────── AssetRepository ────────────────────────────────────────

// UpsertAsset inserts or updates a single asset by symbol.
func (s *Store) UpsertAsset(ctx context.Context, a domain.Asset) error {
	if a.ID == "" {
		a.ID = a.Symbol
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO assets (id, symbol, name) VALUES (?,?,?) ON CONFLICT(symbol) DO UPDATE SET name=excluded.name`,
		a.ID, a.Symbol, a.Name)
	return err
}

// GetAsset returns the asset with the given symbol, or an error if not found.
func (s *Store) GetAsset(ctx context.Context, symbol string) (*domain.Asset, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, symbol, name FROM assets WHERE symbol=?`, symbol)
	var a domain.Asset
	if err := row.Scan(&a.ID, &a.Symbol, &a.Name); err != nil {
		return nil, err
	}
	return &a, nil
}

// ListAssets returns all stored assets ordered by symbol.
func (s *Store) ListAssets(ctx context.Context) ([]domain.Asset, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, symbol, name FROM assets ORDER BY symbol`)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // rows.Close in defer; final error captured by rows.Err()
	var assets []domain.Asset
	for rows.Next() {
		var a domain.Asset
		if err := rows.Scan(&a.ID, &a.Symbol, &a.Name); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

// DeleteAsset removes an asset and all its associated price records.
func (s *Store) DeleteAsset(ctx context.Context, symbol string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM assets WHERE symbol=?`, symbol)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM price_records WHERE symbol=?`, symbol)
	return err
}

// UpsertPriceRecords bulk-upserts price records, replacing rows with matching (symbol, date).
func (s *Store) UpsertPriceRecords(ctx context.Context, records []domain.PriceRecord) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // Rollback is a no-op after Commit; error is intentionally ignored
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO price_records (symbol,date,open,high,low,close,volume,adjusted_close)
		 VALUES (?,?,?,?,?,?,?,?)
		 ON CONFLICT(symbol,date) DO UPDATE SET
		   open=excluded.open, high=excluded.high, low=excluded.low,
		   close=excluded.close, volume=excluded.volume, adjusted_close=excluded.adjusted_close`)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:errcheck // stmt.Close in defer; any error is non-actionable here
	for _, r := range records {
		_, err := stmt.ExecContext(ctx, r.Symbol, r.Date.Format("2006-01-02"),
			r.Open, r.High, r.Low, r.Close, r.Volume, r.AdjustedClose)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetPriceRecords returns up to limit price records for the given symbol in ascending date order.
// A limit of 0 returns all records.
func (s *Store) GetPriceRecords(ctx context.Context, symbol string, limit int) ([]domain.PriceRecord, error) {
	q := `SELECT symbol,date,open,high,low,close,volume,adjusted_close
	      FROM price_records WHERE symbol=? ORDER BY date ASC`
	args := []any{symbol}
	if limit > 0 {
		q += " LIMIT ?"
		args = append(args, limit)
	}
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // rows.Close in defer; final error captured by rows.Err()
	var recs []domain.PriceRecord
	for rows.Next() {
		var r domain.PriceRecord
		var dateStr string
		if err := rows.Scan(&r.Symbol, &dateStr, &r.Open, &r.High, &r.Low, &r.Close, &r.Volume, &r.AdjustedClose); err != nil {
			return nil, err
		}
		r.Date, _ = time.Parse("2006-01-02", dateStr)
		recs = append(recs, r)
	}
	return recs, rows.Err()
}

// ──────────────────── ExperimentRepository ───────────────────────────────────

// SaveExperiment inserts or updates an experiment record (upsert by ID).
func (s *Store) SaveExperiment(ctx context.Context, exp domain.Experiment) error {
	port, _ := json.Marshal(exp.Portfolio)
	cfg, _ := json.Marshal(exp.Config)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO experiments (id,name,description,portfolio,config,created_at,updated_at)
		 VALUES (?,?,?,?,?,?,?)
		 ON CONFLICT(id) DO UPDATE SET name=excluded.name, description=excluded.description,
		   portfolio=excluded.portfolio, config=excluded.config, updated_at=excluded.updated_at`,
		exp.ID, exp.Name, exp.Description, string(port), string(cfg),
		exp.CreatedAt.Format(time.RFC3339), exp.UpdatedAt.Format(time.RFC3339))
	return err
}

// GetExperiment returns the experiment with the given ID.
func (s *Store) GetExperiment(ctx context.Context, id string) (*domain.Experiment, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id,name,description,portfolio,config,created_at,updated_at FROM experiments WHERE id=?`, id)
	return scanExperiment(row)
}

// ListExperiments returns all experiments ordered by creation date (descending).
func (s *Store) ListExperiments(ctx context.Context) ([]domain.Experiment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id,name,description,portfolio,config,created_at,updated_at FROM experiments ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // rows.Close in defer; final error captured by rows.Err()
	var exps []domain.Experiment
	for rows.Next() {
		e, err := scanExperiment(rows)
		if err != nil {
			return nil, err
		}
		exps = append(exps, *e)
	}
	return exps, rows.Err()
}

// DeleteExperiment removes an experiment and all its associated runs.
func (s *Store) DeleteExperiment(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM experiments WHERE id=?`, id)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM runs WHERE experiment_id=?`, id)
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanExperiment(row scanner) (*domain.Experiment, error) {
	var e domain.Experiment
	var portJSON, cfgJSON, createdStr, updatedStr string
	if err := row.Scan(&e.ID, &e.Name, &e.Description, &portJSON, &cfgJSON, &createdStr, &updatedStr); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(portJSON), &e.Portfolio)
	_ = json.Unmarshal([]byte(cfgJSON), &e.Config)
	e.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	e.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	return &e, nil
}

// ──────────────────── SimulationRepository ───────────────────────────────────

// SaveRun inserts or updates a simulation run record (upsert by ID).
func (s *Store) SaveRun(ctx context.Context, run domain.Run) error {
	statsJSON, _ := json.Marshal(run.Stats)
	var finishedAt *string
	if run.FinishedAt != nil {
		str := run.FinishedAt.Format(time.RFC3339)
		finishedAt = &str
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO runs (id,experiment_id,started_at,finished_at,status,error,stats)
		 VALUES (?,?,?,?,?,?,?)
		 ON CONFLICT(id) DO UPDATE SET
		   finished_at=excluded.finished_at, status=excluded.status,
		   error=excluded.error, stats=excluded.stats`,
		run.ID, run.ExperimentID, run.StartedAt.Format(time.RFC3339),
		finishedAt, string(run.Status), run.Error, string(statsJSON))
	return err
}

// GetRun returns the simulation run with the given ID.
func (s *Store) GetRun(ctx context.Context, runID string) (*domain.Run, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id,experiment_id,started_at,finished_at,status,error,stats FROM runs WHERE id=?`, runID)
	return scanRun(row)
}

// ListRuns returns all runs for the given experiment, most recent first.
func (s *Store) ListRuns(ctx context.Context, experimentID string) ([]domain.Run, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id,experiment_id,started_at,finished_at,status,error,stats FROM runs WHERE experiment_id=? ORDER BY started_at DESC`,
		experimentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // rows.Close in defer; final error captured by rows.Err()
	var runs []domain.Run
	for rows.Next() {
		r, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, *r)
	}
	return runs, rows.Err()
}

func scanRun(row scanner) (*domain.Run, error) {
	var r domain.Run
	var startedStr string
	var finishedStr *string
	var statsJSON string
	if err := row.Scan(&r.ID, &r.ExperimentID, &startedStr, &finishedStr, &r.Status, &r.Error, &statsJSON); err != nil {
		return nil, err
	}
	r.StartedAt, _ = time.Parse(time.RFC3339, startedStr)
	if finishedStr != nil {
		t, _ := time.Parse(time.RFC3339, *finishedStr)
		r.FinishedAt = &t
	}
	_ = json.Unmarshal([]byte(statsJSON), &r.Stats)
	return &r, nil
}
