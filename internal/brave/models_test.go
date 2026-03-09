package brave

import (
	"encoding/json"
	"testing"
)

func TestParseRichDisplay(t *testing.T) {
	tests := []struct {
		name            string
		input           *RichResult
		wantNil         bool
		wantSubtype     string
		wantWeather     bool
		wantStock       bool
		wantCrypto      bool
		wantCurrency    bool
		wantCalculator  bool
		wantDefinitions bool
		wantFallback    bool
	}{
		{
			name: "weather",
			input: &RichResult{
				Type:    "rich",
				Subtype: "weather",
				Data: json.RawMessage(`{
					"location": "London",
					"temp_c": 12.5,
					"temp_f": 54.5,
					"condition": "Partly Cloudy",
					"humidity": 72,
					"wind_kph": 18.0,
					"forecast_day": "Monday"
				}`),
			},
			wantSubtype: "weather",
			wantWeather: true,
		},
		{
			name: "stock",
			input: &RichResult{
				Type:    "rich",
				Subtype: "stock",
				Data: json.RawMessage(`{
					"symbol": "AAPL",
					"name": "Apple Inc.",
					"price": 189.50,
					"change": 2.30,
					"change_percent": 1.23,
					"currency": "USD"
				}`),
			},
			wantSubtype: "stock",
			wantStock:   true,
		},
		{
			name: "unknown_subtype",
			input: &RichResult{
				Type:    "rich",
				Subtype: "sports",
				Data:    json.RawMessage(`{"team":"Arsenal","score":3}`),
			},
			wantSubtype:  "sports",
			wantFallback: true,
		},
		{
			name:    "nil_input",
			input:   nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseRichDisplay(tt.input)

			if tt.wantNil {
				if got != nil {
					t.Errorf("ParseRichDisplay() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("ParseRichDisplay() = nil, want non-nil")
			}

			if got.Subtype != tt.wantSubtype {
				t.Errorf("Subtype = %q, want %q", got.Subtype, tt.wantSubtype)
			}

			if tt.wantWeather && got.Weather == nil {
				t.Error("Weather = nil, want non-nil")
			}
			if !tt.wantWeather && got.Weather != nil {
				t.Error("Weather = non-nil, want nil")
			}

			if tt.wantStock && got.Stock == nil {
				t.Error("Stock = nil, want non-nil")
			}
			if !tt.wantStock && got.Stock != nil {
				t.Error("Stock = non-nil, want nil")
			}

			if tt.wantFallback && got.FallbackJSON == "" {
				t.Error("FallbackJSON = empty, want non-empty")
			}
			if !tt.wantFallback && got.FallbackJSON != "" {
				t.Errorf("FallbackJSON = %q, want empty", got.FallbackJSON)
			}

			// Verify typed fields are nil when fallback is expected.
			if tt.wantFallback {
				if got.Weather != nil || got.Stock != nil || got.Crypto != nil ||
					got.Currency != nil || got.Calculator != nil || got.Definitions != nil {
					t.Error("expected all typed fields to be nil for unknown subtype")
				}
			}
		})
	}
}

func ExampleParseRichDisplay() {
	r := &RichResult{
		Type:    "rich",
		Subtype: "calculator",
		Data:    json.RawMessage(`{"expression":"2+2","result":"4"}`),
	}
	d := ParseRichDisplay(r)
	_ = d.Calculator.Result // "4"
}
