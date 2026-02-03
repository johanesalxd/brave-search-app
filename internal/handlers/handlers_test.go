package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/johanesalxd/brave-search-app/internal/brave"
)

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
