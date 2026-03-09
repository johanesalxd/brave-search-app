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

func TestClient_SearchWithRich(t *testing.T) {
	const testCallbackKey = "test-callback-key"

	tests := []struct {
		name           string
		query          string
		webStatus      int
		webBody        interface{}
		richStatus     int
		richBody       interface{}
		wantErr        bool
		wantRichResult bool
		wantWebResults bool
	}{
		{
			name:      "rich result returned",
			query:     "weather london",
			webStatus: http.StatusOK,
			webBody: Response{
				Web: WebData{
					Results: []Result{
						{Title: "London Weather", URL: "https://example.com", Description: "Current weather."},
					},
				},
				Rich: &RichField{
					Type: "rich",
					Hint: RichHint{
						Vertical:    "weather",
						CallbackKey: testCallbackKey,
					},
				},
			},
			richStatus:     http.StatusOK,
			richBody:       RichResult{Type: "rich", Subtype: "weather"},
			wantErr:        false,
			wantRichResult: true,
			wantWebResults: true,
		},
		{
			name:      "no rich result",
			query:     "golang",
			webStatus: http.StatusOK,
			webBody: Response{
				Web: WebData{
					Results: []Result{
						{Title: "Go Programming Language", URL: "https://go.dev", Description: "Build simple, secure, scalable systems with Go."},
					},
				},
			},
			wantErr:        false,
			wantRichResult: false,
			wantWebResults: true,
		},
		{
			name:      "rich fetch error falls back gracefully",
			query:     "stock GOOG",
			webStatus: http.StatusOK,
			webBody: Response{
				Web: WebData{
					Results: []Result{
						{Title: "GOOG Stock", URL: "https://example.com", Description: "Stock data."},
					},
				},
				Rich: &RichField{
					Type: "rich",
					Hint: RichHint{
						Vertical:    "stock",
						CallbackKey: testCallbackKey,
					},
				},
			},
			richStatus:     http.StatusInternalServerError,
			wantErr:        false,
			wantRichResult: false,
			wantWebResults: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			richServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.richStatus)
				if tt.richBody != nil {
					json.NewEncoder(w).Encode(tt.richBody)
				}
			}))
			defer richServer.Close()

			webServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.webStatus)
				if tt.webBody != nil {
					json.NewEncoder(w).Encode(tt.webBody)
				}
			}))
			defer webServer.Close()

			client := NewClient("test-key")
			client.baseURL = webServer.URL
			client.richURL = richServer.URL

			resp, richResult, err := client.SearchWithRich(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.SearchWithRich() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if resp == nil {
				t.Fatal("Client.SearchWithRich() returned nil response, want non-nil")
			}

			if tt.wantWebResults && len(resp.Web.Results) == 0 {
				t.Error("Client.SearchWithRich() returned no web results, expected at least one")
			}

			if tt.wantRichResult && richResult == nil {
				t.Error("Client.SearchWithRich() returned nil rich result, expected non-nil")
			}

			if !tt.wantRichResult && richResult != nil {
				t.Errorf("Client.SearchWithRich() returned unexpected rich result: %+v", richResult)
			}
		})
	}
}

func TestClient_NewsSearch(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		wantSource     string
	}{
		{
			name:           "success",
			query:          "golang news",
			responseStatus: http.StatusOK,
			responseBody: map[string]interface{}{
				"results": []map[string]interface{}{
					{
						"title":       "Go 1.22 Released",
						"url":         "https://go.dev/blog/go1.22",
						"description": "The Go team announces Go 1.22.",
						"age":         "2 hours ago",
						"meta_url":    map[string]string{"hostname": "go.dev"},
					},
				},
			},
			wantErr:    false,
			wantSource: "go.dev",
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
			client.newsURL = server.URL

			results, err := client.NewsSearch(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.NewsSearch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(results) == 0 {
					t.Error("Client.NewsSearch() returned no results, expected at least one")
					return
				}
				if results[0].Source != tt.wantSource {
					t.Errorf("Client.NewsSearch() source = %q, want %q", results[0].Source, tt.wantSource)
				}
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
