package quant

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"investmate-backend/internal/models"
)

// Post-Budget 2024 (FY 2025-26) India Tax Constants
const (
	LTCGExemptionLimit    = 125000.0 // ₹1,25,000 per financial year tax-free
	LTCGEquityTaxRate     = 0.125    // 12.5% on LTCG exceeding ₹1.25L
	STCGEquityTaxRate     = 0.20     // 20% on STCG
	DebtSlabTaxRate       = 0.30     // Assume standard 30% slab rate for high earners
	HoldingPeriodEquityDays = 365    // 12 months for equity & arbitrage
	HoldingPeriodGoldDays   = 365    // Listed gold ETFs 12 months under rationalized budget
)

// RawTransaction represents an input transaction for FIFO processing.
type RawTransaction struct {
	ID         string
	ISIN       string
	SchemeName string
	AssetType  string // "Equity", "Debt", "Gold", "Arbitrage"
	Type       string // "buy" or "sell"
	Units      float64
	NAV        float64
	Date       time.Time
}

// TaxHarvestOpportunity represents an actionable tax optimization step.
type TaxHarvestOpportunity struct {
	ISIN               string  `json:"isin"`
	SchemeName         string  `json:"scheme_name"`
	HarvestType        string  `json:"harvest_type"` // "LTCG_EXEMPTION_HARVEST", "STCL_OFFSET", "DEBT_TO_ARBITRAGE_SHIFT"
	UnitsToHarvest     float64 `json:"units_to_harvest"`
	CurrentGain        float64 `json:"current_gain"`
	TaxSaved           float64 `json:"tax_saved"`
	RecommendationNote string  `json:"recommendation_note"`
}

// TaxHarvestSummary summarizes the tax optimization potential for the current fiscal year.
type TaxHarvestSummary struct {
	HarvestableLTCG         float64                 `json:"harvestable_ltcg"`
	PotentialLTCGTaxSaved   float64                 `json:"potential_ltcg_tax_saved"`
	AvailableSTCL           float64                 `json:"available_stcl"`
	PotentialSTCLTaxSaved   float64                 `json:"potential_stcl_tax_saved"`
	DebtToArbitrageSavings  float64                 `json:"debt_to_arbitrage_savings"`
	TotalTaxAlpha           float64                 `json:"total_tax_alpha"`
	Opportunities           []TaxHarvestOpportunity `json:"opportunities"`
}

// NormalizeAssetType infers the asset type from scheme name and ISIN.
func NormalizeAssetType(schemeName, isin string) string {
	lower := strings.ToLower(schemeName)
	if strings.Contains(lower, "arbitrage") {
		return "Arbitrage"
	}
	if strings.Contains(lower, "gold") || strings.Contains(lower, "sgb") || strings.Contains(lower, "silver") {
		return "Gold"
	}
	if strings.Contains(lower, "debt") || strings.Contains(lower, "liquid") ||
		strings.Contains(lower, "gilt") || strings.Contains(lower, "money market") ||
		strings.Contains(lower, "short term") || strings.Contains(lower, "corporate bond") ||
		strings.Contains(lower, "overnight") {
		return "Debt"
	}
	return "Equity"
}

// BuildFIFOLedger processes raw transactions chronologically to generate an immutable FIFO lot ledger.
func BuildFIFOLedger(userID string, txs []RawTransaction, currentNAVs map[string]float64, asOf time.Time) ([]models.TaxLot, TaxHarvestSummary) {
	if asOf.IsZero() {
		asOf = time.Now()
	}

	// Sort transactions chronologically
	sortedTxs := make([]RawTransaction, len(txs))
	copy(sortedTxs, txs)
	sort.Slice(sortedTxs, func(i, j int) bool {
		return sortedTxs[i].Date.Before(sortedTxs[j].Date)
	})

	// Group buy lots per ISIN
	type workingLot struct {
		id          string
		isin        string
		schemeName  string
		assetType   string
		units       float64
		purchaseNAV float64
		date        time.Time
	}

	inventory := make(map[string][]*workingLot)

	for _, tx := range sortedTxs {
		if tx.Units <= 0 {
			continue
		}

		key := tx.ISIN
		if key == "" {
			key = tx.SchemeName
		}
		if key == "" {
			continue
		}

		assetType := tx.AssetType
		if assetType == "" {
			assetType = NormalizeAssetType(tx.SchemeName, tx.ISIN)
		}

		txType := strings.ToLower(strings.TrimSpace(tx.Type))
		if txType == "buy" || txType == "credit" || txType == "deposit" {
			wl := &workingLot{
				id:          fmt.Sprintf("lot_%s_%d", key, tx.Date.UnixNano()),
				isin:        tx.ISIN,
				schemeName:  tx.SchemeName,
				assetType:   assetType,
				units:       tx.Units,
				purchaseNAV: tx.NAV,
				date:        tx.Date,
			}
			inventory[key] = append(inventory[key], wl)
		} else if txType == "sell" || txType == "debit" || txType == "withdrawal" {
			remainingToSell := tx.Units
			lots := inventory[key]
			var updatedLots []*workingLot

			for _, lot := range lots {
				if remainingToSell <= 0 {
					updatedLots = append(updatedLots, lot)
					continue
				}

				if lot.units <= remainingToSell {
					// Entire lot consumed
					remainingToSell -= lot.units
					lot.units = 0
				} else {
					// Lot partially consumed
					lot.units -= remainingToSell
					remainingToSell = 0
					updatedLots = append(updatedLots, lot)
				}
			}
			inventory[key] = updatedLots
		}
	}

	var resultingLots []models.TaxLot

	for _, lots := range inventory {
		for _, lot := range lots {
			if lot.units <= 0.0001 {
				continue
			}

			totalCost := lot.units * lot.purchaseNAV
			curNAV := lot.purchaseNAV
			key := lot.isin
			if val, ok := currentNAVs[key]; ok && val > 0 {
				curNAV = val
			} else if val, ok := currentNAVs[lot.schemeName]; ok && val > 0 {
				curNAV = val
			}

			currentVal := lot.units * curNAV
			unrealizedGain := currentVal - totalCost

			// Calculate LTCG eligibility based on holding period and asset type
			holdingDuration := asOf.Sub(lot.date)
			holdingDays := int(holdingDuration.Hours() / 24)

			isLTCG := false
			switch lot.assetType {
			case "Equity", "Arbitrage":
				isLTCG = holdingDays >= HoldingPeriodEquityDays
			case "Gold":
				isLTCG = holdingDays >= HoldingPeriodGoldDays
			case "Debt":
				// Post-Budget 2024: Debt funds no longer enjoy indexation or LTCG benefit, taxed at slab rate
				isLTCG = false
			default:
				isLTCG = holdingDays >= HoldingPeriodEquityDays
			}

			taxLot := models.TaxLot{
				ID:             lot.id,
				UserID:         userID,
				ISIN:           lot.isin,
				SchemeName:     lot.schemeName,
				AssetType:      lot.assetType,
				Units:          lot.units,
				PurchaseNAV:    lot.purchaseNAV,
				PurchaseDate:   lot.date,
				TotalCost:      totalCost,
				IsLTCG:         isLTCG,
				UnrealizedGain: unrealizedGain,
			}
			resultingLots = append(resultingLots, taxLot)
		}
	}

	// Sort resulting lots by purchase date
	sort.Slice(resultingLots, func(i, j int) bool {
		return resultingLots[i].PurchaseDate.Before(resultingLots[j].PurchaseDate)
	})

	// Calculate Tax Loss/Gain Harvesting Opportunities
	harvestSummary := CalculateTaxHarvesting(resultingLots)

	return resultingLots, harvestSummary
}

// CalculateTaxHarvesting evaluates Post-Budget 2024 India tax harvesting opportunities.
func CalculateTaxHarvesting(lots []models.TaxLot) TaxHarvestSummary {
	var summary TaxHarvestSummary
	remainingExemption := LTCGExemptionLimit

	for _, lot := range lots {
		// 1. Equity & Arbitrage LTCG Harvesting (up to ₹1.25L tax-free per year)
		if (lot.AssetType == "Equity" || lot.AssetType == "Arbitrage") && lot.IsLTCG && lot.UnrealizedGain > 0 {
			gainAvailable := lot.UnrealizedGain
			gainToHarvest := gainAvailable
			if gainToHarvest > remainingExemption {
				gainToHarvest = remainingExemption
			}

			if gainToHarvest > 0 {
				ratio := gainToHarvest / gainAvailable
				unitsToSell := lot.Units * ratio
				taxSaved := gainToHarvest * LTCGEquityTaxRate // 12.5% saved by stepping up basis tax-free

				summary.HarvestableLTCG += gainToHarvest
				summary.PotentialLTCGTaxSaved += taxSaved
				remainingExemption -= gainToHarvest

				summary.Opportunities = append(summary.Opportunities, TaxHarvestOpportunity{
					ISIN:               lot.ISIN,
					SchemeName:         lot.SchemeName,
					HarvestType:        "LTCG_EXEMPTION_HARVEST",
					UnitsToHarvest:     unitsToSell,
					CurrentGain:        gainToHarvest,
					TaxSaved:           taxSaved,
					RecommendationNote: fmt.Sprintf("Sell %.2f units to harvest ₹%.2f LTCG within ₹1.25L tax-free exemption limit; reinvest immediately to reset cost basis.", unitsToSell, gainToHarvest),
				})
			}
		}

		// 2. Short-Term Capital Loss (STCL) Matching
		if (lot.AssetType == "Equity" || lot.AssetType == "Arbitrage") && !lot.IsLTCG && lot.UnrealizedGain < 0 {
			loss := -lot.UnrealizedGain
			taxOffset := loss * STCGEquityTaxRate // 20% tax saved by matching STCL

			summary.AvailableSTCL += loss
			summary.PotentialSTCLTaxSaved += taxOffset

			summary.Opportunities = append(summary.Opportunities, TaxHarvestOpportunity{
				ISIN:               lot.ISIN,
				SchemeName:         lot.SchemeName,
				HarvestType:        "STCL_OFFSET",
				UnitsToHarvest:     lot.Units,
				CurrentGain:        lot.UnrealizedGain,
				TaxSaved:           taxOffset,
				RecommendationNote: fmt.Sprintf("Book short-term loss of ₹%.2f on %.2f units to offset against STCG taxed at 20%%.", loss, lot.Units),
			})
		}

		// 3. Debt Allocation Shift to Arbitrage Funds
		if lot.AssetType == "Debt" {
			currentVal := lot.TotalCost + lot.UnrealizedGain
			if currentVal > 0 {
				// Debt taxed at slab (30%) vs Arbitrage taxed as Equity (12.5% LTCG)
				annualTaxSavings := currentVal * 0.07 * (DebtSlabTaxRate - LTCGEquityTaxRate) // on 7% estimated debt yield
				summary.DebtToArbitrageSavings += annualTaxSavings

				summary.Opportunities = append(summary.Opportunities, TaxHarvestOpportunity{
					ISIN:               lot.ISIN,
					SchemeName:         lot.SchemeName,
					HarvestType:        "DEBT_TO_ARBITRAGE_SHIFT",
					UnitsToHarvest:     lot.Units,
					CurrentGain:        lot.UnrealizedGain,
					TaxSaved:           annualTaxSavings,
					RecommendationNote: fmt.Sprintf("Migrate debt holdings (₹%.2f) to Arbitrage Funds (12.5%% LTCG) or SGBs to avoid 30%% slab taxation.", currentVal),
				})
			}
		}
	}

	summary.TotalTaxAlpha = summary.PotentialLTCGTaxSaved + summary.PotentialSTCLTaxSaved + summary.DebtToArbitrageSavings
	return summary
}
