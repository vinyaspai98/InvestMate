package quant

import (
	"testing"
	"time"

	"investmate-backend/internal/models"
)

func TestGenerateRecommendations_RebalancingAndDebtShift(t *testing.T) {
	lots := []models.TaxLot{
		{
			ID:             "lot1",
			ISIN:           "INF179K01BE2",
			SchemeName:     "HDFC Top 100",
			AssetType:      "Equity",
			Units:          500,
			PurchaseNAV:    600,
			PurchaseDate:   time.Now().AddDate(-2, 0, 0),
			TotalCost:      300000,
			IsLTCG:         true,
			UnrealizedGain: 50000,
		},
		{
			ID:             "lot2",
			ISIN:           "INF179K01234",
			SchemeName:     "HDFC Corporate Debt",
			AssetType:      "Debt",
			Units:          10000,
			PurchaseNAV:    20,
			PurchaseDate:   time.Now().AddDate(-1, 0, 0),
			TotalCost:      200000,
			IsLTCG:         false,
			UnrealizedGain: 15000,
		},
	}

	harvest := CalculateTaxHarvesting(lots)
	result := GenerateRecommendations(lots, harvest, "moderate")

	if len(result.Actions) == 0 {
		t.Fatalf("expected recommendation actions, got 0")
	}

	var foundClosetMigration, foundDebtShift bool
	for _, a := range result.Actions {
		if a.AssetCategory == "Equity (Index Migration)" {
			foundClosetMigration = true
			if a.TERReduction <= 0 {
				t.Errorf("expected positive TER reduction, got %.2f", a.TERReduction)
			}
		}
		if a.AssetCategory == "Debt to Arbitrage Shift" {
			foundDebtShift = true
			if a.ActionTrigger == "" {
				t.Errorf("expected action trigger for debt shift")
			}
		}
	}

	if !foundClosetMigration {
		t.Errorf("expected closet index migration action for HDFC Top 100")
	}
	if !foundDebtShift {
		t.Errorf("expected debt to arbitrage shift action")
	}
	if result.TaxAlphaProjected <= 0 {
		t.Errorf("expected positive tax alpha projected, got %.2f", result.TaxAlphaProjected)
	}
}
