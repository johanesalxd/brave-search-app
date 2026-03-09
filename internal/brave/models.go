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

// RichWeather holds weather data from the rich search vertical.
type RichWeather struct {
	Location    string  `json:"location"`
	TempC       float64 `json:"temp_c"`
	TempF       float64 `json:"temp_f"`
	Condition   string  `json:"condition"`
	Humidity    int     `json:"humidity"`
	WindKph     float64 `json:"wind_kph"`
	ForecastDay string  `json:"forecast_day"`
}

// RichStock holds stock market data from the rich search vertical.
type RichStock struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
	Currency      string  `json:"currency"`
}

// RichCrypto holds cryptocurrency data from the rich search vertical.
type RichCrypto struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	PriceUSD  float64 `json:"price_usd"`
	Change24h float64 `json:"change_24h"`
	MarketCap float64 `json:"market_cap"`
}

// RichCurrency holds currency conversion data from the rich search vertical.
type RichCurrency struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Rate         float64 `json:"rate"`
	Amount       float64 `json:"amount"`
	Result       float64 `json:"result"`
}

// RichCalculator holds calculator result data from the rich search vertical.
type RichCalculator struct {
	Expression string `json:"expression"`
	Result     string `json:"result"`
}

// RichDefinition holds a single word definition entry.
type RichDefinition struct {
	PartOfSpeech string `json:"part_of_speech"`
	Definition   string `json:"definition"`
	Example      string `json:"example"`
}

// RichDefinitions holds word definition data from the rich search vertical.
type RichDefinitions struct {
	Word        string           `json:"word"`
	Definitions []RichDefinition `json:"definitions"`
}

// RichDisplay is passed to the results template and holds the parsed rich data.
// Only one of the typed fields will be non-nil, matching the Subtype.
// FallbackJSON holds pretty-printed JSON for unknown subtypes.
type RichDisplay struct {
	Subtype      string
	Weather      *RichWeather
	Stock        *RichStock
	Crypto       *RichCrypto
	Currency     *RichCurrency
	Calculator   *RichCalculator
	Definitions  *RichDefinitions
	FallbackJSON string
}

// ParseRichDisplay converts a RichResult into a RichDisplay suitable for
// template rendering. It parses the Data field into the appropriate typed
// struct based on Subtype. On any parse failure, it falls back to
// pretty-printed JSON. ParseRichDisplay never returns an error.
func ParseRichDisplay(r *RichResult) *RichDisplay {
	if r == nil {
		return nil
	}
	d := &RichDisplay{Subtype: r.Subtype}
	switch r.Subtype {
	case "weather":
		var v RichWeather
		if json.Unmarshal(r.Data, &v) == nil {
			d.Weather = &v
			return d
		}
	case "stock":
		var v RichStock
		if json.Unmarshal(r.Data, &v) == nil {
			d.Stock = &v
			return d
		}
	case "cryptocurrency":
		var v RichCrypto
		if json.Unmarshal(r.Data, &v) == nil {
			d.Crypto = &v
			return d
		}
	case "currency":
		var v RichCurrency
		if json.Unmarshal(r.Data, &v) == nil {
			d.Currency = &v
			return d
		}
	case "calculator":
		var v RichCalculator
		if json.Unmarshal(r.Data, &v) == nil {
			d.Calculator = &v
			return d
		}
	case "definitions":
		var v RichDefinitions
		if json.Unmarshal(r.Data, &v) == nil {
			d.Definitions = &v
			return d
		}
	}
	// Fallback: pretty-print raw JSON.
	if b, err := json.MarshalIndent(json.RawMessage(r.Data), "", "  "); err == nil {
		d.FallbackJSON = string(b)
	} else {
		d.FallbackJSON = string(r.Data)
	}
	return d
}
