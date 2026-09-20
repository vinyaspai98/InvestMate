package quant

import (
	"testing"
	"time"

	"investmate-backend/internal/models"
)

func TestCalculateActiveShare_Precision(t *testing.T) {
	// Identical weights should have AS = 0%
	w1 := map[string]float64{"RELIANCE": 50.0, "TCS": 50.0}
	w2 := map[string]float64{"RELIANCE": 50.0, "TCS": 50.0}
	asIdentical := CalculateActiveShare(w1, w2)
	if asIdentical != 0.0 {
		t.Errorf("expected 0.0 for identical portfolios, got %.2f", asIdentical)
	}

	// Completely disjoint weights should have AS = 100%
	wDisjointA := map[string]float64{"RELIANCE": 100.0}
	wDisjointB := map[string]float64{"INFY": 100.0}
	asDisjoint := CalculateActiveShare(wDisjointA, wDisjointB)
	if asDisjoint != 100.0 {
		t.Errorf("expected 100.0 for completely disjoint portfolios, got %.2f", asDisjoint)
	}

	// Known overlap: Fund has 70% in RELIANCE, 30% in TCS; Benchmark has 50% in RELIANCE, 50% in TCS
	// |70-50| + |30-50| = 20 + 20 = 40. AS = 0.5 * 40 = 20%
	wFund := map[string]float64{"RELIANCE": 70.0, "TCS": 30.0}
	wIndex := map[string]float64{"RELIANCE": 50.0, "TCS": 50.0}
	asOverlap := CalculateActiveShare(wFund, wIndex)
	if asOverlap != 20.0 {
		t.Errorf("expected 20.0 for known overlap, got %.2f", asOverlap)
	}
}

func TestAuditFundActiveShare_ClosetIndexerDetection(t *testing.T) {
	// HDFC Top 100 has high overlap with Nifty 50 and TER 1.65% -> Should be closet indexer
	auditHDFC := AuditFundActiveShare("HDFC Top 100 Fund Regular Growth", "INF179K01BE2", 500000.0)
	if auditHDFC.ActiveShareScore >= 60.0 {
		t.Errorf("expected HDFC Top 100 to have Active Share < 60%%, got %.2f", auditHDFC.ActiveShareScore)
	}
	if !auditHDFC.IsClosetIndexer {
		t.Errorf("expected HDFC Top 100 with TER %.2f%% and AS %.2f%% to be flagged as closet indexer", auditHDFC.CurrentTER, auditHDFC.ActiveShareScore)
	}
	if auditHDFC.AnnualFeeBleed <= 0 {
		t.Errorf("expected positive annual fee bleed, got %.2f", auditHDFC.AnnualFeeBleed)
	}

	// Midcap fund typically has higher active share vs Nifty 50 or midcap index
	auditMidcap := AuditFundActiveShare("Nippon India Small Cap Fund", "INF204K01012", 200000.0)
	if auditMidcap.ActiveShareScore < 60.0 {
		// Small cap has high active share
		t.Logf("Small cap active share: %.2f", auditMidcap.ActiveShareScore)
	}
}

func TestCalculatePortfolioHealthScore_Diagnostics(t *testing.T) {
	lots := []models.TaxLot{
		{
			ID:             "lot1",
			ISIN:           "INF179K01BE2",
			SchemeName:     "HDFC Top 100",
			AssetType:      "Equity",
			Units:          100,
			PurchaseNAV:    500,
			PurchaseDate:   time.Now().AddDate(-2, 0, 0),
			TotalCost:      50000,
			IsLTCG:         true,
			UnrealizedGain: 20000,
		},
		{
			ID:             "lot2",
			ISIN:           "INF179K01234",
			SchemeName:     "HDFC Short Term Debt Fund",
			AssetType:      "Debt",
			Units:          2000,
			PurchaseNAV:    25,
			PurchaseDate:   time.Now().AddDate(-1, 0, 0),
			TotalCost:      50000,
			IsLTCG:         false,
			UnrealizedGain: 4000,
		},
	}

	harvest := CalculateTaxHarvesting(lots)
	score := CalculatePortfolioHealthScore(lots, harvest)

	if score.OverallScore <= 0 || score.OverallScore > 100 {
		t.Errorf("invalid overall score: %d", score.OverallScore)
	}
	if score.ActiveShareScore <= 0 || score.ActiveShareScore > 100 {
		t.Errorf("invalid active share score: %d", score.ActiveShareScore)
	}
	if score.TaxEfficiencyScore <= 0 || score.TaxEfficiencyScore > 100 {
		t.Errorf("invalid tax efficiency score: %d", score.TaxEfficiencyScore)
	}
	if score.UlcerIndexScore <= 0 || score.UlcerIndexScore > 100 {
		t.Errorf("invalid ulcer index score: %d", score.UlcerIndexScore)
	}
	if score.FactorBalanceScore <= 0 || score.FactorBalanceScore > 100 {
		t.Errorf("invalid factor balance score: %d", score.FactorBalanceScore)
	}
	if score.FeeControlScore <= 0 || score.FeeControlScore > 100 {
		t.Errorf("invalid fee control score: %d", score.FeeControlScore)
	}
	if len(score.AuditDetails) == 0 {
		t.Errorf("expected audit details for equity holding")
	}
}
