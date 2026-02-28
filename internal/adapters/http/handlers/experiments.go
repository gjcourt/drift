package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/gjcourt/drift/internal/domain"
)

// Dashboard renders the home page with a summary of assets and experiments.
func (h *H) Dashboard(w http.ResponseWriter, r *http.Request) {
	assets, _ := h.ingest.ListAssets(r.Context())
	exps, _ := h.results.ListExperiments(r.Context())
	data := map[string]any{
		"Title":       "Dashboard",
		"Assets":      assets,
		"Experiments": exps,
	}
	if err := h.page("dashboard.html").ExecuteTemplate(w, "layout", data); err != nil {
		renderErr(w, err)
	}
}

// ListExperiments renders the experiments index page.
func (h *H) ListExperiments(w http.ResponseWriter, r *http.Request) {
	exps, err := h.results.ListExperiments(r.Context())
	if err != nil {
		renderErr(w, err)
		return
	}
	data := map[string]any{
		"Title":       "Experiments",
		"Experiments": exps,
	}
	if err := h.page("experiments.html").ExecuteTemplate(w, "layout", data); err != nil {
		renderErr(w, err)
	}
}

// NewExperimentForm renders the experiment-builder form.
func (h *H) NewExperimentForm(w http.ResponseWriter, r *http.Request) {
	assets, _ := h.ingest.ListAssets(r.Context())
	data := map[string]any{
		"Title":  "New Experiment",
		"Assets": assets,
	}
	if err := h.page("experiment-builder.html").ExecuteTemplate(w, "layout", data); err != nil {
		renderErr(w, err)
	}
}

// CreateExperiment handles form submission to create a new experiment.
func (h *H) CreateExperiment(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	numPaths, _ := strconv.Atoi(r.FormValue("num_paths"))
	horizon, _ := strconv.Atoi(r.FormValue("horizon_days"))
	lookback, _ := strconv.Atoi(r.FormValue("lookback_days"))
	startVal, _ := strconv.ParseFloat(r.FormValue("start_value"), 64)
	contrib, _ := strconv.ParseFloat(r.FormValue("annual_contribution"), 64)

	if numPaths <= 0 {
		numPaths = 1000
	}
	if horizon <= 0 {
		horizon = 252
	}
	if lookback <= 0 {
		lookback = 756
	}
	if startVal <= 0 {
		startVal = 100_000
	}

	symbols := r.Form["symbols"]
	rawWeights := r.Form["weights"]
	var assets []domain.PortfolioAsset
	for i, sym := range symbols {
		weight := 1.0 / float64(len(symbols))
		if i < len(rawWeights) {
			if parsed, err := strconv.ParseFloat(rawWeights[i], 64); err == nil {
				weight = parsed / 100.0
			}
		}
		assets = append(assets, domain.PortfolioAsset{Symbol: sym, Weight: weight})
	}

	exp := domain.Experiment{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Portfolio:   domain.Portfolio{Assets: assets, Rebalance: domain.RebalanceAnnual},
		Config: domain.SimulationConfig{
			Model:              domain.SimulationModel(r.FormValue("model")),
			NumPaths:           numPaths,
			HorizonDays:        horizon,
			LookbackDays:       lookback,
			StartValue:         startVal,
			AnnualContribution: contrib,
		},
	}
	if exp.Config.Model == "" {
		exp.Config.Model = domain.ModelGBM
	}

	created, err := h.results.CreateExperiment(r.Context(), exp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.FormValue("run_now") == "1" {
		http.Redirect(w, r, "/experiments/"+created.ID+"/run", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/experiments/"+created.ID, http.StatusSeeOther)
	}
}

// ExperimentDetail renders a single experiment page with its run history.
func (h *H) ExperimentDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	exp, err := h.results.GetExperiment(r.Context(), id)
	if err != nil {
		http.Error(w, "experiment not found", http.StatusNotFound)
		return
	}
	runs, _ := h.results.ListRuns(r.Context(), id)
	data := map[string]any{
		"Title":      exp.Name,
		"Experiment": exp,
		"Runs":       runs,
	}
	if err := h.page("experiment-detail.html").ExecuteTemplate(w, "layout", data); err != nil {
		renderErr(w, err)
	}
}
