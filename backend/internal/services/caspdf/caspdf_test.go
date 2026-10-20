package caspdf

import (
	"context"
	"strings"
	"testing"
)

func TestExtractStatementDateFromText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		found    bool
	}{
		{
			name:     "CDSL Standard Statement as on DD-MMM-YYYY",
			input:    "CENTRAL DEPOSITORY SERVICES (INDIA) LIMITED\nStatement as on : 31-Jan-2026\nClient ID: 12345678",
			expected: "2026-01-31",
			found:    true,
		},
		{
			name:     "Valuation as on DD/MM/YYYY",
			input:    "Portfolio Valuation as on 15/09/2026\nAccount Details...",
			expected: "2026-09-15",
			found:    true,
		},
		{
			name:     "Holdings as on DD-Month-YYYY",
			input:    "Summary of Holdings as on : 28-February-2026\nTotal Portfolio Value: INR 500,000",
			expected: "2026-02-28",
			found:    true,
		},
		{
			name:     "No date in text",
			input:    "Some random text without any statement date",
			expected: "",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, ok := ExtractStatementDateFromText(tt.input)
			if ok != tt.found {
				t.Fatalf("ExtractStatementDateFromText() found = %v, want %v", ok, tt.found)
			}
			if tt.found {
				got := date.Format("2006-01-02")
				if got != tt.expected {
					t.Errorf("ExtractStatementDateFromText() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestParseStatementDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2026-01-31", "2026-01-31"},
		{"31-Jan-2026", "2026-01-31"},
		{"15/09/2026", "2026-09-15"},
		{"05-08-2026", "2026-08-05"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parsed, err := ParseStatementDate(tt.input)
			if err != nil {
				t.Fatalf("ParseStatementDate(%q) error: %v", tt.input, err)
			}
			got := parsed.Format("2006-01-02")
			if got != tt.expected {
				t.Errorf("ParseStatementDate(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestHeuristicParser(t *testing.T) {
	sampleText := `
CDSL Consolidated Account Statement
Statement as on : 31-Jan-2026
Equity Shares:
RELIANCE INDUSTRIES LTD INE002A01018 50 2500.00
TATA CONSULTANCY SERVICES INE467B01029 25 3800.00
Mutual Funds / ETFs:
PARAG PARIKH FLEXI CAP INF879O01019 1000 65.50
NIPPON INDIA ETF NIFTY BEES INF732E01015 200 240.00
`

	result := heuristicParser(sampleText, "2026-01-31")
	if result == nil {
		t.Fatal("expected non-nil result from heuristicParser")
	}

	if result.StatementDate != "2026-01-31" {
		t.Errorf("expected StatementDate = 2026-01-31, got %q", result.StatementDate)
	}

	if len(result.Holdings) != 4 {
		t.Errorf("expected 4 holdings, got %d", len(result.Holdings))
	}

	for _, h := range result.Holdings {
		if h.ISIN == "INF879O01019" || h.ISIN == "INF732E01015" {
			if h.Category != "mutualFunds" {
				t.Errorf("expected mutualFunds for %s, got %s", h.ISIN, h.Category)
			}
		} else {
			if h.Category != "stocks" {
				t.Errorf("expected stocks for %s, got %s", h.ISIN, h.Category)
			}
		}
	}
}

func TestDecryptAndExtractTextErrors(t *testing.T) {
	ctx := context.Background()

	// Empty bytes
	_, err := DecryptAndExtractText(ctx, nil, "ABCDE1234F")
	if err == nil {
		t.Error("expected error for empty bytes, got nil")
	}

	// Corrupted bytes
	_, err = DecryptAndExtractText(ctx, []byte("not a valid pdf content"), "ABCDE1234F")
	if err == nil {
		t.Error("expected error for corrupted pdf bytes, got nil")
	}
}

func TestEmbeddedSystemInstruction(t *testing.T) {
	if len(casSystemInstruction) == 0 {
		t.Fatal("expected casSystemInstruction to be embedded, got empty string")
	}

	keywords := []string{"CDSL", "Consolidated Account Statement", "category", "isin", "stocks", "mutualFunds"}
	for _, kw := range keywords {
		if !strings.Contains(casSystemInstruction, kw) {
			t.Errorf("expected casSystemInstruction to contain %q", kw)
		}
	}
}
