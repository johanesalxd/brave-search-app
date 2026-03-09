package brave

import "encoding/json"

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

// RichHint is embedded in web search responses when rich data is available.
type RichHint struct {
	Vertical    string `json:"vertical"`
	CallbackKey string `json:"callback_key"`
}

// RichField appears in the web search Response when enable_rich_callback=1
// and the query has a matching rich vertical.
type RichField struct {
	Type string   `json:"type"`
	Hint RichHint `json:"hint"`
}

// RichResult is the response from the /web/rich endpoint.
type RichResult struct {
	Type    string          `json:"type"`
	Subtype string          `json:"subtype"`
	Data    json.RawMessage `json:"data"`
}

// Response represents the top-level structure of the Brave Search API response.
type Response struct {
	Web  WebData    `json:"web"`
	Rich *RichField `json:"rich,omitempty"`
}

// newsMetaURL is a helper struct for unmarshaling the nested meta_url object
// returned by the Brave News Search API.
type newsMetaURL struct {
	Hostname string `json:"hostname"`
}

// NewsResult represents a single news article from the Brave News Search API.
type NewsResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Age         string `json:"age"`
	Source      string `json:"-"`
}

// UnmarshalJSON implements custom JSON unmarshaling for NewsResult to extract
// the hostname from the nested meta_url object into the Source field.
func (n *NewsResult) UnmarshalJSON(data []byte) error {
	type alias struct {
		Title       string      `json:"title"`
		URL         string      `json:"url"`
		Description string      `json:"description"`
		Age         string      `json:"age"`
		MetaURL     newsMetaURL `json:"meta_url"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	n.Title = a.Title
	n.URL = a.URL
	n.Description = a.Description
	n.Age = a.Age
	n.Source = a.MetaURL.Hostname
	return nil
}

// NewsResponse represents the top-level structure of the Brave News Search API response.
type NewsResponse struct {
	Results []NewsResult `json:"results"`
}
