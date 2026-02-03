package brave

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultAPIBaseURL is the production URL for Brave Search API.
	DefaultAPIBaseURL = "https://api.search.brave.com/res/v1/web/search"
	// DefaultTimeout is the default client timeout.
	DefaultTimeout = 30 * time.Second
)

// Client is a client for the Brave Search API.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Brave Search API client.
// It requires a valid API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: DefaultAPIBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// Search performs a web search using the Brave Search API.
// It returns a list of results or an error if the request fails.
func (c *Client) Search(query string) ([]Result, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	reqURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	q := reqURL.Query()
	q.Set("q", query)
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned non-200 status code: %d", resp.StatusCode)
	}

	var braveResp Response
	if err := json.NewDecoder(resp.Body).Decode(&braveResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return braveResp.Web.Results, nil
}
