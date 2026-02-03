package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/johanesalxd/brave-search-app/internal/brave"
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
func (s *Server) HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	results, err := s.braveClient.Search(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Query   string
		Results []brave.Result
	}{
		Query:   query,
		Results: results,
	}

	if err := s.templates.ExecuteTemplate(w, "results.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
