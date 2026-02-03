package brave

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Search(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
	}{
		{
			name:           "success",
			query:          "golang",
			responseStatus: http.StatusOK,
			responseBody: Response{
				Web: WebData{
					Results: []Result{
						{Title: "Go Programming Language", URL: "https://go.dev", Description: "Build simple, secure, scalable systems with Go."},
					},
				},
			},
			wantErr: false,
		},
		{
			name:           "empty query",
			query:          "",
			responseStatus: http.StatusOK,
			wantErr:        true,
		},
		{
			name:           "api error",
			query:          "error",
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client := NewClient("test-key")
			client.baseURL = server.URL

			results, err := client.Search(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(results) == 0 {
				t.Error("Client.Search() returned no results, expected at least one")
			}
		})
	}
}

func ExampleClient_Search() {
	// In a real application, you would use os.Getenv("BRAVE_API_KEY")
	client := NewClient("your-api-key")

	results, err := client.Search("golang")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, res := range results {
		fmt.Printf("Title: %s\nURL: %s\n", res.Title, res.URL)
	}
}
