package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RunExperiment triggers an async simulation run for the given experiment ID.
func (h *H) RunExperiment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	run, err := h.sim.RunExperiment(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/runs/"+run.ID, http.StatusSeeOther)
}

// RunResults renders the results page for a completed simulation run.
func (h *H) RunResults(w http.ResponseWriter, r *http.Request) {
	runID := chi.URLParam(r, "id")
	run, err := h.sim.GetRun(r.Context(), runID)
	if err != nil {
		http.Error(w, "run not found", http.StatusNotFound)
		return
	}
	exp, _ := h.results.GetExperiment(r.Context(), run.ExperimentID)
	data := map[string]any{
		"Title":      "Results",
		"Run":        run,
		"Experiment": exp,
	}
	if err := h.page("results.html").ExecuteTemplate(w, "layout", data); err != nil {
		renderErr(w, err)
	}
}
