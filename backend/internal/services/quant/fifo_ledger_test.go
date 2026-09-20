package quant

import (
	"testing"
	"time"
)

func TestBuildFIFOLedger_OrderingAndSelling(t *testing.T) {
	now := time.Now()
	// Buy 100 units at 50 2 years ago, Buy 100 units at 60 6 months ago, Sell 50 units 3 months ago
	txs := []RawTransaction{
		{
			ISIN:       "INF123K01010",
			SchemeName: "Test Equity Fund",
			AssetType:  "Equity",
			Type:       "buy",
			Units:      100,
			NAV:        50.0,
			Date:       now.AddDate(-2, 0, 0),
		},
		{
			ISIN:       "INF123K01010",
			SchemeName: "Test Equity Fund",
			AssetType:  "Equity",
			Type:       "buy",
			Units:      100,
			NAV:        60.0,
			Date:       now.AddDate(0, -6, 0),
		},
		{
			ISIN:       "INF123K01010",
			SchemeName: "Test Equity Fund",
			AssetType:  "Equity",
			Type:       "sell",
			Units:      50,
			NAV:        70.0,
			Date:       now.AddDate(0, -3, 0),
		},
	}

	navMap := map[string]float64{"INF123K01010": 80.0}
	lots, harvest := BuildFIFOLedger("user_test", txs, navMap, now)

	if len(lots) != 2 {
		t.Fatalf("expected 2 lots remaining, got %d", len(lots))
	}

	// First lot should have 50 units remaining (100 - 50 = 50) and be LTCG (>12 months)
	if lots[0].Units != 50.0 {
		t.Errorf("expected lot 0 to have 50 units, got %.2f", lots[0].Units)
	}
	if !lots[0].IsLTCG {
		t.Errorf("expected lot 0 to be LTCG")
	}
	if lots[0].PurchaseNAV != 50.0 {
		t.Errorf("expected lot 0 purchase NAV 50.0, got %.2f", lots[0].PurchaseNAV)
	}
	// Total cost = 50 * 50 = 2500, Unrealized gain = 50 * (80 - 50) = 1500
	if lots[0].UnrealizedGain != 1500.0 {
		t.Errorf("expected unrealized gain 1500.0, got %.2f", lots[0].UnrealizedGain)
	}

	// Second lot should have 100 units remaining and NOT be LTCG (6 months old)
	if lots[1].Units != 100.0 {
		t.Errorf("expected lot 1 to have 100 units, got %.2f", lots[1].Units)
	}
	if lots[1].IsLTCG {
		t.Errorf("expected lot 1 to NOT be LTCG")
	}

	// Tax harvesting should identify the LTCG gain of 1500 as harvestable within ₹1.25L limit
	if harvest.HarvestableLTCG != 1500.0 {
		t.Errorf("expected harvestable LTCG 1500.0, got %.2f", harvest.HarvestableLTCG)
	}
	if harvest.PotentialLTCGTaxSaved != 1500.0*0.125 {
		t.Errorf("expected tax saved %.2f, got %.2f", 1500.0*0.125, harvest.PotentialLTCGTaxSaved)
	}
}

func TestCalculateTaxHarvesting_PostBudget2024Rules(t *testing.T) {
	now := time.Now()
	// Test case: Debt fund with gains -> should flag shift to Arbitrage/SGB
	// Short-term equity lot with loss -> should flag STCL offset @ 20%
	// Long-term equity lot with gain of 200,000 -> should cap LTCG exemption at 125,000
	txs := []RawTransaction{
		{
			ISIN:       "INF999K01001",
			SchemeName: "Mega Cap Equity Fund",
			AssetType:  "Equity",
			Type:       "buy",
			Units:      2000,
			NAV:        100.0,
			Date:       now.AddDate(-3, 0, 0), // LTCG
		},
		{
			ISIN:       "INF888K01002",
			SchemeName: "Tech Stock Basket",
			AssetType:  "Equity",
			Type:       "buy",
			Units:      500,
			NAV:        200.0,
			Date:       now.AddDate(0, -2, 0), // STCG (loss)
		},
		{
			ISIN:       "INF777K01003",
			SchemeName: "Corporate Debt Bond Fund",
			AssetType:  "Debt",
			Type:       "buy",
			Units:      1000,
			NAV:        100.0,
			Date:       now.AddDate(-2, 0, 0),
		},
	}

	navMap := map[string]float64{
		"INF999K01001": 200.0, // Gain = 2000 * (200 - 100) = 200,000
		"INF888K01002": 180.0, // Loss = 500 * (180 - 200) = -10,000 (STCL)
		"INF777K01003": 115.0, // Debt value = 115,000
	}

	lots, harvest := BuildFIFOLedger("user_test", txs, navMap, now)

	// Check LTCG limit is capped at ₹1,25,000
	if harvest.HarvestableLTCG != LTCGExemptionLimit {
		t.Errorf("expected harvestable LTCG capped at %.2f, got %.2f", LTCGExemptionLimit, harvest.HarvestableLTCG)
	}
	expectedLTCGSaved := LTCGExemptionLimit * LTCGEquityTaxRate // 125000 * 0.125 = 15625
	if harvest.PotentialLTCGTaxSaved != expectedLTCGSaved {
		t.Errorf("expected LTCG tax saved %.2f, got %.2f", expectedLTCGSaved, harvest.PotentialLTCGTaxSaved)
	}

	// Check STCL matching
	if harvest.AvailableSTCL != 10000.0 {
		t.Errorf("expected STCL 10000.0, got %.2f", harvest.AvailableSTCL)
	}
	expectedSTCLSaved := 10000.0 * STCGEquityTaxRate // 10000 * 0.20 = 2000
	if harvest.PotentialSTCLTaxSaved != expectedSTCLSaved {
		t.Errorf("expected STCL tax saved %.2f, got %.2f", expectedSTCLSaved, harvest.PotentialSTCLTaxSaved)
	}

	// Check debt shift
	if harvest.DebtToArbitrageSavings <= 0 {
		t.Errorf("expected positive debt to arbitrage tax savings")
	}

	// Verify total lots count
	if len(lots) != 3 {
		t.Errorf("expected 3 lots, got %d", len(lots))
	}
}
