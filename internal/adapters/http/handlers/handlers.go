// Package handlers contains all HTTP request handlers for the Drift web UI.
//
// Handlers are grouped by domain area: assets (data manager), experiments
// (experiment builder + list + detail), and simulations (run + results).
//
// Templates are rendered using a clone-per-request approach to avoid the
// Go html/template shared {{define}} namespace problem: the base template
// (layout.html) is parsed once at startup; page templates are parsed into a
// clone of the base on every request via [H.page].
package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/gjcourt/drift/internal/ports/inbound"
)

// H holds all handler dependencies and the base template used to derive
// per-page templates at request time.
type H struct {
	ingest   inbound.DataIngestionService
	results  inbound.ResultsService
	sim      inbound.SimulationService
	baseTmpl *template.Template
	tmplDir  string
}

// New constructs a handler set.
//   - baseTmpl must contain only layout.html (no page templates).
//   - tmplDir is the directory from which page templates are read on demand.
func New(
	ingest inbound.DataIngestionService,
	results inbound.ResultsService,
	sim inbound.SimulationService,
	baseTmpl *template.Template,
	tmplDir string,
) *H {
	return &H{
		ingest:   ingest,
		results:  results,
		sim:      sim,
		baseTmpl: baseTmpl,
		tmplDir:  tmplDir,
	}
}

// page returns a clone of the base template with the named page template
// parsed into it. Each request gets its own independent template.Template so
// that the shared {{define}} namespace does not bleed across pages.
func (h *H) page(name string) *template.Template {
	t := template.Must(h.baseTmpl.Clone())
	template.Must(t.ParseFiles(filepath.Join(h.tmplDir, name)))
	return t
}

// renderErr writes a plain-text 500 response.
func renderErr(w http.ResponseWriter, err error) {
	http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
}
