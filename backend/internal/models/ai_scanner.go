package models

import "time"

// TaxLot represents an immutable FIFO tax purchase lot for an asset.
type TaxLot struct {
	ID             string    `json:"id" firestore:"id"`
	UserID         string    `json:"user_id" firestore:"user_id"`
	ISIN           string    `json:"isin" firestore:"isin"`
	SchemeName     string    `json:"scheme_name" firestore:"scheme_name"`
	AssetType      string    `json:"asset_type" firestore:"asset_type"` // Equity, Debt, Gold, Arbitrage
	Units          float64   `json:"units" firestore:"units"`
	PurchaseNAV    float64   `json:"purchase_nav" firestore:"purchase_nav"`
	PurchaseDate   time.Time `json:"purchase_date" firestore:"purchase_date"`
	TotalCost      float64   `json:"total_cost" firestore:"total_cost"`
	IsLTCG         bool      `json:"is_ltcg" firestore:"is_ltcg"` // >12 months for equity
	UnrealizedGain float64   `json:"unrealized_gain" firestore:"unrealized_gain"`
}

// ActiveShareAudit stores benchmark overlap and fee drag analysis.
type ActiveShareAudit struct {
	ISIN             string  `json:"isin"`
	SchemeName       string  `json:"scheme_name"`
	BenchmarkIndex   string  `json:"benchmark_index"`   // Nifty 50, Nifty Next 50, Midcap 150
	ActiveShareScore float64 `json:"active_share_score"` // Percentage 0.0 - 100.0
	CurrentTER       float64 `json:"current_ter"`         // e.g. 1.85%
	TargetIndexTER   float64 `json:"target_index_ter"`   // e.g. 0.08%
	IsClosetIndexer  bool    `json:"is_closet_indexer"`  // True if ActiveShare < 60.0 && TER > 1.0%
	AnnualFeeBleed   float64 `json:"annual_fee_bleed"`
}

// PortfolioHealthScore aggregates portfolio diagnostics (0 - 100 scale).
type PortfolioHealthScore struct {
	OverallScore       int                `json:"overall_score"`
	ActiveShareScore   int                `json:"active_share_score"`
	TaxEfficiencyScore int                `json:"tax_efficiency_score"`
	UlcerIndexScore    int                `json:"ulcer_index_score"`
	FactorBalanceScore int                `json:"factor_balance_score"`
	FeeControlScore    int                `json:"fee_control_score"`
	AuditDetails       []ActiveShareAudit `json:"audit_details"`
	Source             string             `json:"source,omitempty"`
	TransactionCount   int                `json:"transaction_count,omitempty"`
	HoldingsCount      int                `json:"holdings_count,omitempty"`
}

// ProfitForecastPoint represents projected wealth at a timeline milestone.
type ProfitForecastPoint struct {
	Year                 int     `json:"year"`
	UnoptimizedValue     float64 `json:"unoptimized_value"`
	UnoptimizedNetProfit float64 `json:"unoptimized_net_profit"`
	OptimizedValue       float64 `json:"optimized_value"`
	OptimizedNetProfit   float64 `json:"optimized_net_profit"`
	NetProfitDelta       float64 `json:"net_profit_delta"`
}

// RecommendationAction defines side-by-side asset migration steps.
type RecommendationAction struct {
	AssetCategory       string  `json:"asset_category"`
	CurrentAsset        string  `json:"current_asset"`
	CurrentAllocation   float64 `json:"current_allocation"`
	TargetAsset         string  `json:"target_asset"`
	TargetAllocation    float64 `json:"target_allocation"`
	TERReduction        float64 `json:"ter_reduction"`
	TaxEfficiencyStatus string  `json:"tax_efficiency_status"`
	ActionTrigger       string  `json:"action_trigger"` // e.g., "STP over 6 months", "Immediate Harvest"
}

// ScanCASRequest represents the payload for e-CAS ingestion.
type ScanCASRequest struct {
	FileBytes string `json:"file_bytes"` // Base64 encoded file content
	Password  string `json:"password"`   // PDF password if encrypted
	Text      string `json:"text"`       // Fallback raw text representation
}

// RecommendationsRequest represents target risk parameters.
type RecommendationsRequest struct {
	TargetRisk string `json:"target_risk"` // "conservative", "moderate", "aggressive"
}

// ForecastProfitRequest represents initial capital and projection horizon.
type ForecastProfitRequest struct {
	InitialBase float64 `json:"initial_base"`
	Years       int     `json:"years"`
}

// MonteCarloResponse represents the stochastic trajectory simulation response.
type MonteCarloResponse struct {
	Trajectories [][]float64 `json:"trajectories"`
	SurvivalRate float64     `json:"survival_rate"`
	Percentile10 []float64   `json:"percentile_10"`
	Median       []float64   `json:"median"`
	Percentile90 []float64   `json:"percentile_90"`
	Years        int         `json:"years"`
}

// PortfolioScanSummary represents diagnostic summary figures.
type PortfolioScanSummary struct {
	TotalValue          float64 `json:"total_value" firestore:"total_value"`
	TERBleed            float64 `json:"ter_bleed" firestore:"ter_bleed"`
	PotentialTaxSavings float64 `json:"potential_tax_savings" firestore:"potential_tax_savings"`
}

// PortfolioScanRecord represents the Firestore document saved under users/{userId}/portfolio_scans/{scanId}.
type PortfolioScanRecord struct {
	ID          string               `json:"id" firestore:"id"`
	UserID      string               `json:"user_id" firestore:"user_id"`
	ScannedAt   time.Time            `json:"scanned_at" firestore:"scanned_at"`
	HealthScore PortfolioHealthScore `json:"health_score" firestore:"health_score"`
	Summary     PortfolioScanSummary `json:"summary" firestore:"summary"`
}

// AIRecommendationRecord represents the Firestore document saved under users/{userId}/ai_recommendations/{recId}.
type AIRecommendationRecord struct {
	ID                string                 `json:"id" firestore:"id"`
	UserID            string                 `json:"user_id" firestore:"user_id"`
	GeneratedAt       time.Time              `json:"generated_at" firestore:"generated_at"`
	Actions           []RecommendationAction `json:"actions" firestore:"actions"`
	TaxAlphaProjected float64                `json:"tax_alpha_projected" firestore:"tax_alpha_projected"`
}

// ForecastCacheRecord represents the Firestore document saved under users/{userId}/forecast_cache/{forecastId}.
type ForecastCacheRecord struct {
	ID             string                `json:"id" firestore:"id"`
	UserID         string                `json:"user_id" firestore:"user_id"`
	BaseCapital    float64               `json:"base_capital" firestore:"base_capital"`
	TimelineYears  int                   `json:"timeline_years" firestore:"timeline_years"`
	ForecastPoints []ProfitForecastPoint `json:"forecast_points" firestore:"forecast_points"`
	UpdatedAt      time.Time             `json:"updated_at" firestore:"updated_at"`
}
