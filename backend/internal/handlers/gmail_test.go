package handlers

import (
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
