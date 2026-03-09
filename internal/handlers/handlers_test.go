package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johanesalxd/brave-search-app/internal/brave"
)

// resultsTemplate is a minimal results template for handler tests.
const resultsTemplate = `{{if .HasRich}}RICH:{{.RichResult.Subtype}}{{end}}MODE:{{.Mode}}{{range .Results}}TITLE:{{.Title}}{{end}}`

// newsTemplate is a minimal news template for handler tests.
const newsTemplate = `QUERY:{{.Query}}{{range .Results}}TITLE:{{.Title}}SOURCE:{{.Source}}{{end}}`

// setupTestServer creates a temporary template directory and returns a Server
// backed by the provided brave.Client.
func setupTestServer(t *testing.T, client *brave.Client) *Server {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "templates")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	indexContent := "<html><body>Index</body></html>"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte(indexContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "results.html"), []byte(resultsTemplate), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "news.html"), []byte(newsTemplate), 0644); err != nil {
		t.Fatal(err)
	}

	srv, err := NewServer(client, tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func TestServer_HandleIndex(t *testing.T) {
	// Create a temporary template directory for testing
	tmpDir, err := os.MkdirTemp("", "templates")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create dummy index.html
	indexContent := "<html><body>Index</body></html>"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte(indexContent), 0644); err != nil {
		t.Fatal(err)
	}
	// Create dummy results.html to satisfy ParseGlob
	if err := os.WriteFile(filepath.Join(tmpDir, "results.html"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	// Create dummy news.html to satisfy ParseGlob
	if err := os.WriteFile(filepath.Join(tmpDir, "news.html"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	client := brave.NewClient("test-key")
	server, err := NewServer(client, tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	server.HandleIndex(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if rr.Body.String() != indexContent {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), indexContent)
	}
}

func TestServer_HandleSearch(t *testing.T) {
	const testCallbackKey = "cb-key-123"

	webResultBody := brave.Response{
		Web: brave.WebData{
			Results: []brave.Result{
				{Title: "Go Lang", URL: "https://go.dev", Description: "Go programming language."},
			},
		},
	}
	webResultWithRichBody := brave.Response{
		Web: brave.WebData{
			Results: []brave.Result{
				{Title: "London Weather", URL: "https://example.com", Description: "Weather info."},
			},
		},
		Rich: &brave.RichField{
			Type: "rich",
			Hint: brave.RichHint{Vertical: "weather", CallbackKey: testCallbackKey},
		},
	}
	richBody := brave.RichResult{Type: "rich", Subtype: "weather"}

	tests := []struct {
		name             string
		requestURL       string
		webStatus        int
		webBody          interface{}
		richStatus       int
		richBody         interface{}
		wantStatus       int
		wantBodyContains []string
	}{
		{
			name:             "mode rich with rich result",
			requestURL:       "/search?q=weather+london&mode=rich",
			webStatus:        http.StatusOK,
			webBody:          webResultWithRichBody,
			richStatus:       http.StatusOK,
			richBody:         richBody,
			wantStatus:       http.StatusOK,
			wantBodyContains: []string{"RICH:weather", "MODE:rich", "TITLE:London Weather"},
		},
		{
			name:             "mode standard bypasses rich",
			requestURL:       "/search?q=golang&mode=standard",
			webStatus:        http.StatusOK,
			webBody:          webResultBody,
			wantStatus:       http.StatusOK,
			wantBodyContains: []string{"MODE:standard", "TITLE:Go Lang"},
		},
		{
			name:             "mode rich no rich result falls back to web results",
			requestURL:       "/search?q=golang&mode=rich",
			webStatus:        http.StatusOK,
			webBody:          webResultBody,
			wantStatus:       http.StatusOK,
			wantBodyContains: []string{"MODE:rich", "TITLE:Go Lang"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			richSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.richStatus)
				if tt.richBody != nil {
					json.NewEncoder(w).Encode(tt.richBody)
				}
			}))
			defer richSrv.Close()

			webSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.webStatus)
				if tt.webBody != nil {
					json.NewEncoder(w).Encode(tt.webBody)
				}
			}))
			defer webSrv.Close()

			client := brave.NewClientWithURLs("test-key", webSrv.URL, richSrv.URL, "")
			srv := setupTestServer(t, client)

			req := httptest.NewRequest(http.MethodGet, tt.requestURL, nil)
			rr := httptest.NewRecorder()

			srv.HandleSearch(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("HandleSearch() status = %d, want %d", rr.Code, tt.wantStatus)
			}

			body := rr.Body.String()
			for _, want := range tt.wantBodyContains {
				if !strings.Contains(body, want) {
					t.Errorf("HandleSearch() body does not contain %q; body = %q", want, body)
				}
			}
		})
	}
}

func TestServer_HandleNews(t *testing.T) {
	newsBody := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"title":       "Go 1.22 Released",
				"url":         "https://go.dev/blog/go1.22",
				"description": "The Go team announces Go 1.22.",
				"age":         "2 hours ago",
				"meta_url":    map[string]string{"hostname": "go.dev"},
			},
		},
	}

	tests := []struct {
		name             string
		requestURL       string
		newsStatus       int
		newsBody         interface{}
		wantStatus       int
		wantBodyContains []string
	}{
		{
			name:             "success",
			requestURL:       "/news?q=golang",
			newsStatus:       http.StatusOK,
			newsBody:         newsBody,
			wantStatus:       http.StatusOK,
			wantBodyContains: []string{"QUERY:golang", "TITLE:Go 1.22 Released", "SOURCE:go.dev"},
		},
		{
			name:       "empty query redirects",
			requestURL: "/news?q=",
			wantStatus: http.StatusFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.newsStatus)
				if tt.newsBody != nil {
					json.NewEncoder(w).Encode(tt.newsBody)
				}
			}))
			defer newsSrv.Close()

			client := brave.NewClientWithURLs("test-key", "", "", newsSrv.URL)
			srv := setupTestServer(t, client)

			req := httptest.NewRequest(http.MethodGet, tt.requestURL, nil)
			rr := httptest.NewRecorder()

			srv.HandleNews(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("HandleNews() status = %d, want %d", rr.Code, tt.wantStatus)
			}

			body := rr.Body.String()
			for _, want := range tt.wantBodyContains {
				if !strings.Contains(body, want) {
					t.Errorf("HandleNews() body does not contain %q; body = %q", want, body)
				}
			}
		})
	}
}
