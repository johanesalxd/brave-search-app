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
	// RichAPIBaseURL is the production URL for the Brave Rich Search API.
	RichAPIBaseURL = "https://api.search.brave.com/res/v1/web/rich"
	// NewsAPIBaseURL is the production URL for the Brave News Search API.
	NewsAPIBaseURL = "https://api.search.brave.com/res/v1/news/search"
	// DefaultTimeout is the default client timeout.
	DefaultTimeout = 30 * time.Second
)

// Client is a client for the Brave Search API.
type Client struct {
	apiKey     string
	baseURL    string
	richURL    string
	newsURL    string
	httpClient *http.Client
}

// NewClient creates a new Brave Search API client.
// It requires a valid API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: DefaultAPIBaseURL,
		richURL: RichAPIBaseURL,
		newsURL: NewsAPIBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// NewClientWithURLs creates a new Brave Search API client with custom base URLs.
// It is intended for use in tests that need to point the client at a mock server.
func NewClientWithURLs(apiKey, baseURL, richURL, newsURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		richURL: richURL,
		newsURL: newsURL,
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

// SearchWithRich performs a web search with rich data enrichment enabled.
// If the query triggers a rich vertical, it fetches the rich result and
// returns it alongside the web response. If no rich result is available,
// richResult is nil and err is nil.
func (c *Client) SearchWithRich(query string) (*Response, *RichResult, error) {
	if query == "" {
		return nil, nil, fmt.Errorf("search query cannot be empty")
	}

	reqURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	q := reqURL.Query()
	q.Set("q", query)
	q.Set("enable_rich_callback", "1")
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("api returned non-200 status code: %d", resp.StatusCode)
	}

	var braveResp Response
	if err := json.NewDecoder(resp.Body).Decode(&braveResp); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if braveResp.Rich == nil || braveResp.Rich.Hint.CallbackKey == "" {
		return &braveResp, nil, nil
	}

	richResult, err := c.fetchRich(braveResp.Rich.Hint.CallbackKey)
	if err != nil {
		// graceful fallback: return web results without rich data
		return &braveResp, nil, nil
	}

	return &braveResp, richResult, nil
}

// fetchRich retrieves rich result data from the Brave Rich Search endpoint
// using the provided callback key obtained from a prior web search response.
func (c *Client) fetchRich(callbackKey string) (*RichResult, error) {
	reqURL, err := url.Parse(c.richURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rich URL: %w", err)
	}

	q := reqURL.Query()
	q.Set("callback_key", callbackKey)
	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create rich request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rich request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rich api returned non-200 status code: %d", resp.StatusCode)
	}

	var richResult RichResult
	if err := json.NewDecoder(resp.Body).Decode(&richResult); err != nil {
		return nil, fmt.Errorf("failed to decode rich response: %w", err)
	}

	return &richResult, nil
}

// NewsSearch performs a news search using the Brave News Search API.
// It returns a list of news results or an error if the request fails.
func (c *Client) NewsSearch(query string) ([]NewsResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	reqURL, err := url.Parse(c.newsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse news URL: %w", err)
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

	var newsResp NewsResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return newsResp.Results, nil
}
