package quant

import (
	"math"
	"strings"

	"investmate-backend/internal/models"
)

// BenchmarkIndex represents an index with constituent weights.
type BenchmarkIndex struct {
	Name         string
	TargetTER    float64 // Direct ETF TER (e.g. 0.08%)
	Constituents map[string]float64 // Ticker/Name -> weight in percentage (sum = 100)
}

// Pre-loaded constituent profiles for Indian benchmarks
var BenchmarkIndices = map[string]BenchmarkIndex{
	"Nifty 50": {
		Name:      "Nifty 50",
		TargetTER: 0.08,
		Constituents: map[string]float64{
			"HDFCBANK":   12.8,
			"RELIANCE":   9.6,
			"ICICIBANK":  8.1,
			"INFY":       5.7,
			"ITC":        4.3,
			"TCS":        3.9,
			"LT":         3.6,
			"BHARTIARTL": 3.4,
			"AXISBANK":   3.1,
			"KOTAKBANK":  2.8,
			"SBIN":       2.7,
			"HINDUNILVR": 2.4,
			"BAJFINANCE": 2.1,
			"M&M":        2.0,
			"MARUTI":     1.7,
			"SUNPHARMA":  1.6,
			"TITAN":      1.4,
			"TATAMOTORS": 1.4,
			"NTPC":       1.3,
			"POWERGRID":  1.2,
			"OTHER":      25.0,
		},
	},
	"Nifty Next 50": {
		Name:      "Nifty Next 50",
		TargetTER: 0.12,
		Constituents: map[string]float64{
			"BEL":        4.5,
			"TRENT":      4.2,
			"HAL":        3.8,
			"VBL":        3.6,
			"CHOLAFIN":   3.2,
			"PFC":        3.0,
			"RECLTD":     2.9,
			"SIEMENS":    2.8,
			"DLF":        2.7,
			"ABB":        2.6,
			"OTHER":      66.7,
		},
	},
	"Nifty Midcap 150": {
		Name:      "Nifty Midcap 150",
		TargetTER: 0.15,
		Constituents: map[string]float64{
			"MAXHEALTH":  2.2,
			"FEDERALBNK": 2.1,
			"AUROPHARMA": 1.9,
			"COFORGE":    1.8,
			"PERSISTENT": 1.7,
			"POLYCAB":    1.6,
			"SUZLON":     1.5,
			"DIXON":      1.4,
			"HINDPETRO":  1.3,
			"OTHER":      84.5,
		},
	},
}

// KnownMutualFundProfiles provides benchmark, TER and representative holdings for popular funds.
type FundProfile struct {
	ISINPrefix     string
	BenchmarkIndex string
	TypicalTER     float64
	Constituents   map[string]float64
	KnownActiveShare *float64
}

var KnownFundDatabase = map[string]FundProfile{
	"hdfc top 100": {
		BenchmarkIndex: "Nifty 50",
		TypicalTER:     1.65,
		Constituents: map[string]float64{
			"HDFCBANK": 11.5, "RELIANCE": 9.0, "ICICIBANK": 8.5, "INFY": 6.0,
			"ITC": 4.5, "TCS": 3.8, "LT": 3.5, "BHARTIARTL": 3.2, "OTHER": 50.0,
		},
	},
	"icici prudential bluechip": {
		BenchmarkIndex: "Nifty 50",
		TypicalTER:     1.72,
		Constituents: map[string]float64{
			"ICICIBANK": 9.5, "RELIANCE": 8.8, "HDFCBANK": 8.0, "INFY": 6.5,
			"L&T": 4.0, "BHARTIARTL": 3.8, "OTHER": 49.4,
		},
	},
	"sbi bluechip": {
		BenchmarkIndex: "Nifty 50",
		TypicalTER:     1.68,
		Constituents: map[string]float64{
			"HDFCBANK": 10.0, "ICICIBANK": 8.2, "RELIANCE": 7.5, "INFY": 5.8,
			"ITC": 4.2, "TCS": 3.5, "OTHER": 60.8,
		},
	},
	"mirae asset large cap": {
		BenchmarkIndex: "Nifty 50",
		TypicalTER:     1.55,
		Constituents: map[string]float64{
			"HDFCBANK": 11.0, "ICICIBANK": 9.0, "RELIANCE": 8.5, "INFY": 6.2,
			"TCS": 4.0, "OTHER": 61.3,
		},
	},
	"axis bluechip": {
		BenchmarkIndex: "Nifty 50",
		TypicalTER:     1.62,
		Constituents: map[string]float64{
			"ICICIBANK": 9.8, "HDFCBANK": 9.2, "BAJFINANCE": 8.0, "INFY": 7.5,
			"TCS": 6.0, "OTHER": 59.5,
		},
	},
	"parag parikh flexi cap": {
		BenchmarkIndex: "Nifty 50",
		TypicalTER:     1.33,
		Constituents: map[string]float64{
			"HDFCBANK": 7.5, "ITC": 6.8, "BAJAJHLDNG": 6.5, "ICICIBANK": 6.0,
			"POWERGRID": 5.5, "FOREIGN_EQUITY": 16.0, "OTHER": 51.7,
		},
	},
	"quant active fund": {
		BenchmarkIndex: "Nifty 50",
		TypicalTER:     1.75,
		Constituents: map[string]float64{
			"RELIANCE": 9.5, "JIOFIN": 7.0, "ADANIPOWER": 5.5, "HDFCBANK": 5.0, "OTHER": 73.0,
		},
	},
	"motilal oswal midcap": {
		BenchmarkIndex: "Nifty Midcap 150",
		TypicalTER:     1.82,
		Constituents: map[string]float64{
			"PERSISTENT": 6.5, "COFORGE": 5.8, "DIXON": 5.2, "OTHER": 82.5,
		},
	},
	"nippon india small cap": {
		BenchmarkIndex: "Nifty Midcap 150",
		TypicalTER:     1.58,
		Constituents: map[string]float64{
			"TUBEINVEST": 3.0, "HDFCBANK": 2.5, "OTHER": 94.5,
		},
	},
}

// CalculateActiveShare computes the Active Share percentage using:
// AS = 0.5 * sum(|w_fund,i - w_index,i|)
func CalculateActiveShare(fundWeights, indexWeights map[string]float64) float64 {
	allTickers := make(map[string]bool)
	for t := range fundWeights {
		allTickers[t] = true
	}
	for t := range indexWeights {
		allTickers[t] = true
	}

	var sumAbsDiff float64
	for t := range allTickers {
		wf := fundWeights[t]
		wi := indexWeights[t]
		sumAbsDiff += math.Abs(wf - wi)
	}

	activeShare := 0.5 * sumAbsDiff
	if activeShare > 100.0 {
		activeShare = 100.0
	}
	if activeShare < 0.0 {
		activeShare = 0.0
	}
	return math.Round(activeShare*100) / 100
}

// MatchFundProfile looks up or infers fund properties (benchmark, TER, constituents).
func MatchFundProfile(schemeName, isin string) (BenchmarkIndex, float64, map[string]float64) {
	lowerName := strings.ToLower(schemeName)
	for nameKey, profile := range KnownFundDatabase {
		if strings.Contains(lowerName, nameKey) {
			bench := BenchmarkIndices[profile.BenchmarkIndex]
			return bench, profile.TypicalTER, profile.Constituents
		}
	}

	// Default benchmark resolution based on scheme keywords
	var bench BenchmarkIndex
	var ter float64
	if strings.Contains(lowerName, "midcap") || strings.Contains(lowerName, "smallcap") {
		bench = BenchmarkIndices["Nifty Midcap 150"]
		ter = 1.75
	} else if strings.Contains(lowerName, "next 50") {
		bench = BenchmarkIndices["Nifty Next 50"]
		ter = 1.60
	} else {
		bench = BenchmarkIndices["Nifty 50"]
		ter = 1.70
	}

	// If direct plan, subtract ~0.75% from typical regular TER
	if strings.Contains(lowerName, "direct") {
		ter -= 0.75
		if ter < 0.20 {
			ter = 0.20
		}
	}

	// Synthetic fund constituent simulation mirroring benchmark with slight tilt
	syntheticWeights := make(map[string]float64)
	for k, v := range bench.Constituents {
		// Large-cap active funds typically hug index closely with 45-65% active share
		syntheticWeights[k] = v * 0.85
	}
	syntheticWeights["ACTIVE_BETS"] = 15.0

	return bench, ter, syntheticWeights
}

// AuditFundActiveShare performs Active Share audit for a single holding.
func AuditFundActiveShare(schemeName, isin string, holdingValue float64) models.ActiveShareAudit {
	bench, currentTER, fundConstituents := MatchFundProfile(schemeName, isin)
	activeShare := CalculateActiveShare(fundConstituents, bench.Constituents)

	// Threshold Flag: ActiveShare < 60.0% && CurrentTER > 1.0%
	isClosetIndexer := (activeShare < 60.0) && (currentTER > 1.0)
	feeBleedRate := currentTER - bench.TargetTER
	if feeBleedRate < 0 {
		feeBleedRate = 0
	}
	annualFeeBleed := holdingValue * (feeBleedRate / 100.0)

	return models.ActiveShareAudit{
		ISIN:             isin,
		SchemeName:       schemeName,
		BenchmarkIndex:   bench.Name,
		ActiveShareScore: activeShare,
		CurrentTER:       currentTER,
		TargetIndexTER:   bench.TargetTER,
		IsClosetIndexer:  isClosetIndexer,
		AnnualFeeBleed:   math.Round(annualFeeBleed*100) / 100,
	}
}

// CalculatePortfolioHealthScore computes the 5-dimension diagnostic radar score.
func CalculatePortfolioHealthScore(lots []models.TaxLot, harvestSummary TaxHarvestSummary) models.PortfolioHealthScore {
	if len(lots) == 0 {
		return models.PortfolioHealthScore{
			OverallScore:       70,
			ActiveShareScore:   75,
			TaxEfficiencyScore: 70,
			UlcerIndexScore:    80,
			FactorBalanceScore: 70,
			FeeControlScore:    75,
			AuditDetails:       []models.ActiveShareAudit{},
		}
	}

	// Group holdings by scheme/ISIN
	type holdingAgg struct {
		schemeName string
		isin       string
		assetType  string
		totalVal   float64
		cost       float64
	}
	holdings := make(map[string]*holdingAgg)
	var totalPortfolioVal float64
	assetAlloc := map[string]float64{"Equity": 0, "Debt": 0, "Gold": 0, "Arbitrage": 0}

	for _, lot := range lots {
		val := lot.TotalCost + lot.UnrealizedGain
		if val < 0 {
			val = 0
		}
		totalPortfolioVal += val
		assetAlloc[lot.AssetType] += val

		key := lot.ISIN
		if key == "" {
			key = lot.SchemeName
		}
		if _, exists := holdings[key]; !exists {
			holdings[key] = &holdingAgg{
				schemeName: lot.SchemeName,
				isin:       lot.ISIN,
				assetType:  lot.AssetType,
			}
		}
		holdings[key].totalVal += val
		holdings[key].cost += lot.TotalCost
	}

	if totalPortfolioVal <= 0 {
		totalPortfolioVal = 1
	}

	// 1. Audit active share for equity / mutual fund holdings
	var auditDetails []models.ActiveShareAudit
	var weightedActiveShare float64
	var weightedTER float64
	var totalEquityVal float64

	for _, h := range holdings {
		if h.assetType == "Equity" || h.assetType == "Arbitrage" {
			audit := AuditFundActiveShare(h.schemeName, h.isin, h.totalVal)
			auditDetails = append(auditDetails, audit)

			totalEquityVal += h.totalVal
			weightedActiveShare += audit.ActiveShareScore * h.totalVal
			weightedTER += audit.CurrentTER * h.totalVal
		}
	}

	var activeSharePct float64
	var avgTER float64
	if totalEquityVal > 0 {
		activeSharePct = weightedActiveShare / totalEquityVal
		avgTER = weightedTER / totalEquityVal
	} else {
		activeSharePct = 75.0
		avgTER = 0.50
	}

	// Score 1: Active Share Score (0-100)
	// Higher active share is better for active fees, or direct index is rewarded
	activeShareScore := int(math.Min(100, math.Max(30, activeSharePct)))
	// Penalize if closet indexer is present
	for _, a := range auditDetails {
		if a.IsClosetIndexer {
			activeShareScore -= 8
		}
	}
	if activeShareScore < 20 {
		activeShareScore = 20
	}
	if activeShareScore > 100 {
		activeShareScore = 100
	}

	// Score 2: Tax Efficiency Score (0-100)
	// Deduct for unharvested LTCG under limit and unmigrated debt drag
	taxEfficiencyScore := 90
	if harvestSummary.HarvestableLTCG > 10000 {
		taxEfficiencyScore -= 15 // Unharvested free gain penalty
	}
	if harvestSummary.DebtToArbitrageSavings > 5000 {
		taxEfficiencyScore -= 20 // Taxable debt slab penalty
	}
	if harvestSummary.AvailableSTCL > 10000 {
		taxEfficiencyScore += 5 // Has usable loss harvest
	}
	if taxEfficiencyScore < 30 {
		taxEfficiencyScore = 30
	}
	if taxEfficiencyScore > 100 {
		taxEfficiencyScore = 100
	}

	// Score 3: Ulcer Index / Stress Risk Score (0-100)
	// Measures downside portfolio stress. Highly concentrated equity portfolio has low score.
	equityShare := assetAlloc["Equity"] / totalPortfolioVal
	ulcerScore := 85
	if equityShare > 0.85 {
		ulcerScore -= 25 // High drawdown volatility risk
	} else if equityShare < 0.30 {
		ulcerScore -= 10 // Too conservative inflation risk
	}
	if ulcerScore < 30 {
		ulcerScore = 30
	}

	// Score 4: Factor Balance Score (0-100)
	// Rewards multi-asset diversification (Equity + Debt/Arbitrage + Gold)
	factorScore := 50
	if assetAlloc["Equity"] > 0 {
		factorScore += 20
	}
	if assetAlloc["Debt"] > 0 || assetAlloc["Arbitrage"] > 0 {
		factorScore += 15
	}
	if assetAlloc["Gold"] > 0 {
		factorScore += 15
	}

	// Score 5: Fee Control Score (0-100)
	// Index funds (<0.15%): 95-100, 0.5%: 80, 1.5%: 50, >2.0%: 30
	feeControlScore := int(100 - (avgTER * 28))
	if feeControlScore < 25 {
		feeControlScore = 25
	}
	if feeControlScore > 100 {
		feeControlScore = 100
	}

	// Overall Composite Score (weighted average)
	overall := int(0.25*float64(activeShareScore) +
		0.25*float64(taxEfficiencyScore) +
		0.20*float64(feeControlScore) +
		0.15*float64(ulcerScore) +
		0.15*float64(factorScore))

	if overall > 100 {
		overall = 100
	}
	if overall < 20 {
		overall = 20
	}

	return models.PortfolioHealthScore{
		OverallScore:       overall,
		ActiveShareScore:   activeShareScore,
		TaxEfficiencyScore: taxEfficiencyScore,
		UlcerIndexScore:    ulcerScore,
		FactorBalanceScore: factorScore,
		FeeControlScore:    feeControlScore,
		AuditDetails:       auditDetails,
	}
}
