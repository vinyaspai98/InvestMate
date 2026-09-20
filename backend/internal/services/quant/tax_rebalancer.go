package quant

import (
	"fmt"
	"math"
	"strings"

	"investmate-backend/internal/models"
)

// TargetAllocation holds standard Black-Litterman asset class weight targets.
type TargetAllocation struct {
	Equity    float64
	Debt      float64
	Gold      float64
	Arbitrage float64
}

// RiskProfileAllocations defines Black-Litterman equilibrium weights for Indian portfolios.
var RiskProfileAllocations = map[string]TargetAllocation{
	"conservative": {
		Equity:    0.35,
		Debt:      0.15,
		Arbitrage: 0.35, // Arbitrage prioritized over taxable debt for tax efficiency
		Gold:      0.15,
	},
	"moderate": {
		Equity:    0.65,
		Debt:      0.05,
		Arbitrage: 0.20,
		Gold:      0.10,
	},
	"aggressive": {
		Equity:    0.80,
		Debt:      0.00,
		Arbitrage: 0.10,
		Gold:      0.10,
	},
}

// RebalanceResult holds recommendations and calculated tax alpha.
type RebalanceResult struct {
	Actions           []models.RecommendationAction
	TaxAlphaProjected float64
}

// GenerateRecommendations builds side-by-side asset migration actions and tax optimizations.
func GenerateRecommendations(
	lots []models.TaxLot,
	harvestSummary TaxHarvestSummary,
	targetRisk string,
) RebalanceResult {
	riskKey := strings.ToLower(strings.TrimSpace(targetRisk))
	targetWeights, ok := RiskProfileAllocations[riskKey]
	if !ok {
		riskKey = "moderate"
		targetWeights = RiskProfileAllocations["moderate"]
	}

	// Calculate total portfolio value and current asset class values
	var totalVal float64
	currentClassVal := map[string]float64{
		"Equity":    0,
		"Debt":      0,
		"Gold":      0,
		"Arbitrage": 0,
	}

	type assetAggregate struct {
		schemeName     string
		isin           string
		assetType      string
		totalVal       float64
		cost           float64
		unrealizedGain float64
		isLTCG         bool
	}
	assetMap := make(map[string]*assetAggregate)

	for _, lot := range lots {
		val := lot.TotalCost + lot.UnrealizedGain
		if val < 0 {
			val = 0
		}
		totalVal += val
		currentClassVal[lot.AssetType] += val

		key := lot.ISIN
		if key == "" {
			key = lot.SchemeName
		}
		if _, exists := assetMap[key]; !exists {
			assetMap[key] = &assetAggregate{
				schemeName: lot.SchemeName,
				isin:       lot.ISIN,
				assetType:  lot.AssetType,
				isLTCG:     lot.IsLTCG,
			}
		}
		assetMap[key].totalVal += val
		assetMap[key].cost += lot.TotalCost
		assetMap[key].unrealizedGain += lot.UnrealizedGain
	}

	if totalVal <= 0 {
		totalVal = 1000000.0 // Default 10L baseline if empty
	}

	var actions []models.RecommendationAction
	var annualTERSavings float64

	// 1. Audit active holdings for Closet Indexers & High TER Fee bleed
	for _, a := range assetMap {
		allocPct := math.Round((a.totalVal/totalVal)*10000) / 100 // e.g. 24.50%

		if a.assetType == "Equity" {
			audit := AuditFundActiveShare(a.schemeName, a.isin, a.totalVal)

			if audit.IsClosetIndexer || audit.CurrentTER > 1.20 {
				targetAsset := "Nifty 50 Direct Index Plan"
				targetTER := 0.08
				if strings.Contains(strings.ToLower(a.schemeName), "next 50") {
					targetAsset = "Nifty Next 50 Direct Index Plan"
					targetTER = 0.12
				} else if strings.Contains(strings.ToLower(a.schemeName), "midcap") {
					targetAsset = "Nifty Midcap 150 Direct Index Plan"
					targetTER = 0.15
				}

				terDiff := audit.CurrentTER - targetTER
				annualTERSavings += a.totalVal * (terDiff / 100.0)

				trigger := "STP over 6 months to minimize exit load"
				if a.unrealizedGain <= LTCGExemptionLimit && a.isLTCG {
					trigger = "Immediate Harvest & Switch (Within ₹1.25L Tax Exemption)"
				}

				actions = append(actions, models.RecommendationAction{
					AssetCategory:       "Equity (Index Migration)",
					CurrentAsset:        fmt.Sprintf("%s (TER %.2f%%)", a.schemeName, audit.CurrentTER),
					CurrentAllocation:   allocPct,
					TargetAsset:         fmt.Sprintf("%s (TER %.2f%%)", targetAsset, targetTER),
					TargetAllocation:    allocPct, // maintain weight, upgrade vehicle
					TERReduction:        math.Round(terDiff*100) / 100,
					TaxEfficiencyStatus: "Saves TER drag; high tracking efficiency",
					ActionTrigger:       trigger,
				})
			}
		} else if a.assetType == "Debt" {
			// 2. Post-Budget 2024: Migrate taxable debt funds to Arbitrage Funds or SGBs
			actions = append(actions, models.RecommendationAction{
				AssetCategory:       "Debt to Arbitrage Shift",
				CurrentAsset:        fmt.Sprintf("%s (Taxed at 30%% Slab)", a.schemeName),
				CurrentAllocation:   allocPct,
				TargetAsset:         "Arbitrage Fund Direct (Taxed as Equity 12.5% LTCG)",
				TargetAllocation:    allocPct,
				TERReduction:        0.45,
				TaxEfficiencyStatus: "Shifts taxation from 30% slab to 12.5% LTCG; saves ~1.2% net yield",
				ActionTrigger:       "Switch to Arbitrage Fund via STP",
			})
		}
	}

	// 3. Tax Harvesting Specific Actions
	if harvestSummary.HarvestableLTCG > 0 {
		actions = append(actions, models.RecommendationAction{
			AssetCategory:       "Capital Gains Harvesting",
			CurrentAsset:        fmt.Sprintf("Eligible LTCG Equity Holdings (Gain: ₹%.0f)", harvestSummary.HarvestableLTCG),
			CurrentAllocation:   math.Round((harvestSummary.HarvestableLTCG/totalVal)*10000) / 100,
			TargetAsset:         "Same Assets Reinvested (Stepped-up NAV)",
			TargetAllocation:    math.Round((harvestSummary.HarvestableLTCG/totalVal)*10000) / 100,
			TERReduction:        0.0,
			TaxEfficiencyStatus: fmt.Sprintf("Tax-Free ₹1.25L Exemption Utilized; Saves ₹%.0f tax", harvestSummary.PotentialLTCGTaxSaved),
			ActionTrigger:       "Immediate Sell & Re-buy before March 31",
		})
	}

	if harvestSummary.AvailableSTCL > 0 {
		actions = append(actions, models.RecommendationAction{
			AssetCategory:       "Tax Loss Harvesting",
			CurrentAsset:        fmt.Sprintf("Short-Term Loss Lots (Loss: ₹%.0f)", harvestSummary.AvailableSTCL),
			CurrentAllocation:   math.Round((harvestSummary.AvailableSTCL/totalVal)*10000) / 100,
			TargetAsset:         "Reinvested into Equivalent Core Fund",
			TargetAllocation:    math.Round((harvestSummary.AvailableSTCL/totalVal)*10000) / 100,
			TERReduction:        0.0,
			TaxEfficiencyStatus: fmt.Sprintf("Offsets STCG taxed at 20%%; Saves ₹%.0f tax liability", harvestSummary.PotentialSTCLTaxSaved),
			ActionTrigger:       "Realize Loss & Roll Forward",
		})
	}

	// 4. 5/25 Threshold Drift Rebalancing check
	// Absolute drift >= 5% or relative drift >= 25%
	curEquityWeight := currentClassVal["Equity"] / totalVal
	targetEquityWeight := targetWeights.Equity
	driftAbs := math.Abs(curEquityWeight - targetEquityWeight)
	driftRel := 0.0
	if targetEquityWeight > 0 {
		driftRel = driftAbs / targetEquityWeight
	}

	if driftAbs >= 0.05 || driftRel >= 0.25 {
		actionMsg := "Direct new monthly SIPs to Debt/Arbitrage to rebalance"
		if curEquityWeight < targetEquityWeight {
			actionMsg = "Direct new monthly SIPs to Equity Index ETFs to rebalance"
		}

		actions = append(actions, models.RecommendationAction{
			AssetCategory:       "5/25 Portfolio Drift Rebalancing",
			CurrentAsset:        fmt.Sprintf("Equity Allocation: %.1f%%", curEquityWeight*100),
			CurrentAllocation:   math.Round(curEquityWeight*10000) / 100,
			TargetAsset:         fmt.Sprintf("Target Equilibrium: %.1f%% (%s)", targetEquityWeight*100, riskKey),
			TargetAllocation:    math.Round(targetEquityWeight*10000) / 100,
			TERReduction:        0.0,
			TaxEfficiencyStatus: "Avoids lump-sum taxable sale; rebalances through ongoing cash flow",
			ActionTrigger:       actionMsg,
		})
	}

	// If no specific recommendations triggered, provide default portfolio optimization plan
	if len(actions) == 0 {
		actions = append(actions, models.RecommendationAction{
			AssetCategory:       "Core Asset Allocation",
			CurrentAsset:        "Active Large Cap Holdings (TER 1.65%)",
			CurrentAllocation:   60.0,
			TargetAsset:         "Nifty 50 Direct Index ETF (TER 0.08%)",
			TargetAllocation:    60.0,
			TERReduction:        1.57,
			TaxEfficiencyStatus: "Optimal Low Fee Architecture",
			ActionTrigger:       "Maintain systematic SIP in direct index",
		})
	}

	totalTaxAlpha := harvestSummary.TotalTaxAlpha + annualTERSavings

	return RebalanceResult{
		Actions:           actions,
		TaxAlphaProjected: math.Round(totalTaxAlpha*100) / 100,
	}
}
