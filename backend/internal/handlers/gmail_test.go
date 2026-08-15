package handlers

import (
	"context"
	"strings"
	"testing"
)

func TestCleanInvestmentName(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		isMutualFund bool
		expected     string
	}{
		// Stocks
		{
			name:         "L&T Stock",
			input:        "LARSEN & TOUBRO LIMITED-EQUITY SHARES OF RS.2/- EACH",
			isMutualFund: false,
			expected:     "LARSEN & TOUBRO LIMITED",
		},
		{
			name:         "NTPC Stock",
			input:        "NTPC LIMITED-EQUITY SHARES",
			isMutualFund: false,
			expected:     "NTPC LIMITED",
		},
		{
			name:         "Tata Motors Stock",
			input:        "TATA MOTORS LTD - NEW EQUITY SHARES OF RS. 2/- AFTER SUB-DIVISION",
			isMutualFund: false,
			expected:     "TATA MOTORS LTD",
		},
		{
			name:         "Simple Stock No Suffix",
			input:        "RELIANCE INDUSTRIES LTD",
			isMutualFund: false,
			expected:     "RELIANCE INDUSTRIES LTD",
		},
		// Mutual Funds
		{
			name:         "Motilal Oswal MF 1",
			input:        "MOTILAL OSWAL AMC LTD#MOTILAL OSWAL MF-MOTILAL OSWAL MIDCAP 30 FUND-DIRECT PLAN-GROWTH",
			isMutualFund: true,
			expected:     "MOTILAL OSWAL MIDCAP 30 FUND",
		},
		{
			name:         "Nippon MF 1",
			input:        "NIPPON LIFE INDIA AM LTD#NIPPON INDIA MF-NIPPON INDIA LARGE CAP FUND DIRECT GROWTH PL-GROWTH OPT",
			isMutualFund: true,
			expected:     "NIPPON INDIA LARGE CAP FUND",
		},
		{
			name:         "HDFC Balanced Advantage MF",
			input:        "HDFC BALANCED ADVANTAGE FUND-DIRECT PLAN - GROWTH",
			isMutualFund: true,
			expected:     "HDFC BALANCED ADVANTAGE FUND",
		},
		{
			name:         "ICICI Prudential MF",
			input:        "ICICI PRUDENTIAL BLUECHIP FUND - REGULAR PLAN - GROWTH",
			isMutualFund: true,
			expected:     "ICICI PRUDENTIAL BLUECHIP FUND",
		},
		{
			name:         "Parag Parikh MF",
			input:        "PPFAS AMC LTD#PARAG PARIKH MF-PARAG PARIKH FLEXI CAP FUND-DIRECT PLAN-GROWTH",
			isMutualFund: true,
			expected:     "PARAG PARIKH FLEXI CAP FUND",
		},
		{
			name:         "Mirae Asset Gold ETF",
			input:        "MIRAE ASSET MF-MIRAE ASSET GOLD ETF",
			isMutualFund: true,
			expected:     "MIRAE ASSET GOLD ETF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanInvestmentName(tt.input, tt.isMutualFund)
			if result != tt.expected {
				t.Errorf("cleanInvestmentName(%q, %t) = %q; expected %q", tt.input, tt.isMutualFund, result, tt.expected)
			}
		})
	}
}

func TestGeminiSystemInstructionPrompt(t *testing.T) {
	if len(geminiSystemInstruction) == 0 {
		t.Fatal("expected geminiSystemInstruction prompt to be loaded from embedded file, but got empty string")
	}
	expectedKeywords := []string{
		"CDSL depository",
		"transactions",
		"category",
		"isin",
		"ticker",
		"quantity",
	}
	for _, kw := range expectedKeywords {
		if !strings.Contains(geminiSystemInstruction, kw) {
			t.Errorf("expected geminiSystemInstruction to contain keyword %q", kw)
		}
	}
}

func TestFetchStockPriceYahoo(t *testing.T) {
	h := &GmailHandler{}
	ctx := context.Background()

	// Test historical price for LT on 2025-04-16
	price := h.fetchStockPrice(ctx, "LT", "2025-04-16")
	if price <= 0 {
		t.Logf("fetchStockPrice for LT on 2025-04-16 returned %.2f (network/API may be unavailable in test env)", price)
	} else {
		t.Logf("fetchStockPrice for LT on 2025-04-16: %.2f", price)
		if price < 2500 || price > 4000 {
			t.Errorf("unexpected historical price for LT on 2025-04-16: %.2f", price)
		}
	}
}

func TestGenerateMFTicker(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"PARAG PARIKH FLEXI CAP FUND-DIRECT PLAN-GROWTH", "PARAG_PARIKH_FLEXI_"},
		{"HDFC BALANCED ADVANTAGE FUND", "HDFC_BALANCED_ADVANT"},
		{"NIPPON INDIA LARGE CAP FUND", "NIPPON_INDIA_LARGE_C"},
		{"MIRAE ASSET GOLD ETF", "MIRAE_ASSET_GOLD_ETF"},
	}
	for _, tt := range tests {
		result := generateMFTicker(tt.input)
		if !strings.HasPrefix(result, tt.expected[:10]) {
			t.Errorf("generateMFTicker(%q) = %q; expected prefix %q", tt.input, result, tt.expected[:10])
		}
	}
}

func TestFetchMutualFundNAV(t *testing.T) {
	h := &GmailHandler{}
	ctx := context.Background()

	// Test historical NAV for Parag Parikh Flexi Cap on 2024-03-15
	nav := h.fetchMutualFundNAV(ctx, "PARAG PARIKH FLEXI CAP FUND", "INF879O01019", "2024-03-15")
	if nav <= 0 {
		t.Logf("fetchMutualFundNAV returned %.2f (network/API may be unavailable in test env)", nav)
	} else {
		t.Logf("fetchMutualFundNAV for Parag Parikh on 2024-03-15: %.2f", nav)
		if nav < 50 || nav > 100 {
			t.Errorf("unexpected historical NAV: %.2f", nav)
		}
	}

	// Test historical NAV for Mirae Asset Gold ETF on 2026-01-06
	etfNav := h.fetchMutualFundNAV(ctx, "MIRAE ASSET MF-MIRAE ASSET GOLD ETF", "INF769K01JP9", "2026-01-06")
	if etfNav <= 0 {
		t.Logf("fetchMutualFundNAV for Mirae Gold ETF returned %.2f (network/API may be unavailable in test env)", etfNav)
	} else {
		t.Logf("fetchMutualFundNAV for Mirae Gold ETF on 2026-01-06: %.2f", etfNav)
		if etfNav < 100 || etfNav > 200 {
			t.Errorf("unexpected historical NAV for Mirae Gold ETF: %.2f", etfNav)
		}
	}
}



