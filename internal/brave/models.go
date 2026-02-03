package brave

// Result represents a single search result item.
type Result struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// WebData represents the "web" section of the Brave Search response.
type WebData struct {
	Results []Result `json:"results"`
}

// Response represents the top-level structure of the Brave Search API response.
type Response struct {
	Web WebData `json:"web"`
}
