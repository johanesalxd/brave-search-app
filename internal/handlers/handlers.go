package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/johanesalxd/brave-search-app/internal/brave"
)

// SearchMode represents the active search mode for a query.
type SearchMode string

const (
	// ModeStandard performs a regular web search without rich data enrichment.
	ModeStandard SearchMode = "standard"
	// ModeRich performs a web search with rich data enrichment enabled.
	// This is the default mode.
	ModeRich SearchMode = "rich"
)

// Server handles HTTP requests for the Brave Search application.
type Server struct {
	braveClient *brave.Client
	templates   *template.Template
}

// NewServer creates a new Server instance.
// It parses templates from the provided directory.
func NewServer(braveClient *brave.Client, templateDir string) (*Server, error) {
	tmpl, err := template.ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, err
	}

	return &Server{
		braveClient: braveClient,
		templates:   tmpl,
	}, nil
}

// HandleIndex renders the home page.
func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.templates.ExecuteTemplate(w, "index.html", nil)
}

// HandleSearch processes search queries and renders results.
// It reads a "mode" query parameter ("standard" or "rich") to determine
// whether to attempt rich data enrichment. The default mode is "rich".
func (s *Server) HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	modeParam := r.URL.Query().Get("mode")
	mode := ModeRich
	if modeParam == string(ModeStandard) {
		mode = ModeStandard
	}

	data := struct {
		Query       string
		Results     []brave.Result
		RichDisplay *brave.RichDisplay
		Mode        SearchMode
		HasRich     bool
	}{
		Query: query,
		Mode:  mode,
	}

	if mode == ModeStandard {
		results, err := s.braveClient.Search(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data.Results = results
	} else {
		resp, richResult, err := s.braveClient.SearchWithRich(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if resp != nil {
			data.Results = resp.Web.Results
		}
		data.RichDisplay = brave.ParseRichDisplay(richResult)
		data.HasRich = data.RichDisplay != nil
	}

	if err := s.templates.ExecuteTemplate(w, "results.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleNews processes news search queries and renders news results.
func (s *Server) HandleNews(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	results, err := s.braveClient.NewsSearch(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Query   string
		Results []brave.NewsResult
	}{
		Query:   query,
		Results: results,
	}

	if err := s.templates.ExecuteTemplate(w, "news.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
