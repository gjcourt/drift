package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// DataManager renders the data-manager page listing all ingested assets.
func (h *H) DataManager(w http.ResponseWriter, r *http.Request) {
	assets, err := h.ingest.ListAssets(r.Context())
	if err != nil {
		renderErr(w, err)
		return
	}
	data := map[string]any{
		"Title":  "Data Manager",
		"Assets": assets,
	}
	if err := h.page("data-manager.html").ExecuteTemplate(w, "layout", data); err != nil {
		renderErr(w, err)
	}
}

// UploadCSV handles multipart CSV uploads. On success it sends an HX-Redirect
// header so HTMX refreshes to the data-manager page; plain form submissions
// follow the same redirect via the Location header.
func (h *H) UploadCSV(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field required", http.StatusBadRequest)
		return
	}
	defer file.Close() //nolint:errcheck // Close on read-only multipart form file; error is non-actionable

	count, err := h.ingest.IngestCSV(r.Context(), file, header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// HTMX partial redirect
	w.Header().Set("HX-Redirect", "/data")
	http.Redirect(w, r, "/data", http.StatusSeeOther)
	fmt.Fprintf(w, "Ingested %d records", count) //nolint:errcheck // response already redirected; write error is inconsequential
}

// DeleteAsset removes an asset and all its associated price records.
func (h *H) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	if err := h.ingest.DeleteAsset(r.Context(), symbol); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
