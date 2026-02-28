// Package httpAdapter wires the Chi router and all HTTP handlers together.
package httpadapter

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/gjcourt/drift/internal/adapters/http/handlers"
	"github.com/gjcourt/drift/internal/domain"
	"github.com/gjcourt/drift/internal/ports/inbound"
)

// New builds the main HTTP handler. tmplDir is the directory containing *.html
// templates; staticDir is the directory served under /static/.
func New(
	ingest inbound.DataIngestionService,
	results inbound.ResultsService,
	sim inbound.SimulationService,
	tmplDir string,
	staticDir string,
) http.Handler {
	base, err := loadBase(tmplDir)
	if err != nil {
		slog.Error("load base template", "dir", tmplDir, "err", err)
	}

	h := handlers.New(ingest, results, sim, base, tmplDir)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", h.Dashboard)

	r.Route("/data", func(r chi.Router) {
		r.Get("/", h.DataManager)
		r.Post("/upload", h.UploadCSV)
		r.Delete("/{symbol}", h.DeleteAsset)
	})

	r.Route("/experiments", func(r chi.Router) {
		r.Get("/", h.ListExperiments)
		r.Get("/new", h.NewExperimentForm)
		r.Post("/", h.CreateExperiment)
		r.Get("/{id}", h.ExperimentDetail)
		r.Post("/{id}/run", h.RunExperiment)
	})

	r.Get("/runs/{id}", h.RunResults)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	return r
}

// loadBase parses only the layout template (no page-specific templates).
// Page templates are parsed per-request via handlers.H.page() to avoid
// the Go html/template shared {{define}} namespace problem.
func loadBase(dir string) (*template.Template, error) {
	funcs := template.FuncMap{
		// mul multiplies two float64 values; used in templates for percentage display.
		"mul": func(a, b float64) float64 { return a * b },
		// statsJSON serialises ResultStats to a JSON literal safe for inline <script> use.
		"statsJSON": func(s domain.ResultStats) template.JS {
			b, _ := json.Marshal(s)
			return template.JS(b)
		},
	}
	layoutFile := filepath.Join(dir, "layout.html")
	return template.New("").Funcs(funcs).ParseFiles(layoutFile)
}
